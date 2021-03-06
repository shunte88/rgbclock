package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"math"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/fsnotify/fsnotify"
	"github.com/golang/freetype/truetype"
	rgbmatrix "github.com/mcuadros/go-rpi-rgb-led-matrix"
	"github.com/spf13/viper"

	//rgbmatrix "github.com/shunte88/go-rpi-rgb-led-matrix"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomonobold"
)

var idx = []int{0, 1, 2, 3}
var capture = false
var base string
var lat float64
var lng float64

func init() {

	base, err := filepath.Abs(filepath.Dir(os.Args[0]))
	checkFatal(err)

	cpu := runtime.NumCPU()
	if cpu > 1 {
		runtime.GOMAXPROCS(cpu - 1)
	}

	weatherserveruri = os.Getenv(weatherServerURI)
	if "" == weatherserveruri {
		weatherserveruri = "http://192.168.1.249:5000/weather/current"
	}

	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/rgbclock")
	viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	checkFatal(err)

	fontfile = viper.GetString("RGB.fontfile")

	layout = viper.GetString("RGB.layout")
	W = viper.GetInt(layout + ".width")
	H = viper.GetInt(layout + ".height")

	rows = viper.GetInt("RGB.rows")
	cols = viper.GetInt("RGB.cols")
	folding = viper.GetBool("RGB.folding")

	daybright = viper.GetInt("RGB.daybright")
	if daybright < 0 || daybright > 100 {
		daybright = 20
	}
	nightbright = viper.GetInt("RGB.nightbright")
	if nightbright < 0 || nightbright > 100 {
		nightbright = daybright
	}

	scroll = viper.GetInt("RGB.scroll_limit")
	hardware = viper.GetString("RGB.hardware")

	showbright = viper.GetBool("RGB.showbright")
	instrument = viper.GetBool("RGB.instrument")
	experiment = viper.GetBool("RGB.experiment")

	colorgrad1 = viper.GetString("RGB.colorgrad1")
	colorgrad2 = viper.GetString("RGB.colorgrad2")

	detail = viper.GetBool(layout + ".detail")
	parallel = viper.GetInt(layout + ".parallel")
	chain = viper.GetInt(layout + ".chain")
	clockw = viper.GetInt(layout + ".clock.width")
	clockh = viper.GetInt(layout + ".clock.height")

	miAlpha = viper.GetFloat64(layout + ".icon.main.alpha")
	miW = viper.GetInt(layout + ".icon.main.width")
	miScale = viper.GetFloat64(layout + ".icon.main.scale")

	wiAlpha = viper.GetFloat64(layout + ".icon.wind.alpha")
	wiW = viper.GetInt(layout + ".icon.wind.width")
	wiScale = viper.GetFloat64(layout + ".icon.wind.scale")

	iconStyle = viper.GetString(layout + ".style")

	lat = viper.GetFloat64("moon.lat")
	lng = viper.GetFloat64("moon.lng")

	feeds := viper.Get("feeds").([]interface{})
	newsDetail := viper.GetBool("news.detail")
	seedTime := viper.GetString("news.window.time")  // time to display, seed - repeat duration affects actual time
	newsDur := viper.GetInt("news.window.duration")  // minutes to display
	newsRepeat := viper.GetInt("news.window.repeat") // minutes until next display

	if `` != seedTime {
		news = InitNews(News{
			Feeds:    feeds,
			Detail:   newsDetail,
			SeedTime: seedTime,
			Width:    126,
			Limit:    62,
			Velocity: 1,
			Duration: (time.Duration(newsDur) * time.Minute),
			Repeat:   (time.Duration(newsRepeat) * time.Minute),
		})
	}
	// init icon map (dynamic scaling)
	mapInit()

	remaining = viper.GetBool("LMS.remaining")
	baseimage := viper.GetString("LMS.visualize.baseimage")
	meterbase := path.Join(viper.GetString("LMS.visualize.basefolder"), baseimage)
	baseimage = `LMS.visualize.` +
		strings.Replace(strings.Replace(baseimage, `.svg`, `.needle`, -1), `.png`, `.needle`, -1)

	lms = NewLMSServer(LMSConfig{
		Host:         viper.GetString("LMS.IP"),
		Port:         viper.GetInt("LMS.port"),
		Player:       viper.GetString("LMS.player"),
		BaseFolder:   path.Join(base, `/cache/`),
		Meter:        viper.GetString("LMS.visualize.meter"),
		MeterMode:    viper.GetString("LMS.visualize.metermode"),
		MeterLayout:  viper.GetString("LMS.visualize.layout"),
		MeterBase:    meterbase,
		NeedleColor:  viper.GetString(baseimage + `.color`),
		NeedleWidth:  viper.GetFloat64(baseimage + `.width`),
		NeedleLength: viper.GetFloat64(baseimage + `.length`),
		NeedleWell:   viper.GetBool(baseimage + `.well`),
		SSESActive:   viper.GetBool("LMS.sses.active"),
		SSESHost:     viper.GetString("LMS.sses.IP"),
		SSESPort:     viper.GetInt("LMS.sses.port"),
		SSESEndpoint: viper.GetString("LMS.sses.endpoint"),
	})

	offset := viper.GetInt("transport.offset")
	route := viper.GetString("transport.route")
	stop := viper.GetString("transport.stop")
	activeDays := viper.GetIntSlice("transport.active.days") // 0=Sunday
	activeFrom, err := parseTime(viper.GetString("transport.active.from"))
	if nil != err {
		activeFrom, _ = parseTime(`04:30 AM`)
	}
	activeUntil, err := parseTime(viper.GetString("transport.active.until"))
	if nil != err {
		activeFrom, _ = parseTime(`09:30 AM`)
	}
	api := os.Getenv(viper.GetString("transport.ApiEnv"))

	transit = NewMBTAClient(api, route, stop, activeFrom, activeUntil, activeDays, time.Duration(offset)*time.Minute)

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		daybright = viper.GetInt("RGB.daybright")
		if daybright < 0 || daybright > 100 {
			daybright = 20
		}
		nightbright = viper.GetInt("RGB.nightbright")
		if nightbright < 0 || nightbright > 100 {
			nightbright = daybright
		}

		capture = viper.GetBool("capture")

		showbright = viper.GetBool("RGB.showbright")
		instrument = viper.GetBool("RGB.instrument")
		experiment = viper.GetBool("RGB.experiment")

		activeFrom, err = parseTime(viper.GetString("transport.active.from"))
		if nil != err {
			activeFrom, _ = parseTime(`04:30 AM`)
		}
		activeUntil, err = parseTime(viper.GetString("transport.active.until"))
		if nil != err {
			activeFrom, _ = parseTime(`09:30 AM`)
		}
		activeDays = viper.GetIntSlice("transport.days") // 0=Sunday

		transit.SetActiveHours(activeFrom, activeUntil, activeDays)

		lat = viper.GetFloat64("moon.lat")
		lng = viper.GetFloat64("moon.lng")

	})

}

func main() {

	/*
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
	*/

	defer lms.Close()

	weather()

	mode = true

	config := &rgbmatrix.DefaultConfig
	config.Rows = rows
	config.Cols = cols
	config.Parallel = parallel
	config.ChainLength = chain
	config.Brightness = daymode.brightness
	config.HardwareMapping = hardware
	config.ShowRefreshRate = false
	config.InverseColors = false
	config.DisableHardwarePulsing = false

	// fixed assets
	imPrecip, _ = cacheImage(`brolly`, imPrecip, 0.00, ``)
	imHumid, _ = cacheImage(`humidity`, imHumid, 0.00, ``)

	channelIdent(`R`, 20, 20)
	channelIdent(`L`, 20, 20)

	togweather := false

	// concurrent updates
	stop := sched(weather, 30*time.Second)
	toggle := sched(toggleMode, 15*time.Second)
	rotator := sched(rotator, 3*time.Second)

	wf := float64(clockw)
	hf := float64(clockh)
	r := wf * 0.48
	var cx float64 = wf / 2.00
	var cy float64 = hf / 2.00
	length := wf * 0.07
	lw := wf * 0.04

	ea := 0.000

	font, err := truetype.Parse(gomonobold.TTF)
	checkFatal(err)

	if `` != fontfile {
		fb, err := ioutil.ReadFile(fontfile)
		if nil == err {
			font, _ = truetype.Parse(fb)
		}
	}

	sface := truetype.NewFace(font, &truetype.Options{
		Size: wf * 0.145,
	})
	dpface := truetype.NewFace(font, &truetype.Options{
		Size: hf * 0.11,
	})
	dptface := truetype.NewFace(font, &truetype.Options{
		Size: hf * 0.055,
	})
	dtface := truetype.NewFace(font, &truetype.Options{
		Size: hf * 0.08,
	})
	lmsface := truetype.NewFace(font, &truetype.Options{
		Size: hf * 0.066,
	})

	transit.SetFace(truetype.NewFace(font, &truetype.Options{
		Size: hf * 0.066,
		DPI:  72,
	}))
	transit.Start()
	defer transit.Stop()

	m, err := rgbmatrix.NewRGBLedMatrix(config)
	checkFatal(err)

	rgbc := rgbmatrix.NewCanvas(m)
	defer rgbc.Close()
	bounds := rgbc.Bounds()

	dc := gg.NewContext(W, H)

	var temps []string

	orangered := `#ff0000` // `#ff4500`
	darkred := `#660000`

	tm := time.Now()
	lastTempo := tm.Format(`15 04`)
	tempo, err := imageTime(tm, hf*.2, orangered)
	checkFatal(err)

	// eye-candy
	grad := gg.NewRadialGradient(cx, cy, r+2, 0, 0, r+2)

	grad.AddColorStop(0, parseHexColor(colorgrad1))
	grad.AddColorStop(1, parseHexColor(colorgrad2))

	lms.SetMaxLen(scroll)
	lms.SetFace(truetype.NewFace(font, &truetype.Options{
		Size: hf * 0.066,
		DPI:  72,
	}), "#ff9900c0")

	lastBrightness := daymode.brightness
	var icache draw.Image

	lms.Start()
	defer lms.Stop()

	if nil != news {
		news.SetFace(truetype.NewFace(font, &truetype.Options{
			Size: hf * 0.055,
			DPI:  72,
		}))
	}
	defer newsStop()

	cpu := NewCPUStat(truetype.NewFace(font, &truetype.Options{
		Size: hf * 0.150,
		DPI:  72,
	}), `#00ffff77`)
	defer cpu.Stop()

	angle := 0.20
	inca := angle
	dump := 0

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Printf("shutdown inc. GC (%v)\n", sig)
		os.Exit(0)
	}()

	for {

		if lastBrightness != daymode.brightness {
			err = rgbc.SetBrightness(uint32(daymode.brightness))
			lastBrightness = daymode.brightness
			if nil != err {
				fmt.Println("brightness", err)
			}
		}

		if nil != icache {
			dc.DrawImageAnchored(icache, 0, 0, 0, 0)
		} else {

			dc.SetHexColor("#000000")
			dc.Clear()

			cornerScroll(dc, `corner-scroll`, 0.9, `#ffcc0008`, false)
			cornerScroll(dc, `alt-corner-scroll`, 0.0, `#d4af3708`, true)

			// keep it crisp, fold scroll under face
			dc.DrawCircle(cx, cy, r+1)
			dc.Fill()

			var imGlobal iconCache
			imGlobal, _ = cacheImage(`global`, imGlobal, 0.666, ``)
			dc.DrawImageAnchored(imGlobal.image, 0, 0, 0, 0)
			dc.SetFillStyle(grad)
			dc.DrawCircle(cx, cy, r+1)
			dc.Fill()

			dc.SetHexColor(orangered)
			dc.SetLineWidth(lw)
			for i := 0; i < 60; i += 5 {

				var th float64 = math.Pi/30.00*float64(i) - math.Pi/3.00
				x, y := math.Cos(th), math.Sin(th)

				x1, y1 := (float64(r)-float64(length))*x+cx, (float64(r)-float64(length))*y+cy
				x2, y2 := float64(r)*x+cx, float64(r)*y+cy

				dc.DrawLine(x1, y1, x2, y2)
				dc.Stroke()

			}

			// cache the clock face - zero struggles
			icache = imaging.New(clockw, clockh, color.NRGBA{0, 0, 0, 0})
			icache = imaging.Paste(icache, dc.Image(), image.Pt(0, 0))

		}

		tm = time.Now()
		s := float64(tm.Second())

		h, err := strconv.ParseFloat(tm.Format("5.000"), 64)
		if nil == err {
			ea = 6 * h
		} else {
			ea = 6 * s
		}

		// glitchy under one second ??
		if ea < 8 {
			dc.SetHexColor(darkred)
			dc.DrawCircle(float64(cx), float64(cy), float64(r))
			dc.Stroke()
		}

		dc.SetHexColor(orangered)
		dc.DrawArc(float64(cx), float64(cy), float64(r), degToRadians(0), degToRadians(ea))
		dc.Stroke()

		// svg time - needs crispier
		ts := tm.Format("15 04")
		if ts != lastTempo {
			lastTempo = ts
			tempo, err = imageTime(tm, hf*.2, orangered)
			checkFatal(err)
		}
		if 0 == int(s)%2 {
			dc.DrawImageAnchored(tempo[0], int(cx), int(cy), 0.5, 0.5)
		} else {
			dc.DrawImageAnchored(tempo[1], int(cx), int(cy), 0.5, 0.5)
		}

		// place weather icon
		if imIcon.image != nil {
			dc.DrawImageAnchored(imIcon.image, int(wf/2), int(0.71875*hf), 0.5, 0.5)
		}

		if !mode {
			temps = strings.Split(w.Current.Temperature, " ")
		} else {
			p := w.Current.Daypart0.Precipitation
			wdx := float64(cx + (wf * 0.09))
			wdy := float64(0.25 * hf)
			if 0 == int(s)%2 {
				if togweather && `0%` != p {
					temps[1] = p
					// precipitation
					if imPrecip.image != nil {
						if `100%` == p {
							wdx += 5.00
						}
						dc.DrawImageAnchored(imPrecip.image, int(wdx), int(wdy), 0.5, 0.5)
					}
				} else {
					temps[1] = w.Current.Humidity
					// humidity
					if imHumid.image != nil {
						dc.DrawImageAnchored(imHumid.image, int(wdx), int(wdy), 0.5, 0.5)
					}
				}
			} else {
				togweather = !togweather
				temps = strings.Split(w.Current.Wind+" -- mph", " ") // fix for "Calm"
				// place wind icon
				if imWindDir.image != nil {
					dc.DrawImageAnchored(imWindDir.image, int(wdx), int(wdy), 0.5, 0.5)
				}
			}
		}
		dc.SetFontFace(sface)
		if !mode {
			dc.SetHexColor("#0099ff")
			wdy := float64(0.27 * hf)
			dc.DrawStringAnchored(fmt.Sprintf("%v", math.Round(w.Current.tempF)), 8+(wf*.25), wdy, 0.5, 0.5)
			if imThermo.image != nil {
				dc.DrawImageAnchored(imThermo.image, int(cx), 4+int(wdy), 0.5, 0.5)
			}
		}
		dc.SetHexColor("#66ff99")
		if !mode {
			wdy := float64(0.27 * hf)
			dc.DrawStringAnchored(fmt.Sprintf("%v", math.Round(w.Current.tempC)), -8+(wf*.75), wdy, 0.5, 0.5)
		} else {
			dc.DrawStringAnchored(temps[1], wf/3, hf*0.27, 0.5, 0.5)
		}

		if detail && W > 64 {
			placeWeatherDetail(dc, hf, dpface)
		}

		if `full` == layout {
			dc.SetHexColor("#ff9900")
			dc.SetFontFace(dtface)
			temps = hackaDate(tm)
			dpos := 2 + (5 * (hf / 8))
			dc.DrawStringAnchored(temps[idx[0]], wf/4, dpos, 0.5, 0.5)
			dc.DrawStringAnchored(temps[idx[1]], 3*(wf/4), dpos, 0.5, 0.5)

			mx := int(wf * 0.171875)
			moonI, err := NewLuna(tm, lat, lng).PhaseIcon(mx, mx)
			if err == nil && moonI != nil {
				dc.DrawImageAnchored(moonI, int(3*(wf/4))+2, int(3*(hf/4))+2, .5, .5)
			}

			if instrument {
				if cpu.CPUStatsTemp() != nil {
					hh := int(cpu.CPUStatsTemp().Bounds().Max.Y / 2)
					dc.DrawImageAnchored(cpu.CPUStatsTemp(), int(wf/2), hh, .5, .5)
				}
				if cpu.CPUStatsUsage() != nil {
					hh := H - int(cpu.CPUStatsUsage().Bounds().Max.Y/2)
					dc.DrawImageAnchored(cpu.CPUStatsUsage(), int(wf/2), hh, .5, .5)
				}
				if cpu.MemStats() != nil {
					hh := H - int(2.8*float64(cpu.MemStats().Bounds().Max.Y/2))
					dc.DrawImageAnchored(cpu.MemStats(), int(wf/2), hh, .5, .5)
				}
			}
			if showbright {
				dc.DrawStringAnchored(evut, cx, hf-(length-2), 0.5, 0.5)
				dc.DrawStringAnchored(imIcon.last, cx, (hf-(length-2))-8, 0.5, 0.5)
			}
			//dc.SetHexColor(w.Current.trendColor)
			//dc.DrawStringAnchored(w.Current.trend, 32, 2+cy+(hf/4), 0.5, 0.5)

		}

		if `play` == lms.Player.Mode {

			pinClockTop(dc)

			if mode {
				placeWeatherDetail(dc, hf/2, dptface)
			} else {
				dst := imaging.Resize(lms.Coverart(), 64, 64, imaging.Lanczos)
				dc.DrawImage(dst, 65, 0)
			}

			dst := imaging.Resize(lms.Coverart(), 126, 49, imaging.Lanczos)
			dst = imaging.Blur(imaging.AdjustBrightness(dst, -40), 6.5)
			dc.DrawImage(dst, 1, 66)

			if lms.VUActive() && !mode {

				dc.DrawImageAnchored(lms.VU(), int(cx), int(cy)+24, .5, .5)

			} else {

				dc.SetHexColor("#ff9900cc")
				dc.SetFontFace(lmsface)

				pos := int(cy + 11)
				dc.DrawImageAnchored(lms.Player.Albumartist.Image(), int(W/2), pos, 0.5, 0.5)
				pos += 9
				dc.DrawImageAnchored(lms.Player.Album.Image(), int(W/2), pos, 0.5, 0.5)
				pos += 9
				dc.DrawImageAnchored(lms.Player.Title.Image(), int(W/2), pos, 0.5, 0.5)
				pos += 9
				dc.DrawImageAnchored(lms.Player.Artist.Image(), int(W/2), pos, 0.5, 0.5)
				dc.DrawStringAnchored(fmt.Sprintf("• %v •", lms.Player.Year), float64(W/2), float64(cy+44), 0.5, 0.5)
				pos += 9
				dc.DrawImageAnchored(lms.PlayModifiers(), 1, pos, 0, 0.5)
				vol := lms.Volume()
				dc.DrawImageAnchored(vol, W-(vol.Bounds().Max.X+2), pos, 0, 0.5)

			}

			placeBorderZone(dc, lmsface, lw, 68, 50)
			drawHorizontalBar(dc, 10, 115, lms.Player.Percent)
			base := float64(H - 9 + 4)
			dc.SetHexColor("#ff9900")
			dc.DrawStringAnchored(lms.Player.TimeStr, 2, base, 0, 0.5)
			if remaining {
				dc.DrawStringAnchored(lms.Player.RemStr, float64(W-2), base, 1, 0.5)
			} else {
				dc.DrawStringAnchored(lms.Player.DurStr, float64(W-2), base, 1, 0.5)
			}
			dc.SetHexColor("#0099ffcc")
			dc.DrawStringAnchored(lms.Player.Bitty, float64(W/2), base, 0.5, 0.5)
			base = 0.390625 * wf
			dc.DrawImageAnchored(lms.VolumePopup(int(base), int(base)), int(wf/2), int(hf/4), .5, .5)

		} else if transit.Display {
			pinClockTop(dc)
			placeWeatherDetail(dc, hf/2, dptface)
			// active transit here - top 3
			noData := true
			pos := int(cy + 16)
			for pred := range transit.Predictions() {
				dc.DrawImageAnchored(pred, int(W/2), pos, 0.5, 0.5)
				pos += 16
				noData = false
			}
			if noData {
				dc.SetHexColor("#ff9900cc")
				dc.SetFontFace(sface)
				dc.DrawStringAnchored(`WAITING`, float64(W/2), float64(pos+3), 0.5, 0.5)
				dc.DrawStringAnchored(`FOR MBTA`, float64(W/2), float64(pos+19), 0.5, 0.5)
			}
			placeBorderZone(dc, lmsface, lw, 60, 55)
		} else {
			if news.Display() {
				pinClockTop(dc)
				placeWeatherDetail(dc, hf/2, dptface)
				pos := int(cy + 2)
				dc.DrawImageAnchored(news.Image(), 1, pos, 0, 0)
				placeBorderZone(dc, lmsface, lw, 60, 59)
			}
		}
		dc.SetLineWidth(lw)

		if experiment {
			dc.DrawImageAnchored(imaging.Rotate(dc.Image(), -angle, image.Black), int(cx), int(cy), 0.5, 0.5)

			angle = angle + inca
			if angle > 360 || angle < 0 {
				inca *= -1
			}
		}

		if capture && dump < 500 {
			go dumpImage(dump, dc.Image())
			dump++
		}

		if folding {
			itmp := imaging.Rotate180(dc.Image())
			dst := imaging.New(4*64, 64, color.NRGBA{0, 0, 0, 0})
			dst = imaging.Paste(dst, dc.Image(), image.Pt(0, 0))
			dst = imaging.Paste(dst, itmp, image.Pt(128, 0))
			draw.Draw(rgbc, dst.Bounds(), dst, image.ZP, draw.Over)
		} else {
			draw.Draw(rgbc, bounds, dc.Image(), image.ZP, draw.Over)
		}

		rgbc.Render()
		time.Sleep(60 * time.Millisecond)

	}

	stop <- true
	toggle <- true
	rotator <- true

}

func placeWeatherDetail(dc *gg.Context, hf float64, dpface font.Face) {
	dc.SetFontFace(dpface)
	placeDetail(dc, w.Current.Daypart1, imIconDP1.image, hf)
	placeDetail(dc, w.Current.Daypart2, imIconDP2.image, hf)
	placeDetail(dc, w.Current.Daypart3, imIconDP3.image, hf)
	placeDetail(dc, w.Current.Daypart4, imIconDP4.image, hf)
}

func placeBorderZone(dc *gg.Context, lmsface font.Face, lw, y1, y2 float64) {
	dc.SetFontFace(lmsface)
	dc.SetHexColor("#000000")
	dc.SetLineWidth(lw - 2)
	dc.DrawRectangle(0, 66, 128, y1)
	dc.Stroke()
	dc.SetLineWidth(0.5)
	dc.SetHexColor("#ff9900")
	dc.DrawRectangle(0, 66, 128, y2)
	dc.Stroke()
}

func pinClockTop(dc *gg.Context) {
	dst := imaging.Resize(dc.Image(), 65, 65, imaging.CatmullRom) //Lanczos)
	dc.DrawImage(dst, 0, 0)
	dc.SetHexColor("#000000")
	dc.DrawRectangle(65, 0, 65, 65)
	dc.Fill()
	dc.DrawRectangle(0, 65, 127, 63)
	dc.Fill()
}

func placeDetail(dc *gg.Context, d Daypart, wi draw.Image, hf float64) {
	f, err := strconv.ParseFloat(d.ID, 64)
	if err != nil {
		return
	}
	dx := (f - 1.00) * 15
	pdy1 := (hf * 0.11) + dx + 1
	pdy2 := (hf * 0.24) + dx

	if "1" != d.ID {
		dc.SetLineWidth(0.15)
		dc.SetHexColor("#86acac")
		dc.DrawLine(67, pdy1-7, 127, pdy1-7)
		dc.Stroke()
	}

	dc.SetHexColor("#2c3e50")
	dc.DrawString(d.Label, 68, pdy1)
	dc.SetHexColor("#0f3443")
	dc.DrawString(fmt.Sprintf("% 4s %sF", d.Hilo, d.Temperature), 68, pdy2)
	if f > 1 {
		f += 1.00
	}
	if wi != nil {
		dc.DrawImageAnchored(wi, W-23, int(pdy1-12+(f-1)), 0, 0)
	}
}

func degToRadians(deg float64) float64 {
	return gg.Radians(-90.00 + deg)
}

func hackaDate(t time.Time) []string {
	suffix := "th"
	switch t.Day() {
	case 1, 21, 31:
		suffix = "st"
	case 2, 22:
		suffix = "nd"
	case 3, 23:
		suffix = "rd"
	}
	return strings.Split(t.Format("Mon Jan 2"+suffix+" 2006"), " ")
}

func rotator() {
	idx = append(idx[1:], idx[0:1]...)
}

func drawHorizontalBar(dc *gg.Context, x, y, pcnt float64) {
	dc.SetLineWidth(0.3)
	l := float64(W) - (2 * x)
	lp := (l - 2.00) * (pcnt / 100.00)
	dc.SetHexColor("#000000")
	dc.DrawRectangle(x+1, y+1, l-2, 2)
	dc.Fill()
	dc.SetHexColor("#ff9900") // bar color
	dc.DrawRectangle(x+1, y+1, lp, 2)
	dc.Fill()
	dc.SetHexColor("#ff9900")
	dc.DrawRectangle(x, y, l, 4)
	dc.Stroke()
	dc.SetLineWidth(1)
}

func cornerScroll(dc *gg.Context, iconName string, scale float64, color string, globe bool) {
	var ic iconCache
	ic, _ = cacheImage(iconName, ic, scale, color)
	dc.DrawImageAnchored(ic.image, 0, -1, 0, 0)
	dst := imaging.FlipH(ic.image)
	lx := W - dst.Bounds().Max.X
	ly := H - dst.Bounds().Max.Y + 1
	dc.DrawImageAnchored(dst, lx, -1, 0, 0)
	dst = imaging.FlipV(dst)
	dc.DrawImageAnchored(dst, lx, ly, 0, 0)
	dst = imaging.FlipH(dst)
	dc.DrawImageAnchored(dst, 0, ly, 0, 0)
}
