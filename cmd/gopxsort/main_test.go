package main

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/MaRcR11/gopxsort/pipeline"
)

func TestParseOptions(t *testing.T) {
	spec, _ := pipeline.Lookup("sort")
	o, err := parseOptions([]string{"-i", "in.png", "--output=out.png", "--angle", "-30", "--reverse", "key=hue", "--vertical", "false", "--seed", "4"}, spec)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"angle=-30", "reverse=true", "key=hue", "vertical=false"}
	if o.input != "in.png" || o.output != "out.png" || o.seed != 4 || !o.seedSet || !slices.Equal(o.rest, want) {
		t.Fatalf("got %+v", o)
	}
	if _, err := parseOptions([]string{"--bogus", "1"}, spec); err == nil {
		t.Fatal("expected error for unknown option")
	}
	j, _ := pipeline.Lookup("jpeg")
	if o, err := parseOptions([]string{"--quality", "40", "-q", "80"}, j); err != nil || o.quality != 80 || !slices.Equal(o.rest, []string{"quality=40"}) {
		t.Fatalf("op parameter vs global flag: %+v %v", o, err)
	}
	b, _ := pipeline.Lookup("brightness")
	if o, err := parseOptions([]string{"-0.2"}, b); err != nil || !slices.Equal(o.rest, []string{"-0.2"}) {
		t.Fatalf("negative positional: %+v %v", o, err)
	}
}

func TestEndToEnd(t *testing.T) {
	quiet = true
	dir := t.TempDir()
	in := filepath.Join(dir, "in.png")
	img := image.NewNRGBA(image.Rect(0, 0, 16, 8))
	for i := range img.Pix {
		img.Pix[i] = uint8(i * 13)
	}
	f, _ := os.Create(in)
	png.Encode(f, img)
	f.Close()

	cmds := [][]string{
		{"-i", in, "-o", filepath.Join(dir, "legacy.jpg"), "-t", "100"},
		{"sort", "-i", in, "-o", filepath.Join(dir, "sort.png"), "--angle", "90", "--key", "hue"},
		{"run", "-i", in, "-o", filepath.Join(dir, "run.png"), "sort path=radial | glitch | dither atkinson palette=cga"},
		{"pixelate", "-i", in, "-o", filepath.Join(dir, "px.gif"), "4"},
		{"@vhs", "-i", in, "-o", filepath.Join(dir, "vhs.bmp")},
		{"run", "-i", in, "-o", filepath.Join(dir, "anim.gif"), "--frames", "3", "--pingpong", "sort lower={0.1..0.6}"},
		{"run", "-i", filepath.Join(dir, "anim.gif"), "-o", filepath.Join(dir, "frame%02d.png"), "invert"},
		{"run", "-i", filepath.Join(dir, "*.png"), "-o", filepath.Join(dir, "batch"), "--format", "jpg", "pixelate 2"},
	}
	for _, args := range cmds {
		if err := run(args, os.Stdout); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
	}
	for _, name := range []string{"anim.gif", "frame00.png", "frame02.png", "batch/in.jpg", "batch/frame01.jpg"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing output: %v", err)
		}
	}
	bad := [][]string{
		{"sort", "-i", in},
		{"sort", "-i", in, "-o", filepath.Join(dir, "x.webp")},
		{"-i", in, "-o", filepath.Join(dir, "x.png"), "-t", "300"},
		{"frobnicate"},
	}
	for _, args := range bad {
		if err := run(args, os.Stdout); err == nil {
			t.Fatalf("%v: expected error", args)
		}
	}
}
