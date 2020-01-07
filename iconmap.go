package main

import (
	"fmt"
	"os"
)

// SVG - svg path model
type SVG struct {
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
	Doc    string `xml:",innerxml"`
}

type icon struct {
	filename string
	asis     bool
	color    string
	scale    float64
	rotate   float64
	alpha    float64
	width    int
	height   int
	shadow   bool
	blur     bool
	popcolor string
}

var iconMap map[string]icon

func mapInit() {
	windDegAsIs := true
	// configured miScale is based on a 30x30 viewport, adjust for asis svg setups - see wic-rain for example
	im := map[string]icon{
		"icon-0":            icon{filename: "wic-tornado", asis: true, color: "red", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                          // tornado
		"icon-1":            icon{filename: "wic-tornado", asis: true, color: "red", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                          // tropical storm
		"icon-2":            icon{filename: "wic-hurricane", asis: true, color: "red", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                        // hurricane
		"icon-3":            icon{filename: "wic-thunderstorm", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                   // severe thunderstorms
		"icon-4":            icon{filename: "wic-lightning", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                      // thunderstorms
		"icon-5":            icon{filename: "wic-rain-mix", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                       // mixed rain and snow
		"icon-6":            icon{filename: "wic-rain-mix", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                       // mixed rain and sleet
		"icon-7":            icon{filename: "wic-day-sleet-storm", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 3, shadow: true},                // mixed snow and sleet
		"icon-8":            icon{filename: "wic-day-sleet", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                      // freezing drizzle
		"icon-9":            icon{filename: "wic-sprinkle", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                       // drizzle
		"icon-10":           icon{filename: "wic-rain-wind", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                      // freezing rain
		"icon-11":           icon{filename: "wic-sprinkle", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                       // light rain
		"icon-12":           icon{filename: "wic-rain", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                           // heavy rain
		"icon-13":           icon{filename: "wic-day-snow-wind", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                  // snow flurries
		"icon-14":           icon{filename: "wic-day-snow", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                       // light snow showers
		"icon-15":           icon{filename: "wic-snow-wind", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                      // blowing snow
		"icon-16":           icon{filename: "wic-snow", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                           // snow
		"icon-17":           icon{filename: "wic-day-hail", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                       // hail
		"icon-18":           icon{filename: "wic-sleet", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                          // sleet
		"icon-19":           icon{filename: "wic-dust", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                           // dust
		"icon-20":           icon{filename: "wic-fog", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                            // foggy
		"icon-21":           icon{filename: "wic-day-haze", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                       // haze
		"icon-22":           icon{filename: "wic-smoke", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                          // smoky
		"icon-23":           icon{filename: "wic-windy", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                          // blustery
		"icon-24":           icon{filename: "wic-windy", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                          // windy
		"icon-25":           icon{filename: "wic-cloudy", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                         // cold
		"icon-26":           icon{filename: "wic-cloudy", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                         // cloudy
		"icon-27":           icon{filename: "wic-night-cloudy", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                   // mostly cloudy (night)
		"icon-28":           icon{filename: "wic-cloudy", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                         // mostly cloudy (day)
		"icon-29":           icon{filename: "wic-night-partly-cloudy", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},            // partly cloudy (night)
		"icon-30":           icon{filename: "wic-day-cloudy", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                     // partly cloudy (day)
		"icon-31":           icon{filename: "wic-night-clear", asis: true, color: "linen", width: 60, height: 60, alpha: .8, scale: miScale / 2, shadow: true},                         // clear (night)
		"icon-32":           icon{filename: "wic-day-sunny", asis: true, color: "yellow", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true, popcolor: "yellow"}, // sunny
		"icon-33":           icon{filename: "wic-stars", asis: true, color: "linen", width: 60, height: 60, alpha: .8, scale: miScale / 2, shadow: true},                               // fair (night)
		"icon-34":           icon{filename: "wic-day-sunny", asis: true, color: "yellow", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                     // fair (day)
		"icon-35":           icon{filename: "wic-hail", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                           // mixed rain and hail
		"icon-36":           icon{filename: "wic-hot", asis: true, color: "yellow", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                           // hot
		"icon-37":           icon{filename: "wic-thunderstorm", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                   // isolated thunderstorms
		"icon-38":           icon{filename: "wic-storm-showers", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                  // scattered thunderstorms
		"icon-39":           icon{filename: "wic-rain", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                           // scattered rain
		"icon-40":           icon{filename: "wic-rain", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                           // heavy rain
		"icon-41":           icon{filename: "wic-snowflake-cold", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                 // heavy snow
		"icon-42":           icon{filename: "wic-snow", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                           // scattered snow showers
		"icon-43":           icon{filename: "wic-snow-wind", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                      // blowing heavy snow
		"icon-44":           icon{filename: "wic-day-cloudy", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                     // partly cloudy (day)
		"icon-45":           icon{filename: "wic-night-thunderstorm", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},             // thundershowers (night)
		"icon-46":           icon{filename: "wic-night-snow", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},                     // snow showers (night)
		"icon-47":           icon{filename: "wic-night-thunderstorm", asis: true, color: "linen", width: 60, height: 60, alpha: miAlpha, scale: miScale / 2, shadow: true},             // isolated thundershowers (night)
		"clock-0":           icon{filename: "wic-time-12", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-1":           icon{filename: "wic-time-1", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-2":           icon{filename: "wic-time-2", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-3":           icon{filename: "wic-time-3", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-4":           icon{filename: "wic-time-4", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-5":           icon{filename: "wic-time-5", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-6":           icon{filename: "wic-time-6", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-7":           icon{filename: "wic-time-7", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-8":           icon{filename: "wic-time-8", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-9":           icon{filename: "wic-time-9", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-10":          icon{filename: "wic-time-10", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-11":          icon{filename: "wic-time-11", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-12":          icon{filename: "wic-time-12", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-13":          icon{filename: "wic-time-1", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-14":          icon{filename: "wic-time-2", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-15":          icon{filename: "wic-time-3", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-16":          icon{filename: "wic-time-4", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-17":          icon{filename: "wic-time-5", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-18":          icon{filename: "wic-time-6", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-19":          icon{filename: "wic-time-7", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-20":          icon{filename: "wic-time-8", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-21":          icon{filename: "wic-time-9", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-22":          icon{filename: "wic-time-10", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"clock-23":          icon{filename: "wic-time-11", asis: true, color: "white", width: 60, height: 60, alpha: 1.0, scale: 1.0},
		"wind-0":            icon{filename: "wi-wind-beaufort-0", color: "yellowgreen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},
		"wind-1":            icon{filename: "wi-wind-beaufort-1", color: "yellowgreen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},
		"wind-2":            icon{filename: "wi-wind-beaufort-2", color: "yellowgreen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},
		"wind-3":            icon{filename: "wi-wind-beaufort-3", color: "yellowgreen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},
		"wind-4":            icon{filename: "wi-wind-beaufort-4", color: "yellowgreen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},
		"wind-5":            icon{filename: "wi-wind-beaufort-5", color: "darkorange", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},
		"wind-6":            icon{filename: "wi-wind-beaufort-6", color: "darkorange", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},
		"wind-7":            icon{filename: "wi-wind-beaufort-7", color: "darkorange", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},
		"wind-8":            icon{filename: "wi-wind-beaufort-8", color: "darkorange", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},
		"wind-9":            icon{filename: "wi-wind-beaufort-9", color: "crimson", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},
		"wind-10":           icon{filename: "wi-wind-beaufort-10", color: "crimson", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},
		"wind-11":           icon{filename: "wi-wind-beaufort-11", color: "crimson", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},
		"wind-12":           icon{filename: "wi-wind-beaufort-12", color: "crimson", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},
		"wind-N":            icon{filename: windDegIcon, asis: windDegAsIs, color: wiColor, width: 60, height: 60, rotate: 180.0, scale: wiScale / 2, alpha: wiAlpha, shadow: true}, // N
		"wind-NNE":          icon{filename: windDegIcon, asis: windDegAsIs, color: wiColor, width: 60, height: 60, rotate: 202.5, scale: wiScale / 2, alpha: wiAlpha, shadow: true}, // NNE
		"wind-NE":           icon{filename: windDegIcon, asis: windDegAsIs, color: wiColor, width: 60, height: 60, rotate: 225.0, scale: wiScale / 2, alpha: wiAlpha, shadow: true}, // NE
		"wind-ENE":          icon{filename: windDegIcon, asis: windDegAsIs, color: wiColor, width: 60, height: 60, rotate: 247.5, scale: wiScale / 2, alpha: wiAlpha, shadow: true}, // ENE
		"wind-E":            icon{filename: windDegIcon, asis: windDegAsIs, color: wiColor, width: 60, height: 60, rotate: 270.0, scale: wiScale / 2, alpha: wiAlpha, shadow: true}, // E
		"wind-ESE":          icon{filename: windDegIcon, asis: windDegAsIs, color: wiColor, width: 60, height: 60, rotate: 292.5, scale: wiScale / 2, alpha: wiAlpha, shadow: true}, // ESE
		"wind-SE":           icon{filename: windDegIcon, asis: windDegAsIs, color: wiColor, width: 60, height: 60, rotate: 315.0, scale: wiScale / 2, alpha: wiAlpha, shadow: true}, // SE
		"wind-SSE":          icon{filename: windDegIcon, asis: windDegAsIs, color: wiColor, width: 60, height: 60, rotate: 337.5, scale: wiScale / 2, alpha: wiAlpha, shadow: true}, // SSE
		"wind-S":            icon{filename: windDegIcon, asis: windDegAsIs, color: wiColor, width: 60, height: 60, rotate: 0.0, scale: wiScale / 2, alpha: wiAlpha, shadow: true},   // S
		"wind-SSW":          icon{filename: windDegIcon, asis: windDegAsIs, color: wiColor, width: 60, height: 60, rotate: 22.5, scale: wiScale / 2, alpha: wiAlpha, shadow: true},  // SSW
		"wind-SW":           icon{filename: windDegIcon, asis: windDegAsIs, color: wiColor, width: 60, height: 60, rotate: 45.0, scale: wiScale / 2, alpha: wiAlpha, shadow: true},  // SW
		"wind-WSW":          icon{filename: windDegIcon, asis: windDegAsIs, color: wiColor, width: 60, height: 60, rotate: 67.5, scale: wiScale / 2, alpha: wiAlpha, shadow: true},  // WSW
		"wind-W":            icon{filename: windDegIcon, asis: windDegAsIs, color: wiColor, width: 60, height: 60, rotate: 90.0, scale: wiScale / 2, alpha: wiAlpha, shadow: true},  // W
		"wind-WNW":          icon{filename: windDegIcon, asis: windDegAsIs, color: wiColor, width: 60, height: 60, rotate: 112.5, scale: wiScale / 2, alpha: wiAlpha, shadow: true}, // WNW
		"wind-NW":           icon{filename: windDegIcon, asis: windDegAsIs, color: wiColor, width: 60, height: 60, rotate: 135.0, scale: wiScale / 2, alpha: wiAlpha, shadow: true}, // NW
		"wind-NNW":          icon{filename: windDegIcon, asis: windDegAsIs, color: wiColor, width: 60, height: 60, rotate: 157.5, scale: wiScale / 2, alpha: wiAlpha, shadow: true}, // NNW
		"moon-0":            icon{filename: "wi-moon-alt-new", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"moon-1":            icon{filename: "wi-moon-alt-waxing-crescent-5", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"moon-2":            icon{filename: "wi-moon-alt-first-quarter", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"moon-3":            icon{filename: "wi-moon-alt-waxing-gibbous-5", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"moon-4":            icon{filename: "wi-moon-alt-full", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"moon-5":            icon{filename: "wi-moon-alt-waning-gibbous-5", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"moon-6":            icon{filename: "wi-moon-alt-third-quarter", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"moon-7":            icon{filename: "wi-moon-alt-waning-crescent-5", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"moon-8":            icon{filename: "wi-moon-alt-new", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"brolly":            icon{filename: "wic-umbrella", asis: true, color: "#66ff99", width: 60, height: 60, scale: (wiScale * .55), alpha: 1, shadow: true},
		"humidity":          icon{filename: "wic-humidity", asis: true, color: "#66ff99", width: 60, height: 60, scale: (wiScale * 0.55), alpha: .8, shadow: true},
		"snowflake":         icon{filename: "wic-snowflake-cold", asis: true, color: "#66ff99", width: 60, height: 60, scale: (wiScale * .6), alpha: 1, shadow: true},
		"Rapid Transit":     icon{filename: "mbta-t-train", color: "red", width: 30, height: 30, scale: 1.0, alpha: 1, shadow: true},
		"Commuter Rail":     icon{filename: "mbta-commuter", color: "purple", width: 30, height: 30, scale: 1.0, alpha: 1, shadow: true},
		"Local Bus":         icon{filename: "mbta-bus", color: "silver", width: 30, height: 30, scale: 1.0, alpha: 1, shadow: false},
		"The Ride":          icon{filename: "mbta-the-ride", color: "#52bbc5", width: 30, height: 30, scale: 1.0, alpha: 1, shadow: false},
		"Ferry":             icon{filename: "mbta-ferry", color: "#008eaa", width: 30, height: 30, scale: 1.0, alpha: 1, shadow: false},
		"corner-scroll":     icon{filename: "cscroll", color: "#0f344340", width: 60, height: 60, scale: 1.0, alpha: 0.25, shadow: false},
		"volume-on":         icon{filename: "volume-on", color: "#ff9900", width: 60, height: 60, scale: 1.0, alpha: .90, shadow: false},
		"volume-0":          icon{filename: "volume-0", color: "#ff9900", width: 60, height: 60, scale: 1.0, alpha: .90, shadow: false},
		"volume-1":          icon{filename: "volume-1", color: "#ff9900", width: 60, height: 60, scale: 1.0, alpha: .90, shadow: false},
		"volume-2":          icon{filename: "volume-2", color: "#ff9900", width: 60, height: 60, scale: 1.0, alpha: .90, shadow: false},
		"volume-3":          icon{filename: "volume-3", color: "#ff9900", width: 60, height: 60, scale: 1.0, alpha: .90, shadow: false},
		"volume-4":          icon{filename: "volume-4", color: "#ff9900", width: 60, height: 60, scale: 1.0, alpha: .90, shadow: false},
		"volume-mute":       icon{filename: "volume-mute", color: "orangered", width: 60, height: 60, scale: 1.0, alpha: .90, shadow: false},
		"repeat-1":          icon{filename: "repeat-one", color: "#ff9900", width: 50, height: 50, scale: 1.0, alpha: .90, shadow: false},
		"repeat-2":          icon{filename: "repeat-all", color: "#ff9900", width: 50, height: 50, scale: 1.0, alpha: .90, shadow: false},
		"shuffle-1":         icon{filename: "shuffle-song", color: "#ff9900", width: 50, height: 50, scale: 1.0, alpha: .90, shadow: false},
		"shuffle-2":         icon{filename: "shuffle-album", color: "#ff9900", width: 50, height: 50, scale: 1.0, alpha: .90, shadow: false},
		"alt-corner-scroll": icon{filename: "cscroll2", color: "#0f344340", width: 60, height: 60, scale: 1.0, alpha: 0.35, shadow: false},
		"globalz":           icon{filename: "globalz", color: "#fffffcc", width: 192, height: 192, scale: 1.0, alpha: 0.025, shadow: true, blur: true},
		"global":            icon{filename: "global", color: "#fffffcc", width: 192, height: 192, scale: 1.0, alpha: 0.025, shadow: true, blur: true},
		"glass":             icon{filename: "glass", asis: true, color: "white", width: 192, height: 192, scale: .666666, alpha: 1, shadow: false, blur: false},
		"skullz":            icon{filename: "arggggh", color: "#6D97ABcc", width: 100, height: 100, scale: 1.0, alpha: 0.5, shadow: true},
		"cpu-temp":          icon{filename: "cputc", asis: true, width: 60, height: 60, scale: 1.0, alpha: 1.0, shadow: false},
		"cpu-metrics":       icon{filename: "cpupc", asis: true, width: 60, height: 60, scale: 1.0, alpha: 1.0, shadow: false},
		"ram-metrics":       icon{filename: "memfree", asis: true, width: 60, height: 60, scale: 1.0, alpha: 1.0, shadow: false},
		"bunny":             icon{filename: "first-bunny", color: "#0f344340", width: 30, height: 30, scale: 1.0, alpha: 1, shadow: true},
	}
	iconMap = im
}

func getIcon(s string) icon {
	var v icon
	v, ok := iconMap[s]
	if !ok || !fileExists(iconFile(v)) {
		fmt.Println("icon missing?", s, v)
		v = icon{filename: "wic-alien", asis: true, color: "red", width: 60, height: 60, scale: 1, alpha: 1, shadow: false}
	}
	return v
}

func iconFile(i icon) string {
	return "svg/" + i.filename + ".svg"
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
