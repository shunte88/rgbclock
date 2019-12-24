package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/disintegration/imaging"
)

func dumpImage(idx int, img image.Image) error {
	fo := fmt.Sprintf("rgbclock%010d.png", idx)
	fh, err := os.Create(fo)
	if err != nil {
		panic(err)
	}
	defer fh.Close()
	return png.Encode(fh, imaging.Resize(img, W*4, H*4, imaging.Lanczos))
}
