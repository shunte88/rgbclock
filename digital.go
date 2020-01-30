package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"time"

	svg "github.com/ajstarks/svgo"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func imageTimeThin(t time.Time, scaleh float64, txtColor string) (img [2]draw.Image, err error) {

	// dynamic SVG

	type segment struct {
		pathd string
		x     float64
	}

	wt, ht := 356, 100
	xw := float64(ht) / scaleh
	sw, sh := float64(wt)/xw, float64(ht)/xw

	var (
		seg7bits = [10][7]int{
			{1, 1, 1, 1, 1, 1, 0}, // 0
			{0, 1, 1, 0, 0, 0, 0}, // 1
			{1, 1, 0, 1, 1, 0, 1}, // 2
			{1, 1, 1, 1, 0, 0, 1}, // 3
			{0, 1, 1, 0, 0, 1, 1}, // 4
			{1, 0, 1, 1, 0, 1, 1}, // 5
			{1, 0, 1, 1, 1, 1, 1}, // 6
			{1, 1, 1, 0, 0, 0, 0}, // 7
			{1, 1, 1, 1, 1, 1, 1}, // 8
			{1, 1, 1, 1, 0, 1, 1}} // 9

		segPos = [5]float64{0.00, 80.00, (3.00 + (float64(wt) / 2.00)), 210.00, 290.00}

		seg7template = [7]segment{
			{pathd: `m%f,8l4,-4l28,0l4,4l-4,4l-28,0l-4,-4z`, x: 20},  // a
			{pathd: `m%f,10l4,4l0,28l-4,4l-4,-4l0,-28l4,-4z`, x: 58}, // b
			{pathd: `m%f,50l4,4l0,28l-4,4l-4,-4l0,-28l4,-4z`, x: 58}, // c
			{pathd: `m%f,88l4,-4l28,0l4,4l-4,4l-28,0l-4,-4z`, x: 20}, // d
			{pathd: `m%f,50l4,4l0,28l-4,4l-4,-4l0,-28l4,-4z`, x: 18}, // e
			{pathd: `m%f,10l4,4l0,28l-4,4l-4,-4l0,-28l4,-4z`, x: 18}, // f
			{pathd: `m%f,48l4,-4l28,0l4,4l-4,4l-28,0l-4,-4z`, x: 20}} // g
		/*
			seg7template2 = [7]segment{
				{pathd: `m%f,2s-3.769851,0.684834 -5.444656,1.555387c-1.552799,0.807825 -4.082876,3.305737 -4.082876,3.305737l12.834008,6.028188l17.889232,-0.194716l6.805204,-9.916841l-28.000911,-0.777755l0,0l-0.000001,0z`, x: 20.602663},
				{pathd: `m%f,2.777755zm2.722328,1.361041l-7.194636,10.695212l-1.167064,27.028564l9.916964,6.610488l2.528844,-2.527612l1.74998,-34.223199s-0.898405,-2.989754 -2.172686,-4.493258c-0.838019,-0.989601 -3.661402,-3.090194 -3.661402,-3.090194l0,0l0,-0.000001z`, x: 48.603574},
				{pathd: `m%f,53.528283l-12.056376,4.278824l-0.972348,27.806195l6.417004,10.305163s3.890624,-2.257721 5.056456,-4.082876c1.149811,-1.802972 1.555264,-6.222288 1.555264,-6.222288l1.360548,-30.140324l-1.360548,-1.944696l0,0.000002z`, x: 54.24233},
				{pathd: `m%f,3.555387zm23.431862,83.859655l-18.311938,0.061619l-12.517286,6.04236c0.000246,0 2.180451,1.883077 3.450665,2.531309c1.192944,0.608796 3.825308,1.229916 3.825308,1.229916l29.965325,0.122006l-6.412074,-9.98721l0,0z`, x: 15.158007},
				{pathd: `m%f,3.555387zm-5.243162,49.48857l-3.231424,2.26758l-0.687422,35.474065l12.443343,-5.911728l1.099283,-27.155499l-9.62378,-4.674418z`, x: 15.158007},
				{pathd: `m%f,41.862571zm-34.028483,-32.083787l-2.333142,35.778463l3.526949,2.165292l9.306812,-5.665252l1.74998,-26.640364l-12.250599,-5.63814l0,0.000001z`, x: 42.964202},
				{pathd: `m%f,44.005683l-10.30578,6.222288l8.5558,4.278085l19.444495,0.972299l11.422316,-4.133896l-10.835333,-7.234935l-18.281498,-0.10384l0,-0.000001z`, x: 22.300884}}
		*/
		//seg7cols  = [7]string{`red`, `orange`, `yellow`, `green`, `blue`, `indigo`, `violet`}
		seg7style = [2]string{
			`style="fill:gray;fill-opacity:0.25;stroke-width:0.5;stroke:%s;stroke-opacity:0.25;stroke-alignment:inside;"`, // off
			`style="fill:%[1]s;fill-opacity:1;stroke-width:2;stroke:%[1]s;stroke-opacity:0.6;stroke-alignment:outside;"`}  // on
	)

	seg7style[0] = fmt.Sprintf(seg7style[0], txtColor)
	seg7style[1] = fmt.Sprintf(seg7style[1], txtColor)

	img[0] = image.NewRGBA(image.Rect(0, 0, int(sw), int(sh)))
	img[1] = image.NewRGBA(image.Rect(0, 0, int(sw), int(sh)))

	var iconMem1 = new(bytes.Buffer)
	var iconMem0 = new(bytes.Buffer)

	var canvas1 = svg.New(iconMem1)
	var canvas0 = svg.New(iconMem0)

	canvas1.Start(wt, ht)
	canvas0.Start(wt, ht)

	canvas1.Group("id=\"time\" transform=\"skewX(-8)\"")
	canvas0.Group("id=\"time\" transform=\"skewX(-8)\"")
	//canvas1.Group("id=\"time\" transform=\"skewX(0)\"")
	//canvas0.Group("id=\"time\" transform=\"skewX(0)\"")
	for i, c := range t.Format(`15:04`) {
		if c != ':' {
			s7 := int(c) - int('0')
			canvas1.Group(fmt.Sprintf("id=\"tempo%d_%d\"", i, s7))
			canvas0.Group(fmt.Sprintf("id=\"tempo%d_%d\"", i, s7))
			for t, b := range seg7bits[s7] {
				canvas1.Path(fmt.Sprintf(seg7template[t].pathd, seg7template[t].x+segPos[i]),
					seg7style[b])
				canvas0.Path(fmt.Sprintf(seg7template[t].pathd, seg7template[t].x+segPos[i]),
					seg7style[b])
			}
			canvas1.Gend()
			canvas0.Gend()
		} else {
			canvas1.Circle(int(segPos[i]), 30, 8, seg7style[1])
			canvas1.Circle(int(segPos[i]), 60, 8, seg7style[1])
			canvas0.Circle(int(segPos[i]), 30, 8, seg7style[0])
			canvas0.Circle(int(segPos[i]), 60, 8, seg7style[0])
		}
	}
	canvas1.Gend()
	canvas0.Gend()
	canvas1.End()
	canvas0.End()

	//fmt.Println(iconMem1.String())

	iconI0, err := oksvg.ReadIconStream(iconMem0)
	if err != nil {
		return img, err
	}
	iconI1, err := oksvg.ReadIconStream(iconMem1)
	if err != nil {
		return img, err
	}

	gv0 := rasterx.NewScannerGV(wt, ht, img[0], img[0].Bounds())
	r0 := rasterx.NewDasher(wt, ht, gv0)
	iconI0.SetTarget(0, 0, float64(sw), float64(sh))
	iconI0.Draw(r0, 1.0)

	gv1 := rasterx.NewScannerGV(wt, ht, img[1], img[1].Bounds())
	r1 := rasterx.NewDasher(wt, ht, gv1)
	iconI1.SetTarget(0, 0, float64(sw), float64(sh))
	iconI1.Draw(r1, 1.0)

	return img, nil

}

func imageTime(t time.Time, scaleh float64, txtColor string) (img [2]draw.Image, err error) {

	// dynamic SVG

	type segment struct {
		pathd string
		x     float64
	}

	wt, ht := 420, 120
	xw := float64(ht) / scaleh
	sw, sh := float64(wt)/xw, float64(ht)/xw

	var (
		seg7bits = [10][7]int{
			{1, 1, 1, 1, 1, 1, 0}, // 0
			{0, 1, 1, 0, 0, 0, 0}, // 1
			{1, 1, 0, 1, 1, 0, 1}, // 2
			{1, 1, 1, 1, 0, 0, 1}, // 3
			{0, 1, 1, 0, 0, 1, 1}, // 4
			{1, 0, 1, 1, 0, 1, 1}, // 5
			{1, 0, 1, 1, 1, 1, 1}, // 6
			{1, 1, 1, 0, 0, 0, 0}, // 7
			{1, 1, 1, 1, 1, 1, 1}, // 8
			{1, 1, 1, 1, 0, 1, 1}} // 9

		segPos = [5]float64{0.00, 10 + 80.00, (3.00 + (float64(wt) / 2.00)), 30 + 210.00, 40 + 290.00}

		seg7template = [7]segment{
			{pathd: `m%f,9.5l8,-8l25,0l8,8l-8,8l-25,0l-8,-8z`, x: 30},   // a
			{pathd: `m%f,12.5l8,8l0,25l-8,8l-8,-8l0,-25l8,-8z`, x: 73},  // b
			{pathd: `m%f,59.5l8,8l0,25l-8,8l-8,-8l0,-25l8,-8z`, x: 73},  // c
			{pathd: `m%f,102.5l8,-8l25,0l8,8l-8,8l-25,0l-8,-8z`, x: 30}, // d
			{pathd: `m%f,100.5l-8,-8l0,-25l8,-8l8,8l0,25l-8,8z`, x: 28}, // e
			{pathd: `m%f,53.5l-8,-8l0,-25l8,-8l8,8l0,25l-8,8z`, x: 28},  // f
			{pathd: `m%f,56.5l8,-8l25,0l8,8l-8,8l-25,0l-8,-8z`, x: 30}}  // g

		seg7style = [2]string{
			`style="fill:gray;fill-opacity:0.25;stroke-width:0.5;stroke:%s;stroke-opacity:0.25;stroke-alignment:inside;"`, // off
			`style="fill:%[1]s;fill-opacity:1;stroke-width:2;stroke:%[1]s;stroke-opacity:0.4;stroke-alignment:outside;"`}  // on
	)

	seg7style[0] = fmt.Sprintf(seg7style[0], txtColor)
	seg7style[1] = fmt.Sprintf(seg7style[1], txtColor)

	img[0] = image.NewRGBA(image.Rect(0, 0, int(sw), int(sh)))
	img[1] = image.NewRGBA(image.Rect(0, 0, int(sw), int(sh)))

	var iconMem1 = new(bytes.Buffer)
	var iconMem0 = new(bytes.Buffer)

	var canvas1 = svg.New(iconMem1)
	var canvas0 = svg.New(iconMem0)

	canvas1.Start(wt, ht)
	canvas0.Start(wt, ht)

	iconMem1.WriteString(`<defs>
  <filter id="blur" x="0" y="0" width="200%" height="200%">
    <feOffset result="offOut" in="SourceAlpha" dx="5" dy="5" />
    <feGaussianBlur result="blurOut" in="offOut" stdDeviation="0.5" />
    <feBlend in="SourceGraphic" in2="blurOut" mode="normal" />
  </filter>
</defs>`)

	canvas1.Group("id=\"time\" filter=\"url(#blur)\" transform=\"skewX(-12)\"")
	canvas0.Group("id=\"time\" filter=\"url(#blur)\" transform=\"skewX(-12)\"")

	for i, c := range t.Format(`15:04`) {
		if c != ':' {
			s7 := int(c) - int('0')
			canvas1.Group(fmt.Sprintf("id=\"tempo%d_%d\"", i, s7))
			canvas0.Group(fmt.Sprintf("id=\"tempo%d_%d\"", i, s7))
			for t, b := range seg7bits[s7] {
				canvas1.Path(fmt.Sprintf(seg7template[t].pathd, seg7template[t].x+segPos[i]),
					seg7style[b])
				canvas0.Path(fmt.Sprintf(seg7template[t].pathd, seg7template[t].x+segPos[i]),
					seg7style[b])
			}
			canvas1.Gend()
			canvas0.Gend()
		} else {
			canvas1.Circle(int(segPos[i]), 40, 10, seg7style[1])
			canvas1.Circle(int(segPos[i]), 80, 10, seg7style[1])
			canvas0.Circle(int(segPos[i]), 40, 10, seg7style[0])
			canvas0.Circle(int(segPos[i]), 80, 10, seg7style[0])
		}
	}
	canvas1.Gend()
	canvas0.Gend()
	canvas1.End()
	canvas0.End()

	//fmt.Println(iconMem1.String())

	iconI0, err := oksvg.ReadIconStream(iconMem0)
	if err != nil {
		return img, err
	}
	iconI1, err := oksvg.ReadIconStream(iconMem1)
	if err != nil {
		return img, err
	}

	gv0 := rasterx.NewScannerGV(wt, ht, img[0], img[0].Bounds())
	r0 := rasterx.NewDasher(wt, ht, gv0)
	iconI0.SetTarget(0, 0, float64(sw), float64(sh))
	iconI0.Draw(r0, 1.0)

	gv1 := rasterx.NewScannerGV(wt, ht, img[1], img[1].Bounds())
	r1 := rasterx.NewDasher(wt, ht, gv1)
	iconI1.SetTarget(0, 0, float64(sw), float64(sh))
	iconI1.Draw(r1, 1.0)

	return img, nil

}

func imageTimeDotty(t time.Time, scaleh float64, txtColor string) (img [2]draw.Image, err error) {

	// dynamic SVG

	type segment struct {
		pathd string
		x     float64
	}

	wt, ht := (90 * 5), 130
	xw := float64(ht) / scaleh
	sw, sh := float64(wt)/xw, float64(ht)/xw

	var (
		seg7bits = [10][7]int{
			{1, 1, 1, 1, 1, 1, 0}, // 0
			{0, 1, 1, 0, 0, 0, 0}, // 1
			{1, 1, 0, 1, 1, 0, 1}, // 2
			{1, 1, 1, 1, 0, 0, 1}, // 3
			{0, 1, 1, 0, 0, 1, 1}, // 4
			{1, 0, 1, 1, 0, 1, 1}, // 5
			{1, 0, 1, 1, 1, 1, 1}, // 6
			{1, 1, 1, 0, 0, 0, 0}, // 7
			{1, 1, 1, 1, 1, 1, 1}, // 8
			{1, 1, 1, 1, 0, 1, 1}} // 9

		segPos = [5]float64{0.00, 20 + 80.00, (3.00 + (float64(wt) / 2.00)), 50 + 210.00, 70 + 290.00}

		seg7template = [7]segment{
			{pathd: `m%f,6.0625c0,-2.762431 2.237569,-5 5,-5c2.762431,0 5,2.237569 5,5c0,2.762431 -2.237569,5 -5,5c-2.762431,0 -5,-2.237569 -5,-5zm11,0c0,-2.762431 2.237569,-5 5,-5c2.762431,0 5,2.237569 5,5c0,2.762431 -2.237569,5 -5,5c-2.762431,0 -5,-2.237569 -5,-5zm11,0c0,-2.762431 2.237569,-5 5,-5c2.762431,0 5,2.237569 5,5c0,2.762431 -2.237569,5 -5,5c-2.762431,0 -5,-2.237569 -5,-5zm11,0c0,-2.762431 2.237569,-5 5,-5c2.762431,0 5,2.237569 5,5c0,2.762431 -2.237569,5 -5,5c-2.762431,0 -5,-2.237569 -5,-5z`, x: 5 + 18.5},     // a
			{pathd: `m%f,17.062496c0,-2.76243 2.23757,-5 5,-5c2.76243,0 5,2.23757 5,5c0,2.76243 -2.23757,5 -5,5c-2.76243,0 -5,-2.23757 -5,-5zm0,11c0,-2.76243 2.23757,-5 5,-5c2.76243,0 5,2.23757 5,5c0,2.76243 -2.23757,5 -5,5c-2.76243,0 -5,-2.23757 -5,-5zm0,11c0,-2.76243 2.23757,-5 5,-5c2.76243,0 5,2.23757 5,5c0,2.76243 -2.23757,5 -5,5c-2.76243,0 -5,-2.23757 -5,-5zm0,11c0,-2.76243 2.23757,-5 5,-5c2.76243,0 5,2.23757 5,5c0,2.76243 -2.23757,5 -5,5c-2.76243,0 -5,-2.23757 -5,-5z`, x: 5 + 61.75},                                 // b
			{pathd: `m%f,70.89521c0,-2.76243 2.23757,-5 5,-5c2.76243,0 5,2.23757 5,5c0,2.76243 -2.23757,5 -5,5c-2.76243,0 -5,-2.23757 -5,-5zm0,11c0,-2.76243 2.23757,-5 5,-5c2.76243,0 5,2.23757 5,5c0,2.76243 -2.23757,5 -5,5c-2.76243,0 -5,-2.23757 -5,-5zm0,11c0,-2.76243 2.23757,-5 5,-5c2.76243,0 5,2.23757 5,5c0,2.76243 -2.23757,5 -5,5c-2.76243,0 -5,-2.23757 -5,-5zm0,11c0,-2.76243 2.23757,-5 5,-5c2.76243,0 5,2.23757 5,5c0,2.76243 -2.23757,5 -5,5c-2.76243,0 -5,-2.23757 -5,-5z`, x: 5 + 61.75},                                  // c
			{pathd: `m%f,113.937496c0,-2.762431 2.237569,-5 5,-5c2.762431,0 5,2.237569 5,5c0,2.762431 -2.237569,5 -5,5c-2.762431,0 -5,-2.237569 -5,-5zm11,0c0,-2.762431 2.237569,-5 5,-5c2.762431,0 5,2.237569 5,5c0,2.762431 -2.237569,5 -5,5c-2.762431,0 -5,-2.237569 -5,-5zm11,0c0,-2.762431 2.237569,-5 5,-5c2.762431,0 5,2.237569 5,5c0,2.762431 -2.237569,5 -5,5c-2.762431,0 -5,-2.237569 -5,-5zm11,0c0,-2.762431 2.237569,-5 5,-5c2.762431,0 5,2.237569 5,5c0,2.762431 -2.237569,5 -5,5c-2.762431,0 -5,-2.237569 -5,-5z`, x: 5 + 18.5}, // d
			{pathd: `m%f,70.89521c0,-2.76243 2.23757,-5 5,-5c2.76243,0 5,2.23757 5,5c0,2.76243 -2.23757,5 -5,5c-2.76243,0 -5,-2.23757 -5,-5zm0,11c0,-2.76243 2.23757,-5 5,-5c2.76243,0 5,2.23757 5,5c0,2.76243 -2.23757,5 -5,5c-2.76243,0 -5,-2.23757 -5,-5zm0,11c0,-2.76243 2.23757,-5 5,-5c2.76243,0 5,2.23757 5,5c0,2.76243 -2.23757,5 -5,5c-2.76243,0 -5,-2.23757 -5,-5zm0,11c0,-2.76243 2.23757,-5 5,-5c2.76243,0 5,2.23757 5,5c0,2.76243 -2.23757,5 -5,5c-2.76243,0 -5,-2.23757 -5,-5z`, x: 5 + 8.25},                                   // e
			{pathd: `m%f,17.062496c0,-2.76243 2.23757,-5 5,-5c2.76243,0 5,2.23757 5,5c0,2.76243 -2.23757,5 -5,5c-2.76243,0 -5,-2.23757 -5,-5zm0,11c0,-2.76243 2.23757,-5 5,-5c2.76243,0 5,2.23757 5,5c0,2.76243 -2.23757,5 -5,5c-2.76243,0 -5,-2.23757 -5,-5zm0,11c0,-2.76243 2.23757,-5 5,-5c2.76243,0 5,2.23757 5,5c0,2.76243 -2.23757,5 -5,5c-2.76243,0 -5,-2.23757 -5,-5zm0,11c0,-2.76243 2.23757,-5 5,-5c2.76243,0 5,2.23757 5,5c0,2.76243 -2.23757,5 -5,5c-2.76243,0 -5,-2.23757 -5,-5z`, x: 5 + 8.25},                                  // f
			{pathd: `m%f,59.9375c0,-2.762431 2.237569,-5 5,-5c2.762431,0 5,2.237569 5,5c0,2.762431 -2.237569,5 -5,5c-2.762431,0 -5,-2.237569 -5,-5zm11,0c0,-2.762431 2.237569,-5 5,-5c2.762431,0 5,2.237569 5,5c0,2.762431 -2.237569,5 -5,5c-2.762431,0 -5,-2.237569 -5,-5zm11,0c0,-2.762431 2.237569,-5 5,-5c2.762431,0 5,2.237569 5,5c0,2.762431 -2.237569,5 -5,5c-2.762431,0 -5,-2.237569 -5,-5zm11,0c0,-2.762431 2.237569,-5 5,-5c2.762431,0 5,2.237569 5,5c0,2.762431 -2.237569,5 -5,5c-2.762431,0 -5,-2.237569 -5,-5z`, x: 5 + 18.5}}    // g

		seg7style = [2]string{
			`style="fill:gray;fill-opacity:0.25;stroke-width:0.5;stroke:%s;stroke-opacity:0.25;stroke-alignment:inside;"`, // off
			`style="fill:%[1]s;fill-opacity:1;stroke-width:2;stroke:%[1]s;stroke-opacity:0.6;stroke-alignment:outside;"`}  // on
	)

	seg7style[0] = fmt.Sprintf(seg7style[0], txtColor)
	seg7style[1] = fmt.Sprintf(seg7style[1], txtColor)

	img[0] = image.NewRGBA(image.Rect(0, 0, int(sw), int(sh)))
	img[1] = image.NewRGBA(image.Rect(0, 0, int(sw), int(sh)))

	var iconMem1 = new(bytes.Buffer)
	var iconMem0 = new(bytes.Buffer)

	var canvas1 = svg.New(iconMem1)
	var canvas0 = svg.New(iconMem0)

	canvas1.Start(wt, ht)
	canvas0.Start(wt, ht)

	canvas1.Group("id=\"time\" transform=\"skewX(-8)\"")
	canvas0.Group("id=\"time\" transform=\"skewX(-8)\"")
	for i, c := range t.Format(`15:04`) {
		if c != ':' {
			s7 := int(c) - int('0')
			canvas1.Group(fmt.Sprintf("id=\"tempo%d_%d\"", i, s7))
			canvas0.Group(fmt.Sprintf("id=\"tempo%d_%d\"", i, s7))
			for t, b := range seg7bits[s7] {
				canvas1.Path(fmt.Sprintf(seg7template[t].pathd, seg7template[t].x+segPos[i]),
					seg7style[b])
				canvas0.Path(fmt.Sprintf(seg7template[t].pathd, seg7template[t].x+segPos[i]),
					seg7style[b])
			}
			canvas1.Gend()
			canvas0.Gend()
		} else {
			canvas1.Circle(int(segPos[i]), 40, 8, seg7style[1])
			canvas1.Circle(int(segPos[i]), 80, 8, seg7style[1])
			canvas0.Circle(int(segPos[i]), 40, 8, seg7style[0])
			canvas0.Circle(int(segPos[i]), 80, 8, seg7style[0])
		}
	}
	canvas1.Gend()
	canvas0.Gend()
	canvas1.End()
	canvas0.End()

	//fmt.Println(iconMem1.String())

	iconI0, err := oksvg.ReadIconStream(iconMem0)
	if err != nil {
		return img, err
	}
	iconI1, err := oksvg.ReadIconStream(iconMem1)
	if err != nil {
		return img, err
	}

	gv0 := rasterx.NewScannerGV(wt, ht, img[0], img[0].Bounds())
	r0 := rasterx.NewDasher(wt, ht, gv0)
	iconI0.SetTarget(0, 0, float64(sw), float64(sh))
	iconI0.Draw(r0, 1.0)

	gv1 := rasterx.NewScannerGV(wt, ht, img[1], img[1].Bounds())
	r1 := rasterx.NewDasher(wt, ht, gv1)
	iconI1.SetTarget(0, 0, float64(sw), float64(sh))
	iconI1.Draw(r1, 1.0)

	return img, nil

}
