---
title: How sorting works
weight: 3
---

A sort is made of four independent parts:

1. A **path** decides how the image is walked: straight lines at any angle, waves, flow fields, spirals and more.
2. An **interval** cuts every line into spans. Only pixels inside a span are moved.
3. A **key** orders the pixels of a span, for example by lightness or hue.
4. A **mode** decides what happens to a span: sort it, shuffle it, average it, and so on.

On top of that, a **mask** limits sorting to part of the image, and **tiles** sort blocks instead of single pixels.

## Path

| Path | Parameters | |
| ---- | ---------- | - |
| `linear` | `angle` | straight lines; 0 = left→right, 90 = bottom→top, any angle works |
| `wave` | `amplitude`, `wavelength`, `vertical` | sine-shaped lines |
| `flow` | `flow-scale`, `warp`, `vertical` | lines bent by a smooth noise field |
| `contour` | `warp`, `contour-blur`, `vertical` | lines bent by the image's brightness, so streaks wrap around bright shapes |
| `spiral` | `spacing`, `cx`, `cy` | a spiral around a center |
| `concentric` | `cx`, `cy` | rings around a center |
| `radial` | `cx`, `cy` | rays leaving a center |
| `hilbert` | | one line along a space-filling Hilbert curve |

[Path examples →](../gallery/paths)

## Interval

| Interval | Parameters | |
| -------- | ---------- | - |
| `threshold` | `lower`, `upper`, `threshold-key` | sort runs of pixels whose value lies between `lower` and `upper` (0–1); this is the default |
| `random` | `length` | spans of random length |
| `waves` | `length` | spans of almost equal length with a random phase per line |
| `edges` | `edge` | sort the regions between edges |
| `none` | | sort every line as a whole |

`--interval-image file.png` takes the intervals from another image instead: its edges or brightness decide where spans start and end.

[Interval examples →](../gallery/intervals)

## Key

`lightness` (default), `luminance`, `hue`, `saturation`, `value` / `brightness`, `intensity`, `minimum`, `chroma`, `red`, `green`, `blue`, `alpha`.

[Key examples →](../gallery/keys)

## Mode

| Mode | |
| ---- | - |
| `sort` | order the span by the key (default); `--reverse` sorts high to low |
| `shuffle` | randomly permute the span |
| `average` | fill the span with its average color |
| `gradient` | replace the span with a gradient between its lowest and highest pixel |
| `smear` | stretch the first pixel across the span |
| `channels` | sort red, green and blue separately |
| `mirror` | lowest values at both ends, highest in the middle |

[Mode examples →](../gallery/modes)

## Mask, tiles and randomness

- `--mask file.png` only sorts where the mask is white; `--invert-mask` sorts the black areas.
- `--tile 24` sorts 24×24 blocks instead of pixels. Tiles support the `sort`, `shuffle` and `mirror` modes.
- `--randomness 0.5` skips half of the spans at random.
- `--seed` makes random intervals, paths and modes reproducible.

[Tile and mask examples →](../gallery/tiles-and-masks)
