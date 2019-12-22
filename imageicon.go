package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"image/draw"
	"io"
	"os"

	svg "github.com/ajstarks/svgo"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func getImageIcon(i icon) (img draw.Image, err error) {

	var s SVG

	var iconMem = new(bytes.Buffer)
	var canvas = svg.New(iconMem)

	f, err := os.Open(iconFile(i))
	if err != nil {
		return img, err
	}
	defer f.Close()

	if err = xml.NewDecoder(f).Decode(&s); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse (%v)\n", err)
		return img, err
	}
	canvas.Start(i.width, i.height)
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
	canvas.Group(transform, style)
	io.WriteString(canvas.Writer, s.Doc)
	canvas.Gend()
	canvas.End()
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

	var s SVG

	var iconMem = new(bytes.Buffer)
	var canvas = svg.New(iconMem)

	f, err := os.Open(iconFile(i))
	if err != nil {
		return img, err
	}
	defer f.Close()

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

	//openPath := `<path `
	//closeEmpty := `/>`
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
	/*
		if 1==0 {
		if 1==1 {
			io.WriteString(canvas.Writer, strings.Replace(s.Doc,closeEmpty,transform+closeEmpty,-1))
		} else {
			io.WriteString(canvas.Writer, strings.Replace(s.Doc,openPath,openPath+transform,-1))
		}
		}
	*/
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
