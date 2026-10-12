package effects

import (
	"image"
	"math"

	"github.com/MaRcR11/sortmancy/internal/imgutil"
)

func GaussianBlur(img image.Image, sigma float64) *image.NRGBA {
	out := imgutil.ToNRGBA(img)
	if sigma <= 0 {
		return out
	}
	const passes = 3
	ideal := math.Sqrt(12*sigma*sigma/passes + 1)
	lower := int(math.Floor(ideal))
	if lower%2 == 0 {
		lower--
	}
	upper := lower + 2
	m := int(math.Round((12*sigma*sigma - passes*float64(lower*lower) - 4*passes*float64(lower) - 3*passes) / (-4*float64(lower) - 4)))
	for i := 0; i < passes; i++ {
		size := upper
		if i < m {
			size = lower
		}
		out = BoxBlur(out, max((size-1)/2, 0))
	}
	return out
}

func MotionBlur(img image.Image, angle float64, length int) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	if length < 2 {
		return src
	}
	w, h := src.Rect.Dx(), src.Rect.Dy()
	rad := angle * math.Pi / 180
	ux, uy := math.Cos(rad), -math.Sin(rad)
	offs := make([][2]int, length)
	for i := range offs {
		t := float64(i) - float64(length-1)/2
		offs[i] = [2]int{int(math.Round(t * ux)), int(math.Round(t * uy))}
	}
	dst := image.NewNRGBA(src.Rect)
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				var sum [4]int
				for _, o := range offs {
					j := (imgutil.Clamp(y+o[1], 0, h-1)*w + imgutil.Clamp(x+o[0], 0, w-1)) * 4
					for c := 0; c < 4; c++ {
						sum[c] += int(src.Pix[j+c])
					}
				}
				i := (y*w + x) * 4
				for c := 0; c < 4; c++ {
					dst.Pix[i+c] = uint8(sum[c] / length)
				}
			}
		}
	})
	return dst
}

func ZoomBlur(img image.Image, strength float64, samples int) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	samples = max(samples, 2)
	cx, cy := float64(w-1)/2, float64(h-1)/2
	dst := image.NewNRGBA(src.Rect)
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				var sum [4]int
				for k := 0; k < samples; k++ {
					s := 1 - strength*float64(k)/float64(samples-1)
					sx := imgutil.Clamp(int(math.Round(cx+(float64(x)-cx)*s)), 0, w-1)
					sy := imgutil.Clamp(int(math.Round(cy+(float64(y)-cy)*s)), 0, h-1)
					j := (sy*w + sx) * 4
					for c := 0; c < 4; c++ {
						sum[c] += int(src.Pix[j+c])
					}
				}
				i := (y*w + x) * 4
				for c := 0; c < 4; c++ {
					dst.Pix[i+c] = uint8(sum[c] / samples)
				}
			}
		}
	})
	return dst
}

func Median(img image.Image, radius int) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	if radius <= 0 {
		return src
	}
	w, h := src.Rect.Dx(), src.Rect.Dy()
	dst := image.NewNRGBA(src.Rect)
	imgutil.Parallel(h, func(y0, y1 int) {
		var hist [3][256]int
		for y := y0; y < y1; y++ {
			for c := range hist {
				hist[c] = [256]int{}
			}
			count := 0
			add := func(x, sign int) {
				if x < 0 || x >= w {
					return
				}
				for yy := max(y-radius, 0); yy <= min(y+radius, h-1); yy++ {
					j := (yy*w + x) * 4
					for c := 0; c < 3; c++ {
						hist[c][src.Pix[j+c]] += sign
					}
					count += sign
				}
			}
			for x := -radius; x <= radius; x++ {
				add(x, 1)
			}
			for x := 0; x < w; x++ {
				i := (y*w + x) * 4
				half := count / 2
				for c := 0; c < 3; c++ {
					seen := 0
					for v := 0; v < 256; v++ {
						if seen += hist[c][v]; seen > half {
							dst.Pix[i+c] = uint8(v)
							break
						}
					}
				}
				dst.Pix[i+3] = src.Pix[i+3]
				add(x-radius, -1)
				add(x+radius+1, 1)
			}
		}
	})
	return dst
}

func morph(img image.Image, radius int, better func(a, b uint8) bool) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	if radius <= 0 {
		return src
	}
	w, h := src.Rect.Dx(), src.Rect.Dy()
	pass := func(in *image.NRGBA, horizontal bool) *image.NRGBA {
		out := image.NewNRGBA(in.Rect)
		imgutil.Parallel(h, func(y0, y1 int) {
			for y := y0; y < y1; y++ {
				for x := 0; x < w; x++ {
					i := (y*w + x) * 4
					for c := 0; c < 3; c++ {
						best := in.Pix[i+c]
						for k := -radius; k <= radius; k++ {
							j := i
							if horizontal {
								j = (y*w + imgutil.Clamp(x+k, 0, w-1)) * 4
							} else {
								j = (imgutil.Clamp(y+k, 0, h-1)*w + x) * 4
							}
							if v := in.Pix[j+c]; better(v, best) {
								best = v
							}
						}
						out.Pix[i+c] = best
					}
					out.Pix[i+3] = in.Pix[i+3]
				}
			}
		})
		return out
	}
	return pass(pass(src, true), false)
}

func Dilate(img image.Image, radius int) *image.NRGBA {
	return morph(img, radius, func(a, b uint8) bool { return a > b })
}

func Erode(img image.Image, radius int) *image.NRGBA {
	return morph(img, radius, func(a, b uint8) bool { return a < b })
}
