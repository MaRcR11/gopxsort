package effects

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/MaRcR11/gopxsort/internal/imgutil"
)

func Map(img image.Image, fn func(c color.NRGBA) color.NRGBA) *image.NRGBA {
	dst := imgutil.ToNRGBA(img)
	w, h := dst.Rect.Dx(), dst.Rect.Dy()
	imgutil.Parallel(h, func(y0, y1 int) {
		for i := y0 * w; i < y1*w; i++ {
			imgutil.Set(dst.Pix, i, fn(imgutil.At(dst.Pix, i)))
		}
	})
	return dst
}

func mapLUT(img image.Image, lut *[256]uint8) *image.NRGBA {
	return Map(img, func(c color.NRGBA) color.NRGBA {
		return color.NRGBA{lut[c.R], lut[c.G], lut[c.B], c.A}
	})
}

func buildLUT(fn func(v float64) float64) *[256]uint8 {
	var lut [256]uint8
	for i := range lut {
		lut[i] = imgutil.Clamp8(fn(float64(i)))
	}
	return &lut
}

func Invert(img image.Image) *image.NRGBA {
	return mapLUT(img, buildLUT(func(v float64) float64 { return 255 - v }))
}

func Grayscale(img image.Image) *image.NRGBA {
	return Map(img, func(c color.NRGBA) color.NRGBA {
		v := imgutil.Clamp8(imgutil.Luma(c) * 255)
		return color.NRGBA{v, v, v, c.A}
	})
}

func Threshold(img image.Image, level float64) *image.NRGBA {
	return Map(img, func(c color.NRGBA) color.NRGBA {
		if imgutil.Luma(c) >= level {
			return color.NRGBA{255, 255, 255, c.A}
		}
		return color.NRGBA{0, 0, 0, c.A}
	})
}

func Posterize(img image.Image, levels int) *image.NRGBA {
	n := float64(max(levels, 2) - 1)
	return mapLUT(img, buildLUT(func(v float64) float64 {
		return math.Round(v/255*n) / n * 255
	}))
}

func Brightness(img image.Image, amount float64) *image.NRGBA {
	return mapLUT(img, buildLUT(func(v float64) float64 { return v + amount*255 }))
}

func Contrast(img image.Image, factor float64) *image.NRGBA {
	return mapLUT(img, buildLUT(func(v float64) float64 { return (v-127.5)*factor + 127.5 }))
}

func Gamma(img image.Image, gamma float64) *image.NRGBA {
	if gamma <= 0 {
		gamma = 1
	}
	return mapLUT(img, buildLUT(func(v float64) float64 { return 255 * math.Pow(v/255, 1/gamma) }))
}

func Saturation(img image.Image, factor float64) *image.NRGBA {
	return Map(img, func(c color.NRGBA) color.NRGBA {
		l := imgutil.Luma(c) * 255
		s := func(v uint8) uint8 { return imgutil.Clamp8(l + (float64(v)-l)*factor) }
		return color.NRGBA{s(c.R), s(c.G), s(c.B), c.A}
	})
}

func HueShift(img image.Image, degrees float64) *image.NRGBA {
	shift := degrees / 360
	return Map(img, func(c color.NRGBA) color.NRGBA {
		h, s, l := imgutil.RGBToHSL(c.R, c.G, c.B)
		r, g, b := imgutil.HSLToRGB(h+shift, s, l)
		return color.NRGBA{r, g, b, c.A}
	})
}

func Solarize(img image.Image, threshold float64) *image.NRGBA {
	t := threshold * 255
	return mapLUT(img, buildLUT(func(v float64) float64 {
		if v >= t {
			return 255 - v
		}
		return v
	}))
}

func ChannelSwap(img image.Image, order string) (*image.NRGBA, error) {
	order = strings.ToLower(order)
	if len(order) != 3 {
		return nil, fmt.Errorf("channel order %q must have three letters", order)
	}
	var idx [3]int
	for i, ch := range order {
		switch ch {
		case 'r':
			idx[i] = 0
		case 'g':
			idx[i] = 1
		case 'b':
			idx[i] = 2
		default:
			return nil, fmt.Errorf("channel order %q may only contain r, g and b", order)
		}
	}
	return Map(img, func(c color.NRGBA) color.NRGBA {
		ch := [3]uint8{c.R, c.G, c.B}
		return color.NRGBA{ch[idx[0]], ch[idx[1]], ch[idx[2]], c.A}
	}), nil
}

func GradientMap(img image.Image, stops []color.Color) *image.NRGBA {
	if len(stops) == 0 {
		return imgutil.ToNRGBA(img)
	}
	cs := make([]color.NRGBA, len(stops))
	for i, s := range stops {
		cs[i] = color.NRGBAModel.Convert(s).(color.NRGBA)
	}
	var lut [256]color.NRGBA
	for i := range lut {
		t := float64(i) / 255 * float64(len(cs)-1)
		j := min(int(t), len(cs)-1)
		k := min(j+1, len(cs)-1)
		f := t - float64(j)
		mix := func(a, b uint8) uint8 { return imgutil.Clamp8(float64(a) + (float64(b)-float64(a))*f) }
		lut[i] = color.NRGBA{mix(cs[j].R, cs[k].R), mix(cs[j].G, cs[k].G), mix(cs[j].B, cs[k].B), 255}
	}
	return Map(img, func(c color.NRGBA) color.NRGBA {
		out := lut[imgutil.Clamp8(imgutil.Luma(c)*255)]
		out.A = c.A
		return out
	})
}

func ParseColor(s string) (color.NRGBA, error) {
	h := strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	if len(h) == 6 {
		h += "ff"
	}
	var c color.NRGBA
	if len(h) != 8 {
		return c, fmt.Errorf("invalid color %q", s)
	}
	if _, err := fmt.Sscanf(h, "%02x%02x%02x%02x", &c.R, &c.G, &c.B, &c.A); err != nil {
		return c, fmt.Errorf("invalid color %q", s)
	}
	return c, nil
}
