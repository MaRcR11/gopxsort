package effects

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"math"
	"math/rand"

	"github.com/MaRcR11/sortcery/internal/imgutil"
)

func ChannelShift(img image.Image, r, g, b image.Point) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	dst := image.NewNRGBA(src.Rect)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	offs := [3]image.Point{r, g, b}
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				i := (y*w + x) * 4
				for ch, o := range offs {
					sx, sy := imgutil.Wrap(x-o.X, w), imgutil.Wrap(y-o.Y, h)
					dst.Pix[i+ch] = src.Pix[(sy*w+sx)*4+ch]
				}
				dst.Pix[i+3] = src.Pix[i+3]
			}
		}
	})
	return dst
}

func ChromaticAberration(img image.Image, strength float64) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	dst := image.NewNRGBA(src.Rect)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	cx, cy := float64(w-1)/2, float64(h-1)/2
	sample := func(x, y float64, ch int) uint8 {
		sx := imgutil.Clamp(int(math.Round(x)), 0, w-1)
		sy := imgutil.Clamp(int(math.Round(y)), 0, h-1)
		return src.Pix[(sy*w+sx)*4+ch]
	}
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				dx, dy := float64(x)-cx, float64(y)-cy
				i := (y*w + x) * 4
				dst.Pix[i] = sample(cx+dx*(1-strength), cy+dy*(1-strength), 0)
				dst.Pix[i+1] = src.Pix[i+1]
				dst.Pix[i+2] = sample(cx+dx*(1+strength), cy+dy*(1+strength), 2)
				dst.Pix[i+3] = src.Pix[i+3]
			}
		}
	})
	return dst
}

func Scanlines(img image.Image, spacing int, intensity float64) *image.NRGBA {
	dst := imgutil.ToNRGBA(img)
	spacing = max(spacing, 1)
	f := 1 - intensity
	w := dst.Rect.Dx()
	for y := 0; y < dst.Rect.Dy(); y += spacing {
		for i := y * w; i < (y+1)*w; i++ {
			p := dst.Pix[i*4 : i*4+3]
			for j := range p {
				p[j] = imgutil.Clamp8(float64(p[j]) * f)
			}
		}
	}
	return dst
}

func Noise(img image.Image, amount float64, mono bool, seed int64) *image.NRGBA {
	dst := imgutil.ToNRGBA(img)
	w, h := dst.Rect.Dx(), dst.Rect.Dy()
	a := amount * 255
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			rng := rand.New(rand.NewSource(seed ^ int64(y+1)*0x5DEECE66D))
			for i := y * w; i < (y+1)*w; i++ {
				p := dst.Pix[i*4 : i*4+3]
				n := (rng.Float64()*2 - 1) * a
				for j := range p {
					if !mono && j > 0 {
						n = (rng.Float64()*2 - 1) * a
					}
					p[j] = imgutil.Clamp8(float64(p[j]) + n)
				}
			}
		}
	})
	return dst
}

func Glitch(img image.Image, slices int, maxShift float64, seed int64) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	dst := imgutil.ToNRGBA(src)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	rng := rand.New(rand.NewSource(seed))
	maxPx := max(int(maxShift*float64(w)), 1)
	maxBand := max(h/20, 1)
	for s := 0; s < slices; s++ {
		y0 := rng.Intn(h)
		y1 := min(y0+1+rng.Intn(maxBand), h)
		shift := rng.Intn(2*maxPx+1) - maxPx
		channel := -1
		if rng.Intn(3) == 0 {
			channel = rng.Intn(3)
		}
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				si := (y*w + imgutil.Wrap(x-shift, w)) * 4
				di := (y*w + x) * 4
				if channel >= 0 {
					dst.Pix[di+channel] = src.Pix[si+channel]
				} else {
					copy(dst.Pix[di:di+4], src.Pix[si:si+4])
				}
			}
		}
	}
	return dst
}

func Wave(img image.Image, amplitude, wavelength float64, vertical bool) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	dst := image.NewNRGBA(src.Rect)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	if wavelength == 0 {
		wavelength = 1
	}
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				sx, sy := x, y
				if vertical {
					sy = imgutil.Wrap(y+int(math.Round(amplitude*math.Sin(2*math.Pi*float64(x)/wavelength))), h)
				} else {
					sx = imgutil.Wrap(x+int(math.Round(amplitude*math.Sin(2*math.Pi*float64(y)/wavelength))), w)
				}
				copy(dst.Pix[(y*w+x)*4:(y*w+x)*4+4], src.Pix[(sy*w+sx)*4:])
			}
		}
	})
	return dst
}

func Melt(img image.Image, maxShift int, seed int64) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	dst := image.NewNRGBA(src.Rect)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	rng := rand.New(rand.NewSource(seed))

	walk := make([]float64, w)
	v := 0.0
	for x := range walk {
		v += rng.NormFloat64()
		walk[x] = v
	}
	radius := max(w/30, 1)
	sag := make([]float64, w)
	lo, hi := math.Inf(1), math.Inf(-1)
	for x := range sag {
		sum, n := 0.0, 0
		for k := max(x-radius, 0); k <= min(x+radius, w-1); k++ {
			sum += walk[k]
			n++
		}
		sag[x] = sum / float64(n)
		lo, hi = math.Min(lo, sag[x]), math.Max(hi, sag[x])
	}

	drips := make([]float64, w)
	for d := 0; d < max(w/12, 1); d++ {
		cx := rng.Intn(w)
		half := 2 + rng.Intn(max(w/60, 3))
		length := rng.Float64()
		for x := max(cx-half, 0); x <= min(cx+half, w-1); x++ {
			t := float64(x-cx) / float64(half)
			drips[x] = math.Max(drips[x], length*(0.5+0.5*math.Cos(math.Pi*t)))
		}
	}
	shift := make([]int, w)
	for x := range shift {
		base := 0.0
		if hi > lo {
			base = (sag[x] - lo) / (hi - lo)
		}
		shift[x] = int((0.3*base + 0.7*drips[x]) * float64(maxShift))
	}
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				sy := max(y-shift[x], 0)
				copy(dst.Pix[(y*w+x)*4:(y*w+x)*4+4], src.Pix[(sy*w+x)*4:])
			}
		}
	})
	return dst
}

func JPEGCorrupt(img image.Image, amount, quality int, seed int64) (*image.NRGBA, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, imgutil.ToNRGBA(img), &jpeg.Options{Quality: imgutil.Clamp(quality, 1, 100)}); err != nil {
		return nil, err
	}
	data := buf.Bytes()
	start := scanStart(data)
	end := len(data) - 2
	if start < 0 || end-start < 16 {
		return nil, errors.New("jpeg: no scan data to corrupt")
	}
	rng := rand.New(rand.NewSource(seed))
	corrupt := make([]byte, len(data))
	for attempt := 0; attempt < 32; attempt++ {
		copy(corrupt, data)
		n := max(amount>>(attempt/4), 1)
		for i := 0; i < n; i++ {
			p := start + rng.Intn(end-start)

			if corrupt[p] == 0xFF || corrupt[p-1] == 0xFF {
				continue
			}
			corrupt[p] = byte(rng.Intn(0xFF))
		}
		out, err := jpeg.Decode(bytes.NewReader(corrupt))
		if err == nil {
			return imgutil.ToNRGBA(out), nil
		}
	}
	return imgutil.ToNRGBA(img), nil
}

func scanStart(b []byte) int {
	for i := len(b) - 4; i >= 2; i-- {
		if b[i] == 0xFF && b[i+1] == 0xDA {
			return i + 2 + int(b[i+2])<<8 + int(b[i+3])
		}
	}
	return -1
}
