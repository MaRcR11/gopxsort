package effects

import (
	"image"
	"image/color"
	"math/rand"
	"slices"
	"testing"
)

func randomImage(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	rand.New(rand.NewSource(1)).Read(img.Pix)
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = 255
	}
	return img
}

func TestInvolutions(t *testing.T) {
	src := randomImage(31, 17)
	cases := map[string]func(image.Image) *image.NRGBA{
		"invert": Invert,
		"fliph":  FlipHorizontal,
		"flipv":  FlipVertical,
		"rotate180": func(img image.Image) *image.NRGBA {
			out, _ := Rotate(img, 180)
			return out
		},
	}
	for name, fn := range cases {
		if out := fn(fn(src)); !slices.Equal(out.Pix, src.Pix) {
			t.Errorf("%s applied twice is not the identity", name)
		}
	}
}

func TestRotate(t *testing.T) {
	src := randomImage(31, 17)
	out := image.Image(src)
	for i := 0; i < 4; i++ {
		r, err := Rotate(out, 90)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			if r.Bounds().Dx() != 17 || r.Bounds().Dy() != 31 {
				t.Fatalf("rotated size %v", r.Bounds())
			}

			if r.NRGBAAt(0, 0) != src.NRGBAAt(0, 16) {
				t.Fatal("rotation is not clockwise")
			}
		}
		out = r
	}
	if !slices.Equal(out.(*image.NRGBA).Pix, src.Pix) {
		t.Fatal("four quarter turns are not the identity")
	}
	if _, err := Rotate(src, 45); err == nil {
		t.Fatal("expected error for 45 degrees")
	}
}

func TestChannelShift(t *testing.T) {
	src := randomImage(20, 10)
	if out := ChannelShift(src, image.Point{}, image.Point{}, image.Point{}); !slices.Equal(out.Pix, src.Pix) {
		t.Fatal("zero shift changed the image")
	}
	out := ChannelShift(src, image.Pt(3, 1), image.Point{}, image.Point{})
	if out.NRGBAAt(3, 1).R != src.NRGBAAt(0, 0).R || out.NRGBAAt(3, 1).G != src.NRGBAAt(3, 1).G {
		t.Fatal("red channel not shifted correctly")
	}
}

func TestDitherUsesOnlyPalette(t *testing.T) {
	src := randomImage(40, 30)
	pal := Palettes["gameboy"]
	allowed := map[color.NRGBA]bool{}
	for _, c := range pal {
		allowed[c.(color.NRGBA)] = true
	}
	for _, name := range DitherNames() {
		m, _ := DitherByName(name)
		out := Dither(src, DitherOptions{Method: m, Palette: pal})
		for i := 0; i < 40*30; i++ {
			if c := out.NRGBAAt(i%40, i/40); !allowed[c] {
				t.Fatalf("%s produced %v outside the palette", name, c)
			}
		}
	}
}

func TestDitherLevels(t *testing.T) {
	out := Dither(randomImage(30, 30), DitherOptions{Method: FloydSteinberg, Levels: 2})
	for i, v := range out.Pix {
		if i%4 != 3 && v != 0 && v != 255 {
			t.Fatalf("value %d is not 0 or 255", v)
		}
	}
}

func TestBayerMatrix(t *testing.T) {
	m := bayer(4)
	seen := map[float64]bool{}
	for _, row := range m {
		for _, v := range row {
			if seen[v] || v < -0.5 || v >= 0.5 {
				t.Fatalf("bad bayer value %v", v)
			}
			seen[v] = true
		}
	}
}

func TestPixelate(t *testing.T) {
	out := Pixelate(randomImage(20, 20), 5)
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			if out.NRGBAAt(x, y) != out.NRGBAAt(0, 0) {
				t.Fatal("block is not uniform")
			}
		}
	}
}

func TestPosterize(t *testing.T) {
	out := Posterize(randomImage(30, 30), 3)
	for i, v := range out.Pix {
		if i%4 != 3 && v != 0 && v != 128 && v != 255 {
			t.Fatalf("unexpected level %d", v)
		}
	}
}

func TestChannelSwap(t *testing.T) {
	src := randomImage(5, 5)
	out, err := ChannelSwap(src, "bgr")
	if err != nil {
		t.Fatal(err)
	}
	if c, s := out.NRGBAAt(2, 2), src.NRGBAAt(2, 2); c.R != s.B || c.B != s.R || c.G != s.G {
		t.Fatal("channels not swapped")
	}
	if _, err := ChannelSwap(src, "rgx"); err == nil {
		t.Fatal("expected error for invalid order")
	}
}

func TestParseColor(t *testing.T) {
	cases := map[string]color.NRGBA{
		"#fff":      {255, 255, 255, 255},
		"ff0000":    {255, 0, 0, 255},
		"#00ff0080": {0, 255, 0, 128},
	}
	for in, want := range cases {
		if got, err := ParseColor(in); err != nil || got != want {
			t.Errorf("ParseColor(%q) = %v, %v", in, got, err)
		}
	}
	if _, err := ParseColor("nope"); err == nil {
		t.Error("expected error")
	}
}

func TestShapePreserved(t *testing.T) {
	src := randomImage(33, 21)
	outs := map[string]*image.NRGBA{
		"aberration": ChromaticAberration(src, 0.02),
		"scanlines":  Scanlines(src, 2, 0.5),
		"noise":      Noise(src, 0.2, false, 1),
		"glitch":     Glitch(src, 10, 0.2, 1),
		"wave":       Wave(src, 4, 10, true),
		"melt":       Melt(src, 10, 1),
		"blur":       BoxBlur(src, 3),
		"sharpen":    Sharpen(src, 1, 1),
		"edges":      EdgeDetect(src),
		"vignette":   Vignette(src, 0.5),
		"mirror":     Mirror(src, MirrorQuad),
		"gradient":   GradientMap(src, Palettes["vapor"]),
	}
	jpg, err := JPEGCorrupt(src, 10, 50, 1)
	if err != nil {
		t.Fatal(err)
	}
	outs["jpeg"] = jpg
	for name, out := range outs {
		if out.Bounds() != src.Bounds() {
			t.Errorf("%s changed the bounds to %v", name, out.Bounds())
		}
	}
}

func TestInputNotModified(t *testing.T) {
	src := randomImage(10, 10)
	orig := slices.Clone(src.Pix)
	Invert(src)
	Pixelate(src, 3)
	Glitch(src, 5, 0.5, 1)
	if !slices.Equal(src.Pix, orig) {
		t.Fatal("input image was modified")
	}
}

func TestResize(t *testing.T) {
	src := randomImage(40, 20)
	for _, f := range []Filter{Nearest, Bilinear} {
		out := Resize(src, 13, 7, f)
		if out.Bounds().Dx() != 13 || out.Bounds().Dy() != 7 {
			t.Fatalf("filter %d: size %v", f, out.Bounds())
		}
		up := Scale(src, 2, f)
		if up.Bounds().Dx() != 80 || up.Bounds().Dy() != 40 {
			t.Fatalf("filter %d: scaled size %v", f, up.Bounds())
		}
	}
	if out := Resize(src, 40, 20, Bilinear); !slices.Equal(out.Pix, src.Pix) {
		t.Fatal("same-size resize changed pixels")
	}
}

func TestCrop(t *testing.T) {
	src := randomImage(20, 10)
	out := Crop(src, image.Rect(5, 2, 15, 30))
	if out.Bounds().Dx() != 10 || out.Bounds().Dy() != 8 {
		t.Fatalf("size %v", out.Bounds())
	}
	if out.NRGBAAt(0, 0) != src.NRGBAAt(5, 2) {
		t.Fatal("wrong origin")
	}
}

func TestMix(t *testing.T) {
	a, b := randomImage(10, 10), Invert(randomImage(10, 10))
	if out := Mix(a, b, 0); !slices.Equal(out.Pix, a.Pix) {
		t.Fatal("amount 0 should return the base")
	}
	if out := Mix(a, b, 1); !slices.Equal(out.Pix, b.Pix) {
		t.Fatal("amount 1 should return the overlay")
	}
	if out := Mix(a, randomImage(5, 3), 0.5); out.Bounds() != a.Bounds() {
		t.Fatal("mix changed the size")
	}
}

func TestMedianCut(t *testing.T) {
	src := randomImage(50, 50)
	pal := MedianCut(src, 8)
	if len(pal) != 8 {
		t.Fatalf("got %d colors", len(pal))
	}
	out := Quantize(src, 8, DitherNone)
	seen := map[color.NRGBA]bool{}
	for i := 0; i < 50*50; i++ {
		seen[out.NRGBAAt(i%50, i/50)] = true
	}
	if len(seen) > 8 {
		t.Fatalf("quantized image has %d colors", len(seen))
	}
	two := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	two.SetNRGBA(0, 0, color.NRGBA{255, 0, 0, 255})
	two.SetNRGBA(1, 0, color.NRGBA{0, 0, 255, 255})
	if got := MedianCut(two, 16); len(got) != 2 {
		t.Fatalf("expected 2 colors for a 2-color image, got %d", len(got))
	}
}

func TestHalftone(t *testing.T) {
	white := Map(randomImage(16, 16), func(c color.NRGBA) color.NRGBA { return color.NRGBA{255, 255, 255, 255} })
	out := Halftone(white, 8, false)
	for i, v := range out.Pix {
		if v != 255 {
			t.Fatalf("white input should give white paper, byte %d = %d", i, v)
		}
	}
}

func TestOilPaintFlat(t *testing.T) {
	flat := Map(randomImage(12, 12), func(c color.NRGBA) color.NRGBA { return color.NRGBA{10, 20, 30, 255} })
	if out := OilPaint(flat, 3); !slices.Equal(out.Pix, flat.Pix) {
		t.Fatal("oil paint changed a flat image")
	}
}

func TestMoreEffectsKeepSize(t *testing.T) {
	src := randomImage(33, 21)
	outs := map[string]*image.NRGBA{
		"bloom":        Bloom(src, 0.5, 6, 1),
		"blocks":       BlockShuffle(src, 8, 0.5, 1),
		"kaleidoscope": Kaleidoscope(src, 6, 15),
		"halftone":     Halftone(src, 5, true),
	}
	for name, out := range outs {
		if out.Bounds() != src.Bounds() {
			t.Errorf("%s changed the bounds to %v", name, out.Bounds())
		}
	}
}

func TestDistortionsWithZeroStrengthAreIdentity(t *testing.T) {
	src := randomImage(31, 19)
	cases := map[string]*image.NRGBA{
		"displace":  Displace(src, nil, 0, 0),
		"echo":      Echo(src, 0, image.Pt(5, 0), 0.5),
		"bend":      DataBend(src, 100, 0),
		"swirl":     Swirl(src, 0, 0.5),
		"ripple":    Ripple(src, 0, 10),
		"bulge":     Bulge(src, 0, 0.5),
		"curvature": Curvature(src, 0),
		"slitscan":  SlitScan(src, 1, false),
		"interlace": Interlace(src, 0, 2, 0),
	}
	for name, out := range cases {
		if !slices.Equal(out.Pix, src.Pix) {
			t.Errorf("%s with zero strength changed the image", name)
		}
	}
}

func TestDistortionsKeepSize(t *testing.T) {
	src := randomImage(33, 21)
	outs := map[string]*image.NRGBA{
		"displace":  Displace(src, Invert(src), 10, 5),
		"echo":      Echo(src, 3, image.Pt(4, 2), 0.7),
		"bend":      DataBend(src, 301, 0.5),
		"swirl":     Swirl(src, 180, 0.5),
		"ripple":    Ripple(src, 4, 8),
		"bulge":     Bulge(src, -0.5, 0.5),
		"polar":     Polar(src, false),
		"unpolar":   Polar(src, true),
		"curvature": Curvature(src, 0.2),
		"slitscan":  SlitScan(src, 0.3, true),
		"interlace": Interlace(src, 3, 3, 0.3),
		"scatter":   Scatter(src, 3, 1),
		"tracking":  Tracking(src, 3, 0.1, 0.3, 1),
	}
	for name, out := range outs {
		if out.Bounds() != src.Bounds() {
			t.Errorf("%s changed the bounds to %v", name, out.Bounds())
		}
	}
}

func TestSlitScanStretchesTheSlit(t *testing.T) {
	src := randomImage(20, 5)
	out := SlitScan(src, 0.5, false)
	col := 9
	for y := 0; y < 5; y++ {
		for x := col; x < 20; x++ {
			if out.NRGBAAt(x, y) != src.NRGBAAt(col, y) {
				t.Fatalf("pixel (%d,%d) is not the slit color", x, y)
			}
		}
	}
}
