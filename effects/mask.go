package effects

import (
	"fmt"
	"image"
	"math"

	"github.com/MaRcR11/sortmancy/internal/imgutil"
)

func RegionMask(width, height int, r image.Rectangle) *image.Gray {
	m := image.NewGray(image.Rect(0, 0, width, height))
	r = r.Intersect(m.Rect)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			m.Pix[y*m.Stride+x] = 255
		}
	}
	return m
}

func ParseRegion(s string) (image.Rectangle, error) {
	var x, y, w, h int
	if _, err := fmt.Sscanf(s, "%d,%d,%d,%d", &x, &y, &w, &h); err != nil || w <= 0 || h <= 0 {
		return image.Rectangle{}, fmt.Errorf("region %q must be x,y,width,height with positive width and height", s)
	}
	return image.Rect(x, y, x+w, y+h), nil
}

func maskWeights(mask image.Image, w, h int, invert bool, feather int) []float64 {
	at := imgutil.SampleNearest(mask, w, h)
	weights := make([]float64, w*h)
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				v := imgutil.LumaOf(at(x, y))
				if invert {
					v = 1 - v
				}
				weights[y*w+x] = v
			}
		}
	})
	if feather > 0 {
		r := max(feather/2, 1)
		weights = imgutil.BlurFloat(imgutil.BlurFloat(weights, w, h, r), w, h, r)
	}
	return weights
}

func ApplyMask(original, processed, mask image.Image, invert bool, feather int) (*image.NRGBA, error) {
	orig := imgutil.ToNRGBA(original)
	out := imgutil.ToNRGBA(processed)
	if orig.Rect != out.Rect {
		return nil, fmt.Errorf("mask: the effect changed the size from %dx%d to %dx%d, so it cannot be masked",
			orig.Rect.Dx(), orig.Rect.Dy(), out.Rect.Dx(), out.Rect.Dy())
	}
	w, h := orig.Rect.Dx(), orig.Rect.Dy()
	weights := maskWeights(mask, w, h, invert, feather)
	imgutil.Parallel(h, func(y0, y1 int) {
		for i := y0 * w; i < y1*w; i++ {
			k := weights[i]
			if k >= 1 {
				continue
			}
			for c := 0; c < 4; c++ {
				a, b := float64(orig.Pix[i*4+c]), float64(out.Pix[i*4+c])
				out.Pix[i*4+c] = uint8(math.Round(a + (b-a)*k))
			}
		}
	})
	return out, nil
}
