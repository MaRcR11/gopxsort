---
title: Installation
weight: 1
---

## Command line tool

{{< tabs >}}
{{< tab name="Go" >}}
With Go 1.21 or newer:

```bash
go install github.com/MaRcR11/sortmancy/cmd/sortmancy@latest
```
{{< /tab >}}
{{< tab name="Binary" >}}
Download a precompiled binary for Windows, Linux or macOS from the
[releases page](https://github.com/MaRcR11/sortmancy/releases), unpack it and put `sortmancy` on your `PATH`.
{{< /tab >}}
{{< tab name="From source" >}}
```bash
git clone https://github.com/MaRcR11/sortmancy.git
cd sortmancy
go build -o sortmancy ./cmd/sortmancy
```
{{< /tab >}}
{{< tab name="Docker" >}}
The image contains only the binary (about 4 MB). Mount the folder with your images to `/work`:

```bash
docker run --rm -v "$PWD:/work" ghcr.io/marcr11/sortmancy @vhs -i in.jpg -o out.png
```
{{< /tab >}}
{{< /tabs >}}

Check that it works:

```bash
sortmancy version
sortmancy list
```

## Library

```bash
go get github.com/MaRcR11/sortmancy
```

The library only depends on the Go standard library and `golang.org/x/image`.

## Supported formats

| Format | Read | Write |
| ------ | ---- | ----- |
| PNG | yes | yes |
| JPEG | yes | yes |
| GIF, including animations | yes | yes |
| BMP | yes | yes |
| TIFF | yes | yes |
| WebP | yes | no |
