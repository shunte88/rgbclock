package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"
	"sync"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// InfoLabel with built-in marquee scroll
type InfoLabel struct {
	text       string
	buffer     []byte
	maxlen     int
	active     bool
	graphic    bool // image rather than text
	yy         int
	height     int
	x          int // scroll x
	v          int // scroll velocity (pix)
	marquee    chan bool
	dur        time.Duration
	b          image.Rectangle
	r          image.Rectangle
	place      image.Rectangle
	pt         image.Point
	slice      *image.RGBA
	canvas     *image.RGBA
	color      color.Color
	face       font.Face
	fontHeight float64
	mux        sync.Mutex
}

// NewInfoLabel instatiates an info label
func NewInfoLabel(maxlen, velocity int, d time.Duration, graphical, start bool) *InfoLabel {
	il := InfoLabel{
		text:       ``,
		dur:        d,
		graphic:    graphical,
		active:     false,
		buffer:     []byte(``),
		v:          velocity, //6,
		slice:      image.NewRGBA(image.Rect(0, 0, 1, 1)),
		canvas:     image.NewRGBA(image.Rect(0, 0, 1, 1)),
		face:       basicfont.Face7x13,
		fontHeight: 13,
		color:      image.Black,
	}
	il.maxlen = maxlen
	if start {
		il.Start()
	}
	return &il
}

// Start the label marquee
func (il *InfoLabel) Start() {
	if !il.active {
		il.marquee = sched(il.Marquee, il.dur)
		il.active = true
	}
}

// Active if scrolling
func (il *InfoLabel) Active() bool {
	return il.active
}

// SetFace defines font face
func (il *InfoLabel) SetFace(f font.Face, x string) {
	t := il.text
	gt := false
	if il.face != f {
		gt = true
		il.face = f
		fmx := il.face.Metrics()
		il.yy = 2 + int(float64(fmx.Ascent>>6)-float64(fmx.Descent>>6))
		il.fontHeight = float64((fmx.Height >> 6) + 2)
	}
	r, g, b, a := il.parseHexColor(x)
	il.color = color.NRGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
	if gt {
		il.text = ``
		il.SetText(t)
	}
}

// Stop the label marquee
func (il *InfoLabel) Stop() {
	il.marquee <- true
	il.active = false
}

func (il *InfoLabel) measureString(s string) float64 {
	d := &font.Drawer{
		Face: il.face,
	}
	adv := d.MeasureString(s)
	return float64(adv >> 6)
}

func (il *InfoLabel) parseHexColor(x string) (r, g, b, a int) {
	x = strings.TrimPrefix(x, "#")
	a = 255
	if len(x) == 3 {
		format := "%1x%1x%1x"
		fmt.Sscanf(x, format, &r, &g, &b)
		r |= r << 4
		g |= g << 4
		b |= b << 4
	}
	if len(x) == 6 {
		format := "%02x%02x%02x"
		fmt.Sscanf(x, format, &r, &g, &b)
	}
	if len(x) == 8 {
		format := "%02x%02x%02x%02x"
		fmt.Sscanf(x, format, &r, &g, &b, &a)
	}
	return
}

// SetMaxlen allows for adjusting based on font metric - we should aoutomate
func (il *InfoLabel) SetMaxlen(m int) {
	if il.maxlen != m {
		il.mux.Lock()
		il.maxlen = m
		// reset buffer just in case
		if m < len(il.text) {
			t := il.text
			il.text = ``
			il.SetText(t)
		}
		il.mux.Unlock()
	}
}

// GetText get label source text
func (il *InfoLabel) GetText() string {
	return il.text
}

// SetText sets label source text
func (il *InfoLabel) SetText(t string) {
	if il.text != t {
		il.mux.Lock()
		il.text = t
		if !il.graphic {
			il.buffer = []byte(` ` + t)
		} else {
			ww := int(il.measureString(` `))
			w := int(il.measureString(t)) + (2 * ww)
			il.canvas = image.NewRGBA(image.Rect(0, 0, w, int(il.fontHeight+2)))
			il.b = il.canvas.Bounds()
			il.r = il.b
			il.r.Max.X = il.v
			il.place = image.Rect(il.b.Max.X-il.v, 0, il.b.Max.X, il.b.Max.Y)
			il.slice = image.NewRGBA(il.r)
			il.pt = image.Pt(il.v, 0)

			mx := il.face.Metrics()
			yy := 2 + int(float64(mx.Ascent>>6)-float64(mx.Descent>>6))

			point := fixed.Point26_6{fixed.Int26_6(ww * 64), fixed.Int26_6(yy * 64)}

			d := &font.Drawer{
				Dst:  il.canvas,
				Src:  image.NewUniform(il.color),
				Face: il.face,
				Dot:  point,
			}
			d.DrawString(t)

		}
		il.mux.Unlock()
	}
}

// Marquee scrolls text by manipulating text buffer (rotates)
func (il *InfoLabel) Marquee() {
	if !il.graphic {
		if len(il.text) > il.maxlen {
			il.buffer = append(il.buffer[1:], il.buffer[0:1]...)
		}
	} else {
		if len(il.text) > il.maxlen {
			draw.Draw(il.slice, il.r, il.canvas, il.r.Min, draw.Src)
			draw.Draw(il.canvas, il.b, il.canvas, il.b.Min.Add(il.pt), draw.Src)
			draw.Draw(il.canvas, il.place, il.slice, il.r.Min, draw.Src)
		}
	}
}

// Image returns current canvas graphic-mode
func (il *InfoLabel) Image() *image.RGBA {
	return il.canvas
}

// Display gets the current marquee text-mode
func (il *InfoLabel) Display() string {
	return string(il.buffer)
}
