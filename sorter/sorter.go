package sorter

import (
	"cmp"
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"slices"
	"strings"

	"github.com/MaRcR11/gopxsort/internal/imgutil"
)

type Mode int

const (
	ModeSort Mode = iota
	ModeShuffle
	ModeAverage
	ModeGradient
	ModeSmear
)

var modeNames = []string{"sort", "shuffle", "average", "gradient", "smear"}

func (m Mode) String() string {
	if int(m) >= 0 && int(m) < len(modeNames) {
		return modeNames[m]
	}
	return fmt.Sprintf("Mode(%d)", int(m))
}

func ModeByName(name string) (Mode, bool) {
	i := slices.Index(modeNames, strings.ToLower(name))
	return Mode(i), i >= 0
}

func ModeNames() []string { return slices.Clone(modeNames) }

type Options struct {
	Key        Key
	Interval   Interval
	Path       Path
	Mode       Mode
	Reverse    bool
	Randomness float64
	Mask       image.Image
	InvertMask bool
	Seed       int64
}

func DefaultOptions() Options {
	return Options{
		Key:      Lightness,
		Interval: Threshold(0.25, 0.8),
		Path:     Linear(0),
		Mode:     ModeSort,
	}
}

func (o Options) withDefaults() Options {
	d := DefaultOptions()
	if o.Key == nil {
		o.Key = d.Key
	}
	if o.Interval == nil {
		o.Interval = d.Interval
	}
	if o.Path == nil {
		o.Path = d.Path
	}
	return o
}

func Sort(img image.Image, opts Options) *image.NRGBA {
	dst := imgutil.ToNRGBA(img)
	SortInPlace(dst, opts)
	return dst
}

func SortInPlace(img *image.NRGBA, opts Options) {
	o := opts.withDefaults()
	w, h := img.Rect.Dx(), img.Rect.Dy()
	if w == 0 || h == 0 {
		return
	}
	if img.Rect.Min != (image.Point{}) || img.Stride != 4*w {
		panic("sorter: SortInPlace needs an NRGBA image with origin (0, 0) and no row padding")
	}

	lines := o.Path.Lines(w, h)
	split := o.Interval.Bind(img)
	var mask []bool
	if o.Mask != nil {
		mask = maskBits(o.Mask, w, h, o.InvertMask)
	}

	imgutil.ParallelEach(len(lines), func(li int) {
		offs := lines[li]
		line := &Line{
			Index:   li,
			Offsets: offs,
			Pixels:  make([]color.NRGBA, len(offs)),
			Rand:    rand.New(rand.NewSource(lineSeed(o.Seed, li))),
			width:   w,
		}
		for i, p := range offs {
			line.Pixels[i] = imgutil.At(img.Pix, p)
		}

		spans := split(line)
		if mask != nil {
			spans = clipToMask(spans, offs, mask)
		}
		out := slices.Clone(line.Pixels)
		var buf []keyed
		for _, s := range spans {
			s.Start, s.End = max(s.Start, 0), min(s.End, len(out))
			if s.End-s.Start < 2 {
				continue
			}
			if o.Randomness > 0 && line.Rand.Float64() < o.Randomness {
				continue
			}
			buf = apply(o, out[s.Start:s.End], line.Rand, buf)
		}
		for i, p := range offs {
			imgutil.Set(img.Pix, p, out[i])
		}
	})
}

type keyed struct {
	k float64
	c color.NRGBA
}

func apply(o Options, seg []color.NRGBA, rng *rand.Rand, buf []keyed) []keyed {
	switch o.Mode {
	case ModeShuffle:
		rng.Shuffle(len(seg), func(i, j int) { seg[i], seg[j] = seg[j], seg[i] })

	case ModeAverage:
		var r, g, b, a int
		for _, c := range seg {
			r, g, b, a = r+int(c.R), g+int(c.G), b+int(c.B), a+int(c.A)
		}
		n := len(seg)
		avg := color.NRGBA{uint8(r / n), uint8(g / n), uint8(b / n), uint8(a / n)}
		for i := range seg {
			seg[i] = avg
		}

	case ModeGradient:
		lo, hi := seg[0], seg[0]
		kLo, kHi := o.Key(lo), o.Key(lo)
		for _, c := range seg[1:] {
			if k := o.Key(c); k < kLo {
				lo, kLo = c, k
			} else if k > kHi {
				hi, kHi = c, k
			}
		}
		if o.Reverse {
			lo, hi = hi, lo
		}
		last := float64(len(seg) - 1)
		for i := range seg {
			seg[i] = lerp(lo, hi, float64(i)/last)
		}

	case ModeSmear:
		src := seg[0]
		if o.Reverse {
			src = seg[len(seg)-1]
		}
		for i := range seg {
			seg[i] = src
		}

	default:
		buf = buf[:0]
		for _, c := range seg {
			buf = append(buf, keyed{o.Key(c), c})
		}
		if o.Reverse {
			slices.SortStableFunc(buf, func(a, b keyed) int { return cmp.Compare(b.k, a.k) })
		} else {
			slices.SortStableFunc(buf, func(a, b keyed) int { return cmp.Compare(a.k, b.k) })
		}
		for i := range seg {
			seg[i] = buf[i].c
		}
	}
	return buf
}

func lerp(a, b color.NRGBA, t float64) color.NRGBA {
	mix := func(x, y uint8) uint8 { return imgutil.Clamp8(float64(x) + (float64(y)-float64(x))*t) }
	return color.NRGBA{mix(a.R, b.R), mix(a.G, b.G), mix(a.B, b.B), mix(a.A, b.A)}
}

func lineSeed(seed int64, line int) int64 {
	z := uint64(seed) + uint64(line+1)*0x9E3779B97F4A7C15
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return int64(z ^ (z >> 31))
}

func maskBits(m image.Image, w, h int, invert bool) []bool {
	at := imgutil.SampleNearest(m, w, h)
	bits := make([]bool, w*h)
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				bits[y*w+x] = (imgutil.LumaOf(at(x, y)) >= 0.5) != invert
			}
		}
	})
	return bits
}

func clipToMask(spans []Span, offs []int, mask []bool) []Span {
	var out []Span
	for _, s := range spans {
		s.Start, s.End = max(s.Start, 0), min(s.End, len(offs))
		for _, r := range runs(s.End-s.Start, func(i int) bool { return mask[offs[s.Start+i]] }) {
			out = append(out, Span{s.Start + r.Start, s.Start + r.End})
		}
	}
	return out
}
