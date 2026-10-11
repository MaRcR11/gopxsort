package sortcery_test

import (
	"image"
	"image/color"
	"log"

	"github.com/MaRcR11/sortcery"
	"github.com/MaRcR11/sortcery/effects"
	"github.com/MaRcR11/sortcery/imageio"
	"github.com/MaRcR11/sortcery/pipeline"
	"github.com/MaRcR11/sortcery/sorter"
)

func Example_sort() {
	img, err := imageio.Load("img/city/city.jpg")
	if err != nil {
		log.Fatal(err)
	}
	out := sorter.Sort(img, sorter.Options{
		Key:      sorter.Hue,
		Interval: sorter.Threshold(0.3, 0.9),
		Path:     sorter.Linear(90),
		Reverse:  true,
	})
	if err := imageio.Save("city_hue.png", out, imageio.Options{}); err != nil {
		log.Fatal(err)
	}
}

func Example_custom() {
	img, _ := imageio.Load("img/city/city.jpg")
	redness := func(c color.NRGBA) float64 { return float64(c.R) - float64(c.B) }
	everyTenPixels := sorter.IntervalFunc(func(l *sorter.Line) []sorter.Span {
		var spans []sorter.Span
		for s := 0; s < len(l.Pixels); s += 10 {
			spans = append(spans, sorter.Span{Start: s, End: min(s+10, len(l.Pixels))})
		}
		return spans
	})
	out := sorter.Sort(img, sorter.Options{Key: redness, Interval: everyTenPixels})
	_ = imageio.Save("city_custom.png", out, imageio.Options{})
}

func Example_effects() {
	img, _ := imageio.Load("img/city/city.jpg")
	var out image.Image = sorter.Sort(img, sorter.Options{Interval: sorter.Edges(0.2), Path: sorter.Linear(270)})
	out = effects.ChannelShift(out, image.Pt(6, 0), image.Point{}, image.Pt(-6, 0))
	out = effects.Dither(out, effects.DitherOptions{Method: effects.Atkinson, Palette: effects.Palettes["pico8"]})
	_ = imageio.Save("city_fx.png", out, imageio.Options{})
}

func Example_pipeline() {
	p, err := pipeline.ParseSeed("sort path=concentric mode=smear | glitch slices=20 | scanlines", 42)
	if err != nil {
		log.Fatal(err)
	}
	img, _ := imageio.Load("img/city/city.jpg")
	out, err := p.Apply(img)
	if err != nil {
		log.Fatal(err)
	}
	_ = imageio.Save("city_pipeline.png", out, imageio.Options{})
}

func Example_applyFile() {
	err := sortcery.ApplyFile("img/city/city.jpg", "out.jpg", "sort angle=45 | vignette")
	if err != nil {
		log.Fatal(err)
	}
}
