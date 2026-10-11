package effects

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/MaRcR11/sortcery/internal/imgutil"
)

func Halftone(img image.Image, size int, colored bool) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	size = max(size, 2)
	dst := image.NewNRGBA(src.Rect)
	maxR := float64(size) / math.Sqrt2
	cellsY := (h + size - 1) / size
	imgutil.Parallel(cellsY, func(c0, c1 int) {
		for cy := c0; cy < c1; cy++ {
			y0, y1 := cy*size, min(cy*size+size, h)
			for x0 := 0; x0 < w; x0 += size {
				x1 := min(x0+size, w)
				var sum [4]int
				for y := y0; y < y1; y++ {
					for x := x0; x < x1; x++ {
						for c := 0; c < 4; c++ {
							sum[c] += int(src.Pix[(y*w+x)*4+c])
						}
					}
				}
				n := (y1 - y0) * (x1 - x0)
				avg := color.NRGBA{uint8(sum[0] / n), uint8(sum[1] / n), uint8(sum[2] / n), uint8(sum[3] / n)}
				l := imgutil.Luma(avg)
				ink, paper := color.NRGBA{0, 0, 0, 255}, color.NRGBA{255, 255, 255, 255}
				radius := maxR * math.Sqrt(1-l)
				if colored {
					ink, paper = avg, color.NRGBA{0, 0, 0, 255}
					ink.A = 255
					radius = maxR * math.Sqrt(l)
				}
				mx, my := float64(x0)+float64(size)/2, float64(y0)+float64(size)/2
				for y := y0; y < y1; y++ {
					for x := x0; x < x1; x++ {
						d := math.Hypot(float64(x)+0.5-mx, float64(y)+0.5-my)
						cov := math.Max(0, math.Min(1, radius-d+0.5))
						mix := func(a, b uint8) uint8 { return imgutil.Clamp8(float64(a) + (float64(b)-float64(a))*cov) }
						imgutil.Set(dst.Pix, y*w+x, color.NRGBA{mix(paper.R, ink.R), mix(paper.G, ink.G), mix(paper.B, ink.B), avg.A})
					}
				}
			}
		}
	})
	return dst
}

func Bloom(img image.Image, threshold float64, radius int, intensity float64) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	bright := Map(src, func(c color.NRGBA) color.NRGBA {
		l := imgutil.Luma(c)
		if l <= threshold || threshold >= 1 {
			return color.NRGBA{0, 0, 0, c.A}
		}
		f := (l - threshold) / (1 - threshold)
		return color.NRGBA{imgutil.Clamp8(float64(c.R) * f), imgutil.Clamp8(float64(c.G) * f), imgutil.Clamp8(float64(c.B) * f), c.A}
	})
	r := max(radius/3, 1)
	for i := 0; i < 3; i++ {
		bright = BoxBlur(bright, r)
	}
	for i := range src.Pix {
		if i%4 == 3 {
			continue
		}
		a, b := float64(src.Pix[i])/255, math.Min(float64(bright.Pix[i])/255*intensity, 1)
		src.Pix[i] = imgutil.Clamp8((1 - (1-a)*(1-b)) * 255)
	}
	return src
}

func OilPaint(img image.Image, radius int) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	radius = max(radius, 1)
	stride := w + 1
	sat := make([][5]int64, stride*(h+1))
	for y := 0; y < h; y++ {
		var row [5]int64
		for x := 0; x < w; x++ {
			c := imgutil.At(src.Pix, y*w+x)
			l := int64(imgutil.Clamp8(imgutil.Luma(c) * 255))
			v := [5]int64{int64(c.R), int64(c.G), int64(c.B), l, l * l}
			for k := range row {
				row[k] += v[k]
				sat[(y+1)*stride+x+1][k] = sat[y*stride+x+1][k] + row[k]
			}
		}
	}
	sum := func(x0, y0, x1, y1 int) (s [5]int64, n int64) {
		x0, y0 = max(x0, 0), max(y0, 0)
		x1, y1 = min(x1, w-1)+1, min(y1, h-1)+1
		for k := range s {
			s[k] = sat[y1*stride+x1][k] - sat[y0*stride+x1][k] - sat[y1*stride+x0][k] + sat[y0*stride+x0][k]
		}
		return s, int64((x1 - x0) * (y1 - y0))
	}
	dst := image.NewNRGBA(src.Rect)
	imgutil.Parallel(h, func(ya, yb int) {
		for y := ya; y < yb; y++ {
			for x := 0; x < w; x++ {
				quads := [4][4]int{
					{x - radius, y - radius, x, y},
					{x, y - radius, x + radius, y},
					{x - radius, y, x, y + radius},
					{x, y, x + radius, y + radius},
				}
				bestVar := math.Inf(1)
				var best [5]int64
				var bestN int64 = 1
				for _, q := range quads {
					s, n := sum(q[0], q[1], q[2], q[3])
					mean := float64(s[3]) / float64(n)
					if v := float64(s[4])/float64(n) - mean*mean; v < bestVar {
						bestVar, best, bestN = v, s, n
					}
				}
				i := (y*w + x) * 4
				dst.Pix[i] = uint8(best[0] / bestN)
				dst.Pix[i+1] = uint8(best[1] / bestN)
				dst.Pix[i+2] = uint8(best[2] / bestN)
				dst.Pix[i+3] = src.Pix[i+3]
			}
		}
	})
	return dst
}

func BlockShuffle(img image.Image, size int, amount float64, seed int64) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	dst := imgutil.ToNRGBA(src)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	size = max(size, 1)
	bw, bh := w/size, h/size
	if bw == 0 || bh == 0 {
		return dst
	}
	rng := rand.New(rand.NewSource(seed))
	for by := 0; by < bh; by++ {
		for bx := 0; bx < bw; bx++ {
			if rng.Float64() >= amount {
				continue
			}
			sx, sy := rng.Intn(bw)*size, rng.Intn(bh)*size
			for y := 0; y < size; y++ {
				di := ((by*size+y)*w + bx*size) * 4
				si := ((sy+y)*w + sx) * 4
				copy(dst.Pix[di:di+size*4], src.Pix[si:si+size*4])
			}
		}
	}
	return dst
}
