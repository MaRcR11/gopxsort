package sorter

import (
	"cmp"
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"slices"
	"strings"

	"github.com/MaRcR11/sortcery/internal/imgutil"
)

type Mode int

const (
	ModeSort Mode = iota
	ModeShuffle
	ModeAverage
	ModeGradient
	ModeSmear
	ModeChannels
	ModeMirror
)

var modeNames = []string{"sort", "shuffle", "average", "gradient", "smear", "channels", "mirror"}

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
	Tile       int
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

	if o.Tile > 1 {
		sortTiles(img, o)
		return
	}

	lines := pathLines(o.Path, img)
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

func pathLines(p Path, img *image.NRGBA) [][]int {
	if ip, ok := p.(ImagePath); ok {
		return ip.LinesFor(img)
	}
	return p.Lines(img.Rect.Dx(), img.Rect.Dy())
}

type keyed struct {
	k float64
	i int
}

func permutation(o Options, seg []color.NRGBA, rng *rand.Rand, buf []keyed) ([]int, []keyed) {
	perm := make([]int, len(seg))
	if o.Mode == ModeShuffle {
		for i := range perm {
			perm[i] = i
		}
		rng.Shuffle(len(perm), func(i, j int) { perm[i], perm[j] = perm[j], perm[i] })
		return perm, buf
	}
	buf = buf[:0]
	for i, c := range seg {
		buf = append(buf, keyed{o.Key(c), i})
	}
	if o.Reverse {
		slices.SortStableFunc(buf, func(a, b keyed) int { return cmp.Compare(b.k, a.k) })
	} else {
		slices.SortStableFunc(buf, func(a, b keyed) int { return cmp.Compare(a.k, b.k) })
	}
	if o.Mode == ModeMirror {
		n := len(buf)
		for i, k := range buf {
			if i%2 == 0 {
				perm[i/2] = k.i
			} else {
				perm[n-1-i/2] = k.i
			}
		}
		return perm, buf
	}
	for i, k := range buf {
		perm[i] = k.i
	}
	return perm, buf
}

func apply(o Options, seg []color.NRGBA, rng *rand.Rand, buf []keyed) []keyed {
	switch o.Mode {
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

	case ModeChannels:
		ch := make([]uint8, len(seg))
		for c := 0; c < 3; c++ {
			for i, p := range seg {
				ch[i] = [3]uint8{p.R, p.G, p.B}[c]
			}
			if o.Reverse {
				slices.SortFunc(ch, func(a, b uint8) int { return int(b) - int(a) })
			} else {
				slices.Sort(ch)
			}
			for i := range seg {
				switch c {
				case 0:
					seg[i].R = ch[i]
				case 1:
					seg[i].G = ch[i]
				default:
					seg[i].B = ch[i]
				}
			}
		}

	default:
		var perm []int
		perm, buf = permutation(o, seg, rng, buf)
		src := slices.Clone(seg)
		for i, p := range perm {
			seg[i] = src[p]
		}
	}
	return buf
}

func sortTiles(img *image.NRGBA, o Options) {
	size := o.Tile
	w, h := img.Rect.Dx(), img.Rect.Dy()
	tw, th := w/size, h/size
	if tw == 0 || th == 0 {
		return
	}
	small := image.NewNRGBA(image.Rect(0, 0, tw, th))
	for ty := 0; ty < th; ty++ {
		for tx := 0; tx < tw; tx++ {
			var sum [4]int
			for y := ty * size; y < (ty+1)*size; y++ {
				for x := tx * size; x < (tx+1)*size; x++ {
					for c := 0; c < 4; c++ {
						sum[c] += int(img.Pix[(y*w+x)*4+c])
					}
				}
			}
			n := size * size
			imgutil.Set(small.Pix, ty*tw+tx, color.NRGBA{uint8(sum[0] / n), uint8(sum[1] / n), uint8(sum[2] / n), uint8(sum[3] / n)})
		}
	}
	lines := pathLines(o.Path, small)
	split := o.Interval.Bind(small)
	var mask []bool
	if o.Mask != nil {
		mask = maskBits(o.Mask, tw, th, o.InvertMask)
	}
	src := imgutil.ToNRGBA(img)
	moveTile := func(dst, from int) {
		dx, dy := (dst%tw)*size, (dst/tw)*size
		sx, sy := (from%tw)*size, (from/tw)*size
		for y := 0; y < size; y++ {
			copy(img.Pix[((dy+y)*w+dx)*4:((dy+y)*w+dx+size)*4], src.Pix[((sy+y)*w+sx)*4:])
		}
	}
	imgutil.ParallelEach(len(lines), func(li int) {
		offs := lines[li]
		line := &Line{
			Index:   li,
			Offsets: offs,
			Pixels:  make([]color.NRGBA, len(offs)),
			Rand:    rand.New(rand.NewSource(lineSeed(o.Seed, li))),
			width:   tw,
		}
		for i, p := range offs {
			line.Pixels[i] = imgutil.At(small.Pix, p)
		}
		spans := split(line)
		if mask != nil {
			spans = clipToMask(spans, offs, mask)
		}
		var buf []keyed
		for _, s := range spans {
			s.Start, s.End = max(s.Start, 0), min(s.End, len(offs))
			if s.End-s.Start < 2 {
				continue
			}
			if o.Randomness > 0 && line.Rand.Float64() < o.Randomness {
				continue
			}
			var perm []int
			perm, buf = permutation(o, line.Pixels[s.Start:s.End], line.Rand, buf)
			for i, p := range perm {
				if i != p {
					moveTile(offs[s.Start+i], offs[s.Start+p])
				}
			}
		}
	})
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
