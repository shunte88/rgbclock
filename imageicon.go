package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"image/draw"
	"io"
	"io/ioutil"
	"os"
	"strings"

	svg "github.com/ajstarks/svgo"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func getImageIcon(i icon) (img draw.Image, err error) {

	f, err := os.Open(iconFile(i))
	if err != nil {
		return img, err
	}
	defer f.Close()

	var iconMem = new(bytes.Buffer)

	if i.asis {

		iconMem.ReadFrom(f)

	} else {

		var s SVG

		var canvas = svg.New(iconMem)

		if err = xml.NewDecoder(f).Decode(&s); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to parse (%v)\n", err)
			return img, err
		}
		canvas.Start(i.width, i.height)
		style := fmt.Sprintf(iconStyle, i.color)
		transform := ``
		if i.rotate+i.scale > 0.0 {
			if 0.00 != i.scale {
				transform += fmt.Sprintf("scale(%[1]f %[1]f)", i.scale)
			}
			if 0.00 != i.rotate {
				transform += fmt.Sprintf("rotate(%[1]f %[2]v %[2]v)", i.rotate, 15)
			}
			transform = fmt.Sprintf(`transform="%v"`, transform)
		}
		canvas.Group(transform, style)
		io.WriteString(canvas.Writer, s.Doc)
		canvas.Gend()
		canvas.End()

	}

	iconI, err := oksvg.ReadIconStream(iconMem)
	if err != nil {
		return img, err
	}

	img = image.NewRGBA(image.Rect(0, 0, i.width, i.height))
	gv := rasterx.NewScannerGV(i.width, i.height, img, img.Bounds())
	r := rasterx.NewDasher(i.width, i.height, gv)
	if 0 == i.alpha {
		i.alpha = 1.0
	}
	iconI.Draw(r, i.alpha)

	return img, nil

}

func getImageIconWIP(i icon) (img draw.Image, err error) {

	f, err := os.Open(iconFile(i))
	if err != nil {
		return img, err
	}
	defer f.Close()

	sx, sy := i.width, i.height
	var iconMem = new(bytes.Buffer)

	if i.asis {

		if i.filename == windDegIcon && i.rotate != 0 {
			body, err := ioutil.ReadAll(f)
			if err != nil {
				return img, err
			}
			bs := strings.Replace(string(body), `rotate(0`, fmt.Sprintf("rotate(%f", i.rotate), -1)
			iconMem.WriteString(bs)
		} else {
			iconMem.ReadFrom(f)
		}
	} else {

		var s SVG

		var canvas = svg.New(iconMem)

		if err = xml.NewDecoder(f).Decode(&s); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to parse (%v)\n", err)
			return img, err
		}

		canvas.Start(i.width, i.height)

		gfs := svg.Filterspec{Result: "SHADOW"}
		bfs := svg.Filterspec{Result: "BLUR"}
		ofs := svg.Filterspec{}

		canvas.Def()
		canvas.Filter("BLUR", `width="150%" height="150%"`)
		canvas.FeGaussianBlur(bfs, 4, 4)
		canvas.Fend()
		canvas.Filter("SHADOW", `x="-15%" y="-15%" width="150%" height="150%"`)
		canvas.FeGaussianBlur(gfs, 1, 1)
		canvas.FeOffset(ofs, 2, 2)
		canvas.Fend()

		if `` != i.popcolor {
			canvas.Group(`id="POP"`)
			canvas.Circle(int(i.width/2), int(i.height/2), int(i.width/3)-6)
			canvas.Gend()
		}
		gid := `id="ICON"`
		style := fmt.Sprintf(iconStyle, i.color)
		transform := ""
		if i.rotate+i.scale > 0.0 {
			if 0.00 != i.scale {
				transform += fmt.Sprintf("scale(%[1]f %[1]f)", i.scale)
			}
			if 0.00 != i.rotate {
				transform += fmt.Sprintf("rotate(%[1]f %[2]v %[2]v)", i.rotate, 15)
			}
			transform = fmt.Sprintf(`transform="%v"`, transform)
		}
		canvas.Group(gid)
		io.WriteString(canvas.Writer, s.Doc)
		canvas.Gend()
		canvas.DefEnd()

		if i.filename == windDegIcon {
			canvas.Circle(4+int(i.width/3), 4+int(i.height/3), int(i.width/3)-2,
				`style="fill: #0099FF" fill-opacity="0.4"`)
		} else if `` != i.popcolor {
			canvas.Use(0, 0, "#POP", fmt.Sprintf(`fill-opacity="0.15" style=" fill: %s; filter: url(#BLUR)"`, i.popcolor))
		}
		if i.shadow {
			canvas.Use(0, 0, "#ICON", `style="filter: url(#SHADOW); fill: black"`, transform)
		}
		if i.blur {
			canvas.Use(0, 0, "#ICON", fmt.Sprintf("style=\"filter: url(#BLUR); fill: %s\"", i.color), transform)
		}
		canvas.Use(0, 0, "#ICON", style, transform)
		canvas.End()

		//fmt.Println(iconMem.String())

	}

	iconI, err := oksvg.ReadIconStream(iconMem)
	if err != nil {
		return img, err
	}

	if 0 == i.scale {
		i.scale = 1.0
	}
	if 0 == i.alpha {
		i.alpha = 1.0
	}

	if i.asis && 1 != i.scale {
		sx, sy = int(i.scale*float64(sx)), int(i.scale*float64(sy))
	}
	img = image.NewRGBA(image.Rect(0, 0, sx, sy))
	gv := rasterx.NewScannerGV(sx, sy, img, img.Bounds())
	r := rasterx.NewDasher(sx, sy, gv)
	if i.asis && 0 != i.scale {
		iconI.SetTarget(0, 0, float64(sx), float64(sy))
	}
	iconI.Draw(r, i.alpha)

	return img, nil

}
