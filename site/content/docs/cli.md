---
title: Command line
weight: 2
---

```text
sortcery <operation> -i <input> -o <output> [--param value | param=value ...]
sortcery run -i <input> -o <output> "<op> [params] | <op> [params] | ..."
sortcery @<preset> -i <input> -o <output>
sortcery -i <input> -o <output> [-t <0-255>]      (classic threshold sort)

sortcery list          list all operations
sortcery presets       list all presets
sortcery help <op>     show the parameters of an operation
sortcery version       print the version
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
sortcery sort -i in.jpg -o out.png --angle 90 --key hue
sortcery sort -i in.jpg -o out.png angle=90 key=hue
sortcery pixelate -i in.jpg -o out.png 12
```

The last form uses positional values in the order `sortcery help <op>` lists them. Flags without a value such as `--reverse` switch a boolean parameter on.
When an operation has a parameter with the same name as a global option, for example `jpeg --quality`, the operation's parameter wins. The short `-q` always sets the output quality.

The [operations reference](../reference) lists every parameter.

## Pipelines

`run` chains operations with `|`. Each stage works on the result of the previous one:

```bash
sortcery run -i in.jpg -o out.png "sort angle=270 interval=edges edge=0.15 | shift | scanlines"
```

Two operations are special inside pipelines:

- `mix 0.5` blends the current result with the original input, so `pixelate 24 | mix 0.5` applies the effect at half strength.
- `resize original` scales back to the size of the input. `resize scale=0.2 | sort angle=90 | resize original filter=nearest` gives chunky, low-resolution sorting.

## Presets

Presets are ready-made pipelines. Run one directly, or use it as a stage:

```bash
sortcery @vhs -i in.jpg -o out.png
sortcery run -i in.jpg -o out.png "@chunky | @crt"
```

`sortcery presets` prints all of them; the [reference](../reference#presets) and the [gallery](../gallery/presets) show what they do.

## Animation

Write `{from..to}` in any parameter to sweep it across the frames:

```bash
sortcery run -i in.jpg -o anim.gif --frames 30 --pingpong "sort angle={0..90} lower={0.05..0.5}"
sortcery run -i in.jpg -o frames/f%03d.png --frames 60 "glitch seed={1..60}"
```

Animated GIF input is processed frame by frame, and parameter ranges still work:

```bash
sortcery run -i animated.gif -o sorted.gif "sort angle=90"
```

## Batch processing and pipes

```bash
sortcery run -i "photos/*.jpg" -o sorted/ "sort angle=90"
sortcery @vaporwave -i photos/ -o out/ --format png
cat in.png | sortcery @crt -i - -o - > out.png
```

## Reproducible results

Random operations print the seed they used. Pass it back with `--seed` to get exactly the same image again:

```bash
sortcery glitch -i in.jpg -o out.png --seed 42
```
