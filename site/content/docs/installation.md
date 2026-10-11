---
title: Installation
weight: 1
---

## Command line tool

{{< tabs >}}
{{< tab name="Go" >}}
With Go 1.21 or newer:

```bash
go install github.com/MaRcR11/sortcery/cmd/sortcery@latest
```
{{< /tab >}}
{{< tab name="Binary" >}}
Download a precompiled binary for Windows, Linux or macOS from the
[releases page](https://github.com/MaRcR11/sortcery/releases), unpack it and put `sortcery` on your `PATH`.
{{< /tab >}}
{{< tab name="From source" >}}
```bash
git clone https://github.com/MaRcR11/sortcery.git
cd sortcery
go build -o sortcery ./cmd/sortcery
```
{{< /tab >}}
{{< /tabs >}}

Check that it works:

```bash
sortcery version
sortcery list
```

## Library

```bash
go get github.com/MaRcR11/sortcery
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
