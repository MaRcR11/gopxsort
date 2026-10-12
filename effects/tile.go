package effects

import (
	"image"
	"image/draw"
	"math"

	"github.com/MaRcR11/sortmancy/internal/imgutil"
)

func Tile(img image.Image, cols, rows int, mirror bool) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	cols, rows = max(cols, 1), max(rows, 1)
	dst := image.NewNRGBA(src.Rect)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			x0, x1 := c*w/cols, (c+1)*w/cols
			y0, y1 := r*h/rows, (r+1)*h/rows
			if x1 <= x0 || y1 <= y0 {
				continue
			}
			tile := Resize(src, x1-x0, y1-y0, Bilinear)
			if mirror && c%2 == 1 {
				tile = FlipHorizontal(tile)
			}
			if mirror && r%2 == 1 {
				tile = FlipVertical(tile)
			}
			draw.Draw(dst, image.Rect(x0, y0, x1, y1), tile, image.Point{}, draw.Src)
		}
	}
	return dst
}

func Droste(img image.Image, scale float64, steps int) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	dst := imgutil.ToNRGBA(src)
	scale = math.Max(0.05, math.Min(0.95, scale))
	s := 1.0
	for i := 0; i < steps; i++ {
		s *= scale
		tw, th := int(float64(w)*s), int(float64(h)*s)
		if tw < 2 || th < 2 {
			break
		}
		x0, y0 := (w-tw)/2, (h-th)/2
		draw.Draw(dst, image.Rect(x0, y0, x0+tw, y0+th), Resize(dst, tw, th, Bilinear), image.Point{}, draw.Src)
	}
	return dst
}
