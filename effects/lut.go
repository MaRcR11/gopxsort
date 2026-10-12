package effects

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/MaRcR11/sortmancy/internal/imgutil"
)

type LUT struct {
	Size int
	Data []float32
}

func NewLUT(size int, fn func(r, g, b float64) (float64, float64, float64)) *LUT {
	size = max(size, 2)
	l := &LUT{Size: size, Data: make([]float32, size*size*size*3)}
	step := 1 / float64(size-1)
	i := 0
	for b := 0; b < size; b++ {
		for g := 0; g < size; g++ {
			for r := 0; r < size; r++ {
				or, og, ob := fn(float64(r)*step, float64(g)*step, float64(b)*step)
				l.Data[i], l.Data[i+1], l.Data[i+2] = float32(or), float32(og), float32(ob)
				i += 3
			}
		}
	}
	return l
}

func ParseCubeLUT(r io.Reader) (*LUT, error) {
	sc := bufio.NewScanner(r)
	size := 0
	lo, hi := [3]float64{0, 0, 0}, [3]float64{1, 1, 1}
	var data []float32
	triple := func(fields []string) ([3]float64, error) {
		var v [3]float64
		if len(fields) != 3 {
			return v, fmt.Errorf("expected three numbers, got %q", strings.Join(fields, " "))
		}
		for i, f := range fields {
			x, err := strconv.ParseFloat(f, 64)
			if err != nil {
				return v, fmt.Errorf("%q is not a number", f)
			}
			v[i] = x
		}
		return v, nil
	}
	for line := 1; sc.Scan(); line++ {
		text := strings.TrimSpace(sc.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		fields := strings.Fields(text)
		switch strings.ToUpper(fields[0]) {
		case "TITLE":
			continue
		case "LUT_1D_SIZE":
			return nil, fmt.Errorf("cube line %d: 1D LUTs are not supported", line)
		case "LUT_3D_SIZE":
			n, err := strconv.Atoi(fields[len(fields)-1])
			if err != nil || n < 2 || n > 256 {
				return nil, fmt.Errorf("cube line %d: invalid LUT_3D_SIZE", line)
			}
			size = n
			continue
		case "DOMAIN_MIN", "DOMAIN_MAX":
			v, err := triple(fields[1:])
			if err != nil {
				return nil, fmt.Errorf("cube line %d: %v", line, err)
			}
			if strings.ToUpper(fields[0]) == "DOMAIN_MIN" {
				lo = v
			} else {
				hi = v
			}
			continue
		}
		v, err := triple(fields)
		if err != nil {
			return nil, fmt.Errorf("cube line %d: %v", line, err)
		}
		data = append(data, float32(v[0]), float32(v[1]), float32(v[2]))
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if size == 0 {
		return nil, fmt.Errorf("cube file has no LUT_3D_SIZE")
	}
	if len(data) != size*size*size*3 {
		return nil, fmt.Errorf("cube file has %d entries, expected %d", len(data)/3, size*size*size)
	}
	for i := range data {
		c := i % 3
		data[i] = float32((float64(data[i]) - lo[c]) / (hi[c] - lo[c]))
	}
	return &LUT{Size: size, Data: data}, nil
}

func LoadCubeLUT(path string) (*LUT, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	l, err := ParseCubeLUT(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return l, nil
}

func (l *LUT) WriteCube(w io.Writer, title string) error {
	bw := bufio.NewWriter(w)
	if title != "" {
		fmt.Fprintf(bw, "TITLE %q\n", title)
	}
	fmt.Fprintf(bw, "LUT_3D_SIZE %d\n", l.Size)
	for i := 0; i < len(l.Data); i += 3 {
		fmt.Fprintf(bw, "%.6f %.6f %.6f\n", l.Data[i], l.Data[i+1], l.Data[i+2])
	}
	return bw.Flush()
}

func (l *LUT) Lookup(r, g, b float64) (float64, float64, float64) {
	n := l.Size - 1
	pos := func(v float64) (int, int, float64) {
		v = math.Max(0, math.Min(1, v)) * float64(n)
		i := min(int(v), n-1)
		return i, i + 1, v - float64(i)
	}
	r0, r1, tr := pos(r)
	g0, g1, tg := pos(g)
	b0, b1, tb := pos(b)
	at := func(ri, gi, bi, c int) float64 {
		return float64(l.Data[((bi*l.Size+gi)*l.Size+ri)*3+c])
	}
	var out [3]float64
	for c := 0; c < 3; c++ {
		c00 := at(r0, g0, b0, c)*(1-tr) + at(r1, g0, b0, c)*tr
		c10 := at(r0, g1, b0, c)*(1-tr) + at(r1, g1, b0, c)*tr
		c01 := at(r0, g0, b1, c)*(1-tr) + at(r1, g0, b1, c)*tr
		c11 := at(r0, g1, b1, c)*(1-tr) + at(r1, g1, b1, c)*tr
		c0 := c00*(1-tg) + c10*tg
		c1 := c01*(1-tg) + c11*tg
		out[c] = c0*(1-tb) + c1*tb
	}
	return out[0], out[1], out[2]
}

func ApplyLUT(img image.Image, l *LUT, amount float64) *image.NRGBA {
	return Map(img, func(p color.NRGBA) color.NRGBA {
		r, g, b := float64(p.R)/255, float64(p.G)/255, float64(p.B)/255
		lr, lg, lb := l.Lookup(r, g, b)
		mix := func(a, v float64) uint8 { return imgutil.Clamp8((a + (v-a)*amount) * 255) }
		return color.NRGBA{mix(r, lr), mix(g, lg), mix(b, lb), p.A}
	})
}

func smooth(v float64) float64 { return v * v * (3 - 2*v) }

var builtinLUTs = map[string]func(r, g, b float64) (float64, float64, float64){
	"tealorange": func(r, g, b float64) (float64, float64, float64) {
		l := 0.2126*r + 0.7152*g + 0.0722*b
		s := l - 0.5
		return smooth(r) + 0.18*s + 0.03, g + 0.02*s, smooth(b) - 0.22*s + 0.02
	},
	"bleach": func(r, g, b float64) (float64, float64, float64) {
		l := 0.2126*r + 0.7152*g + 0.0722*b
		mix := func(v float64) float64 { return smooth(v*0.4 + l*0.6) }
		return mix(r), mix(g), mix(b)
	},
	"crossprocess": func(r, g, b float64) (float64, float64, float64) {
		return smooth(r)*1.05 + 0.02, math.Pow(g, 0.9)*0.98 + 0.03, 0.18 + b*0.62
	},
	"noir": func(r, g, b float64) (float64, float64, float64) {
		l := smooth(smooth(0.3*r + 0.59*g + 0.11*b))
		return l, l, l
	},
	"faded": func(r, g, b float64) (float64, float64, float64) {
		l := 0.2126*r + 0.7152*g + 0.0722*b
		f := func(v float64) float64 { return 0.1 + 0.82*(v*0.75+l*0.25) }
		return f(r) + 0.02, f(g), f(b) - 0.01
	},
}

var (
	lutOnce  sync.Once
	lutCache map[string]*LUT
)

func LUTByName(name string) (*LUT, bool) {
	lutOnce.Do(func() {
		lutCache = map[string]*LUT{}
		for n, fn := range builtinLUTs {
			lutCache[n] = NewLUT(33, fn)
		}
	})
	l, ok := lutCache[strings.ToLower(name)]
	return l, ok
}

func LUTNames() []string { return sortedKeys(builtinLUTs) }
