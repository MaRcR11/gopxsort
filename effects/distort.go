package effects

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/MaRcR11/sortmancy/internal/imgutil"
)

func bilinear(src *image.NRGBA, fx, fy float64) color.NRGBA {
	w, h := src.Rect.Dx(), src.Rect.Dy()
	fx = math.Max(0, math.Min(float64(w-1), fx))
	fy = math.Max(0, math.Min(float64(h-1), fy))
	x0, y0 := int(fx), int(fy)
	x1, y1 := min(x0+1, w-1), min(y0+1, h-1)
	tx, ty := fx-float64(x0), fy-float64(y0)
	var out [4]uint8
	for c := 0; c < 4; c++ {
		p := func(x, y int) float64 { return float64(src.Pix[(y*w+x)*4+c]) }
		top := p(x0, y0) + (p(x1, y0)-p(x0, y0))*tx
		bot := p(x0, y1) + (p(x1, y1)-p(x0, y1))*tx
		out[c] = imgutil.Clamp8(top + (bot-top)*ty)
	}
	return color.NRGBA{out[0], out[1], out[2], out[3]}
}

func remapFloat(img image.Image, from func(x, y float64) (float64, float64, bool)) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	dst := image.NewNRGBA(src.Rect)
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				sx, sy, ok := from(float64(x), float64(y))
				if !ok {
					imgutil.Set(dst.Pix, y*w+x, color.NRGBA{0, 0, 0, 255})
					continue
				}
				imgutil.Set(dst.Pix, y*w+x, bilinear(src, sx, sy))
			}
		}
	})
	return dst
}

func Displace(img image.Image, displacementMap image.Image, amountX, amountY float64) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	m := src
	if displacementMap != nil {
		m = imgutil.ResizeNearest(displacementMap, w, h)
	}
	lum := make([]float64, w*h)
	for i := range lum {
		lum[i] = imgutil.Luma(imgutil.At(m.Pix, i)) - 0.5
	}
	return remapFloat(src, func(x, y float64) (float64, float64, bool) {
		l := lum[int(y)*w+int(x)]
		return x + l*amountX, y + l*amountY, true
	})
}

func Echo(img image.Image, copies int, offset image.Point, decay float64) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	dst := image.NewNRGBA(src.Rect)
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				i := (y*w + x) * 4
				var sum [3]float64
				total := 0.0
				weight := 1.0
				for k := 0; k <= copies; k++ {
					sx := imgutil.Clamp(x-k*offset.X, 0, w-1)
					sy := imgutil.Clamp(y-k*offset.Y, 0, h-1)
					j := (sy*w + sx) * 4
					for c := 0; c < 3; c++ {
						sum[c] += float64(src.Pix[j+c]) * weight
					}
					total += weight
					weight *= decay
				}
				for c := 0; c < 3; c++ {
					dst.Pix[i+c] = imgutil.Clamp8(sum[c] / total)
				}
				dst.Pix[i+3] = src.Pix[i+3]
			}
		}
	})
	return dst
}

func DataBend(img image.Image, delay int, feedback float64) *image.NRGBA {
	dst := imgutil.ToNRGBA(img)
	n := dst.Rect.Dx() * dst.Rect.Dy()
	stream := make([]float64, n*3)
	for i := 0; i < n; i++ {
		for c := 0; c < 3; c++ {
			stream[i*3+c] = float64(dst.Pix[i*4+c])
		}
	}
	delay = max(delay, 1)
	for i := delay; i < len(stream); i++ {
		stream[i] = math.Min(255, stream[i]+feedback*stream[i-delay])
	}
	for i := 0; i < n; i++ {
		for c := 0; c < 3; c++ {
			dst.Pix[i*4+c] = imgutil.Clamp8(stream[i*3+c])
		}
	}
	return dst
}

func center(img image.Image) (cx, cy, r float64) {
	b := img.Bounds()
	cx, cy = float64(b.Dx()-1)/2, float64(b.Dy()-1)/2
	return cx, cy, math.Min(cx, cy)
}

func Swirl(img image.Image, degrees, radius float64) *image.NRGBA {
	cx, cy, r := center(img)
	r *= radius * 2
	turn := degrees * math.Pi / 180
	return remapFloat(img, func(x, y float64) (float64, float64, bool) {
		dx, dy := x-cx, y-cy
		d := math.Hypot(dx, dy)
		if d >= r || r == 0 {
			return x, y, true
		}
		t := 1 - d/r
		a := math.Atan2(dy, dx) + turn*t*t
		return cx + d*math.Cos(a), cy + d*math.Sin(a), true
	})
}

func Ripple(img image.Image, amplitude, wavelength float64) *image.NRGBA {
	cx, cy, _ := center(img)
	if wavelength == 0 {
		wavelength = 1
	}
	return remapFloat(img, func(x, y float64) (float64, float64, bool) {
		dx, dy := x-cx, y-cy
		d := math.Hypot(dx, dy)
		if d == 0 {
			return x, y, true
		}
		s := amplitude * math.Sin(2*math.Pi*d/wavelength)
		return x + dx/d*s, y + dy/d*s, true
	})
}

func Bulge(img image.Image, strength, radius float64) *image.NRGBA {
	cx, cy, r := center(img)
	r *= radius * 2
	return remapFloat(img, func(x, y float64) (float64, float64, bool) {
		dx, dy := x-cx, y-cy
		d := math.Hypot(dx, dy)
		if d >= r || r == 0 {
			return x, y, true
		}
		exp := 1 + strength
		if strength < 0 {
			exp = 1 / (1 - strength)
		}
		f := r * math.Pow(d/r, exp) / math.Max(d, 1e-9)
		return cx + dx*f, cy + dy*f, true
	})
}

func Polar(img image.Image, inverse bool) *image.NRGBA {
	b := img.Bounds()
	w, h := float64(b.Dx()), float64(b.Dy())
	cx, cy := (w-1)/2, (h-1)/2
	maxR := math.Hypot(cx, cy)
	return remapFloat(img, func(x, y float64) (float64, float64, bool) {
		if inverse {
			a := x / w * 2 * math.Pi
			d := (1 - y/h) * maxR
			return cx + d*math.Cos(a), cy - d*math.Sin(a), true
		}
		dx, dy := x-cx, cy-y
		a := math.Atan2(dy, dx)
		if a < 0 {
			a += 2 * math.Pi
		}
		return a / (2 * math.Pi) * (w - 1), (1 - math.Hypot(dx, dy)/maxR) * (h - 1), true
	})
}

func Curvature(img image.Image, strength float64) *image.NRGBA {
	b := img.Bounds()
	w, h := float64(b.Dx()), float64(b.Dy())
	return remapFloat(img, func(x, y float64) (float64, float64, bool) {
		u, v := x/(w-1)*2-1, y/(h-1)*2-1
		k := 1 + strength*(u*u+v*v)
		u, v = u*k, v*k
		if u < -1 || u > 1 || v < -1 || v > 1 {
			return 0, 0, false
		}
		return (u + 1) / 2 * (w - 1), (v + 1) / 2 * (h - 1), true
	})
}

func SlitScan(img image.Image, position float64, vertical bool) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	if vertical {
		row := imgutil.Clamp(int(position*float64(h-1)), 0, h-1)
		return remap(src, func(x, y int) (int, int) { return x, min(y, row) })
	}
	col := imgutil.Clamp(int(position*float64(w-1)), 0, w-1)
	return remap(src, func(x, y int) (int, int) { return min(x, col), y })
}

func Interlace(img image.Image, shift, spacing int, darken float64) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	w := src.Rect.Dx()
	spacing = max(spacing, 2)
	dst := remap(src, func(x, y int) (int, int) {
		if y%spacing == spacing-1 {
			return imgutil.Wrap(x-shift, w), y
		}
		return x, y
	})
	for y := spacing - 1; y < dst.Rect.Dy(); y += spacing {
		for i := y * w * 4; i < (y+1)*w*4; i++ {
			if i%4 != 3 {
				dst.Pix[i] = imgutil.Clamp8(float64(dst.Pix[i]) * (1 - darken))
			}
		}
	}
	return dst
}

func Scatter(img image.Image, radius int, seed int64) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	dst := image.NewNRGBA(src.Rect)
	radius = max(radius, 1)
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			rng := rand.New(rand.NewSource(seed ^ int64(y+1)*0x2545F4914F6CDD1D))
			for x := 0; x < w; x++ {
				sx := imgutil.Clamp(x+rng.Intn(2*radius+1)-radius, 0, w-1)
				sy := imgutil.Clamp(y+rng.Intn(2*radius+1)-radius, 0, h-1)
				copy(dst.Pix[(y*w+x)*4:(y*w+x)*4+4], src.Pix[(sy*w+sx)*4:])
			}
		}
	})
	return dst
}

func Tracking(img image.Image, bands int, shift float64, noise float64, seed int64) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	dst := imgutil.ToNRGBA(src)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	rng := rand.New(rand.NewSource(seed))
	for b := 0; b < bands; b++ {
		y0 := rng.Intn(h)
		height := max(h/40, 2) + rng.Intn(max(h/15, 1))
		maxShift := shift * float64(w)
		for y := y0; y < min(y0+height, h); y++ {
			t := float64(y-y0) / float64(height)
			off := int(maxShift * math.Sin(t*math.Pi) * (0.6 + 0.4*rng.Float64()))
			for x := 0; x < w; x++ {
				sx := imgutil.Wrap(x-off, w)
				si, di := (y*w+sx)*4, (y*w+x)*4
				for c := 0; c < 3; c++ {
					v := float64(src.Pix[si+c]) + 40*math.Sin(t*math.Pi)
					if rng.Float64() < noise*math.Sin(t*math.Pi) {
						v = 200 + 55*rng.Float64()
					}
					dst.Pix[di+c] = imgutil.Clamp8(v)
				}
			}
		}
	}
	return dst
}
