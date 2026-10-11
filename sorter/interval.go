package sorter

import (
	"image"
	"image/color"
	"math/rand"

	"github.com/MaRcR11/sortcery/internal/imgutil"
)

type Span struct{ Start, End int }

type Line struct {
	Index   int
	Pixels  []color.NRGBA
	Offsets []int
	Rand    *rand.Rand
	width   int
}

func (l *Line) Point(i int) image.Point {
	return image.Pt(l.Offsets[i]%l.width, l.Offsets[i]/l.width)
}

type Splitter func(l *Line) []Span

type Interval interface {
	Bind(img *image.NRGBA) Splitter
}

type IntervalFunc func(l *Line) []Span

func (f IntervalFunc) Bind(*image.NRGBA) Splitter { return Splitter(f) }

func None() Interval {
	return IntervalFunc(func(l *Line) []Span {
		if len(l.Pixels) == 0 {
			return nil
		}
		return []Span{{0, len(l.Pixels)}}
	})
}

func Threshold(lower, upper float64) Interval {
	return ThresholdBy(Lightness, lower, upper)
}

func ThresholdBy(key Key, lower, upper float64) Interval {
	return IntervalFunc(func(l *Line) []Span {
		return runs(len(l.Pixels), func(i int) bool {
			v := key(l.Pixels[i])
			return v >= lower && v <= upper
		})
	})
}

func Random(length int) Interval {
	length = max(length, 1)
	return IntervalFunc(func(l *Line) []Span {
		return chop(len(l.Pixels), func() int { return 1 + l.Rand.Intn(2*length) })
	})
}

func Waves(length int) Interval {
	length = max(length, 1)
	jitter := max(length/4, 1)
	return IntervalFunc(func(l *Line) []Span {
		first := true
		return chop(len(l.Pixels), func() int {
			if first {
				first = false
				return 1 + l.Rand.Intn(length)
			}
			return length + l.Rand.Intn(jitter+1)
		})
	})
}

func Edges(threshold float64) Interval { return edges{threshold} }

type edges struct{ threshold float64 }

func (e edges) Bind(img *image.NRGBA) Splitter {
	mag := imgutil.Sobel(img)
	return func(l *Line) []Span {
		return runs(len(l.Offsets), func(i int) bool { return mag[l.Offsets[i]] <= e.threshold })
	}
}

func runs(n int, in func(i int) bool) []Span {
	var spans []Span
	start := -1
	for i := 0; i < n; i++ {
		if in(i) {
			if start < 0 {
				start = i
			}
		} else if start >= 0 {
			spans = append(spans, Span{start, i})
			start = -1
		}
	}
	if start >= 0 {
		spans = append(spans, Span{start, n})
	}
	return spans
}

func chop(n int, next func() int) []Span {
	var spans []Span
	for pos := 0; pos < n; {
		end := min(pos+max(next(), 1), n)
		spans = append(spans, Span{pos, end})
		pos = end
	}
	return spans
}

func FromImage(src image.Image, inner Interval) Interval { return fromImage{src, inner} }

type fromImage struct {
	src   image.Image
	inner Interval
}

func (f fromImage) Bind(img *image.NRGBA) Splitter {
	other := imgutil.ResizeNearest(f.src, img.Rect.Dx(), img.Rect.Dy())
	split := f.inner.Bind(other)
	return func(l *Line) []Span {
		c := *l
		c.Pixels = make([]color.NRGBA, len(l.Offsets))
		for i, p := range l.Offsets {
			c.Pixels[i] = imgutil.At(other.Pix, p)
		}
		return split(&c)
	}
}
