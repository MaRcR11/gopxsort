package sorter

import (
	"image"
	"image/color"
	"math/rand"
	"slices"
	"strconv"
	"testing"
)

func randomImage(w, h int, seed int64) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	rng := rand.New(rand.NewSource(seed))
	rng.Read(img.Pix)
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = 255
	}
	return img
}

func paths() map[string]Path {
	ps := map[string]Path{
		"wave":       Wave(5, 17, false),
		"wave-v":     Wave(5, 17, true),
		"concentric": Concentric(0.5, 0.5),
		"corner":     Concentric(0, 1),
		"radial":     Radial(0.3, 0.6),
		"flow":       FlowField(10, 6, false, 3),
		"flow-v":     FlowField(10, 6, true, 3),
		"hilbert":    Hilbert(),
		"spiral":     Spiral(0.5, 0.5, 3),
	}
	for _, a := range []float64{0, 10, 30, 45, 60, 89.9, 90, 135, 180, 225, 270, 333, -45} {
		ps["linear"+strconv.FormatFloat(a, 'f', -1, 64)] = Linear(a)
	}
	return ps
}

func TestPathsCoverEveryPixelOnce(t *testing.T) {
	const w, h = 37, 23
	for name, p := range paths() {
		seen := make([]int, w*h)
		for _, l := range p.Lines(w, h) {
			for _, i := range l {
				seen[i]++
			}
		}
		for i, n := range seen {
			if n != 1 {
				t.Fatalf("%s: pixel (%d,%d) visited %d times", name, i%w, i/w, n)
			}
		}
	}
}

func TestLinearLinesAreConnected(t *testing.T) {
	const w, h = 40, 30
	for _, a := range []float64{0, 20, 45, 70, 90, 160, 200, 300} {
		for _, l := range Linear(a).Lines(w, h) {
			for i := 1; i < len(l); i++ {
				dx, dy := l[i]%w-l[i-1]%w, l[i]/w-l[i-1]/w
				if dx < -1 || dx > 1 || dy < -1 || dy > 1 {
					t.Fatalf("angle %v: jump from %d to %d", a, l[i-1], l[i])
				}
			}
		}
	}
}

func TestLinearDirection(t *testing.T) {
	const w, h = 5, 4
	cases := []struct {
		angle  float64
		dx, dy int
	}{{0, 1, 0}, {90, 0, -1}, {180, -1, 0}, {270, 0, 1}}
	for _, c := range cases {
		l := Linear(c.angle).Lines(w, h)[0]
		dx, dy := l[1]%w-l[0]%w, l[1]/w-l[0]/w
		if dx != c.dx || dy != c.dy {
			t.Errorf("angle %v: step (%d,%d), want (%d,%d)", c.angle, dx, dy, c.dx, c.dy)
		}
	}
}

func TestSortRowsWholeLine(t *testing.T) {
	src := randomImage(64, 16, 1)
	out := Sort(src, Options{Interval: None()})
	for y := 0; y < 16; y++ {
		var before, after []color.NRGBA
		for x := 0; x < 64; x++ {
			before = append(before, src.NRGBAAt(x, y))
			after = append(after, out.NRGBAAt(x, y))
		}
		for x := 1; x < 64; x++ {
			if Lightness(after[x-1]) > Lightness(after[x]) {
				t.Fatalf("row %d not sorted at x=%d", y, x)
			}
		}
		less := func(a, b color.NRGBA) int {
			return int(a.R)<<16 | int(a.G)<<8 | int(a.B) - (int(b.R)<<16 | int(b.G)<<8 | int(b.B))
		}
		slices.SortFunc(before, less)
		slices.SortFunc(after, less)
		if !slices.Equal(before, after) {
			t.Fatalf("row %d: pixels were lost or duplicated", y)
		}
	}
}

func TestReverse(t *testing.T) {
	out := Sort(randomImage(50, 3, 2), Options{Interval: None(), Key: Hue, Reverse: true})
	for x := 1; x < 50; x++ {
		if Hue(out.NRGBAAt(x-1, 1)) < Hue(out.NRGBAAt(x, 1)) {
			t.Fatalf("not descending at x=%d", x)
		}
	}
}

func TestThresholdLeavesOtherPixels(t *testing.T) {
	src := randomImage(80, 20, 3)
	out := Sort(src, Options{Interval: Threshold(0.4, 0.6)})
	for i := 0; i < 80*20; i++ {
		c := src.NRGBAAt(i%80, i/80)
		if l := Lightness(c); (l < 0.4 || l > 0.6) && out.NRGBAAt(i%80, i/80) != c {
			t.Fatalf("pixel %d outside the threshold moved", i)
		}
	}
}

func TestMask(t *testing.T) {
	src := randomImage(40, 10, 4)
	mask := image.NewGray(image.Rect(0, 0, 4, 1))
	mask.Pix[0], mask.Pix[1] = 255, 255
	out := Sort(src, Options{Interval: None(), Mask: mask})
	for y := 0; y < 10; y++ {
		for x := 20; x < 40; x++ {
			if out.NRGBAAt(x, y) != src.NRGBAAt(x, y) {
				t.Fatalf("masked pixel (%d,%d) changed", x, y)
			}
		}
	}
	if slices.Equal(out.Pix, src.Pix) {
		t.Fatal("unmasked half was not sorted")
	}
}

func TestSeedIsDeterministic(t *testing.T) {
	src := randomImage(120, 60, 5)
	opts := Options{Interval: Random(10), Mode: ModeShuffle, Randomness: 0.3, Path: Linear(30), Seed: 7}
	a, b := Sort(src, opts), Sort(src, opts)
	if !slices.Equal(a.Pix, b.Pix) {
		t.Fatal("same seed produced different output")
	}
	opts.Seed = 8
	if c := Sort(src, opts); slices.Equal(a.Pix, c.Pix) {
		t.Fatal("different seeds produced identical output")
	}
}

func TestModes(t *testing.T) {
	src := randomImage(30, 2, 6)
	avg := Sort(src, Options{Interval: None(), Mode: ModeAverage})
	smear := Sort(src, Options{Interval: None(), Mode: ModeSmear})
	grad := Sort(src, Options{Interval: None(), Mode: ModeGradient})
	for x := 1; x < 30; x++ {
		if avg.NRGBAAt(x, 0) != avg.NRGBAAt(0, 0) {
			t.Fatal("average span is not uniform")
		}
		if smear.NRGBAAt(x, 0) != src.NRGBAAt(0, 0) {
			t.Fatal("smear did not repeat the first pixel")
		}
	}
	if grad.NRGBAAt(0, 0) == grad.NRGBAAt(29, 0) {
		t.Fatal("gradient endpoints are equal")
	}
}

func TestEdgesInterval(t *testing.T) {

	img := image.NewNRGBA(image.Rect(0, 0, 20, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 20; x++ {
			v := uint8(0)
			if x >= 10 {
				v = 255
			}
			img.SetNRGBA(x, y, color.NRGBA{v, v, v, 255})
		}
	}
	split := Edges(0.5).Bind(img)
	l := &Line{Offsets: []int{}, width: 20}
	for x := 0; x < 20; x++ {
		l.Offsets = append(l.Offsets, 20+x)
	}
	spans := split(l)
	if len(spans) != 2 || spans[0] != (Span{0, 9}) || spans[1] != (Span{11, 20}) {
		t.Fatalf("unexpected spans %v", spans)
	}
}

func TestKeyNames(t *testing.T) {
	for _, n := range KeyNames() {
		if _, ok := KeyByName(n); !ok {
			t.Fatalf("key %q not found", n)
		}
	}
	for _, n := range ModeNames() {
		if m, ok := ModeByName(n); !ok || m.String() != n {
			t.Fatalf("mode %q does not round-trip", n)
		}
	}
}

func BenchmarkSort1080p(b *testing.B) {
	src := randomImage(1920, 1080, 1)
	for i := 0; i < b.N; i++ {
		Sort(src, Options{Path: Linear(30)})
	}
}

func TestSpiralCoversEveryPixelOnce(t *testing.T) {
	const w, h = 41, 29
	for _, spacing := range []float64{1, 3.5, 8} {
		seen := make([]int, w*h)
		for _, l := range Spiral(0.4, 0.6, spacing).Lines(w, h) {
			for _, i := range l {
				seen[i]++
			}
		}
		for i, n := range seen {
			if n != 1 {
				t.Fatalf("spacing %v: pixel %d visited %d times", spacing, i, n)
			}
		}
	}
}

func TestImagePathsCoverEveryPixelOnce(t *testing.T) {
	const w, h = 37, 23
	img := randomImage(w, h, 9)
	for _, vertical := range []bool{false, true} {
		seen := make([]int, w*h)
		for _, l := range Contour(8, 1, vertical).(ImagePath).LinesFor(img) {
			for _, p := range l {
				seen[p]++
			}
		}
		for i, n := range seen {
			if n != 1 {
				t.Fatalf("vertical=%v: pixel %d visited %d times", vertical, i, n)
			}
		}
	}
}

func TestHilbertIsConnected(t *testing.T) {
	lines := Hilbert().Lines(32, 32)
	if len(lines) != 1 || len(lines[0]) != 32*32 {
		t.Fatalf("expected one line through all pixels")
	}
	l := lines[0]
	for i := 1; i < len(l); i++ {
		dx, dy := l[i]%32-l[i-1]%32, l[i]/32-l[i-1]/32
		if dx*dx+dy*dy != 1 {
			t.Fatalf("jump from %d to %d", l[i-1], l[i])
		}
	}
}

func TestModeChannels(t *testing.T) {
	out := Sort(randomImage(40, 2, 10), Options{Interval: None(), Mode: ModeChannels})
	for x := 1; x < 40; x++ {
		a, b := out.NRGBAAt(x-1, 0), out.NRGBAAt(x, 0)
		if a.R > b.R || a.G > b.G || a.B > b.B {
			t.Fatalf("channels not sorted at x=%d", x)
		}
	}
}

func TestModeMirror(t *testing.T) {
	out := Sort(randomImage(41, 1, 11), Options{Interval: None(), Mode: ModeMirror})
	k := func(x int) float64 { return Lightness(out.NRGBAAt(x, 0)) }
	for x := 1; x <= 20; x++ {
		if k(x-1) > k(x) || k(40-x+1) > k(40-x) {
			t.Fatalf("not rising towards the middle at %d", x)
		}
	}
}

func TestTileSort(t *testing.T) {
	const size, tw, th = 4, 8, 3
	img := image.NewNRGBA(image.Rect(0, 0, tw*size+2, th*size))
	rng := rand.New(rand.NewSource(12))
	for ty := 0; ty < th; ty++ {
		for tx := 0; tx < tw; tx++ {
			c := color.NRGBA{uint8(rng.Intn(256)), uint8(rng.Intn(256)), uint8(rng.Intn(256)), 255}
			for y := 0; y < size; y++ {
				for x := 0; x < size; x++ {
					img.SetNRGBA(tx*size+x, ty*size+y, c)
				}
			}
		}
	}
	out := Sort(img, Options{Interval: None(), Tile: size})
	for ty := 0; ty < th; ty++ {
		for tx := 0; tx < tw; tx++ {
			c := out.NRGBAAt(tx*size, ty*size)
			for y := 0; y < size; y++ {
				for x := 0; x < size; x++ {
					if out.NRGBAAt(tx*size+x, ty*size+y) != c {
						t.Fatal("tile was split up")
					}
				}
			}
			if tx > 0 && Lightness(out.NRGBAAt((tx-1)*size, ty*size)) > Lightness(c) {
				t.Fatalf("tiles in row %d are not sorted", ty)
			}
		}
	}
	if out.NRGBAAt(tw*size, 0) != img.NRGBAAt(tw*size, 0) {
		t.Fatal("pixels outside full tiles changed")
	}
}

func TestFromImage(t *testing.T) {
	src := randomImage(40, 10, 13)
	guide := image.NewGray(image.Rect(0, 0, 2, 1))
	guide.Pix[1] = 255
	out := Sort(src, Options{Interval: FromImage(guide, Threshold(0.5, 1))})
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			if out.NRGBAAt(x, y) != src.NRGBAAt(x, y) {
				t.Fatalf("pixel (%d,%d) outside the guide's bright half moved", x, y)
			}
		}
	}
	if slices.Equal(out.Pix, src.Pix) {
		t.Fatal("nothing was sorted")
	}
}
