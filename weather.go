package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"image/png"
	"os"
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
		//fmt.Println(">"+current+"<", i)
		ic.m.Lock()
		ic.image, err = getImageIconWIP(i)
		if err != nil {
			fmt.Println(err)
			ic.image, err = getImageIcon(i)
		}
		ic.m.Unlock()
		if err != nil {
			return ic, err
		}
		ic.m.Lock()
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

	imIconDP1, err = cacheImage(w.Current.Daypart1.Icon, imIconDP1, 0.7, ``)
	checkFatal(err)

	imIconDP2, err = cacheImage(w.Current.Daypart2.Icon, imIconDP2, 0.7, ``)
	checkFatal(err)

	imIconDP3, err = cacheImage(w.Current.Daypart3.Icon, imIconDP3, 0.7, ``)
	checkFatal(err)

	imIconDP4, err = cacheImage(w.Current.Daypart4.Icon, imIconDP4, 0.7, ``)
	checkFatal(err)

	/*
			if w.Current.Beafort != lastWindIcon {
				i := getIcon(fmt.Sprintf("wind-%d", w.Current.Beafort))
				imWind, err = getImageIcon(i)
		                checkFatal(err)
			}
			lastWindIcon = w.Current.Beafort
	*/

	test := fmt.Sprintf("wind-%s", strings.Split(w.Current.Wind, " ")[0])
	imWindDir, err = cacheImage(test, imWindDir, 0.00, ``)
	checkFatal(err)

	test = w.Current.Sunrise + "-" + w.Current.Sunset
	if lastHorizon != test {
		sunrise, err = parseTime(w.Current.Sunrise)
		checkFatal(err)
		sunset, err = parseTime(w.Current.Sunset)
		checkFatal(err)
	}
	lastHorizon = test

	check, _ := parseTime(time.Now().Format("03:04 PM"))
	if check.After(sunrise) && check.Before(sunset.Add(-time.Minute)) {
		brightness = daybright
		evut = nextEvent(check, sunset)
	} else {
		brightness = nightbright
		nooner, _ := parseTime(`11:59 AM`)
		if check.Before(nooner) {
			evut = nextEvent(check, sunrise)
		} else {
			evut = nextEvent(check, sunrise.Add(24*time.Hour))
		}
	}

	hr, _, _ := time.Now().Clock()
	if hr != lastHour {
		m := NewLuna(time.Now())
		test := fmt.Sprintf("moon-%d", m.PhaseFix())
		imMoon, err = cacheImage(test, imMoon, 0.00, ``)
		checkFatal(err)
	}
	lastHour = hr

	// a little tweaking
	w.Current.Humidity = strings.Replace(w.Current.Humidity, `%`, ``, -1)

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
