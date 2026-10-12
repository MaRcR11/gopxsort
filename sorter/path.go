package sorter

import (
	"math"
	"sort"

	"github.com/MaRcR11/sortmancy/internal/imgutil"
)

type Path interface {
	Lines(width, height int) [][]int
}

type PathFunc func(width, height int) [][]int

func (f PathFunc) Lines(width, height int) [][]int { return f(width, height) }

func Linear(angle float64) Path {
	return PathFunc(func(w, h int) [][]int {
		rad := angle * math.Pi / 180

		cos, sin := math.Cos(rad), -math.Sin(rad)
		if math.Abs(cos) >= math.Abs(sin) {
			t := sin / cos
			lines := shear(w, h, false, func(x int) int { return int(math.Round(float64(x) * t)) })
			if cos < 0 {
				reverseAll(lines)
			}
			return lines
		}
		t := cos / sin
		lines := shear(h, w, true, func(y int) int { return int(math.Round(float64(y) * t)) })
		if sin < 0 {
			reverseAll(lines)
		}
		return lines
	})
}

func Wave(amplitude, wavelength float64, vertical bool) Path {
	if wavelength == 0 {
		wavelength = 1
	}
	off := func(a int) int {
		return int(math.Round(amplitude * math.Sin(2*math.Pi*float64(a)/wavelength)))
	}
	return PathFunc(func(w, h int) [][]int {
		if !vertical {
			return shear(w, h, false, off)
		}
		lines := shear(h, w, true, off)
		reverseAll(lines)
		return lines
	})
}

func shear(major, minor int, transposed bool, off func(int) int) [][]int {
	if major == 0 || minor == 0 {
		return nil
	}
	offs := make([]int, major)
	lo, hi := 0, 0
	for a := range offs {
		o := off(a)
		offs[a] = o
		lo, hi = min(lo, o), max(hi, o)
	}

	n := minor + hi - lo
	counts := make([]int, n)
	for _, o := range offs {
		for c := 0; c < minor; c++ {
			counts[c-o+hi]++
		}
	}
	backing := make([]int, major*minor)
	lines := make([][]int, 0, n)
	index := make([]int, n)
	pos := 0
	for k, cnt := range counts {
		if cnt == 0 {
			continue
		}
		index[k] = len(lines)
		lines = append(lines, backing[pos:pos:pos+cnt])
		pos += cnt
	}
	width := major
	if transposed {
		width = minor
	}
	for a, o := range offs {
		for c := 0; c < minor; c++ {
			x, y := a, c
			if transposed {
				x, y = c, a
			}
			k := index[c-o+hi]
			lines[k] = append(lines[k], y*width+x)
		}
	}
	return lines
}

func Concentric(cx, cy float64) Path {
	return PathFunc(func(w, h int) [][]int {
		dist, ang := polar(w, h, cx, cy)
		maxR := 0
		for _, d := range dist {
			maxR = max(maxR, int(math.Round(d)))
		}
		lines := group(len(dist), maxR+1, func(i int) int { return int(math.Round(dist[i])) })
		for _, l := range lines {
			sortBy(l, ang)
		}
		return lines
	})
}

func Radial(cx, cy float64) Path {
	return PathFunc(func(w, h int) [][]int {
		dist, ang := polar(w, h, cx, cy)
		maxR := 0.0
		for _, d := range dist {
			maxR = math.Max(maxR, d)
		}
		rays := max(int(2*math.Pi*maxR), 8)
		lines := group(len(dist), rays, func(i int) int {
			return imgutil.Wrap(int(math.Round(ang[i]/(2*math.Pi)*float64(rays))), rays)
		})
		for _, l := range lines {
			sortBy(l, dist)
		}
		return lines
	})
}

func polar(w, h int, cx, cy float64) (dist, ang []float64) {
	px, py := cx*float64(w-1), cy*float64(h-1)
	dist = make([]float64, w*h)
	ang = make([]float64, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx, dy := float64(x)-px, py-float64(y)
			i := y*w + x
			dist[i] = math.Hypot(dx, dy)
			a := math.Atan2(dy, dx)
			if a < 0 {
				a += 2 * math.Pi
			}
			ang[i] = a
		}
	}
	return dist, ang
}

func group(n, buckets int, bucket func(i int) int) [][]int {
	counts := make([]int, buckets)
	keys := make([]int, n)
	for i := range keys {
		keys[i] = bucket(i)
		counts[keys[i]]++
	}
	backing := make([]int, n)
	out := make([][]int, buckets)
	pos := 0
	for k, c := range counts {
		out[k] = backing[pos : pos : pos+c]
		pos += c
	}
	for i, k := range keys {
		out[k] = append(out[k], i)
	}
	lines := out[:0]
	for _, l := range out {
		if len(l) > 0 {
			lines = append(lines, l)
		}
	}
	return lines
}

func sortBy(idx []int, by []float64) {
	sort.SliceStable(idx, func(a, b int) bool { return by[idx[a]] < by[idx[b]] })
}

func reverseAll(lines [][]int) {
	for _, l := range lines {
		for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
			l[i], l[j] = l[j], l[i]
		}
	}
}

func Spiral(cx, cy, spacing float64) Path {
	spacing = math.Max(spacing, 1)
	return PathFunc(func(w, h int) [][]int {
		dist, ang := polar(w, h, cx, cy)
		lanes := int(math.Ceil(spacing))
		pos := make([]float64, len(dist))
		lines := group(len(dist), lanes, func(i int) int {
			turn := ang[i] / (2 * math.Pi)
			t := dist[i]/spacing - turn
			n := math.Round(t)
			pos[i] = n + turn
			return min(int((t-n+0.5)*float64(lanes)), lanes-1)
		})
		for _, l := range lines {
			sortBy(l, pos)
		}
		return lines
	})
}
