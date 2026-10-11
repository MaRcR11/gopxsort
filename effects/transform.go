package effects

import (
	"image"
	"math"
	"strings"

	"github.com/MaRcR11/sortcery/internal/imgutil"
)

type Filter int

const (
	Nearest Filter = iota
	Bilinear
)

func FilterByName(name string) (Filter, bool) {
	switch strings.ToLower(name) {
	case "nearest":
		return Nearest, true
	case "bilinear", "linear", "smooth":
		return Bilinear, true
	}
	return 0, false
}

func FilterNames() []string { return []string{"bilinear", "nearest"} }

func Resize(img image.Image, w, h int, filter Filter) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	sw, sh := src.Rect.Dx(), src.Rect.Dy()
	w, h = max(w, 1), max(h, 1)
	if sw == 0 || sh == 0 {
		return image.NewNRGBA(image.Rect(0, 0, w, h))
	}
	if w == sw && h == sh {
		return src
	}
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	fx, fy := float64(sw)/float64(w), float64(sh)/float64(h)
	if filter == Nearest {
		imgutil.Parallel(h, func(y0, y1 int) {
			for y := y0; y < y1; y++ {
				sy := min(int((float64(y)+0.5)*fy), sh-1)
				for x := 0; x < w; x++ {
					sx := min(int((float64(x)+0.5)*fx), sw-1)
					copy(dst.Pix[(y*w+x)*4:(y*w+x)*4+4], src.Pix[(sy*sw+sx)*4:])
				}
			}
		})
		return dst
	}
	if r := int(math.Max(fx, fy) / 2); r >= 1 {
		src = BoxBlur(src, r)
	}
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			syf := math.Max((float64(y)+0.5)*fy-0.5, 0)
			sy0 := min(int(syf), sh-1)
			sy1 := min(sy0+1, sh-1)
			ty := syf - float64(sy0)
			for x := 0; x < w; x++ {
				sxf := math.Max((float64(x)+0.5)*fx-0.5, 0)
				sx0 := min(int(sxf), sw-1)
				sx1 := min(sx0+1, sw-1)
				tx := sxf - float64(sx0)
				for c := 0; c < 4; c++ {
					p := func(px, py int) float64 { return float64(src.Pix[(py*sw+px)*4+c]) }
					top := p(sx0, sy0) + (p(sx1, sy0)-p(sx0, sy0))*tx
					bot := p(sx0, sy1) + (p(sx1, sy1)-p(sx0, sy1))*tx
					dst.Pix[(y*w+x)*4+c] = imgutil.Clamp8(top + (bot-top)*ty)
				}
			}
		}
	})
	return dst
}

func Scale(img image.Image, factor float64, filter Filter) *image.NRGBA {
	b := img.Bounds()
	return Resize(img, int(math.Round(float64(b.Dx())*factor)), int(math.Round(float64(b.Dy())*factor)), filter)
}

func Crop(img image.Image, r image.Rectangle) *image.NRGBA {
	b := img.Bounds()
	r = r.Add(b.Min).Intersect(b)
	if r.Empty() {
		return image.NewNRGBA(image.Rect(0, 0, 0, 0))
	}
	return imgutil.ToNRGBA(subImage(img, r))
}

func subImage(img image.Image, r image.Rectangle) image.Image {
	if s, ok := img.(interface {
		SubImage(image.Rectangle) image.Image
	}); ok {
		return s.SubImage(r)
	}
	dst := image.NewNRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			dst.Set(x, y, img.At(r.Min.X+x, r.Min.Y+y))
		}
	}
	return dst
}

func Mix(base, over image.Image, amount float64) *image.NRGBA {
	dst := imgutil.ToNRGBA(base)
	w, h := dst.Rect.Dx(), dst.Rect.Dy()
	o := imgutil.ToNRGBA(over)
	if o.Rect.Dx() != w || o.Rect.Dy() != h {
		o = Resize(o, w, h, Bilinear)
	}
	imgutil.Parallel(h, func(y0, y1 int) {
		for i := y0 * w * 4; i < y1*w*4; i++ {
			a, b := float64(dst.Pix[i]), float64(o.Pix[i])
			dst.Pix[i] = imgutil.Clamp8(a + (b-a)*amount)
		}
	})
	return dst
}

func Kaleidoscope(img image.Image, segments int, rotation float64) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	segments = max(segments, 1)
	seg := 2 * math.Pi / float64(segments)
	rot := rotation * math.Pi / 180
	cx, cy := float64(w-1)/2, float64(h-1)/2
	reflect := func(v, n int) int {
		if n == 1 {
			return 0
		}
		period := 2 * (n - 1)
		v = imgutil.Wrap(v, period)
		if v >= n {
			v = period - v
		}
		return v
	}
	dst := image.NewNRGBA(src.Rect)
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				dx, dy := float64(x)-cx, float64(y)-cy
				r := math.Hypot(dx, dy)
				a := math.Atan2(dy, dx) - rot
				a -= math.Floor(a/(2*math.Pi)) * 2 * math.Pi
				k := int(a / seg)
				a -= float64(k) * seg
				if k%2 == 1 {
					a = seg - a
				}
				a += rot
				sx := reflect(int(math.Round(cx+r*math.Cos(a))), w)
				sy := reflect(int(math.Round(cy+r*math.Sin(a))), h)
				copy(dst.Pix[(y*w+x)*4:(y*w+x)*4+4], src.Pix[(sy*w+sx)*4:])
			}
		}
	})
	return dst
}
