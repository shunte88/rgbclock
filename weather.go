package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"image/draw"
	"image/png"
	"os"

	svg "github.com/ajstarks/svgo"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// Daypart forecast period detail
type Daypart struct {
	Hilo          string `json:"hilo"`
	Icon          string `json:"icon"`
	ID            string `json:"id"`
	Label         string `json:"label"`
	Temperature   string `json:"temperature"`
	Precipitation string `json:"pcntprecip"`
}

// Weather server details
type Weather struct {
	Current struct {
		Beafort     int     `json:"beafort"`
		Daypart0    Daypart `json:"daypart-0"`
		Daypart1    Daypart `json:"daypart-1"`
		Daypart2    Daypart `json:"daypart-2"`
		Daypart3    Daypart `json:"daypart-3"`
		Daypart4    Daypart `json:"daypart-4"`
		DewPoint    string  `json:"dew point"`
		Feels       string  `json:"feels"`
		Humidity    string  `json:"humidity"`
		Icon        string  `json:"Icon"`
		Joke        string  `json:"joke"`
		Phrase      string  `json:"phrase"`
		Pressure    string  `json:"pressure"`
		Price       float64 `json:"price,omitempty"`
		Sunrise     string  `json:"sunrise"`
		Sunset      string  `json:"sunset"`
		Temperature string  `json:"temp"`
		Ticker      string  `json:"ticker,omitempty"`
		Visibility  string  `json:"visibility"`
		Wind        string  `json:"wind"`
		tempF       float64
		tempC       float64
	} `json:"current"`
}

func parseTime(t string) (time.Time, error) {
	t = strings.ToUpper(fmt.Sprintf("%08s", t))
	return time.Parse("03:04 PM", t)
}

func cacheImage(current string, ic iconCache, scale float64, color string) (iconCache, error) {

	var err error
	if ic.last != current {

		i := getIcon(current)
		if 0.00 != scale {
			i.scale = scale
		}
		if `` != color {
			i.color = color
		}

		ic.m.Lock()
		ic.image, err = getImageIconWIP(i)
		if err != nil {
			fmt.Println(err)
			ic.image, err = getImageIcon(i)
		}
		if err != nil {
			ic.m.Unlock()
			return ic, err
		}
		ic.last = current
		ic.m.Unlock()
		return ic, nil
	}
	return ic, nil
}

func cacheThermo(current string, ic iconCache, sw, sh int) (iconCache, error) {

	var err error
	if ic.last != current {

		ic.m.Lock()
		ic.image, err = thermometer(sw, sh)
		if err != nil {
			ic.m.Unlock()
			return ic, err
		}
		ic.last = current
		ic.m.Unlock()
		return ic, nil
	}
	return ic, nil
}

func cacheImageMethod(current string, ic iconCache, sw, sh int, f func(int, int) (draw.Image, error)) (iconCache, error) {

	var err error
	if ic.last != current {

		ic.m.Lock()
		ic.image, err = f(sw, sh)
		if err != nil {
			ic.m.Unlock()
			return ic, err
		}
		ic.last = current
		ic.m.Unlock()
		return ic, nil
	}
	return ic, nil
}

var lastHour int = -1
var snap bool = true

func weather() {

	snap = false
	test := ``
	var netClient = &http.Client{
		Timeout: time.Second * 5,
	}
	res, err := netClient.Get(weatherserveruri)
	if nil != err {
		return
	}

	resp, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if nil != err {
		return
	}

	err = json.Unmarshal(resp, &w)
	checkFatal(err)

	w.Current.Temperature =
		strings.Replace(w.Current.Temperature, ".0", "", -1)

	test = w.Current.Sunrise + "-" + w.Current.Sunset
	if lastHorizon != test {
		sunrise, err = parseTime(w.Current.Sunrise)
		checkFatal(err)
		sunset, err = parseTime(w.Current.Sunset)
		checkFatal(err)
	}
	lastHorizon = test

	refI := daymode
	check, _ := parseTime(time.Now().Format("03:04 PM"))
	if check.After(sunrise) && check.Before(sunset.Add(-time.Minute)) {
		daymode = daylight{daybright, true}
		evut = nextEvent(check, sunset)
	} else {
		daymode = daylight{nightbright, false}
		nooner, _ := parseTime(`11:59 AM`)
		if check.Before(nooner) {
			evut = nextEvent(check, sunrise)
		} else {
			evut = nextEvent(check, sunrise.Add(24*time.Hour))
		}
	}

	// force refresh on boundary - for conditions that have a noctuque in use - icon-33 24 hr issue
	if refI != daymode {
		imIcon, _ = cacheImage(`icon-0`, imIcon, 0.00, ``)
	}
	imIcon, err = cacheImage(w.Current.Icon, imIcon, 0.00, ``)
	checkFatal(err)
	// debug here
	if snap {
		ft, err := os.Create("test.png")
		checkFatal(err)
		defer ft.Close()
		png.Encode(ft, imIcon.image)
		snap = false
	}

	imIconDP1, err = cacheImage(w.Current.Daypart1.Icon, imIconDP1, 0.35, ``)
	checkFatal(err)

	imIconDP2, err = cacheImage(w.Current.Daypart2.Icon, imIconDP2, 0.35, ``)
	checkFatal(err)

	imIconDP3, err = cacheImage(w.Current.Daypart3.Icon, imIconDP3, 0.35, ``)
	checkFatal(err)

	imIconDP4, err = cacheImage(w.Current.Daypart4.Icon, imIconDP4, 0.35, ``)
	checkFatal(err)

	/*
			if w.Current.Beafort != lastWindIcon {
				i := getIcon(fmt.Sprintf("wind-%d", w.Current.Beafort))
				imWind, err = getImageIcon(i)
		                checkFatal(err)
			}
			lastWindIcon = w.Current.Beafort
	*/

	test = fmt.Sprintf("wind-%s", strings.Split(w.Current.Wind, " ")[0])
	imWindDir, err = cacheImage(test, imWindDir, 0.00, ``)
	checkFatal(err)

	hr, _, _ := time.Now().Clock()

	lastHour = hr

	// a little tweaking
	w.Current.Humidity = strings.Replace(w.Current.Humidity, `%`, ``, -1)

	re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
	temps := re.FindAllString(w.Current.Temperature, -1)
	w.Current.tempF, _ = strconv.ParseFloat(temps[0], 64)
	w.Current.tempC, _ = strconv.ParseFloat(temps[1], 64)

	tx := int(float64(clockw) * 0.4)
	imThermo, err = cacheThermo(temps[0], imThermo, tx, tx)

}

func nextEvent(t1, t2 time.Time) string {
	ret := t2.Sub(t1).String()
	ret = strings.Replace(
		strings.Replace(
			strings.Replace(ret, `h`, `:`, -1), `m0s`, ``, -1), `s`, ``, -1)
	if ret[0] == '-' {
		ret = ``
	} else {
		t := strings.Split(ret, `:`)
		for i, s := range t {
			if len(s) < 2 {
				t[i] = `0` + s
			}
		}
		if len(t) > 1 {
			ret = t[0] + `:` + t[1]
		} else {
			ret = `00:` + t[0]
		}
	}
	return ret
}

//var dump = true

func thermometer(sw, sh int) (img draw.Image, err error) {

	// dynamic SVG

	wt, ht := 120, 120

	img = image.NewRGBA(image.Rect(0, 0, sw, sh))
	var iconMem = new(bytes.Buffer)

	var canvas = svg.New(iconMem)
	canvas.Start(wt, ht)

	bfs := svg.Filterspec{In: `SourceGraphic`}
	canvas.Def()
	canvas.Filter(`blur_f`, `filterUnits="objectBoundingBox" x="-50%" y="-50%" width="200%" height="200%"`)
	canvas.FeGaussianBlur(bfs, 6, 6)
	canvas.Fend()
	canvas.Filter(`blur_z`, `filterUnits="objectBoundingBox" x="-50%" y="-50%" width="200%" height="200%"`)
	canvas.FeGaussianBlur(bfs, 2, 2)
	canvas.Fend()
	canvas.Filter(`blur_k`, `filterUnits="objectBoundingBox"`)
	canvas.FeGaussianBlur(bfs, 1, 1)
	canvas.Fend()
	canvas.DefEnd()

	degF := `m12.04517,17c0,-1.49271 0.53737,-2.77643 1.58227,-3.82133s2.32862,-1.58227 3.82133,-1.58227c1.46285,0 2.74658,0.53737 3.79148,1.58227c1.04489,1.07475 1.58227,2.32862 1.58227,3.82133c0,1.49271 -0.53737,2.77643 -1.58227,3.85118c-1.04489,1.07475 -2.32862,1.61212 -3.79148,1.61212s-2.74658,-0.53737 -3.82133,-1.61212c-1.04489,-1.07475 -1.58227,-2.35848 -1.58227,-3.85118zm2.62716,0c0,0.77621 0.26869,1.433 0.80606,2.00023c0.56723,0.56723 1.22402,0.83592 2.00023,0.83592s1.433,-0.26869 2.00023,-0.83592s0.83592,-1.22402 0.83592,-2.00023c0,-0.77621 -0.26869,-1.433 -0.83592,-1.97037c-0.56723,-0.53737 -1.22402,-0.83592 -2.00023,-0.83592c-0.77621,0 -1.433,0.26869 -2.00023,0.80606c-0.53737,0.53737 -0.80606,1.19417 -0.80606,2.00023zm13.16568,20.5695c0,0.41796 0.14927,0.80606 0.44781,1.1046s0.68665,0.44781 1.1046,0.44781c0.41796,0 0.80606,-0.14927 1.1046,-0.44781c0.29854,-0.29854 0.44781,-0.68665 0.44781,-1.1046l0,-11.31472l8.53828,0c0.41796,0 0.80606,-0.14927 1.1046,-0.47767s0.44781,-0.68665 0.44781,-1.13446c0,-0.44781 -0.14927,-0.80606 -0.44781,-1.13446c-0.29854,-0.29854 -0.68665,-0.44781 -1.13446,-0.44781l-8.53828,0l0,-8.15018l11.40428,0c0.41796,0 0.77621,-0.14927 1.07475,-0.44781s0.41796,-0.68665 0.41796,-1.13446s-0.14927,-0.80606 -0.41796,-1.13446s-0.62694,-0.44781 -1.07475,-0.44781l-14.24042,0c-0.20898,0 -0.29854,0.11942 -0.29854,0.3284l0,25.49544l0.05971,0z`
	degC := `m75.50005,17c0,-1.49186 0.53707,-2.77486 1.58137,-3.81916c1.07414,-1.07414 2.3273,-1.58137 3.81916,-1.58137c1.46202,0 2.74502,0.53707 3.78933,1.58137c1.0443,1.07414 1.58137,2.3273 1.58137,3.81916c0,1.49186 -0.53707,2.77486 -1.58137,3.81916c-1.0443,1.07414 -2.3273,1.58137 -3.78933,1.58137c-1.49186,0 -2.77486,-0.53707 -3.81916,-1.58137s-1.58137,-2.3273 -1.58137,-3.81916zm2.62567,0c0,0.77577 0.26853,1.43219 0.8056,1.99909c0.56691,0.56691 1.22333,0.83544 1.99909,0.83544c0.77577,0 1.43219,-0.26853 1.99909,-0.83544s0.83544,-1.22333 0.83544,-1.99909c0,-0.77577 -0.26853,-1.43219 -0.83544,-1.99909s-1.22333,-0.83544 -1.99909,-0.83544c-0.77577,0 -1.43219,0.26853 -1.99909,0.83544c-0.53707,0.53707 -0.8056,1.22333 -0.8056,1.99909zm11.60667,13.18805c0,2.29747 0.62658,4.3264 1.90958,6.11663c0.65642,0.92495 1.58137,1.67088 2.77486,2.23779c1.16365,0.53707 2.50633,0.83544 3.99819,0.83544c4.35623,0 7.10126,-1.67088 8.20523,-4.98281c0.11935,-0.41772 0.05967,-0.83544 -0.17902,-1.22333c-0.2387,-0.38788 -0.56691,-0.59674 -0.98463,-0.68626c-0.41772,-0.11935 -0.83544,-0.05967 -1.19349,0.20886c-0.35805,0.2387 -0.59674,0.56691 -0.68626,1.01447c0,0.02984 0,0.05967 -0.02984,0.14919l-0.05967,0.20886c-0.32821,0.56691 -0.77577,1.01447 -1.34267,1.34267c-0.92495,0.56691 -2.14828,0.83544 -3.66998,0.83544c-0.92495,0 -1.7604,-0.14919 -2.47649,-0.4774c-1.19349,-0.50723 -2.02893,-1.40235 -2.53616,-2.65551c-0.32821,-0.8056 -0.50723,-1.79023 -0.50723,-2.89421l0,-9.60758c0,-0.44756 0.02984,-0.89512 0.08951,-1.34267c0.11935,-1.13381 0.56691,-2.17812 1.34267,-3.10307c0.86528,-1.0443 2.23779,-1.55153 4.11753,-1.55153c1.55153,0 2.77486,0.26853 3.66998,0.8056c0.59674,0.35805 1.0443,0.8056 1.34267,1.34267c0.02984,0.05967 0.02984,0.14919 0.05967,0.2387c0.02984,0.08951 0.02984,0.14919 0.02984,0.17902c0.11935,0.41772 0.35805,0.71609 0.68626,0.89512c0.35805,0.20886 0.74593,0.2387 1.19349,0.14919c0.41772,-0.08951 0.74593,-0.32821 0.98463,-0.68626c0.2387,-0.35805 0.29837,-0.74593 0.17902,-1.19349l0,-0.02984l-0.2387,-0.68626c-0.14919,-0.32821 -0.41772,-0.77577 -0.83544,-1.283c-0.38788,-0.53707 -0.86528,-0.95479 -1.34267,-1.31284c-0.62658,-0.44756 -1.43219,-0.8056 -2.44665,-1.13381c-1.01447,-0.29837 -2.11844,-0.44756 -3.31193,-0.44756c-1.5217,0 -2.83453,0.26853 -4.02802,0.8056c-1.16365,0.53707 -2.0886,1.25316 -2.71519,2.17812c-1.283,1.7604 -1.93942,3.81916 -1.93942,6.1763l0,9.57774l-0.05967,0z`

	// rear plane shadow -
	canvas.Group(`style="fill:black;filter:url(#blur_f);"`)
	canvas.Path("m39.90978,78.17153c0,-3.4036 0.80084,-6.56694 2.36249,-9.53009s3.76398,-5.40572 6.60699,-7.3678l0,-39.72202c0,-3.20339 1.08115,-5.92626 3.28348,-8.12859s4.9252,-3.36355 8.12859,-3.36355c3.24344,0 5.96631,1.12118 8.16864,3.32351c2.20233,2.24237 3.32351,4.9252 3.32351,8.12859l0,39.72202c2.84301,1.96208 5.00529,4.4447 6.56694,7.3678s2.32246,6.12647 2.32246,9.53009c0,3.68389 -0.92097,7.12753 -2.72287,10.25083s-4.28452,5.60593 -7.40783,7.40783s-6.5269,2.72287 -10.25083,2.72287c-3.68389,0 -7.08749,-0.92097 -10.21079,-2.72287s-5.60593,-4.28452 -7.44787,-7.40783s-2.72287,-6.5269 -2.72287,-10.21079l-0.00004,0zm7.04746,0c0,3.72394 1.3214,6.92733 3.92416,9.57013c2.60276,2.6428 5.7661,3.96418 9.45,3.96418c3.72394,0 6.92733,-1.3214 9.61015,-4.00423s4.04427,-5.84617 4.04427,-9.49004c0,-2.48262 -0.64068,-4.80509 -1.92203,-6.92733c-1.28136,-2.12224 -3.04321,-3.76398 -5.28559,-4.9252l-1.12118,-0.56059c-0.40043,-0.16016 -0.60063,-0.56059 -0.60063,-1.16122l0,-43.08557c0,-1.28136 -0.44047,-2.36249 -1.36145,-3.24344c-0.92097,-0.84088 -2.04215,-1.28136 -3.4036,-1.28136c-1.28136,0 -2.40253,0.44047 -3.32351,1.28136c-0.92097,0.84088 -1.36145,1.92203 -1.36145,3.24344l0,43.00548c0,0.60063 -0.2002,1.00106 -0.56059,1.16122l-1.08115,0.56059c-2.20233,1.16122 -3.92416,2.80296 -5.16547,4.9252c-1.24131,2.12224 -1.84195,4.40466 -1.84195,6.96737l0.00002,0.00001zm3.1233,0c0,2.84301 0.96102,5.28559 2.9231,7.28771s4.28452,3.00317 7.04746,3.00317s5.12543,-1.00106 7.16758,-3.00317c2.04215,-2.00212 3.04321,-4.4447 3.04321,-7.24767c0,-2.52267 -0.88093,-4.76504 -2.60276,-6.68708c-1.72183,-1.92203 -3.84407,-3.08326 -6.32669,-3.4036l-2.44258,-0.08011c-2.44258,0.36038 -4.52479,1.48156 -6.2466,3.4036c-1.72183,1.96208 -2.56271,4.16441 -2.56271,6.72712l0,0.00002l-0.00001,0.00001z")
	canvas.Path(degF)
	canvas.Path(degC)
	canvas.Gend()

	// glassback -
	canvas.Path("m39.90978,78.17153c0,-3.4036 0.80084,-6.56694 2.36249,-9.53009s3.76398,-5.40572 6.60699,-7.3678l0,-39.72202c0,-3.20339 1.08115,-5.92626 3.28348,-8.12859s4.9252,-3.36355 8.12859,-3.36355c3.24344,0 5.96631,1.12118 8.16864,3.32351c2.20233,2.24237 3.32351,4.9252 3.32351,8.12859l0,39.72202c2.84301,1.96208 5.00529,4.4447 6.56694,7.3678s2.32246,6.12647 2.32246,9.53009c0,3.68389 -0.92097,7.12753 -2.72287,10.25083s-4.28452,5.60593 -7.40783,7.40783s-6.5269,2.72287 -10.25083,2.72287c-3.68389,0 -7.08749,-0.92097 -10.21079,-2.72287s-5.60593,-4.28452 -7.44787,-7.40783s-2.72287,-6.5269 -2.72287,-10.21079l-0.00004,0z",
		`style="fill:green;fill-opacity:0.45"`)

	// glass -
	canvas.Path("m39.90978,78.17153c0,-3.4036 0.80084,-6.56694 2.36249,-9.53009s3.76398,-5.40572 6.60699,-7.3678l0,-39.72202c0,-3.20339 1.08115,-5.92626 3.28348,-8.12859s4.9252,-3.36355 8.12859,-3.36355c3.24344,0 5.96631,1.12118 8.16864,3.32351c2.20233,2.24237 3.32351,4.9252 3.32351,8.12859l0,39.72202c2.84301,1.96208 5.00529,4.4447 6.56694,7.3678s2.32246,6.12647 2.32246,9.53009c0,3.68389 -0.92097,7.12753 -2.72287,10.25083s-4.28452,5.60593 -7.40783,7.40783s-6.5269,2.72287 -10.25083,2.72287c-3.68389,0 -7.08749,-0.92097 -10.21079,-2.72287s-5.60593,-4.28452 -7.44787,-7.40783s-2.72287,-6.5269 -2.72287,-10.21079l-0.00004,0zm7.04746,0c0,3.72394 1.3214,6.92733 3.92416,9.57013c2.60276,2.6428 5.7661,3.96418 9.45,3.96418c3.72394,0 6.92733,-1.3214 9.61015,-4.00423s4.04427,-5.84617 4.04427,-9.49004c0,-2.48262 -0.64068,-4.80509 -1.92203,-6.92733c-1.28136,-2.12224 -3.04321,-3.76398 -5.28559,-4.9252l-1.12118,-0.56059c-0.40043,-0.16016 -0.60063,-0.56059 -0.60063,-1.16122l0,-43.08557c0,-1.28136 -0.44047,-2.36249 -1.36145,-3.24344c-0.92097,-0.84088 -2.04215,-1.28136 -3.4036,-1.28136c-1.28136,0 -2.40253,0.44047 -3.32351,1.28136c-0.92097,0.84088 -1.36145,1.92203 -1.36145,3.24344l0,43.00548c0,0.60063 -0.2002,1.00106 -0.56059,1.16122l-1.08115,0.56059c-2.20233,1.16122 -3.92416,2.80296 -5.16547,4.9252c-1.24131,2.12224 -1.84195,4.40466 -1.84195,6.96737z",
		`style="fill:lightskyblue;fill-opacity:0.3;stroke-width:0.4;stroke-opacity:0.2;stroke:midnightblue;"`)
	// capillary -
	canvas.Path("m46.95724,78.17153c0,3.72394 1.3214,6.92733 3.92416,9.57013c2.60276,2.6428 5.7661,3.96418 9.45,3.96418c3.72394,0 6.92733,-1.3214 9.61015,-4.00423s4.04427,-5.84617 4.04427,-9.49004c0,-2.48262 -0.64068,-4.80509 -1.92203,-6.92733c-1.28136,-2.12224 -3.04321,-3.76398 -5.28559,-4.9252l-1.12118,-0.56059c-0.40043,-0.16016 -0.60063,-0.56059 -0.60063,-1.16122l0,-43.08557c0,-1.28136 -0.44047,-2.36249 -1.36145,-3.24344c-0.92097,-0.84088 -2.04215,-1.28136 -3.4036,-1.28136c-1.28136,0 -2.40253,0.44047 -3.32351,1.28136c-0.92097,0.84088 -1.36145,1.92203 -1.36145,3.24344l0,43.00548c0,0.60063 -0.2002,1.00106 -0.56059,1.16122l-1.08115,0.56059c-2.20233,1.16122 -3.92416,2.80296 -5.16547,4.9252c-1.24131,2.12224 -1.84195,4.40466 -1.84195,6.96737z",
		`style="fill:linen;filter:url(#blur_shadow);fill-opacity:0.4;"`)

	canvas.Group(`style="stroke-width:0.4;stroke-opacity:0.2;stroke:white;"`)
	canvas.Path(degF, `style="fill:#0099ff;"`)
	canvas.Path(degC, `style="fill:#66ff99;"`)
	canvas.Gend()

	alcoCol := `red`
	alcoWidth := `3.0`
	if w.Current.tempF <= 32 {
		alcoCol = `blue`
	}
	// bulb -
	canvas.Circle(60, 78, 10, fmt.Sprintf("style=\"fill:%s\"", alcoCol))

	// temperature -
	// ( 31 = -20) .. (11 = 120)
	canvas.Group()
	canvas.Line(60, 78, 60, 62, fmt.Sprintf("style=\"stroke-width:%s;stroke:%s;stroke-linecap:round\"", alcoWidth, alcoCol))
	if w.Current.tempF > -20 {
		y2 := int(40.0 * (float64(20+w.Current.tempF) / 140.00))
		canvas.Line(60, 78, 60, 62-y2, fmt.Sprintf("style=\"stroke-width:%s;stroke:%s;stroke-linecap:round\"", alcoWidth, alcoCol))
	}
	canvas.Gend()

	// pop -
	canvas.Circle(60, 78, 8,
		`style="fill:orangered;filter:url(#blur_k);filter-opacity:0.1;fill-opacity:0.2;"`)
	canvas.Circle(60, 78, 7,
		`style="fill:white;filter:url(#blur_k);filter-opacity:0.1;fill-opacity:0.2;"`)
	canvas.Circle(56, 74, 4,
		`style="fill:white;filter:url(#blur_k);filter-opacity:0.1;fill-opacity:0.1;"`)

	// ticks
	for tick := 22; tick < 64; tick += 3 {
		dash := tick % 2
		canvas.Line(50, tick, 56+(10*dash), tick, `style="stroke-width:1;stroke:greenyellow;stroke-linecap:round;stroke-opacity:0.3;"`)
	}

	// glass 2 -
	canvas.Path("m39.90978,78.17153c0,-3.4036 0.80084,-6.56694 2.36249,-9.53009s3.76398,-5.40572 6.60699,-7.3678l0,-39.72202c0,-3.20339 1.08115,-5.92626 3.28348,-8.12859s4.9252,-3.36355 8.12859,-3.36355c3.24344,0 5.96631,1.12118 8.16864,3.32351c2.20233,2.24237 3.32351,4.9252 3.32351,8.12859l0,39.72202c2.84301,1.96208 5.00529,4.4447 6.56694,7.3678s2.32246,6.12647 2.32246,9.53009c0,3.68389 -0.92097,7.12753 -2.72287,10.25083s-4.28452,5.60593 -7.40783,7.40783s-6.5269,2.72287 -10.25083,2.72287c-3.68389,0 -7.08749,-0.92097 -10.21079,-2.72287s-5.60593,-4.28452 -7.44787,-7.40783s-2.72287,-6.5269 -2.72287,-10.21079l-0.00004,0z",
		`style="fill:lightskyblue;fill-opacity:0.2;stroke-width:1;stroke-opacity:0.4;stroke:midnightblue;"`)

	canvas.End()

	iconI, err := oksvg.ReadIconStream(iconMem)
	if err != nil {
		return img, err
	}

	gv := rasterx.NewScannerGV(wt, ht, img, img.Bounds())
	r := rasterx.NewDasher(wt, ht, gv)
	iconI.SetTarget(0, 0, float64(sw), float64(sh))
	iconI.Draw(r, 1.0)

	return img, nil

}
