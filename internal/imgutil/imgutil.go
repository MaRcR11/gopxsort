package imgutil

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
)

func ToNRGBA(img image.Image) *image.NRGBA {
	b := img.Bounds()
	r := image.Rect(0, 0, b.Dx(), b.Dy())
	switch src := img.(type) {
	case *image.NRGBA:
		dst := image.NewNRGBA(r)
		for y := 0; y < r.Dy(); y++ {
			copy(dst.Pix[y*dst.Stride:(y+1)*dst.Stride], src.Pix[src.PixOffset(b.Min.X, b.Min.Y+y):])
		}
		return dst
	case *image.YCbCr, *image.Gray:

		rgba := image.NewRGBA(r)
		draw.Draw(rgba, r, src, b.Min, draw.Src)
		return &image.NRGBA{Pix: rgba.Pix, Stride: rgba.Stride, Rect: r}
	}
	dst := image.NewNRGBA(r)
	draw.Draw(dst, r, img, b.Min, draw.Src)
	return dst
}

func Parallel(n int, fn func(start, end int)) {
	workers := runtime.GOMAXPROCS(0)
	if workers > n {
		workers = n
	}
	if workers <= 1 {
		if n > 0 {
			fn(0, n)
		}
		return
	}
	chunk := (n + workers - 1) / workers
	var wg sync.WaitGroup
	for s := 0; s < n; s += chunk {
		e := min(s+chunk, n)
		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			fn(s, e)
		}(s, e)
	}
	wg.Wait()
}

func ParallelEach(n int, fn func(i int)) {
	workers := min(runtime.GOMAXPROCS(0), n)
	if workers <= 1 {
		for i := 0; i < n; i++ {
			fn(i)
		}
		return
	}
	var next int64 = -1
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for {
				i := int(atomic.AddInt64(&next, 1))
				if i >= n {
					return
				}
				fn(i)
			}
		}()
	}
	wg.Wait()
}

func At(pix []uint8, i int) color.NRGBA {
	p := pix[4*i : 4*i+4 : 4*i+4]
	return color.NRGBA{p[0], p[1], p[2], p[3]}
}

func Set(pix []uint8, i int, c color.NRGBA) {
	p := pix[4*i : 4*i+4 : 4*i+4]
	p[0], p[1], p[2], p[3] = c.R, c.G, c.B, c.A
}

func Clamp8(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 255 {
		return 255
	}
	return uint8(v + 0.5)
}

func Clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func Wrap(v, n int) int {
	v %= n
	if v < 0 {
		v += n
	}
	return v
}

func Luma(c color.NRGBA) float64 {
	return (0.2126*float64(c.R) + 0.7152*float64(c.G) + 0.0722*float64(c.B)) / 255
}

func LumaOf(c color.Color) float64 {
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	return Luma(n)
}

func RGBToHSL(r, g, b uint8) (h, s, l float64) {
	rf, gf, bf := float64(r)/255, float64(g)/255, float64(b)/255
	mx := max(rf, gf, bf)
	mn := min(rf, gf, bf)
	l = (mx + mn) / 2
	d := mx - mn
	if d == 0 {
		return 0, 0, l
	}
	if l > 0.5 {
		s = d / (2 - mx - mn)
	} else {
		s = d / (mx + mn)
	}
	switch mx {
	case rf:
		h = (gf - bf) / d
		if gf < bf {
			h += 6
		}
	case gf:
		h = (bf-rf)/d + 2
	default:
		h = (rf-gf)/d + 4
	}
	return h / 6, s, l
}

func HSLToRGB(h, s, l float64) (r, g, b uint8) {
	if s == 0 {
		v := Clamp8(l * 255)
		return v, v, v
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	return Clamp8(hueToRGB(p, q, h+1.0/3) * 255),
		Clamp8(hueToRGB(p, q, h) * 255),
		Clamp8(hueToRGB(p, q, h-1.0/3) * 255)
}

func hueToRGB(p, q, t float64) float64 {
	t -= math.Floor(t)
	switch {
	case t < 1.0/6:
		return p + (q-p)*6*t
	case t < 0.5:
		return q
	case t < 2.0/3:
		return p + (q-p)*(2.0/3-t)*6
	}
	return p
}

func Sobel(img *image.NRGBA) []float64 {
	w, h := img.Rect.Dx(), img.Rect.Dy()
	lum := make([]float64, w*h)
	for i := range lum {
		lum[i] = Luma(At(img.Pix, i))
	}
	out := make([]float64, w*h)
	at := func(x, y int) float64 {
		return lum[Clamp(y, 0, h-1)*w+Clamp(x, 0, w-1)]
	}
	Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				gx := at(x+1, y-1) + 2*at(x+1, y) + at(x+1, y+1) -
					at(x-1, y-1) - 2*at(x-1, y) - at(x-1, y+1)
				gy := at(x-1, y+1) + 2*at(x, y+1) + at(x+1, y+1) -
					at(x-1, y-1) - 2*at(x, y-1) - at(x+1, y-1)
				out[y*w+x] = math.Min(math.Hypot(gx, gy)/4, 1)
			}
		}
	})
	return out
}

func SampleNearest(img image.Image, w, h int) func(x, y int) color.Color {
	b := img.Bounds()
	return func(x, y int) color.Color {
		return img.At(b.Min.X+x*b.Dx()/w, b.Min.Y+y*b.Dy()/h)
	}
}

func BlurFloat(v []float64, w, h, r int) []float64 {
	if r <= 0 {
		return v
	}
	tmp := make([]float64, len(v))
	out := make([]float64, len(v))
	n := float64(2*r + 1)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			s := 0.0
			for k := -r; k <= r; k++ {
				s += v[y*w+Clamp(x+k, 0, w-1)]
			}
			tmp[y*w+x] = s / n
		}
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			s := 0.0
			for k := -r; k <= r; k++ {
				s += tmp[Clamp(y+k, 0, h-1)*w+x]
			}
			out[y*w+x] = s / n
		}
	}
	return out
}

func ResizeNearest(img image.Image, w, h int) *image.NRGBA {
	at := SampleNearest(img, w, h)
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(x, y, at(x, y))
		}
	}
	return dst
}

func ToGray(img *image.NRGBA) *image.Gray {
	g := image.NewGray(img.Rect)
	for i := range g.Pix {
		g.Pix[i] = Clamp8(Luma(At(img.Pix, i)) * 255)
	}
	return g
}
