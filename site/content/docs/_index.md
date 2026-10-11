---
title: Introduction
weight: 1
---

sortcery is a pixel sorting and glitch art toolkit for Go. You can use it as a command line tool or import it as a library.

![Sorted by hue between edges](/gallery/sort_hue_edges.jpg)

## Quick start

Install the command line tool:

```bash
go install github.com/MaRcR11/sortcery/cmd/sortcery@latest
```

Sort an image from top to bottom by hue:

```bash
sortcery sort -i photo.jpg -o sorted.png --angle 270 --key hue
```

Chain several operations, or use a preset:

```bash
sortcery run -i photo.jpg -o out.png "sort angle=45 lower=0.4 | shift | scanlines"
sortcery @vaporwave -i photo.jpg -o out.png
```

Or use it from Go:

```go
img, _ := imageio.Load("photo.jpg")
out := sorter.Sort(img, sorter.Options{
	Key:  sorter.Hue,
	Path: sorter.Linear(270),
})
_ = imageio.Save("sorted.png", out, imageio.Options{})
```

## Next steps

{{< cards >}}
  {{< card link="installation" title="Installation" icon="download" >}}
  {{< card link="cli" title="Command line" icon="terminal" >}}
  {{< card link="sorting" title="How sorting works" icon="sort-ascending" >}}
  {{< card link="library" title="Go library" icon="code" >}}
  {{< card link="reference" title="Operations reference" icon="book-open" >}}
  {{< card link="gallery" title="Gallery" icon="photograph" >}}
{{< /cards >}}
