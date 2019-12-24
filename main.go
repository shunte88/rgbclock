package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
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

func init() {

	weatherserveruri = os.Getenv(weatherServerURI)
	if "" == weatherserveruri {
		weatherserveruri = "http://192.168.1.249:5000/weather/current"
	}

	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/rgbclock")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	checkFatal(err)

	fontfile = viper.GetString("RGB.fontfile")
	fontfile2 = viper.GetString("RGB.fontfile2")
	W = viper.GetInt("RGB.width")
	H = viper.GetInt("RGB.height")
	rows = viper.GetInt("RGB.rows")
	cols = viper.GetInt("RGB.cols")
	folding = viper.GetBool("RGB.folding")

	parallel = viper.GetInt("RGB.parallel")
	chain = viper.GetInt("RGB.chain")

	daybright = viper.GetInt("RGB.daybright")
	if daybright < 0 || daybright > 100 {
		daybright = 20
	}
	nightbright = viper.GetInt("RGB.nightbright")
	if nightbright < 0 || nightbright > 100 {
		nightbright = daybright
	}

	layout = viper.GetString("RGB.layout")
	scroll = viper.GetInt("RGB.scroll_limit")
	hardware = viper.GetString("RGB.hardware")
	showbright = viper.GetBool("RGB.showbright")
	experiment = viper.GetBool("RGB.experiment")

	colorgrad1 = viper.GetString("RGB.colorgrad1")
	colorgrad2 = viper.GetString("RGB.colorgrad2")

	detail = viper.GetBool(layout + ".detail")
	clockw = viper.GetInt(layout + ".clock.width")
	clockh = viper.GetInt(layout + ".clock.height")

	miAlpha = viper.GetFloat64(layout + ".icon.main.alpha")
	miW = viper.GetInt(layout + ".icon.main.width")
	miScale = viper.GetFloat64(layout + ".icon.main.scale")

	wiAlpha = viper.GetFloat64(layout + ".icon.wind.alpha")
	wiW = viper.GetInt(layout + ".icon.wind.width")
	wiScale = viper.GetFloat64(layout + ".icon.wind.scale")

	iconStyle = viper.GetString(layout + ".style")

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
	lmsIP := viper.GetString("LMS.IP")
	lmsPort := viper.GetInt("LMS.port")
	lmsPlayer := viper.GetString("LMS.player")
	lms = NewLMSServer(lmsIP, lmsPort, lmsPlayer)

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

		showbright = viper.GetBool("RGB.showbright")
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

	// 4->9 (125%)
	weather()

	mode = true

	config := &rgbmatrix.DefaultConfig
	config.Rows = rows
	config.Cols = cols
	config.Parallel = parallel
	config.ChainLength = chain
	config.Brightness = brightness
	config.HardwareMapping = hardware
	config.ShowRefreshRate = false
	config.InverseColors = false
	config.DisableHardwarePulsing = false

	// fixed assets
	imPrecip, _ = cacheImage(`brolly`, imPrecip, 0.00, ``)
	imHumid, _ = cacheImage(`humidity`, imHumid, 0.00, ``)

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
	lcdfont := font
	fb, err := ioutil.ReadFile(fontfile)
	if nil == err {
		lcdfont, _ = truetype.Parse(fb)
	}
	if `` != fontfile2 {
		fb, err = ioutil.ReadFile(fontfile2)
		if nil == err {
			font, _ = truetype.Parse(fb)
		}
	}

	face := truetype.NewFace(lcdfont, &truetype.Options{
		Size: wf * 0.23,
	})
	zface := truetype.NewFace(lcdfont, &truetype.Options{
		Size: wf * 0.24,
	})
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

	mbtaface := truetype.NewFace(font, &truetype.Options{
		Size: hf * 0.066,
	})

	transit.SetFace(mbtaface)
	transit.Start()
	defer transit.Stop()

	m, err := rgbmatrix.NewRGBLedMatrix(config)
	checkFatal(err)

	rgbc := rgbmatrix.NewCanvas(m)
	defer rgbc.Close()
	bounds := rgbc.Bounds()

	dc := gg.NewContext(W, H)

	var temps []string

	// eye-candy
	grad := gg.NewRadialGradient(cx, cy, r+2, 0, 0, r+2)

	grad.AddColorStop(0, parseHexColor(colorgrad1))
	grad.AddColorStop(1, parseHexColor(colorgrad2))

	// using a 2nd copy of the font for InfoLabel as the mutex seems to cause problems
	xlmsface := truetype.NewFace(font, &truetype.Options{
		Size: hf * 0.066,
		DPI:  72,
	})

	lms.SetMaxLen(scroll)
	lms.SetFace(xlmsface, "#ff9900c0")

	lastBrightness := brightness
	var icache draw.Image

	lms.Start()
	defer lms.Stop()

	if nil != news {
		newsface := truetype.NewFace(font, &truetype.Options{
			Size: hf * 0.055,
			DPI:  72,
		})
		news.SetFace(newsface)
	}

	defer newsStop()

	angle := 0.20
	inca := angle
	dump := 501

	for {

		if lastBrightness != brightness {
			err = rgbc.SetBrightness(uint32(brightness))
			lastBrightness = brightness
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

			dc.SetHexColor("#ff0000")
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

		t := time.Now()
		s := float64(t.Second())

		h, err := strconv.ParseFloat(t.Format("5.000"), 64)
		if nil == err {
			ea = 6 * h
		} else {
			ea = 6 * s
		}

		// glitchy under one second ??
		if ea < 8 {
			dc.SetHexColor("#660000")
			dc.DrawCircle(float64(cx), float64(cy), float64(r))
			dc.Stroke()
		}

		dc.SetHexColor("#ff0000")
		dc.DrawArc(float64(cx), float64(cy), float64(r), degToRadians(0), degToRadians(ea))
		dc.Stroke()

		// tighter time format, specically with non-monospaced fonts
		ts := t.Format("15 04")
		colon := `:`
		if 0 == int(s)%2 {
			colon = ` `
		}

		placeTime(dc, wf, hf, zface, ts, colon, "#660000")
		placeTime(dc, wf, hf, face, ts, colon, "#ff0000")

		// place weather icon
		if imIcon.image != nil {
			dc.DrawImageAnchored(imIcon.image, int(wf/2), int(0.71875*hf), 0.5, 0.5)
		}

		if !mode {
			temps = strings.Split(w.Current.Temperature, " ")
		} else {
			p := w.Current.Daypart0.Precipitation
			if 0 == int(s)%2 {
				if togweather && `0%` != p {
					temps[1] = p
					// precipitation
					if imPrecip.image != nil {
						wdx := float64(cx + (wf * 0.09))
						wdy := float64(0.28 * hf)
						if `100%` == p {
							wdx += 5.00
						}
						dc.DrawImageAnchored(imPrecip.image, int(wdx), int(wdy), 0.5, 0.5)
					}
				} else {
					temps[1] = w.Current.Humidity
					// humidity
					if imHumid.image != nil {
						wdx := float64(cx + (wf * 0.09))
						wdy := float64(0.28 * hf)
						dc.DrawImageAnchored(imHumid.image, int(wdx), int(wdy), 0.5, 0.5)
					}
				}
			} else {
				togweather = !togweather
				temps = strings.Split(w.Current.Wind, " ")
				// place wind icon
				if imWindDir.image != nil {
					wdx := float64(cx + (wf * 0.09375))
					wdy := float64(0.312 * hf)
					dc.DrawImageAnchored(imWindDir.image, int(wdx), int(wdy), 0.5, 0.5)
				}
			}
		}
		dc.SetFontFace(sface)
		if !mode {
			dc.SetHexColor("#0099ff")
			dc.DrawStringAnchored(temps[0], wf/2, hf*0.2, 0.5, 0.5)
		}
		dc.SetHexColor("#66ff99")
		if !mode {
			dc.DrawStringAnchored(temps[1], wf/2, hf*0.32, 0.5, 0.5)
		} else {
			dc.DrawStringAnchored(temps[1], wf/3, hf*0.27, 0.5, 0.5)
		}

		if detail && W > 64 {
			placeWeatherDetail(dc, hf, dpface)
		}

		if `full` == layout {
			dc.SetHexColor("#ff9900")
			dc.SetFontFace(dtface)
			temps = hackaDate(t)
			dpos := 2 + (5 * (hf / 8))
			dc.DrawStringAnchored(temps[idx[0]], wf/4, dpos, 0.5, 0.5)
			dc.DrawStringAnchored(temps[idx[1]], 3*(wf/4), dpos, 0.5, 0.5)
			if showbright {
				dc.DrawStringAnchored(evut, cx, hf-length, 0.5, 0.5)
			}
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
			dc.DrawImageAnchored(lms.Volume(), W-31, pos, 0, 0.5)

			placeBorderZone(dc, lmsface, lw, 68, 50)
			drawHorizontalBar(dc, 10, 115, lms.Player.Percent)
			base := float64(H - 9 + 4)
			dc.SetHexColor("#ff9900")
			dc.DrawStringAnchored(lms.Player.TimeStr, 16, base, 0.5, 0.5)
			if remaining {
				if len(lms.Player.RemStr) == 8 {
					dc.DrawStringAnchored(lms.Player.RemStr, float64(W-22), base, 0.5, 0.5)
				} else {
					dc.DrawStringAnchored(lms.Player.RemStr, float64(W-18), base, 0.5, 0.5)
				}
			} else {
				if len(lms.Player.DurStr) == 8 {
					dc.DrawStringAnchored(lms.Player.DurStr, float64(W-20), base, 0.5, 0.5)
				} else {
					dc.DrawStringAnchored(lms.Player.DurStr, float64(W-16), base, 0.5, 0.5)
				}
			}
			dc.SetHexColor("#0099ffcc")
			dc.DrawStringAnchored(lms.Player.Bitty, float64(W/2), base, 0.5, 0.5)
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

		if dump < 500 {
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

func placeTime(dc *gg.Context, wf, hf float64, face font.Face, ts, colon, color string) {
	dc.SetHexColor(color)
	dc.SetFontFace(face)
	dc.DrawStringAnchored(ts, wf/2, hf*0.47, 0.5, 0.5)
	dc.DrawStringAnchored(colon, wf/2, hf*0.43, 0.5, 0.5)
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
	dc.DrawRectangle(0, 67, 128, y1)
	dc.Stroke()
	dc.SetLineWidth(0.5)
	dc.SetHexColor("#ff9900")
	dc.DrawRectangle(0, 67, 128, y2)
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
	dc.SetLineWidth(0.4)
	l := float64(W) - (2 * x)
	lp := (l - 2.00) * (pcnt / 100.00)
	dc.SetHexColor("#000000")
	dc.DrawRectangle(x+1, y+1, l-2, 2)
	dc.Fill()
	dc.SetHexColor("#ff9900") // bar color
	dc.SetLineWidth(1)
	dc.DrawRectangle(x+1, y+1, lp, 2)
	dc.Fill()
	dc.SetHexColor("#ff9900")
	dc.DrawRectangle(x, y, l, 4)
	dc.Stroke()
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
