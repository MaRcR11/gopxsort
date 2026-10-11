# gopxsort

Pixel sorting and glitch art for Go: a command line tool and a library.

[![Go Reference](https://pkg.go.dev/badge/github.com/MaRcR11/gopxsort.svg)](https://pkg.go.dev/github.com/MaRcR11/gopxsort)
[![CI](https://github.com/MaRcR11/gopxsort/actions/workflows/ci.yml/badge.svg)](https://github.com/MaRcR11/gopxsort/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Original | Sorted by hue between edges | Spiral sort | Animated sweep
-------- | ----------------------------- | ----------- | --------------
![Original](img/city/city.jpg) | ![](img/examples/sort_hue_edges.jpg) | ![](img/examples/sort_path_spiral.jpg) | ![](img/examples/anim_threshold.gif)

- **Pixel sorting**, modelled on the usual tools (satyarth/pixelsort, ASDF, Acerola):
  - 13 sort keys: lightness, luminance, hue, saturation, value, chroma, single channels, and more.
  - Interval functions: threshold, random, waves, edges, none.
  - Paths: straight lines at any angle, sine waves, spirals, concentric rings, radial rays.
  - Span modes: sort, shuffle, average, gradient, smear.
  - Masks, reverse order, randomness, reproducible seeds.
- **39 operations** in total:
  - Color: invert, posterize, hue, gradient map, ...
  - Glitch: RGB shift, chromatic aberration, slice glitch, block shuffle, JPEG corruption, melt, wave, scanlines, noise.
  - Retro: 10 dithering algorithms with retro palettes, median-cut quantization.
  - Artistic: halftone, bloom, oil paint, kaleidoscope.
  - Geometry: resize, crop, pixelate, mirror, flip, rotate.
  - Filters: blur, sharpen, edges, vignette.
- **Pipelines**: chain operations, e.g. `"sort angle=90 | glitch | dither palette=gameboy"`.
  `mix` blends the result with the original, and `resize original` scales back to the input size.
- **14 presets**: ready-made looks such as `@vhs`, `@vaporwave`, `@crt`, `@gameboy` and `@chunky`.
- **Animation**: sweep any parameter with `{from..to}` to create GIFs or frame sequences.
  Animated GIF input is processed frame by frame.
- **Formats**: reads PNG, JPEG, GIF, BMP, TIFF and WebP; writes all of them except WebP.
  stdin/stdout are supported.
- **Batch processing**: process whole folders or glob patterns in one call.
- **Fast**: work is spread across all CPU cores. Sorting a 1080p image takes about 40 ms.
- **Minimal dependencies**: the standard library plus `golang.org/x/image`.

## Installation

Precompiled binaries for Windows, Linux and macOS are on the [releases](https://github.com/MaRcR11/gopxsort/releases) page.
With Go 1.21 or newer:

```bash
go install github.com/MaRcR11/gopxsort/cmd/gopxsort@latest
```

Use it as a library:

```bash
go get github.com/MaRcR11/gopxsort
```

## Command line

```text
gopxsort <operation> -i <input> -o <output> [--param value | param=value ...]
gopxsort run -i <input> -o <output> "<op> [params] | <op> [params] | ..."
gopxsort @<preset> -i <input> -o <output>
gopxsort -i <input> -o <output> [-t <0-255>]      (classic threshold sort)

gopxsort list          list all operations
gopxsort presets       list all presets
gopxsort help <op>     show the parameters of an operation
```

| Option | Description |
| ------ | ----------- |
| `-i, --input`   | input image, a folder, a glob such as `"photos/*.jpg"`, or `-` for stdin |
| `-o, --output`  | output image (the extension selects the format), a folder for batch input, `frame%03d.png` for numbered frames, or `-` for stdout |
| `-f, --format`  | output format, overriding the extension: png, jpg, gif, bmp, tiff |
| `-q, --quality` | JPEG quality 1–100 (default 95) |
| `--seed`        | seed for random operations; printed when omitted so results can be reproduced |
| `--frames`      | render an animation with this many frames |
| `--delay`       | frame delay in ms (default 80, or the input GIF's own delays) |
| `--pingpong`    | sweep parameters there and back so the animation loops seamlessly |
| `--quiet`       | no progress output |
| `-t, --threshold` | classic mode: sort pixels with luminance ≥ t (0–255) |

You can write parameters as `--name value`, as `name=value`, or positionally in the order that `gopxsort help <op>` lists them:

```bash
gopxsort sort -i in.jpg -o out.png --angle 90 --key hue
gopxsort sort -i in.jpg -o out.png interval=edges edge=0.2 reverse
gopxsort pixelate -i in.jpg -o out.png 12
gopxsort run -i in.jpg -o out.png "sort angle=45 lower=0.4 | shift | scanlines"
gopxsort @vhs -i in.jpg -o out.png
gopxsort run -i in.jpg -o out.png "@chunky | @crt"
```

### Animation

Write `{from..to}` in any parameter to sweep it across the frames:

```bash
gopxsort run -i in.jpg -o anim.gif --frames 30 --pingpong "sort angle={0..90} lower={0.05..0.5}"
gopxsort run -i in.jpg -o frames/f%03d.png --frames 60 "glitch seed={1..60}"
gopxsort run -i animated.gif -o sorted.gif "sort angle=90"
```

### Batch and pipes

```bash
gopxsort run -i "photos/*.jpg" -o sorted/ "sort angle=90"
gopxsort @vaporwave -i photos/ -o out/ --format png
cat in.png | gopxsort @crt -i - -o - > out.png
```

The original interface still works: `gopxsort -i img/city/city.jpg -o out.jpg -t 128`.

### Sorting

`gopxsort help sort` lists all parameters. The main ones:

| Parameter | Values | Default | |
| --------- | ------ | ------- | - |
| `key` | lightness, luminance, hue, saturation, value/brightness, intensity, minimum, chroma, red, green, blue, alpha | lightness | what pixels are ordered by |
| `interval` | threshold, random, waves, edges, none | threshold | how lines are cut into sorted spans |
| `lower`, `upper` | 0–1 | 0.25, 0.8 | threshold: only pixels in this range are sorted |
| `threshold-key` | any key | lightness | threshold: which value is compared |
| `length` | pixels | 50 | random / waves: span length |
| `edge` | 0–1 | 0.25 | edges: edge strength that stops a span |
| `path` | linear, wave, spiral, concentric, radial | linear | how the image is traversed |
| `angle` | degrees | 0 | linear: 0 = left→right, 90 = bottom→top, any angle works |
| `amplitude`, `wavelength`, `vertical` | | 20, 200, false | wave path |
| `spacing` | pixels | 4 | spiral path: distance between turns |
| `cx`, `cy` | 0–1 | 0.5, 0.5 | center of spiral / concentric / radial paths |
| `mode` | sort, shuffle, average, gradient, smear | sort | what happens to each span |
| `reverse` | flag | | sort high→low |
| `randomness` | 0–1 | 0 | chance to skip a span |
| `mask`, `invert-mask` | image file | | only sort the white (or black) areas of a mask |

### All operations

| Group | Operations |
| ----- | ---------- |
| Sorting | `sort` |
| Color | `invert`, `grayscale`, `threshold`, `posterize`, `brightness`, `contrast`, `gamma`, `saturation`, `hue`, `solarize`, `swap` (channel swap), `gradientmap` |
| Glitch | `shift` (RGB split), `aberration`, `glitch` (slice displacement), `blocks` (block shuffle), `jpeg` (data corruption), `melt`, `wave`, `scanlines`, `noise` |
| Retro | `dither` with `floyd`, `atkinson`, `jarvis`, `stucki`, `burkes`, `sierra`, `bayer2/4/8` or `none`, palettes `bw`, `gray4`, `gameboy`, `cga`, `pico8`, `c64`, `sepia`, `vapor`, `websafe` or your own `#hex,#hex,...`; `quantize` (best N colors) |
| Artistic | `halftone`, `bloom`, `oilpaint`, `kaleidoscope` |
| Geometry | `resize`, `crop`, `pixelate`, `mirror`, `flip`, `rotate` |
| Filters & blending | `blur`, `sharpen`, `edges`, `vignette`, `mix` |

### Presets

| Preset | Pipeline |
| ------ | -------- |
| `@classic` | `sort key=luminance threshold-key=luminance lower=0.5 upper=1` |
| `@vertical` | `sort angle=270 interval=edges edge=0.2` |
| `@spiral` | `sort path=spiral spacing=6 interval=edges edge=0.2` |
| `@melt` | `sort angle=270 interval=edges edge=0.15 mode=smear \| melt shift=120` |
| `@glitch` | `sort lower=0.35 randomness=0.5 \| glitch slices=30 \| shift \| jpeg amount=2` |
| `@vhs` | `shift \| wave \| noise \| scanlines \| saturation` |
| `@crt` | `bloom \| aberration \| scanlines \| vignette` |
| `@vaporwave` | `sort key=hue angle=90 \| gradientmap \| mix 0.6 \| scanlines` |
| `@dream` | `sort path=concentric \| bloom \| saturation` |
| `@gameboy` | `resize scale=0.33 \| dither atkinson palette=gameboy \| resize original filter=nearest` |
| `@8bit` | `resize scale=0.25 \| quantize colors=16 \| resize original filter=nearest` |
| `@chunky` | `resize scale=0.125 \| sort angle=90 \| resize original filter=nearest` |
| `@newspaper` | `grayscale \| contrast 1.4 \| halftone size=6` |
| `@oil` | `oilpaint radius=4 \| saturation 1.2` |

`gopxsort presets` prints the full pipelines.

## Library

```go
import (
	"github.com/MaRcR11/gopxsort/effects"
	"github.com/MaRcR11/gopxsort/imageio"
	"github.com/MaRcR11/gopxsort/pipeline"
	"github.com/MaRcR11/gopxsort/sorter"
)

img, _ := imageio.Load("in.jpg")

// Pixel sort with full control.
out := sorter.Sort(img, sorter.Options{
	Key:      sorter.Hue,
	Interval: sorter.Edges(0.2),
	Path:     sorter.Spiral(0.5, 0.5, 6),
	Mode:     sorter.ModeSort,
	Reverse:  true,
	Seed:     42,
})

// Apply effects. Every effect takes an image.Image and returns a new *image.NRGBA.
glitched := effects.ChannelShift(out, image.Pt(8, 0), image.Point{}, image.Pt(-8, 0))
retro := effects.Dither(glitched, effects.DitherOptions{
	Method:  effects.Atkinson,
	Palette: effects.MedianCut(img, 16),
})

_ = imageio.Save("out.png", retro, imageio.Options{})
```

Or describe the whole chain as a string, the same way the CLI does:

```go
p, err := pipeline.ParseSeed("@chunky | glitch slices=20 | mix 0.7", 42)
out, err := p.Apply(img)

// shortest form
err = gopxsort.ApplyFile("in.jpg", "out.png", "sort angle=45 | vignette")
```

Animations:

```go
anim, _ := imageio.LoadAnimation("in.gif")
frames, _ := pipeline.Animate(anim.Frames, "sort angle={0..180}", pipeline.AnimateOptions{Frames: 30, PingPong: true})
out := &imageio.Animation{}
for _, f := range frames {
	out.Frames = append(out.Frames, f)
}
_ = imageio.SaveAnimation("out.gif", out)
```

**Extending:**
- `sorter.Key` is a plain `func(color.NRGBA) float64`.
- `sorter.IntervalFunc` and `sorter.PathFunc` let you define your own spans and traversals.
- `pipeline.Func` adds any Go function to a pipeline, and `pipeline.SourceFuncOp` adds one that also receives the original input.
- `pipeline.Register` makes a new operation available by name, and the CLI then picks it up too.
- Add an entry to `pipeline.Presets` to define your own `@preset`.

## Gallery

All examples start from [img/city/city.jpg](img/city/city.jpg).
**[docs/EXAMPLES.md](docs/EXAMPLES.md)** has 98 examples covering every sort option, operation, preset and animation, each with the full command and the equivalent Go library code.

| | | |
|-|-|-|
| ![](img/examples/sort_vertical.jpg)<br>`sort --angle 270 --lower 0.1 --upper 0.9` | ![](img/examples/sort_random_vertical.jpg)<br>`sort --interval random --length 100 --angle 270 --reverse` | ![](img/examples/sort_mode_smear.jpg)<br>`sort --mode smear --interval edges --edge 0.12 --angle 270` |
| ![](img/examples/sort_path_concentric.jpg)<br>`sort --path concentric --lower 0.1` | ![](img/examples/sort_path_wave.jpg)<br>`sort --path wave --amplitude 40 --wavelength 300 --lower 0.1` | ![](img/examples/sort_mode_gradient.jpg)<br>`sort --mode gradient --interval random --length 60 --angle 90` |
| ![](img/examples/glitch.jpg)<br>`glitch --slices 30 --shift 0.15` | ![](img/examples/jpeg.jpg)<br>`jpeg --amount 15 --quality 40 --seed 2` | ![](img/examples/melt.jpg)<br>`melt --shift 150` |
| ![](img/examples/dither.jpg)<br>`dither --method atkinson --palette gameboy` | ![](img/examples/kaleidoscope.jpg)<br>`kaleidoscope --segments 6` | ![](img/examples/halftone.jpg)<br>`halftone --size 8 --color` |
| ![](img/examples/preset_vaporwave.jpg)<br>`@vaporwave` | ![](img/examples/preset_crt.jpg)<br>`@crt` | ![](img/examples/preset_glitch.jpg)<br>`@glitch` |
| ![](img/examples/pipeline_blocky.jpg)<br>`run "resize scale=0.2 \| sort angle=90 lower=0.1 \| resize original filter=nearest \| aberration"` | ![](img/examples/preset_chunky.jpg)<br>`@chunky` | ![](img/examples/anim_kaleidoscope.gif)<br>`run --frames 16 --pingpong "resize width=480 \| kaleidoscope rotation={0..60}"` |

## Project structure

```text
gopxsort/
├── gopxsort.go          shortcuts: Sort, Apply, ApplyFile
├── sorter/              pixel sorting engine: keys, intervals, paths, modes
├── effects/             color, glitch, dithering, artistic and geometry effects
├── pipeline/            operation registry, "op | op" parser, presets, animation
├── imageio/             load and save images and animated GIFs
├── internal/imgutil/    shared low-level helpers (not public API)
├── cmd/gopxsort/        the command line tool
├── docs/EXAMPLES.md     every example with CLI command and library code
└── img/                 sample image and gallery
```

## Development

```bash
go test ./...                                   # run tests
go test -race ./...                             # run tests with the race detector
go test -bench . ./sorter                       # benchmarks
go test -fuzz FuzzParse -fuzztime 30s ./pipeline   # fuzz the pipeline parser
```

## License

[MIT](LICENSE)
