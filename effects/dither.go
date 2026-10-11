package effects

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"
	"strings"

	"github.com/MaRcR11/gopxsort/internal/imgutil"
)

type DitherMethod int

const (
	DitherNone DitherMethod = iota
	FloydSteinberg
	Atkinson
	JarvisJudiceNinke
	Stucki
	Burkes
	Sierra
	Bayer2
	Bayer4
	Bayer8
)

var ditherNames = map[string]DitherMethod{
	"none":           DitherNone,
	"floyd":          FloydSteinberg,
	"floydsteinberg": FloydSteinberg,
	"atkinson":       Atkinson,
	"jarvis":         JarvisJudiceNinke,
	"stucki":         Stucki,
	"burkes":         Burkes,
	"sierra":         Sierra,
	"bayer2":         Bayer2,
	"bayer4":         Bayer4,
	"bayer8":         Bayer8,
}

func DitherByName(name string) (DitherMethod, bool) {
	m, ok := ditherNames[strings.ToLower(name)]
	return m, ok
}

func DitherNames() []string { return sortedKeys(ditherNames) }

type DitherOptions struct {
	Method  DitherMethod
	Palette color.Palette
	Levels  int
}

type weight struct {
	dx, dy int
	w      float64
}

var kernels = map[DitherMethod][]weight{
	FloydSteinberg: norm(16, weight{1, 0, 7}, weight{-1, 1, 3}, weight{0, 1, 5}, weight{1, 1, 1}),
	Atkinson:       norm(8, weight{1, 0, 1}, weight{2, 0, 1}, weight{-1, 1, 1}, weight{0, 1, 1}, weight{1, 1, 1}, weight{0, 2, 1}),
	JarvisJudiceNinke: norm(48,
		weight{1, 0, 7}, weight{2, 0, 5},
		weight{-2, 1, 3}, weight{-1, 1, 5}, weight{0, 1, 7}, weight{1, 1, 5}, weight{2, 1, 3},
		weight{-2, 2, 1}, weight{-1, 2, 3}, weight{0, 2, 5}, weight{1, 2, 3}, weight{2, 2, 1}),
	Stucki: norm(42,
		weight{1, 0, 8}, weight{2, 0, 4},
		weight{-2, 1, 2}, weight{-1, 1, 4}, weight{0, 1, 8}, weight{1, 1, 4}, weight{2, 1, 2},
		weight{-2, 2, 1}, weight{-1, 2, 2}, weight{0, 2, 4}, weight{1, 2, 2}, weight{2, 2, 1}),
	Burkes: norm(32,
		weight{1, 0, 8}, weight{2, 0, 4},
		weight{-2, 1, 2}, weight{-1, 1, 4}, weight{0, 1, 8}, weight{1, 1, 4}, weight{2, 1, 2}),
	Sierra: norm(32,
		weight{1, 0, 5}, weight{2, 0, 3},
		weight{-2, 1, 2}, weight{-1, 1, 4}, weight{0, 1, 5}, weight{1, 1, 4}, weight{2, 1, 2},
		weight{-1, 2, 2}, weight{0, 2, 3}, weight{1, 2, 2}),
}

func norm(div float64, ws ...weight) []weight {
	for i := range ws {
		ws[i].w /= div
	}
	return ws
}

func Dither(img image.Image, o DitherOptions) *image.NRGBA {
	dst := imgutil.ToNRGBA(img)
	w, h := dst.Rect.Dx(), dst.Rect.Dy()
	quant := quantizer(o)

	switch o.Method {
	case Bayer2, Bayer4, Bayer8:
		m := bayer(map[DitherMethod]int{Bayer2: 2, Bayer4: 4, Bayer8: 8}[o.Method])
		n := len(m)
		spread := ditherSpread(o)
		imgutil.Parallel(h, func(y0, y1 int) {
			for y := y0; y < y1; y++ {
				for x := 0; x < w; x++ {
					i := y*w + x
					c := imgutil.At(dst.Pix, i)
					t := m[y%n][x%n] * spread
					out := quant(float64(c.R)+t, float64(c.G)+t, float64(c.B)+t)
					out.A = c.A
					imgutil.Set(dst.Pix, i, out)
				}
			}
		})

	case DitherNone:
		imgutil.Parallel(h, func(y0, y1 int) {
			for i := y0 * w; i < y1*w; i++ {
				c := imgutil.At(dst.Pix, i)
				out := quant(float64(c.R), float64(c.G), float64(c.B))
				out.A = c.A
				imgutil.Set(dst.Pix, i, out)
			}
		})

	default:
		kernel := kernels[o.Method]
		if kernel == nil {
			kernel = kernels[FloydSteinberg]
		}

		buf := make([]float64, w*h*3)
		for i := 0; i < w*h; i++ {
			buf[i*3], buf[i*3+1], buf[i*3+2] = float64(dst.Pix[i*4]), float64(dst.Pix[i*4+1]), float64(dst.Pix[i*4+2])
		}
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				i := y*w + x
				r, g, b := buf[i*3], buf[i*3+1], buf[i*3+2]
				out := quant(r, g, b)
				out.A = dst.Pix[i*4+3]
				imgutil.Set(dst.Pix, i, out)
				er, eg, eb := r-float64(out.R), g-float64(out.G), b-float64(out.B)
				for _, k := range kernel {
					nx, ny := x+k.dx, y+k.dy
					if nx < 0 || nx >= w || ny >= h {
						continue
					}
					j := (ny*w + nx) * 3
					buf[j] += er * k.w
					buf[j+1] += eg * k.w
					buf[j+2] += eb * k.w
				}
			}
		}
	}
	return dst
}

func quantizer(o DitherOptions) func(r, g, b float64) color.NRGBA {
	if len(o.Palette) > 0 {
		pal := make([]color.NRGBA, len(o.Palette))
		for i, c := range o.Palette {
			pal[i] = color.NRGBAModel.Convert(c).(color.NRGBA)
		}
		return func(r, g, b float64) color.NRGBA {
			best, bestD := pal[0], math.Inf(1)
			for _, p := range pal {
				dr, dg, db := r-float64(p.R), g-float64(p.G), b-float64(p.B)

				if d := 2*dr*dr + 4*dg*dg + 3*db*db; d < bestD {
					best, bestD = p, d
				}
			}
			return best
		}
	}
	n := float64(max(o.Levels, 2) - 1)
	q := func(v float64) uint8 {
		return imgutil.Clamp8(math.Round(math.Max(0, math.Min(255, v))/255*n) / n * 255)
	}
	return func(r, g, b float64) color.NRGBA { return color.NRGBA{q(r), q(g), q(b), 255} }
}

func ditherSpread(o DitherOptions) float64 {
	if len(o.Palette) < 2 {
		return 255 / float64(max(o.Levels, 2)-1)
	}
	pal := make([]color.NRGBA, len(o.Palette))
	for i, c := range o.Palette {
		pal[i] = color.NRGBAModel.Convert(c).(color.NRGBA)
	}
	sum := 0.0
	for i, a := range pal {
		nearest := math.Inf(1)
		for j, b := range pal {
			if i != j {
				dr, dg, db := float64(a.R)-float64(b.R), float64(a.G)-float64(b.G), float64(a.B)-float64(b.B)
				nearest = math.Min(nearest, math.Sqrt(dr*dr+dg*dg+db*db))
			}
		}
		sum += nearest
	}
	return math.Min(sum/float64(len(pal)), 255)
}

func bayer(n int) [][]float64 {
	m := [][]int{{0}}
	for size := 1; size < n; size *= 2 {
		next := make([][]int, size*2)
		for y := range next {
			next[y] = make([]int, size*2)
			for x := range next[y] {
				v := 4 * m[y%size][x%size]
				switch {
				case y < size && x >= size:
					v += 2
				case y >= size && x < size:
					v += 3
				case y >= size && x >= size:
					v++
				}
				next[y][x] = v
			}
		}
		m = next
	}
	out := make([][]float64, n)
	for y := range out {
		out[y] = make([]float64, n)
		for x := range out[y] {
			out[y][x] = (float64(m[y][x])+0.5)/float64(n*n) - 0.5
		}
	}
	return out
}

var Palettes = map[string]color.Palette{
	"bw":      hexPalette("000000", "ffffff"),
	"gray4":   hexPalette("000000", "555555", "aaaaaa", "ffffff"),
	"gameboy": hexPalette("0f380f", "306230", "8bac0f", "9bbc0f"),
	"cga":     hexPalette("000000", "55ffff", "ff55ff", "ffffff"),
	"pico8": hexPalette("000000", "1d2b53", "7e2553", "008751", "ab5236", "5f574f", "c2c3c7", "fff1e8",
		"ff004d", "ffa300", "ffec27", "00e436", "29adff", "83769c", "ff77a8", "ffccaa"),
	"c64": hexPalette("000000", "ffffff", "880000", "aaffee", "cc44cc", "00cc55", "0000aa", "eeee77",
		"dd8855", "664400", "ff7777", "333333", "777777", "aaff66", "0088ff", "bbbbbb"),
	"sepia":   hexPalette("2b1d0e", "5e4227", "9a7b4f", "d4bc8c", "f4ead5"),
	"vapor":   hexPalette("2e2157", "ff71ce", "01cdfe", "05ffa1", "b967ff", "fffb96"),
	"websafe": websafe(),
}

func PaletteNames() []string { return sortedKeys(Palettes) }

func ParsePalette(s string) (color.Palette, error) {
	if p, ok := Palettes[strings.ToLower(s)]; ok {
		return p, nil
	}
	var p color.Palette
	for _, part := range strings.Split(s, ",") {
		c, err := ParseColor(part)
		if err != nil {
			return nil, fmt.Errorf("palette %q: not a known name and %v", s, err)
		}
		p = append(p, c)
	}
	return p, nil
}

func hexPalette(hex ...string) color.Palette {
	p := make(color.Palette, len(hex))
	for i, h := range hex {
		c, err := ParseColor(h)
		if err != nil {
			panic(err)
		}
		p[i] = c
	}
	return p
}

func websafe() color.Palette {
	var p color.Palette
	for r := 0; r < 6; r++ {
		for g := 0; g < 6; g++ {
			for b := 0; b < 6; b++ {
				p = append(p, color.NRGBA{uint8(r * 51), uint8(g * 51), uint8(b * 51), 255})
			}
		}
	}
	return p
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
