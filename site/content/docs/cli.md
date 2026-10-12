---
title: Command line
weight: 2
---

```text
sortmancy <operation> -i <input> -o <output> [--param value | param=value ...]
sortmancy run -i <input> -o <output> "<op> [params] | <op> [params] | ..."
sortmancy @<preset> -i <input> -o <output>
sortmancy -i <input> -o <output> [-t <0-255>]      (classic threshold sort)

sortmancy list          list all operations
sortmancy presets       list all presets
sortmancy help <op>     show the parameters of an operation
sortmancy version       print the version
```

## Options

| Option | Description |
| ------ | ----------- |
| `-i, --input`   | input image, a folder, a glob such as `"photos/*.jpg"`, or `-` for stdin |
| `-o, --output`  | output image (the extension selects the format), a folder for batch input, `frame%03d.png` for numbered frames, or `-` for stdout |
| `-f, --format`  | output format, overriding the extension: `png`, `jpg`, `gif`, `bmp`, `tiff` |
| `-q, --quality` | JPEG quality 1–100 (default 95) |
| `--seed`        | seed for random operations; printed when omitted so results can be reproduced |
| `--frames`      | render an animation with this many frames |
| `--delay`       | frame delay in ms (default 80, or the input GIF's own delays) |
| `--pingpong`    | sweep parameters there and back so the animation loops seamlessly |
| `--quiet`       | no progress output |
| `-t, --threshold` | classic mode only: sort pixels with luminance ≥ t (0–255) |

## Parameters

Parameters of an operation can be written in three ways:

```bash
sortmancy sort -i in.jpg -o out.png --angle 90 --key hue
sortmancy sort -i in.jpg -o out.png angle=90 key=hue
sortmancy pixelate -i in.jpg -o out.png 12
```

The last form uses positional values in the order `sortmancy help <op>` lists them. Flags without a value such as `--reverse` switch a boolean parameter on.
When an operation has a parameter with the same name as a global option, for example `jpeg --quality`, the operation's parameter wins. The short `-q` always sets the output quality.

The [operations reference](../reference) lists every parameter.

## Pipelines

`run` chains operations with `|`. Each stage works on the result of the previous one:

```bash
sortmancy run -i in.jpg -o out.png "sort angle=270 interval=edges edge=0.15 | shift | scanlines"
```

Two operations are special inside pipelines:

- `mix 0.5` blends the current result with the original input, so `pixelate 24 | mix 0.5` applies the effect at half strength.
- `resize original` scales back to the size of the input. `resize scale=0.2 | sort angle=90 | resize original filter=nearest` gives chunky, low-resolution sorting.

## Part of the image

Every operation accepts four common parameters to change only part of the image:

| Parameter | |
| --------- | - |
| `--region x,y,width,height` | only change this rectangle (pixels) |
| `--mask file.png` | only change where the mask is white; gray values blend |
| `--invert-mask` | change the black areas of the mask, or everything outside the region |
| `--feather n` | soften the edge by n pixels |

```bash
sortmancy glitch -i in.jpg -o out.png --region 0,0,1920,430
sortmancy grayscale -i in.jpg -o out.png --region 560,180,800,500 --invert-mask --feather 40
sortmancy run -i in.jpg -o out.png "sort angle=270 region=0,0,960,1080 | neon region=960,0,960,1080"
```

Operations that change the image size, such as `resize` or `ascii`, cannot be masked. For `sort`, `--mask` keeps its own meaning: masked pixels are never moved and also split the sorted spans. [Examples →](../gallery/masks)

## Presets

Presets are ready-made pipelines. Run one directly, or use it as a stage:

```bash
sortmancy @vhs -i in.jpg -o out.png
sortmancy run -i in.jpg -o out.png "@chunky | @crt"
```

`sortmancy presets` prints all of them; the [reference](../reference#presets) and the [gallery](../gallery/presets) show what they do.

## Animation

Write `{from..to}` in any parameter to sweep it across the frames:

```bash
sortmancy run -i in.jpg -o anim.gif --frames 30 --pingpong "sort angle={0..90} lower={0.05..0.5}"
sortmancy run -i in.jpg -o frames/f%03d.png --frames 60 "glitch seed={1..60}"
```

Animated GIF input is processed frame by frame, and parameter ranges still work:

```bash
sortmancy run -i animated.gif -o sorted.gif "sort angle=90"
```

## Batch processing and pipes

```bash
sortmancy run -i "photos/*.jpg" -o sorted/ "sort angle=90"
sortmancy @vaporwave -i photos/ -o out/ --format png
cat in.png | sortmancy @crt -i - -o - > out.png
```

## Reproducible results

Random operations print the seed they used. Pass it back with `--seed` to get exactly the same image again:

```bash
sortmancy glitch -i in.jpg -o out.png --seed 42
```
