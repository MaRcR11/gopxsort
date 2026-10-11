package pipeline

import (
	"image"
	"strings"
	"testing"
)

func testImage() *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, 24, 16))
	for i := range img.Pix {
		img.Pix[i] = uint8(i * 7)
	}
	return img
}

func TestEveryOperationRunsWithDefaults(t *testing.T) {
	for _, s := range Specs() {
		op, err := New(s.Name, nil, 1)
		if err != nil {
			t.Fatalf("%s: %v", s.Name, err)
		}
		out, err := op.Apply(testImage())
		if err != nil {
			t.Fatalf("%s: %v", s.Name, err)
		}
		if out == nil {
			t.Fatalf("%s returned nil", s.Name)
		}
	}
}

func TestParse(t *testing.T) {
	p, err := Parse("sort key=hue angle=90 reverse | pixelate 4 | dither bayer4 palette=gameboy")
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != 3 || p[0].Name() != "sort" || p[2].Name() != "dither" {
		t.Fatalf("unexpected pipeline %v", p)
	}
	if got := p.String(); !strings.Contains(got, "pixelate 4") {
		t.Fatalf("String() = %q", got)
	}
	out, err := p.Apply(testImage())
	if err != nil || out.Bounds().Dx() != 24 {
		t.Fatalf("apply: %v", err)
	}
}

func TestParseErrors(t *testing.T) {
	cases := []string{
		"",
		"sort |",
		"nosuchop",
		"sort colour=red",
		"sort key=sparkle",
		"sort angle=steep",
		"pixelate 4 5",
		"rotate 45",
		"swap rgx",
		"shift r=1",
		"dither palette=#zzz",
	}
	for _, c := range cases {
		if _, err := Parse(c); err == nil {
			t.Errorf("Parse(%q) succeeded, want error", c)
		}
	}
}

func TestFuncOp(t *testing.T) {
	called := false
	p := Pipeline{Func("custom", func(img *image.NRGBA) (*image.NRGBA, error) {
		called = true
		return img, nil
	})}
	if _, err := p.Apply(testImage()); err != nil || !called {
		t.Fatal("custom op did not run")
	}
}

func TestRegisterDuplicatePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	Register(Spec{Name: "sort"})
}

func TestPresetsParse(t *testing.T) {
	for _, name := range PresetNames() {
		p, err := Parse("@" + name)
		if err != nil {
			t.Fatalf("@%s: %v", name, err)
		}
		out, err := p.Apply(testImage())
		if err != nil {
			t.Fatalf("@%s: %v", name, err)
		}
		if out.Bounds() != testImage().Bounds() {
			t.Fatalf("@%s changed the size to %v", name, out.Bounds())
		}
	}
	if _, err := Parse("@nope"); err == nil {
		t.Fatal("expected error for unknown preset")
	}
	if p, err := Parse("@vertical | pixelate 2"); err != nil || len(p) != 2 {
		t.Fatalf("preset as a stage: %v %v", p, err)
	}
}

func TestSourceOps(t *testing.T) {
	p, err := Parse("resize scale=0.5 | invert | mix 0 | resize original")
	if err != nil {
		t.Fatal(err)
	}
	out, err := p.Apply(testImage())
	if err != nil {
		t.Fatal(err)
	}
	if out.Bounds() != testImage().Bounds() {
		t.Fatalf("size %v", out.Bounds())
	}
	if out.Pix[0] != testImage().Pix[0] {
		t.Fatal("mix 0 should restore the input")
	}
	if _, err := Parse("resize width=10 | crop x=2 y=2 width=4 height=3"); err != nil {
		t.Fatal(err)
	}
}

func TestExpand(t *testing.T) {
	cases := []struct {
		tpl  string
		t    float64
		want string
	}{
		{"sort lower={0.1..0.5}", 0.5, "sort lower=0.3"},
		{"sort angle={0..90}", 1, "sort angle=90"},
		{"sort angle={0..90} seed={10..20}", 0.25, "sort angle=23 seed=13"},
		{"glitch shift={-0.2..0.2}", 0, "glitch shift=-0.2"},
		{"pixelate 4", 0.7, "pixelate 4"},
	}
	for _, c := range cases {
		if got := Expand(c.tpl, c.t); got != c.want {
			t.Errorf("Expand(%q, %v) = %q, want %q", c.tpl, c.t, got, c.want)
		}
	}
	if !HasRanges("sort angle={0..90}") || HasRanges("sort angle=90") {
		t.Fatal("HasRanges is wrong")
	}
}

func TestAnimate(t *testing.T) {
	frames, err := Animate([]image.Image{testImage()}, "sort angle={0..90} | pixelate {1..4}", AnimateOptions{Frames: 4, PingPong: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 4 {
		t.Fatalf("got %d frames", len(frames))
	}
	if FrameTime(0, 4, true) != 0 || FrameTime(2, 4, true) != 1 {
		t.Fatal("pingpong timing is wrong")
	}
}

func FuzzParse(f *testing.F) {
	for _, s := range []string{"sort key=hue angle=90", "@vhs | pixelate 4", "dither bayer4 palette=#000,#fff", "resize original"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		p, err := Parse(s)
		if err != nil {
			return
		}
		for _, o := range p {
			if o.Name() == "sort" && strings.Contains(s, "mask") {
				return
			}
		}
		_ = p.String()
	})
}
