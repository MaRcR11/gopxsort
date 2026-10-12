package effects

import (
	"image"
	"image/color"
	"math"

	"github.com/MaRcR11/sortmancy/internal/imgutil"
)

func Toon(img image.Image, levels int, edge float64) *image.NRGBA {
	smoothed := BoxBlur(img, 1)
	n := float64(max(levels, 2) - 1)
	flat := Map(OilPaint(smoothed, 3), func(c color.NRGBA) color.NRGBA {
		l := imgutil.Luma(c)
		q := math.Min(1, math.Round(l*n+0.25)/n)
		if l < 1e-3 {
			v := imgutil.Clamp8(q * 255)
			return color.NRGBA{v, v, v, c.A}
		}
		f := q / l
		return color.NRGBA{imgutil.Clamp8(float64(c.R) * f), imgutil.Clamp8(float64(c.G) * f), imgutil.Clamp8(float64(c.B) * f), c.A}
	})
	mag := imgutil.Sobel(smoothed)
	for i, m := range mag {
		if m > edge {
			imgutil.Set(flat.Pix, i, color.NRGBA{0, 0, 0, flat.Pix[i*4+3]})
		}
	}
	return flat
}

func Neon(img image.Image, edge float64, glow int) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	mag := imgutil.Sobel(src)
	lines := image.NewNRGBA(src.Rect)
	edge = math.Max(edge, 1e-3)
	for i, m := range mag {
		c := imgutil.At(src.Pix, i)
		h, s, _ := imgutil.RGBToHSL(c.R, c.G, c.B)
		r, g, b := imgutil.HSLToRGB(h, math.Max(s, 0.6)*1.2, 0.55)
		f := math.Min(1, m/edge)
		f *= f
		imgutil.Set(lines.Pix, i, color.NRGBA{imgutil.Clamp8(float64(r) * f), imgutil.Clamp8(float64(g) * f), imgutil.Clamp8(float64(b) * f), 255})
	}
	halo := lines
	for k := 0; k < 3; k++ {
		halo = BoxBlur(halo, max(glow/3, 1))
	}
	return Blend(Blend(lines, halo, BlendScreen, 1), halo, BlendScreen, 1)
}

func Emboss(img image.Image, strength, angle float64) *image.NRGBA {
	src := BoxBlur(img, 1)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	rad := angle * math.Pi / 180
	dx, dy := int(math.Round(math.Cos(rad))), -int(math.Round(math.Sin(rad)))
	lum := func(x, y int) float64 {
		return imgutil.Luma(imgutil.At(src.Pix, imgutil.Clamp(y, 0, h-1)*w+imgutil.Clamp(x, 0, w-1)))
	}
	dst := image.NewNRGBA(src.Rect)
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				v := imgutil.Clamp8(128 + strength*255*(lum(x+dx, y+dy)-lum(x-dx, y-dy)))
				imgutil.Set(dst.Pix, y*w+x, color.NRGBA{v, v, v, src.Pix[(y*w+x)*4+3]})
			}
		}
	})
	return dst
}
