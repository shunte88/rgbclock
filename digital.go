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

// imageTimeHero
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
			{pathd: `m%f,7.87738c-5.73067,5.00191 -5.86317,5.16754 -6.32693,6.92318c-0.23188,0.99376 -0.33125,1.92127 -0.19875,2.05377c0.09938,0.1325 10.20259,0.29813 22.39268,0.33125l22.19393,0.1325l3.5444,-3.2794c1.92127,-1.78877 4.73692,-4.40566 6.19443,-5.79692l2.65002,-2.55064l-1.32501,-1.32501l-1.32501,-1.32501l-20.9683,-0.16563l-20.93517,-0.16563l-5.8963,5.16754l0.00001,0.00001z`, x: 27.2902},
			{pathd: `m%f,15.42994c-4.10754,3.84253 -6.19443,6.06193 -6.36005,6.6913c-0.46376,1.75564 -4.24004,24.77771 -4.30629,26.33459c-0.09938,1.42439 0.03312,1.68939 2.21939,4.53816c1.29189,1.62314 2.45127,2.98127 2.6169,3.0144c0.29813,0 3.47815,-2.78253 7.55256,-6.6913c2.4844,-2.38502 2.51752,-2.45127 2.7494,-4.40566c0.09938,-1.09314 1.19251,-7.71819 2.35189,-14.74075c3.04752,-17.9539 2.91503,-16.86077 2.21939,-17.6889c-1.19251,-1.42439 -2.55064,-2.84877 -2.71627,-2.81565c-0.1325,0 -2.94815,2.58377 -6.32693,5.7638l0.00001,0.00001z`, x: 74.92435},
			{pathd: `m%f,67.96663c-2.18627,2.02064 -4.07441,3.90878 -4.20691,4.20691c-0.265,0.59626 -1.75564,9.20883 -2.25252,12.98511c-0.16563,1.19251 -0.76188,4.80316 -1.32501,8.08257c-0.56313,3.24628 -0.99376,6.1613 -0.92751,6.49255c0.09938,0.62938 7.81756,10.26884 8.21507,10.26884c0.1325,0 1.35813,-0.89438 2.71627,-1.95439l2.45127,-1.98752l1.95439,-11.5276c1.09314,-6.32693 1.95439,-12.1901 1.92127,-13.01824c-0.03312,-0.82813 0.23188,-3.31253 0.59626,-5.53192c0.33125,-2.25252 0.62938,-4.40566 0.62938,-4.80316c0,-0.86126 -4.77004,-6.89006 -5.39942,-6.89006c-0.23188,0.03312 -2.18627,1.65626 -4.37253,3.6769z`, x: 64.48989},
			{pathd: `m%f,103.54317c-0.53,1.55689 -0.69563,3.31253 -0.36438,3.80941c0.23188,0.29813 2.25252,2.68315 4.53816,5.2338l4.14066,4.70379l20.8358,0l20.80268,0l2.12002,-1.68939c1.15938,-0.92751 2.12002,-1.82189 2.12002,-1.95439c0,-0.29813 -2.18627,-3.08065 -5.99567,-7.68506l-2.12002,-2.55064l-22.98894,-0.09938c-15.63513,-0.09938 -22.98894,0 -23.08832,0.23188l-0.00001,-0.00001z`, x: 6.15627},
			{pathd: `m%f,66.57536c-0.89438,0.69563 -2.51752,2.12002 -3.64378,3.1469c-1.78877,1.65626 -2.08689,2.08689 -2.28565,3.44503c-0.09938,0.86126 -0.46376,3.04752 -0.79501,4.86941c-0.33125,1.82189 -0.99376,5.83005 -1.49064,8.94382c-0.49688,3.08065 -1.09314,6.72443 -1.35813,8.08257c-0.265,1.32501 -0.46376,2.55064 -0.46376,2.71627c0,0.36438 14.47574,0.3975 14.6745,0.03312c0.09938,-0.1325 0.86126,-4.40566 1.68939,-9.50695c1.95439,-11.6601 2.51752,-15.03888 2.71627,-15.50263c0.06625,-0.23188 -1.15938,-1.98752 -2.71627,-3.97503c-3.34565,-4.17378 -3.74316,-4.30629 -6.32693,-2.25252l0.00001,0z`, x: 17.71699},
			{pathd: `m%f,25.50002c-0.29813,1.59001 -0.46376,3.1469 -0.33125,3.44503c0.19875,0.53 -0.29813,4.14066 -1.52376,11.16322c-0.33125,1.92127 -0.86126,5.00191 -1.15938,6.89006l-0.56313,3.41191l1.32501,1.72251c2.28565,2.91503 2.6169,3.21315 3.97503,3.21315c1.06001,0 1.72251,-0.43063 5.43255,-3.57753c2.31877,-1.95439 4.27316,-3.77628 4.40566,-4.07441c0.19875,-0.56313 4.14066,-24.04895 4.14066,-24.74458c0,-0.29813 -1.72251,-0.3975 -7.58569,-0.3975l-7.58569,0l-0.53,2.94815l-0.00001,0z`, x: 19.14139},
			{pathd: `m%f,57.03529l-3.77628,3.6769l0.66251,1.19251c0.33125,0.66251 1.45751,2.21939 2.45127,3.44503l1.82189,2.28565l14.04512,-0.03312l14.01199,0l3.94191,-3.57753c2.15314,-1.95439 3.94191,-3.77628 3.94191,-4.04128c0,-0.29813 -1.12626,-1.88814 -2.4844,-3.57753l-2.51752,-3.04752l-14.17762,0l-14.14449,0l-3.77628,3.6769l-0.00001,-0.00001z`, x: 28.21771}}

		seg7style = [2]string{
			fmt.Sprintf("style=\"fill:gray;fill-opacity:0.3;stroke-width:0.5;stroke:%s;stroke-opacity:0.25;stroke-alignment:inside;\"", txtColor), // off
			fmt.Sprintf("style=\"fill:%[1]s;fill-opacity:1;stroke-width:2;stroke:%[1]s;stroke-opacity:0.4;stroke-alignment:outside;\"", txtColor)} // on
	)

	img[0] = image.NewRGBA(image.Rect(0, 0, int(sw), int(sh)))
	img[1] = image.NewRGBA(image.Rect(0, 0, int(sw), int(sh)))

	var iconMem1 = new(bytes.Buffer)
	var iconMem0 = new(bytes.Buffer)

	var canvas1 = svg.New(iconMem1)
	var canvas0 = svg.New(iconMem0)

	canvas1.Start(wt, ht)
	canvas0.Start(wt, ht)

	// inject filter definition - svgo does not expose enough
	iconMem1.WriteString(`<defs>
  <filter id="blur" x="0" y="0" width="200%" height="200%">
    <feOffset result="offOut" in="SourceAlpha" dx="5" dy="5" />
    <feGaussianBlur result="blurOut" in="offOut" stdDeviation="0.5" />
    <feBlend in="SourceGraphic" in2="blurOut" mode="normal" />
  </filter>
</defs>`)

	canvas1.Group(`id="time" filter="url(#blur)"`)
	canvas0.Group(`id="time" filter="url(#blur)"`)

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
			canvas1.Circle(int(segPos[i])-6, 80, 10, seg7style[1])
			canvas0.Circle(int(segPos[i]), 40, 10, seg7style[0])
			canvas0.Circle(int(segPos[i])-6, 80, 10, seg7style[0])
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
			fmt.Sprintf("style=\"fill:gray;fill-opacity:0.3;stroke-width:0.5;stroke:%s;stroke-opacity:0.25;stroke-alignment:inside;\"", txtColor), // off
			fmt.Sprintf("style=\"fill:%[1]s;fill-opacity:1;stroke-width:2;stroke:%[1]s;stroke-opacity:0.4;stroke-alignment:outside;\"", txtColor)} // on
	)

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

// imageTimeThick - best
func imageTimeThick(t time.Time, scaleh float64, txtColor string) (img [2]draw.Image, err error) {

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
			fmt.Sprintf("style=\"fill:gray;fill-opacity:0.3;stroke-width:0.5;stroke:%s;stroke-opacity:0.25;stroke-alignment:inside;\"", txtColor), // off
			fmt.Sprintf("style=\"fill:%[1]s;fill-opacity:1;stroke-width:2;stroke:%[1]s;stroke-opacity:0.4;stroke-alignment:outside;\"", txtColor)} // on
	)

	img[0] = image.NewRGBA(image.Rect(0, 0, int(sw), int(sh)))
	img[1] = image.NewRGBA(image.Rect(0, 0, int(sw), int(sh)))

	var iconMem1 = new(bytes.Buffer)
	var iconMem0 = new(bytes.Buffer)

	var canvas1 = svg.New(iconMem1)
	var canvas0 = svg.New(iconMem0)

	canvas1.Start(wt, ht)
	canvas0.Start(wt, ht)

	// inject filter definition - svgo does not expose enough
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
			fmt.Sprintf("style=\"fill:gray;fill-opacity:0.3;stroke-width:0.5;stroke:%s;stroke-opacity:0.25;stroke-alignment:inside;\"", txtColor), // off
			fmt.Sprintf("style=\"fill:%[1]s;fill-opacity:1;stroke-width:2;stroke:%[1]s;stroke-opacity:0.4;stroke-alignment:outside;\"", txtColor)} // on
	)

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
