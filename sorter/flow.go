package sorter

import (
	"image"
	"math"

	"github.com/MaRcR11/gopxsort/internal/imgutil"
)

type ImagePath interface {
	Path
	LinesFor(img *image.NRGBA) [][]int
}

func warp(w, h int, vertical bool, off func(x, y int) float64) [][]int {
	major, minor := w, h
	if vertical {
		major, minor = h, w
	}
	keys := make([]int, w*h)
	lo, hi := math.MaxInt, math.MinInt
	for a := 0; a < major; a++ {
		for c := 0; c < minor; c++ {
			x, y := a, c
			if vertical {
				x, y = c, a
			}
			k := c - int(math.Round(off(x, y)))
			keys[y*w+x] = k
			lo, hi = min(lo, k), max(hi, k)
		}
	}
	counts := make([]int, hi-lo+1)
	for _, k := range keys {
		counts[k-lo]++
	}
	backing := make([]int, w*h)
	lines := make([][]int, len(counts))
	pos := 0
	for k, n := range counts {
		lines[k] = backing[pos : pos : pos+n]
		pos += n
	}
	for a := 0; a < major; a++ {
		for c := 0; c < minor; c++ {
			x, y := a, c
			if vertical {
				x, y = c, a
			}
			k := keys[y*w+x] - lo
			lines[k] = append(lines[k], y*w+x)
		}
	}
	out := lines[:0]
	for _, l := range lines {
		if len(l) > 0 {
			out = append(out, l)
		}
	}
	if vertical {
		reverseAll(out)
	}
	return out
}

func FlowField(scale, amount float64, vertical bool, seed int64) Path {
	scale = math.Max(scale, 1)
	return PathFunc(func(w, h int) [][]int {
		return warp(w, h, vertical, func(x, y int) float64 {
			fx, fy := float64(x)/scale, float64(y)/scale
			n := valueNoise(fx, fy, seed) + 0.5*valueNoise(fx*2, fy*2, seed+1)
			return amount * (n/1.5*2 - 1)
		})
	})
}

func valueNoise(x, y float64, seed int64) float64 {
	x0, y0 := math.Floor(x), math.Floor(y)
	tx, ty := x-x0, y-y0
	tx, ty = tx*tx*(3-2*tx), ty*ty*(3-2*ty)
	ix, iy := int64(x0), int64(y0)
	v00, v10 := hashNoise(ix, iy, seed), hashNoise(ix+1, iy, seed)
	v01, v11 := hashNoise(ix, iy+1, seed), hashNoise(ix+1, iy+1, seed)
	top := v00 + (v10-v00)*tx
	bot := v01 + (v11-v01)*tx
	return top + (bot-top)*ty
}

func hashNoise(x, y, seed int64) float64 {
	z := uint64(x)*0x9E3779B97F4A7C15 ^ uint64(y)*0xC2B2AE3D27D4EB4F ^ uint64(seed)*0x165667B19E3779F9
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	z ^= z >> 31
	return float64(z>>11) / float64(1<<53)
}

type contour struct {
	amount   float64
	blur     int
	vertical bool
}

func Contour(amount float64, blur int, vertical bool) Path {
	return contour{amount: amount, blur: max(blur, 0), vertical: vertical}
}

func (c contour) Lines(w, h int) [][]int {
	return warp(w, h, c.vertical, func(int, int) float64 { return 0 })
}

func (c contour) LinesFor(img *image.NRGBA) [][]int {
	w, h := img.Rect.Dx(), img.Rect.Dy()
	lum := make([]float64, w*h)
	for i := range lum {
		lum[i] = imgutil.Luma(imgutil.At(img.Pix, i))
	}
	lum = imgutil.BlurFloat(lum, w, h, c.blur)
	return warp(w, h, c.vertical, func(x, y int) float64 { return c.amount * lum[y*w+x] })
}

func Hilbert() Path {
	return PathFunc(func(w, h int) [][]int {
		n := 1
		for n < max(w, h) {
			n *= 2
		}
		line := make([]int, 0, w*h)
		for d := 0; d < n*n; d++ {
			x, y := hilbertPoint(n, d)
			if x < w && y < h {
				line = append(line, y*w+x)
			}
		}
		return [][]int{line}
	})
}

func hilbertPoint(n, d int) (x, y int) {
	t := d
	for s := 1; s < n; s *= 2 {
		rx := 1 & (t / 2)
		ry := 1 & (t ^ rx)
		if ry == 0 {
			if rx == 1 {
				x, y = s-1-x, s-1-y
			}
			x, y = y, x
		}
		x += s * rx
		y += s * ry
		t /= 4
	}
	return x, y
}
