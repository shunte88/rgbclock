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
	im := map[string]icon{
		"icon-0":            icon{filename: "wi-tornado", color: "red", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                          // tornado
		"icon-1":            icon{filename: "wi-tornado", color: "red", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                          // tropical storm
		"icon-2":            icon{filename: "wi-hurricane", color: "red", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                        // hurricane
		"icon-3":            icon{filename: "wi-thunderstorm", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                   // severe thunderstorms
		"icon-4":            icon{filename: "wi-lightning", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                      // thunderstorms
		"icon-5":            icon{filename: "wi-rain-mix", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                       // mixed rain and snow
		"icon-6":            icon{filename: "wi-rain-mix", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                       // mixed rain and sleet
		"icon-7":            icon{filename: "wi-day-sleet-storm", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                // mixed snow and sleet
		"icon-8":            icon{filename: "wi-day-sleet", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                      // freezing drizzle
		"icon-9":            icon{filename: "wi-sprinkle", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                       // drizzle
		"icon-10":           icon{filename: "wi-rain-wind", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                      // freezing rain
		"icon-11":           icon{filename: "wi-sprinkle", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                       // light rain
		"icon-12":           icon{filename: "wi-rain", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                           // heavy rain
		"icon-13":           icon{filename: "wi-day-snow-wind", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                  // snow flurries
		"icon-14":           icon{filename: "wi-day-snow", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                       // light snow showers
		"icon-15":           icon{filename: "wi-snow-wind", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                      // blowing snow
		"icon-16":           icon{filename: "wi-snow", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                           // snow
		"icon-17":           icon{filename: "wi-day-hail", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                       // hail
		"icon-18":           icon{filename: "wi-sleet", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                          // sleet
		"icon-19":           icon{filename: "wi-dust", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                           // dust
		"icon-20":           icon{filename: "wi-fog", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                            // foggy
		"icon-21":           icon{filename: "wi-haze", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                           // haze
		"icon-22":           icon{filename: "wi-smoke", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                          // smoky
		"icon-23":           icon{filename: "wi-windy", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                          // blustery
		"icon-24":           icon{filename: "wi-windy", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                          // windy
		"icon-25":           icon{filename: "wi-cloudy", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                         // cold
		"icon-26":           icon{filename: "wi-cloudy", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                         // cloudy
		"icon-27":           icon{filename: "wi-night-cloudy", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                   // mostly cloudy (night)
		"icon-28":           icon{filename: "wi-cloudy", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                         // mostly cloudy (day)
		"icon-29":           icon{filename: "wi-night-partly-cloudy", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},            // partly cloudy (night)
		"icon-30":           icon{filename: "wi-day-cloudy", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                     // partly cloudy (day)
		"icon-31":           icon{filename: "wi-night-clear", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                    // clear (night)
		"icon-32":           icon{filename: "wi-day-sunny", color: "yellow", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true, popcolor: "yellow"}, // sunny
		"icon-33":           icon{filename: "wi-night-clear", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                    // fair (night)
		"icon-34":           icon{filename: "wi-day-sunny", color: "yellow", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                     // fair (day)
		"icon-35":           icon{filename: "wi-hail", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                           // mixed rain and hail
		"icon-36":           icon{filename: "wi-hot", color: "yellow", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                           // hot
		"icon-37":           icon{filename: "wi-thunderstorm", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                   // isolated thunderstorms
		"icon-38":           icon{filename: "wi-storm-showers", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                  // scattered thunderstorms
		"icon-39":           icon{filename: "wi-rain", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                           // scattered rain
		"icon-40":           icon{filename: "wi-rain", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                           // heavy rain
		"icon-41":           icon{filename: "wi-snowflake-cold", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                 // heavy snow
		"icon-42":           icon{filename: "wi-snow", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                           // scattered snow showers
		"icon-43":           icon{filename: "wi-snow-wind", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                      // blowing heavy snow
		"icon-44":           icon{filename: "wi-day-cloudy", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                     // partly cloudy (day)
		"icon-45":           icon{filename: "wi-night-thunderstorm", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},             // thundershowers (night)
		"icon-46":           icon{filename: "wi-night-snow", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},                     // snow showers (night)
		"icon-47":           icon{filename: "wi-night-thunderstorm", color: "linen", width: miW, height: miW, alpha: miAlpha, scale: miScale, shadow: true},             // isolated thundershowers (night)
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
		"wind-N":            icon{filename: windDegIcon, color: wiColor, width: wiW, height: wiW, rotate: 180.0, scale: wiScale, alpha: wiAlpha, shadow: true}, // N
		"wind-NNE":          icon{filename: windDegIcon, color: wiColor, width: wiW, height: wiW, rotate: 202.5, scale: wiScale, alpha: wiAlpha, shadow: true}, // NNE
		"wind-NE":           icon{filename: windDegIcon, color: wiColor, width: wiW, height: wiW, rotate: 225.0, scale: wiScale, alpha: wiAlpha, shadow: true}, // NE
		"wind-ENE":          icon{filename: windDegIcon, color: wiColor, width: wiW, height: wiW, rotate: 247.5, scale: wiScale, alpha: wiAlpha, shadow: true}, // ENE
		"wind-E":            icon{filename: windDegIcon, color: wiColor, width: wiW, height: wiW, rotate: 270.0, scale: wiScale, alpha: wiAlpha, shadow: true}, // E
		"wind-ESE":          icon{filename: windDegIcon, color: wiColor, width: wiW, height: wiW, rotate: 292.5, scale: wiScale, alpha: wiAlpha, shadow: true}, // ESE
		"wind-SE":           icon{filename: windDegIcon, color: wiColor, width: wiW, height: wiW, rotate: 315.0, scale: wiScale, alpha: wiAlpha, shadow: true}, // SE
		"wind-SSE":          icon{filename: windDegIcon, color: wiColor, width: wiW, height: wiW, rotate: 337.5, scale: wiScale, alpha: wiAlpha, shadow: true}, // SSE
		"wind-S":            icon{filename: windDegIcon, color: wiColor, width: wiW, height: wiW, rotate: 0.0, scale: wiScale, alpha: wiAlpha, shadow: true},   // S
		"wind-SSW":          icon{filename: windDegIcon, color: wiColor, width: wiW, height: wiW, rotate: 22.5, scale: wiScale, alpha: wiAlpha, shadow: true},  // SSW
		"wind-SW":           icon{filename: windDegIcon, color: wiColor, width: wiW, height: wiW, rotate: 45.0, scale: wiScale, alpha: wiAlpha, shadow: true},  // SW
		"wind-WSW":          icon{filename: windDegIcon, color: wiColor, width: wiW, height: wiW, rotate: 67.5, scale: wiScale, alpha: wiAlpha, shadow: true},  // WSW
		"wind-W":            icon{filename: windDegIcon, color: wiColor, width: wiW, height: wiW, rotate: 90.0, scale: wiScale, alpha: wiAlpha, shadow: true},  // W
		"wind-WNW":          icon{filename: windDegIcon, color: wiColor, width: wiW, height: wiW, rotate: 112.5, scale: wiScale, alpha: wiAlpha, shadow: true}, // WNW
		"wind-NW":           icon{filename: windDegIcon, color: wiColor, width: wiW, height: wiW, rotate: 135.0, scale: wiScale, alpha: wiAlpha, shadow: true}, // NW
		"wind-NNW":          icon{filename: windDegIcon, color: wiColor, width: wiW, height: wiW, rotate: 157.5, scale: wiScale, alpha: wiAlpha, shadow: true}, // NNW
		"moon-0":            icon{filename: "wi-moon-alt-new", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"moon-1":            icon{filename: "wi-moon-alt-waxing-crescent-5", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"moon-2":            icon{filename: "wi-moon-alt-first-quarter", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"moon-3":            icon{filename: "wi-moon-alt-waxing-gibbous-5", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"moon-4":            icon{filename: "wi-moon-alt-full", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"moon-5":            icon{filename: "wi-moon-alt-waning-gibbous-5", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"moon-6":            icon{filename: "wi-moon-alt-third-quarter", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"moon-7":            icon{filename: "wi-moon-alt-waning-crescent-5", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"moon-8":            icon{filename: "wi-moon-alt-new", color: "bisque", width: 24, height: 24, scale: 0.60, alpha: 1, shadow: true},
		"brolly":            icon{filename: "wi-umbrella", color: "#66ff99", width: wiW, height: wiW, scale: (wiScale * 1.2), alpha: 1, shadow: true},
		"humidity":          icon{filename: "wi-humidity", color: "#66ff99", width: wiW, height: wiW, scale: (wiScale * 1.2), alpha: 1, shadow: true},
		"snowflake":         icon{filename: "wi-snowflake-cold", color: "#66ff99", width: wiW, height: wiW, scale: (wiScale * 1.2), alpha: 1, shadow: true},
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
		"skullz":            icon{filename: "arggggh", color: "#6D97ABcc", width: 100, height: 100, scale: 1.0, alpha: 0.5, shadow: true},
		"bunny":             icon{filename: "first-bunny", color: "#0f344340", width: 30, height: 30, scale: 1.0, alpha: 1, shadow: true},
	}
	iconMap = im
}

func getIcon(s string) icon {
	var v icon
	v, ok := iconMap[s]
	if !ok || !fileExists(iconFile(v)) {
		fmt.Println("icon missing?", s, v)
		v = icon{filename: "wi-alien", color: "red", width: miW, height: miW, scale: 1, alpha: 1, shadow: false}
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
