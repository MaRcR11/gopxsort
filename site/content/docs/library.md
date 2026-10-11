---
title: Go library
weight: 4
---

```bash
go get github.com/MaRcR11/sortcery
```

| Package | |
| ------- | - |
| [`sorter`](https://pkg.go.dev/github.com/MaRcR11/sortcery/sorter) | the pixel sorting engine: keys, intervals, paths, modes |
| [`effects`](https://pkg.go.dev/github.com/MaRcR11/sortcery/effects) | color, glitch, distortion, dithering, artistic and geometry effects |
| [`pipeline`](https://pkg.go.dev/github.com/MaRcR11/sortcery/pipeline) | operation registry, the `op \| op` parser, presets and animation |
| [`imageio`](https://pkg.go.dev/github.com/MaRcR11/sortcery/imageio) | load and save images and animated GIFs |
| [`sortcery`](https://pkg.go.dev/github.com/MaRcR11/sortcery) | shortcuts: `Sort`, `Apply`, `ApplyFile` |

Every example in the [gallery](../gallery) shows the matching Go code next to the command line call.

## Sorting

```go
img, err := imageio.Load("in.jpg")

out := sorter.Sort(img, sorter.Options{
	Key:      sorter.Hue,
	Interval: sorter.Edges(0.2),
	Path:     sorter.Spiral(0.5, 0.5, 6),
	Mode:     sorter.ModeSort,
	Reverse:  true,
	Seed:     42,
})

err = imageio.Save("out.png", out, imageio.Options{})
```

Unset fields use the defaults: lightness key, threshold interval from 0.25 to 0.8, horizontal path, sort mode.

## Effects

Every effect takes an `image.Image` and returns a new `*image.NRGBA`; the input is never modified.

```go
glitched := effects.ChannelShift(out, image.Pt(8, 0), image.Point{}, image.Pt(-8, 0))
retro := effects.Dither(glitched, effects.DitherOptions{
	Method:  effects.Atkinson,
	Palette: effects.MedianCut(img, 16),
})
```

## Pipelines and presets

Describe the whole chain as a string, the same way the command line does:

```go
p, err := pipeline.ParseSeed("@chunky | glitch slices=20 | mix 0.7", 42)
out, err := p.Apply(img)
```

Or in one line, file in and file out:

```go
err := sortcery.ApplyFile("in.jpg", "out.png", "sort angle=45 | vignette")
```

## Animation

```go
anim, err := imageio.LoadAnimation("in.gif")
frames, err := pipeline.Animate(anim.Frames, "sort angle={0..180}", pipeline.AnimateOptions{
	Frames:   30,
	PingPong: true,
})
out := &imageio.Animation{}
for _, f := range frames {
	out.Frames = append(out.Frames, f)
}
err = imageio.SaveAnimation("out.gif", out)
```

## Extending

Custom sort key, a plain function:

```go
redness := func(c color.NRGBA) float64 { return float64(c.R) - float64(c.B) }
out := sorter.Sort(img, sorter.Options{Key: redness})
```

Custom interval and path:

```go
everyTenPixels := sorter.IntervalFunc(func(l *sorter.Line) []sorter.Span {
	var spans []sorter.Span
	for s := 0; s < len(l.Pixels); s += 10 {
		spans = append(spans, sorter.Span{Start: s, End: min(s+10, len(l.Pixels))})
	}
	return spans
})

columns := sorter.PathFunc(func(w, h int) [][]int {
	lines := make([][]int, w)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			lines[x] = append(lines[x], y*w+x)
		}
	}
	return lines
})
```

Add your own function to a pipeline with `pipeline.Func`, or register a named operation so the command line can use it too:

```go
pipeline.Register(pipeline.Spec{
	Name:   "redshift",
	Help:   "Push the image towards red.",
	Params: []pipeline.Param{{Name: "amount", Kind: pipeline.KindFloat, Default: "0.2"}},
	Build: func(a *pipeline.Args) (pipeline.ApplyFunc, error) {
		amount := a.Float("amount")
		return func(img *image.NRGBA) (*image.NRGBA, error) {
			return effects.Map(img, func(c color.NRGBA) color.NRGBA {
				c.R = uint8(min(255, float64(c.R)*(1+amount)))
				return c
			}), nil
		}, nil
	},
})

pipeline.Presets["mylook"] = "redshift 0.3 | @crt"
```
