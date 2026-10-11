package effects

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/MaRcR11/gopxsort/internal/imgutil"
)

func Pixelate(img image.Image, size int) *image.NRGBA {
	dst := imgutil.ToNRGBA(img)
	size = max(size, 1)
	w, h := dst.Rect.Dx(), dst.Rect.Dy()
	blocksY := (h + size - 1) / size
	imgutil.Parallel(blocksY, func(b0, b1 int) {
		for by := b0; by < b1; by++ {
			y0, y1 := by*size, min(by*size+size, h)
			for x0 := 0; x0 < w; x0 += size {
				x1 := min(x0+size, w)
				var sum [4]int
				for y := y0; y < y1; y++ {
					for x := x0; x < x1; x++ {
						for c := 0; c < 4; c++ {
							sum[c] += int(dst.Pix[(y*w+x)*4+c])
						}
					}
				}
				n := (y1 - y0) * (x1 - x0)
				avg := color.NRGBA{uint8(sum[0] / n), uint8(sum[1] / n), uint8(sum[2] / n), uint8(sum[3] / n)}
				for y := y0; y < y1; y++ {
					for x := x0; x < x1; x++ {
						imgutil.Set(dst.Pix, y*w+x, avg)
					}
				}
			}
		}
	})
	return dst
}

type MirrorMode int

const (
	MirrorLeft MirrorMode = iota
	MirrorRight
	MirrorTop
	MirrorBottom
	MirrorQuad
)

var mirrorNames = map[string]MirrorMode{
	"left": MirrorLeft, "right": MirrorRight, "top": MirrorTop, "bottom": MirrorBottom, "quad": MirrorQuad,
}

func MirrorByName(name string) (MirrorMode, bool) {
	m, ok := mirrorNames[strings.ToLower(name)]
	return m, ok
}

func MirrorNames() []string { return sortedKeys(mirrorNames) }

func Mirror(img image.Image, mode MirrorMode) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	return remap(src, func(x, y int) (int, int) {
		switch mode {
		case MirrorLeft:
			if x >= w/2 {
				x = w - 1 - x
			}
		case MirrorRight:
			if x < w/2 {
				x = w - 1 - x
			}
		case MirrorTop:
			if y >= h/2 {
				y = h - 1 - y
			}
		case MirrorBottom:
			if y < h/2 {
				y = h - 1 - y
			}
		case MirrorQuad:
			if x >= w/2 {
				x = w - 1 - x
			}
			if y >= h/2 {
				y = h - 1 - y
			}
		}
		return x, y
	})
}

func FlipHorizontal(img image.Image) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	w := src.Rect.Dx()
	return remap(src, func(x, y int) (int, int) { return w - 1 - x, y })
}

func FlipVertical(img image.Image) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	h := src.Rect.Dy()
	return remap(src, func(x, y int) (int, int) { return x, h - 1 - y })
}

func Rotate(img image.Image, degrees int) (*image.NRGBA, error) {
	if degrees%90 != 0 {
		return nil, fmt.Errorf("rotation must be a multiple of 90 degrees, got %d", degrees)
	}
	src := imgutil.ToNRGBA(img)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	switch imgutil.Wrap(degrees, 360) {
	case 90:
		return remapSize(src, h, w, func(x, y int) (int, int) { return y, h - 1 - x }), nil
	case 180:
		return remap(src, func(x, y int) (int, int) { return w - 1 - x, h - 1 - y }), nil
	case 270:
		return remapSize(src, h, w, func(x, y int) (int, int) { return w - 1 - y, x }), nil
	}
	return src, nil
}

func remap(src *image.NRGBA, from func(x, y int) (int, int)) *image.NRGBA {
	return remapSize(src, src.Rect.Dx(), src.Rect.Dy(), from)
}

func remapSize(src *image.NRGBA, w, h int, from func(x, y int) (int, int)) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	sw := src.Rect.Dx()
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				sx, sy := from(x, y)
				copy(dst.Pix[(y*w+x)*4:(y*w+x)*4+4], src.Pix[(sy*sw+sx)*4:])
			}
		}
	})
	return dst
}

func BoxBlur(img image.Image, radius int) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	if radius <= 0 {
		return src
	}
	w, h := src.Rect.Dx(), src.Rect.Dy()
	tmp := image.NewNRGBA(src.Rect)
	dst := image.NewNRGBA(src.Rect)

	blur1D(src.Pix, tmp.Pix, w, h, radius, 4, w*4)
	blur1D(tmp.Pix, dst.Pix, h, w, radius, w*4, 4)
	return dst
}

func blur1D(src, dst []uint8, n, m, radius, step, lineStep int) {
	imgutil.Parallel(m, func(l0, l1 int) {
		for l := l0; l < l1; l++ {
			base := l * lineStep
			for c := 0; c < 4; c++ {
				sum := 0
				for k := -radius; k <= radius; k++ {
					sum += int(src[base+imgutil.Clamp(k, 0, n-1)*step+c])
				}
				for i := 0; i < n; i++ {
					dst[base+i*step+c] = uint8((sum + radius) / (2*radius + 1))
					out := imgutil.Clamp(i-radius, 0, n-1)
					in := imgutil.Clamp(i+radius+1, 0, n-1)
					sum += int(src[base+in*step+c]) - int(src[base+out*step+c])
				}
			}
		}
	})
}

func Sharpen(img image.Image, amount float64, radius int) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	blur := BoxBlur(src, max(radius, 1))
	for i := range src.Pix {
		if i%4 == 3 {
			continue
		}
		v := float64(src.Pix[i])
		src.Pix[i] = imgutil.Clamp8(v + (v-float64(blur.Pix[i]))*amount)
	}
	return src
}

func EdgeDetect(img image.Image) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	mag := imgutil.Sobel(src)
	for i, m := range mag {
		v := imgutil.Clamp8(m * 255)
		imgutil.Set(src.Pix, i, color.NRGBA{v, v, v, src.Pix[i*4+3]})
	}
	return src
}

func Vignette(img image.Image, strength float64) *image.NRGBA {
	dst := imgutil.ToNRGBA(img)
	w, h := dst.Rect.Dx(), dst.Rect.Dy()
	cx, cy := float64(w-1)/2, float64(h-1)/2
	maxD := math.Hypot(cx, cy)
	if maxD == 0 {
		return dst
	}
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				d := math.Hypot(float64(x)-cx, float64(y)-cy) / maxD
				f := 1 - strength*d*d
				p := dst.Pix[(y*w+x)*4 : (y*w+x)*4+3]
				for j := range p {
					p[j] = imgutil.Clamp8(float64(p[j]) * f)
				}
			}
		}
	})
	return dst
}
