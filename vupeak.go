package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"

	"github.com/fogleman/gg"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func (ls *LMSServer) vuPeak(dBfs [2]int32) {

	w, h := 162, 40 // ls.vulayout.w2m, ls.vulayout.h2m
	var iconMem = new(bytes.Buffer)

	level := [12]int32{-50, -40, -30, -24, -18, -12, -9, -6, -3, 0, 2, 4}
	ypos := [2]float64{9.28, 25.21}

	bs := `<svg width="162" height="40" xmlns="http://www.w3.org/2000/svg">

	     <defs>
	       <g id="led">
		   <rect x="0" y="0" width="8" height="5.5" />
		   <rect x=".5" y=".5" width="1.25" height="4.5" fill="black" fill-opacity=".15" stroke-width=".2" stroke-opacity=".5" stroke-position="inside" stroke="darkgray"/>
		   <rect x="2.33" y=".5" width="1.25" height="4.5" fill="black" fill-opacity=".15" stroke-width=".2" stroke-opacity=".5" stroke-position="inside" stroke="darkgray"/>
		   <rect x="4.13" y=".5" width="1.25" height="4.5" fill="black" fill-opacity=".15" stroke-width=".2" stroke-opacity=".5" stroke-position="inside" stroke="darkgray"/>
		   <rect x="6.0" y=".5" width="1.25" height="4.5" fill="black" fill-opacity=".15" stroke-width=".2" stroke-opacity=".5" stroke-position="inside" stroke="darkgray"/>
			  </g>
	       <path id="over" fill-opacity="0.4" d="m136.51396,17.79571c0.720987,0 1.231205,0.184043 1.530655,0.552737c0.29945,0.368694 0.362619,0.90503 0.188295,1.609617c-0.177969,0.704587 -0.508396,1.240924 -0.990067,1.609617c-0.485922,0.368694 -1.089073,0.552737 -1.81006,0.552737c-0.720987,0 -1.231205,-0.184043 -1.530655,-0.552737c-0.29945,-0.364442 -0.36019,-0.900778 -0.182221,-1.609617c0.174325,-0.708839 0.502322,-1.245176 0.983993,-1.609617c0.485922,-0.368694 1.089073,-0.552737 1.81006,-0.552737zm-0.255109,1.002215c-0.287302,0 -0.526618,0.086859 -0.716735,0.261183c-0.194369,0.174325 -0.327998,0.408782 -0.400886,0.704587l-0.097184,0.388738c-0.072888,0.295805 -0.056488,0.530263 0.048592,0.704587c0.105081,0.174325 0.301879,0.261183 0.589181,0.261183c0.287302,0 0.52844,-0.086859 0.722809,-0.261183c0.198621,-0.174325 0.334072,-0.408782 0.40696,-0.704587l0.097184,-0.388738c0.072888,-0.295805 0.054666,-0.530263 -0.054666,-0.704587c-0.109333,-0.174325 -0.307953,-0.261183 -0.595255,-0.261183zm4.604113,3.249605l-1.554951,0l-0.43733,-4.178931l1.433471,0l0.139703,2.794053l0.024296,0l1.542803,-2.794053l1.37273,0l-2.520722,4.178931zm2.034192,0l1.044733,-4.178931l3.614047,0l-0.255109,1.002215l-2.271686,0l-0.139703,0.577033l1.943689,0l-0.242961,0.959696l-1.943689,0l-0.157925,0.637773l2.314205,0l-0.249035,1.002215l-3.656565,0l-0.000001,-0.000001zm9.074598,-2.897311c-0.064992,0.255109 -0.190117,0.485922 -0.37659,0.692439c-0.190117,0.206517 -0.429434,0.358368 -0.716735,0.455552l0.491996,1.74932l-1.506359,0l-0.358368,-1.524581l-0.49807,0l-0.382664,1.524581l-1.34236,0l1.044733,-4.178931l2.557166,0c0.29945,0 0.540588,0.058918 0.722809,0.176147c0.182221,0.113584 0.303701,0.267257 0.364442,0.461626c0.056488,0.198621 0.056488,0.413034 0,0.643847zm-1.378804,0.054666c0.024296,-0.109333 0.010326,-0.200443 -0.042518,-0.273331c-0.056488,-0.072888 -0.13788,-0.109333 -0.242961,-0.109333l-0.880734,0l-0.188295,0.771402l0.880734,0c0.105081,0 0.202265,-0.036444 0.291553,-0.109333c0.092933,-0.07714 0.153673,-0.170073 0.182221,-0.279405z"/>
	   </defs>

	`
	tick := `<use xlink:href="#led" x="%f" y="%f" fill="%s" fill-opacity="%f" stroke-width="0.1" stroke="darkslategrey"/>`
	for p, l := range level {
		xpos := float64(24.00 + ((float64(p) + 1.00) * 9.791538))
		for channel := 0; channel < 2; channel++ {
			color := `green`
			opacity := .9
			testd := (20.00 + dBfs[channel])
			if l >= testd {
				color = `lightgrey`
				opacity = .4
			} else {
				if l >= 0 && testd > 0 {
					if l >= 4 {
						color = `red`
					} else {
						color = `yellow`
					}
				}
			}
			bs += fmt.Sprintf("\n"+tick, xpos, ypos[channel], color, opacity)
		}
	}

	// Overload
	style := `style="fill:lightgrey;fill-opacity:0.5"`
	if dBfs[0] > 0 || dBfs[1] > 0 {
		style = `style="fill:red;fill-opacity:1"`
	}
	bs += fmt.Sprintf("\n<use xlink:href=\"#over\" %s/>", style)

	bs += " </svg>"

	iconMem.WriteString(bs)

	iconI, err := oksvg.ReadIconStream(iconMem)
	if err != nil {
		fmt.Println(iconMem.String())
		panic(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, ls.vulayout.w2m, ls.vulayout.h2m))
	gv := rasterx.NewScannerGV(w, h, img, img.Bounds())
	r := rasterx.NewDasher(w, h, gv)
	iconI.SetTarget(0, 0, float64(ls.vulayout.w2m), float64(ls.vulayout.h2m))
	iconI.Draw(r, 1.0)

	dc := gg.NewContext(ls.vulayout.w2m, ls.vulayout.h2m)
	dc.DrawImageAnchored(ls.vulayout.baseImage, 0, 0, 0, 0)
	dc.DrawImageAnchored(img, 0, 0, 0, 0)

	ls.mux.Lock()
	draw.Draw(ls.vulayout.vu, ls.vulayout.vu.Bounds(), dc.Image(), image.ZP, draw.Over)
	ls.mux.Unlock()

}
