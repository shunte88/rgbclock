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

func imageTime(t time.Time, scaleh float64, txtColor string) (img [2]draw.Image, err error) {

	// dynamic SVG

	type segment struct {
		pathd string
		x     int
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
			{1, 1, 1, 1, 1, 0, 1}, // 6
			{1, 1, 1, 0, 0, 0, 0}, // 7
			{1, 1, 1, 1, 1, 1, 1}, // 8
			{1, 1, 1, 1, 0, 1, 1}} // 9

		segPos = [5]int{0, 80, int(wt / 2), 210, 290}

		seg7template = [7]segment{
			{pathd: `m%d,8l4,-4l28,0l4,4l-4,4l-28,0l-4,-4z`, x: 20},
			{pathd: `m%d,10l4,4l0,28l-4,4l-4,-4l0,-28l4,-4z`, x: 58},
			{pathd: `m%d,50l4,4l0,28l-4,4l-4,-4l0,-28l4,-4z`, x: 58},
			{pathd: `m%d,88l4,-4l28,0l4,4l-4,4l-28,0l-4,-4z`, x: 20},
			{pathd: `m%d,50l4,4l0,28l-4,4l-4,-4l0,-28l4,-4z`, x: 18},
			{pathd: `m%d,10l4,4l0,28l-4,4l-4,-4l0,-28l4,-4z`, x: 18},
			{pathd: `m%d,48l4,-4l28,0l4,4l-4,4l-28,0l-4,-4z`, x: 20}}

		seg7style = [2]string{
			`style="fill:darkgray;fill-opacity:0.2;stroke-width:0.5;stroke:%s;stroke-opacity:0.2;stroke-alignment:inside;"`, // off
			`style="fill:%[1]s;fill-opacity:1;stroke-width:2;stroke:%[1]s;stroke-opacity:0.8;stroke-alignment:outside;"`}    // on
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

	canvas1.Group("id=\"time\" transform=\"skewX(-6)\"")
	canvas0.Group("id=\"time\" transform=\"skewX(-6)\"")
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
			canvas1.Circle(segPos[i], 30, 8, seg7style[1])
			canvas1.Circle(segPos[i], 60, 8, seg7style[1])
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
