package effects

import (
	"image"
	"image/color"

	"github.com/MaRcR11/sortcery/internal/imgutil"
)

func MedianCut(img image.Image, colors int) color.Palette {
	return imgutil.MedianCut(imgutil.ToNRGBA(img), colors)
}

func Quantize(img image.Image, colors int, method DitherMethod) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	return Dither(src, DitherOptions{Method: method, Palette: imgutil.MedianCut(src, colors)})
}
