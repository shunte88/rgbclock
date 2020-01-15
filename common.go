package main

import (
	"fmt"
	"image/color"
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

type daylight struct {
	brightness int
	isdaylight bool
}

var (
	folding     bool = false
	detail      bool = false
	experiment  bool = false
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
	instrument       bool   = false
	evut             string = ``
	weatherserveruri string = ``
	precipitation    string = ``
	scroll           int    = 22
	w                Weather
	// we can override this via config
	iconStyle   string = `style="fill: %s" fill-opacity="1.0" stroke-opacity="0.4" stroke="black" stroke-width="1"`
	remaining          = false
	imIcon      iconCache
	imWind      draw.Image
	imWindDir   iconCache
	imIconDP1   iconCache
	imIconDP2   iconCache
	imIconDP3   iconCache
	imIconDP4   iconCache
	imPrecip    iconCache
	imSnow      iconCache
	imHumid     iconCache
	fontfile           = `font/LCDM2B__.TTF`
	fontfile2          = `font/Roboto-Thin.ttf`
	lastHorizon        = ``
	daymode            = daylight{20, false}
	colorgrad1  string = `#56ccf240`
	colorgrad2  string = `#2f80ed40`
	sunrise            = time.Now()
	sunset             = time.Now()
	news        *News
	lms         *LMSServer
	transit     *MBTA
)

const (
	weatherServerURI = "WEATHER_SERVER_URI"
	windDegIcon      = "wic-wind-deg"
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

func parseHexColor(x string) (c color.RGBA) {
	x = strings.TrimPrefix(x, "#")
	c.A = 255
	if len(x) == 3 {
		format := "%1x%1x%1x"
		fmt.Sscanf(x, format, &c.R, &c.G, &c.B)
		c.R |= c.R << 4
		c.G |= c.G << 4
		c.B |= c.B << 4
	}
	if len(x) == 6 {
		format := "%02x%02x%02x"
		fmt.Sscanf(x, format, &c.R, &c.G, &c.B)
	}
	if len(x) == 8 {
		format := "%02x%02x%02x%02x"
		fmt.Sscanf(x, format, &c.R, &c.G, &c.B, &c.A)
	}
	return
}

func intInSlice(a []int, x int) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func checkFatal(err error) {
	if err != nil {
		panic(err)
	}
}
