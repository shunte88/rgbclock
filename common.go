package main

import (
	"image/draw"
	"strings"
	"sync"
	"time"
)

type iconCache struct {
	last  string
	image draw.Image
	m     sync.Mutex
}

type feed struct {
	title string
	link  string
}

var (
	feeds       []interface{}
	folding     bool = false
	detail      bool = false
	clockw      int  = 64
	clockh      int  = 64
	mode        bool
	cols        int     = 64
	rows        int     = 64
	W           int     = 64
	H           int     = 64
	parallel    int     = 1
	chain       int     = 1
	layout      string  = "simple"
	hardware    string  = "adafruit-hat-pwm"
	miAlpha     float64 = 0.60
	miW         int     = 27
	miScale     float64 = 0.90
	wiAlpha     float64 = 0.95
	wiW         int     = 24
	wiScale     float64 = 0.65
	daybright   int     = 20
	nightbright int     = 20
	// event timer display
	showbright       bool   = false
	evut             string = ``
	weatherserveruri string = ``
	precipitation    string = ``
	scroll           int    = 22
	w                Weather
	// we can override this via config
	iconStyle string = `style="fill: %s" fill-opacity="1.0" stroke-opacity="0.4" stroke="black" stroke-width="1"`

	imIcon      iconCache
	imWind      draw.Image
	imWindDir   iconCache
	imMoon      iconCache
	imIconDP1   iconCache
	imIconDP2   iconCache
	imIconDP3   iconCache
	imIconDP4   iconCache
	imPrecip    iconCache
	imSnow      iconCache
	imHumid     iconCache
	fontfile    = `font/LCDM2B__.TTF`
	fontfile2   = `font/Roboto-Thin.ttf`
	lastHorizon = ""
	brightness  = 20
	lastNews    = time.Now().Add(-24 * time.Hour)
	sunrise     = time.Now()
	sunset      = time.Now()
	lms         *LMSServer
)

const (
	weatherServerURI = "WEATHER_SERVER_URI"
	windDegIcon      = "wi-wind-deg"
	wiColor          = "#66ff99" //"yellow"
)

func sched(what func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			what()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}

func rotateBuffer(buffer []byte) []byte {
	return append(buffer[1:], buffer[0:1]...)
}

func containsI(a string, b string) bool {
	return strings.Contains(
		strings.ToLower(a),
		strings.ToLower(b),
	)
}

func toggleMode() {
	mode = !mode
}

func checkFatal(err error) {
	if err != nil {
		panic(err)
	}
}
