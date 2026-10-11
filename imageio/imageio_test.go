package imageio

import (
	"bytes"
	"image"
	"image/color"
	"path/filepath"
	"testing"
)

func testImage() *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	for i := range img.Pix {
		img.Pix[i] = uint8(i * 5)
	}
	img.SetNRGBA(3, 4, color.NRGBA{200, 100, 50, 255})
	return img
}

func TestRoundTrip(t *testing.T) {
	for _, name := range []string{"a.png", "b.JPG", "c.jpeg", "d.gif", "e.bmp", "f.tiff", "g.tif"} {
		path := filepath.Join(t.TempDir(), name)
		if err := Save(path, testImage(), Options{Quality: 90}); err != nil {
			t.Fatalf("%s: save: %v", name, err)
		}
		got, err := Load(path)
		if err != nil {
			t.Fatalf("%s: load: %v", name, err)
		}
		if got.Bounds().Dx() != 8 || got.Bounds().Dy() != 8 {
			t.Fatalf("%s: size %v", name, got.Bounds())
		}
	}
}

func TestFormatOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "noext")
	if err := Save(path, testImage(), Options{Format: PNG}); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err != nil {
		t.Fatal(err)
	}
}

func TestFormatFromPath(t *testing.T) {
	if _, err := FormatFromPath("x.xyz"); err == nil {
		t.Fatal("expected error for .xyz")
	}
	if f, err := FormatFromPath("X.JPEG"); err != nil || f != JPEG {
		t.Fatalf("got %v, %v", f, err)
	}
	if f, err := ParseFormat("tif"); err != nil || f != TIFF {
		t.Fatalf("got %v, %v", f, err)
	}
	if err := Save(filepath.Join(t.TempDir(), "x.webp"), testImage(), Options{}); err == nil {
		t.Fatal("expected error writing webp")
	}
}

func TestAnimationRoundTrip(t *testing.T) {
	a := &Animation{}
	for i := 0; i < 3; i++ {
		img := testImage()
		img.Pix[0] = uint8(i * 100)
		a.Frames = append(a.Frames, img)
		a.Delays = append(a.Delays, 5+i)
	}
	var buf bytes.Buffer
	if err := EncodeAnimation(&buf, a); err != nil {
		t.Fatal(err)
	}
	got, err := DecodeAnimation(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Frames) != 3 || got.Delays[2] != 7 {
		t.Fatalf("got %d frames, delays %v", len(got.Frames), got.Delays)
	}
	if got.Frames[1].Bounds().Dx() != 8 {
		t.Fatal("frame size changed")
	}
	if err := SaveAnimation(filepath.Join(t.TempDir(), "x.png"), a); err == nil {
		t.Fatal("expected error saving an animation as png")
	}
}

func TestStillImageAsAnimation(t *testing.T) {
	var buf bytes.Buffer
	if err := Encode(&buf, testImage(), PNG, Options{}); err != nil {
		t.Fatal(err)
	}
	a, err := DecodeAnimation(&buf)
	if err != nil || len(a.Frames) != 1 {
		t.Fatalf("got %v, %v", a, err)
	}
}

func TestPalettedKeepsColors(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 1))
	cols := []color.NRGBA{{255, 0, 0, 255}, {0, 255, 0, 255}, {0, 0, 255, 255}, {255, 255, 255, 255}}
	for x, c := range cols {
		img.SetNRGBA(x, 0, c)
	}
	p := Paletted(img, 256)
	for x, c := range cols {
		r, g, b, _ := p.At(x, 0).RGBA()
		if uint8(r>>8) != c.R || uint8(g>>8) != c.G || uint8(b>>8) != c.B {
			t.Fatalf("pixel %d: got %v want %v", x, p.At(x, 0), c)
		}
	}
}
