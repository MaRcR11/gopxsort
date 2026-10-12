package effects

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/MaRcR11/sortmancy/internal/imgutil"
)

func AutoLevels(img image.Image, clip float64) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	n := src.Rect.Dx() * src.Rect.Dy()
	if n == 0 {
		return src
	}
	var hist [3][256]int
	for i := 0; i < n; i++ {
		for c := 0; c < 3; c++ {
			hist[c][src.Pix[i*4+c]]++
		}
	}
	limit := int(math.Max(0, math.Min(0.49, clip)) * float64(n))
	var luts [3][256]uint8
	for c := 0; c < 3; c++ {
		lo, hi, sum := 0, 255, 0
		for v := 0; v < 256; v++ {
			if sum += hist[c][v]; sum > limit {
				lo = v
				break
			}
		}
		sum = 0
		for v := 255; v >= 0; v-- {
			if sum += hist[c][v]; sum > limit {
				hi = v
				break
			}
		}
		for v := range luts[c] {
			if hi <= lo {
				luts[c][v] = uint8(v)
				continue
			}
			luts[c][v] = imgutil.Clamp8(float64(v-lo) / float64(hi-lo) * 255)
		}
	}
	return Map(src, func(p color.NRGBA) color.NRGBA {
		return color.NRGBA{luts[0][p.R], luts[1][p.G], luts[2][p.B], p.A}
	})
}

func Equalize(img image.Image) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	n := src.Rect.Dx() * src.Rect.Dy()
	if n == 0 {
		return src
	}
	var hist [256]int
	for i := 0; i < n; i++ {
		hist[imgutil.Clamp8(imgutil.Luma(imgutil.At(src.Pix, i))*255)]++
	}
	var cdf [256]float64
	sum, first := 0, -1
	for v := range hist {
		sum += hist[v]
		if first < 0 && hist[v] > 0 {
			first = sum
		}
		cdf[v] = float64(sum)
	}
	var target [256]float64
	for v := range target {
		if n == first {
			target[v] = float64(v)
			continue
		}
		target[v] = math.Max(0, (cdf[v]-float64(first))/float64(n-first)) * 255
	}
	return Map(src, func(p color.NRGBA) color.NRGBA {
		l := imgutil.Luma(p) * 255
		nl := target[imgutil.Clamp8(l)]
		if l < 1 {
			v := imgutil.Clamp8(nl)
			return color.NRGBA{v, v, v, p.A}
		}
		f := nl / l
		return color.NRGBA{imgutil.Clamp8(float64(p.R) * f), imgutil.Clamp8(float64(p.G) * f), imgutil.Clamp8(float64(p.B) * f), p.A}
	})
}

func Levels(img image.Image, black, white, gamma float64) *image.NRGBA {
	if white <= black {
		white = black + 1.0/255
	}
	if gamma <= 0 {
		gamma = 1
	}
	return mapLUT(img, buildLUT(func(v float64) float64 {
		t := math.Max(0, math.Min(1, (v/255-black)/(white-black)))
		return math.Pow(t, 1/gamma) * 255
	}))
}

type CurvePoint struct{ X, Y float64 }

func ParseCurve(s string) ([]CurvePoint, error) {
	var pts []CurvePoint
	for _, part := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' }) {
		xs, ys, ok := strings.Cut(part, ":")
		x, errX := strconv.ParseFloat(xs, 64)
		y, errY := strconv.ParseFloat(ys, 64)
		if !ok || errX != nil || errY != nil {
			return nil, fmt.Errorf("curve point %q must be x:y", part)
		}
		pts = append(pts, CurvePoint{x, y})
	}
	if len(pts) < 2 {
		return nil, fmt.Errorf("a curve needs at least two points")
	}
	return pts, nil
}

func curveLUT(points []CurvePoint) *[256]uint8 {
	pts := append([]CurvePoint(nil), points...)
	sort.Slice(pts, func(i, j int) bool { return pts[i].X < pts[j].X })
	n := len(pts)
	slope := make([]float64, n-1)
	for i := range slope {
		dx := pts[i+1].X - pts[i].X
		if dx <= 0 {
			dx = 1e-9
		}
		slope[i] = (pts[i+1].Y - pts[i].Y) / dx
	}
	tan := make([]float64, n)
	tan[0], tan[n-1] = slope[0], slope[n-2]
	for i := 1; i < n-1; i++ {
		if slope[i-1]*slope[i] <= 0 {
			tan[i] = 0
		} else {
			tan[i] = (slope[i-1] + slope[i]) / 2
		}
	}
	for i := range slope {
		if slope[i] == 0 {
			tan[i], tan[i+1] = 0, 0
			continue
		}
		a, b := tan[i]/slope[i], tan[i+1]/slope[i]
		if s := a*a + b*b; s > 9 {
			t := 3 / math.Sqrt(s)
			tan[i], tan[i+1] = t*a*slope[i], t*b*slope[i]
		}
	}
	eval := func(x float64) float64 {
		if x <= pts[0].X {
			return pts[0].Y
		}
		if x >= pts[n-1].X {
			return pts[n-1].Y
		}
		k := sort.Search(n, func(i int) bool { return pts[i].X > x }) - 1
		h := pts[k+1].X - pts[k].X
		t := (x - pts[k].X) / h
		t2, t3 := t*t, t*t*t
		return (2*t3-3*t2+1)*pts[k].Y + (t3-2*t2+t)*h*tan[k] + (-2*t3+3*t2)*pts[k+1].Y + (t3-t2)*h*tan[k+1]
	}
	return buildLUT(func(v float64) float64 { return eval(v/255) * 255 })
}

func Curves(img image.Image, points []CurvePoint) *image.NRGBA {
	if len(points) < 2 {
		return imgutil.ToNRGBA(img)
	}
	return mapLUT(img, curveLUT(points))
}

func Temperature(img image.Image, warmth, tint float64) *image.NRGBA {
	r, g, b := 1+0.25*warmth, 1+0.2*tint, 1-0.25*warmth
	return Map(img, func(p color.NRGBA) color.NRGBA {
		return color.NRGBA{imgutil.Clamp8(float64(p.R) * r), imgutil.Clamp8(float64(p.G) * g), imgutil.Clamp8(float64(p.B) * b), p.A}
	})
}

func Sepia(img image.Image, amount float64) *image.NRGBA {
	return Map(img, func(p color.NRGBA) color.NRGBA {
		r, g, b := float64(p.R), float64(p.G), float64(p.B)
		sr := 0.393*r + 0.769*g + 0.189*b
		sg := 0.349*r + 0.686*g + 0.168*b
		sb := 0.272*r + 0.534*g + 0.131*b
		mix := func(a, s float64) uint8 { return imgutil.Clamp8(a + (s-a)*amount) }
		return color.NRGBA{mix(r, sr), mix(g, sg), mix(b, sb), p.A}
	})
}

func ChannelMix(img image.Image, m [3][3]float64) *image.NRGBA {
	return Map(img, func(p color.NRGBA) color.NRGBA {
		in := [3]float64{float64(p.R), float64(p.G), float64(p.B)}
		var out [3]uint8
		for i := range out {
			out[i] = imgutil.Clamp8(m[i][0]*in[0] + m[i][1]*in[1] + m[i][2]*in[2])
		}
		return color.NRGBA{out[0], out[1], out[2], p.A}
	})
}

func ParseMatrix(s string) ([3][3]float64, error) {
	var m [3][3]float64
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' })
	if len(parts) != 9 {
		return m, fmt.Errorf("channel matrix needs 9 numbers (rows for red, green, blue), got %d", len(parts))
	}
	for i, p := range parts {
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return m, fmt.Errorf("channel matrix: %q is not a number", p)
		}
		m[i/3][i%3] = v
	}
	return m, nil
}
