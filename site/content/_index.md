---
title: sortmancy
layout: hextra-home
---

{{< hextra/hero-badge link="https://github.com/MaRcR11/sortmancy" >}}
  <span>Free and open source</span>
  {{< icon name="arrow-circle-right" attributes="height=14" >}}
{{< /hextra/hero-badge >}}

<div class="hx:mt-6 hx:mb-6">
{{< hextra/hero-headline >}}
  Pixel sorting and glitch art&nbsp;<br class="hx:sm:block hx:hidden" />for Go
{{< /hextra/hero-headline >}}
</div>

<div class="hx:mb-12">
{{< hextra/hero-subtitle >}}
  A fast command line tool and library with 74 operations,&nbsp;<br class="hx:sm:block hx:hidden" />presets, pipelines and animation.
{{< /hextra/hero-subtitle >}}
</div>

<div class="hx:mb-6">
{{< hextra/hero-button text="Get started" link="docs" >}}
{{< hextra/hero-button text="See the gallery" link="docs/gallery" style="background: transparent; border: 1px solid currentColor;" >}}
</div>

<div class="hx:mt-12"></div>

{{< gallery type="carousel" >}}
  {{< gallery-item src="gallery/sort_hue_edges.jpg" caption="Sorted by hue between edges" >}}
  {{< gallery-item src="gallery/sort_path_spiral.jpg" caption="Spiral path" >}}
  {{< gallery-item src="gallery/sort_path_flow_liquid.jpg" caption="Flow field" >}}
  {{< gallery-item src="gallery/sort_mode_channels_all.jpg" caption="Per-channel sort" >}}
  {{< gallery-item src="gallery/polar.jpg" caption="Polar coordinates" >}}
  {{< gallery-item src="gallery/preset_vaporwave.jpg" caption="Vaporwave preset" >}}
  {{< gallery-item src="gallery/glitch.jpg" caption="Slice glitch" >}}
  {{< gallery-item src="gallery/preset_chunky.jpg" caption="Chunky preset" >}}
  {{< gallery-item src="gallery/neon.jpg" caption="Neon edges" >}}
  {{< gallery-item src="gallery/lowpoly.jpg" caption="Low poly" >}}
  {{< gallery-item src="gallery/preset_cinematic.jpg" caption="Cinematic preset" >}}
{{< /gallery >}}

<div class="hx:mt-12"></div>

{{< hextra/feature-grid >}}
  {{< hextra/feature-card
    title="Pixel sorting"
    icon="sort-ascending"
    subtitle="13 sort keys, 5 interval functions, 8 paths including flow fields, spirals and Hilbert curves, 7 span modes, tiles, masks and guide images."
    link="docs/sorting"
  >}}
  {{< hextra/feature-card
    title="74 operations"
    icon="sparkles"
    subtitle="Color grading, blend modes, LUTs, glitch, distortion, dithering, ASCII art, mosaics and more. Every one works from the command line and from Go."
    link="docs/reference"
  >}}
  {{< hextra/feature-card
    title="Pipelines and presets"
    icon="adjustments"
    subtitle="Chain anything with `|`, blend with the original using `mix`, or start from 20 presets such as `@vhs`, `@cinematic` and `@neon`."
    link="docs/cli"
  >}}
  {{< hextra/feature-card
    title="Animation"
    icon="film"
    subtitle="Sweep any parameter with `{from..to}` to render GIFs or frame sequences. Animated GIFs are processed frame by frame."
    link="docs/cli#animation"
  >}}
  {{< hextra/feature-card
    title="A real Go library"
    icon="code"
    subtitle="Typed options, custom sort keys, intervals and paths, your own operations and presets. Only depends on the standard library and golang.org/x/image."
    link="docs/library"
  >}}
  {{< hextra/feature-card
    title="Fast"
    icon="lightning-bolt"
    subtitle="Work is spread across all CPU cores. Sorting a 1080p image takes about 40 ms."
  >}}
{{< /hextra/feature-grid >}}
