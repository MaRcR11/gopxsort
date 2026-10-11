package imageio

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"github.com/MaRcR11/sortcery/internal/imgutil"
)

type Format string

const (
	PNG  Format = "png"
	JPEG Format = "jpeg"
	GIF  Format = "gif"
	BMP  Format = "bmp"
	TIFF Format = "tiff"
	WEBP Format = "webp"
)

const Stdio = "-"

type Options struct {
	Quality int
	Format  Format
}

type Animation struct {
	Frames    []image.Image
	Delays    []int
	LoopCount int
}

func FormatFromPath(path string) (Format, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		return PNG, nil
	case ".jpg", ".jpeg":
		return JPEG, nil
	case ".gif":
		return GIF, nil
	case ".bmp":
		return BMP, nil
	case ".tif", ".tiff":
		return TIFF, nil
	case ".webp":
		return WEBP, nil
	}
	return "", fmt.Errorf("unsupported image format %q (use .png, .jpg, .gif, .bmp, .tiff or .webp)", filepath.Ext(path))
}

func ParseFormat(s string) (Format, error) {
	return FormatFromPath("x." + strings.TrimPrefix(s, "."))
}

func CanEncode(f Format) bool {
	return f != WEBP && f != ""
}

func IsSupported(path string) bool {
	_, err := FormatFromPath(path)
	return err == nil
}

func readAll(path string) ([]byte, error) {
	if path == Stdio {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func Load(path string) (image.Image, error) {
	data, err := readAll(path)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return img, nil
}

func LoadAnimation(path string) (*Animation, error) {
	data, err := readAll(path)
	if err != nil {
		return nil, err
	}
	return DecodeAnimation(bytes.NewReader(data))
}

func DecodeAnimation(r io.Reader) (*Animation, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(data, []byte("GIF8")) {
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		return &Animation{Frames: []image.Image{img}, Delays: []int{0}}, nil
	}
	g, err := gif.DecodeAll(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return compose(g), nil
}

func compose(g *gif.GIF) *Animation {
	w, h := g.Config.Width, g.Config.Height
	if w == 0 || h == 0 {
		b := g.Image[0].Bounds()
		w, h = b.Max.X, b.Max.Y
	}
	canvas := image.NewNRGBA(image.Rect(0, 0, w, h))
	a := &Animation{LoopCount: g.LoopCount}
	for i, fr := range g.Image {
		disposal := byte(0)
		if i < len(g.Disposal) {
			disposal = g.Disposal[i]
		}
		var prev *image.NRGBA
		if disposal == gif.DisposalPrevious {
			prev = imgutil.ToNRGBA(canvas)
		}
		draw.Draw(canvas, fr.Bounds(), fr, fr.Bounds().Min, draw.Over)
		a.Frames = append(a.Frames, imgutil.ToNRGBA(canvas))
		delay := 0
		if i < len(g.Delay) {
			delay = g.Delay[i]
		}
		a.Delays = append(a.Delays, delay)
		switch disposal {
		case gif.DisposalBackground:
			draw.Draw(canvas, fr.Bounds(), image.Transparent, image.Point{}, draw.Src)
		case gif.DisposalPrevious:
			canvas = prev
		}
	}
	return a
}

func create(path string, fn func(w io.Writer) error) (err error) {
	if path == Stdio {
		w := bufio.NewWriter(os.Stdout)
		if err := fn(w); err != nil {
			return err
		}
		return w.Flush()
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()
	w := bufio.NewWriter(f)
	if err := fn(w); err != nil {
		return err
	}
	return w.Flush()
}

func outputFormat(path string, opts Options) (Format, error) {
	if opts.Format != "" {
		return opts.Format, nil
	}
	if path == Stdio {
		return PNG, nil
	}
	return FormatFromPath(path)
}

func Save(path string, img image.Image, opts Options) error {
	format, err := outputFormat(path, opts)
	if err != nil {
		return err
	}
	if !CanEncode(format) {
		return fmt.Errorf("writing %s is not supported", format)
	}
	return create(path, func(w io.Writer) error { return Encode(w, img, format, opts) })
}

func Encode(w io.Writer, img image.Image, format Format, opts Options) error {
	switch format {
	case PNG:
		return png.Encode(w, img)
	case JPEG:
		q := opts.Quality
		if q <= 0 {
			q = 95
		}
		return jpeg.Encode(w, img, &jpeg.Options{Quality: min(q, 100)})
	case GIF:
		return gif.Encode(w, Paletted(img, 256), nil)
	case BMP:
		return bmp.Encode(w, img)
	case TIFF:
		return tiff.Encode(w, img, &tiff.Options{Compression: tiff.Deflate})
	}
	return fmt.Errorf("writing %s is not supported", format)
}

func SaveAnimation(path string, a *Animation) error {
	if path != Stdio {
		if f, err := FormatFromPath(path); err != nil || f != GIF {
			return fmt.Errorf("animations can only be saved as .gif, got %q", filepath.Ext(path))
		}
	}
	return create(path, func(w io.Writer) error { return EncodeAnimation(w, a) })
}

func EncodeAnimation(w io.Writer, a *Animation) error {
	g := &gif.GIF{LoopCount: a.LoopCount, Image: make([]*image.Paletted, len(a.Frames))}
	imgutil.ParallelEach(len(a.Frames), func(i int) {
		g.Image[i] = Paletted(a.Frames[i], 256)
	})
	for i := range a.Frames {
		delay := 8
		if i < len(a.Delays) && a.Delays[i] > 0 {
			delay = a.Delays[i]
		}
		g.Delay = append(g.Delay, delay)
	}
	return gif.EncodeAll(w, g)
}

func Paletted(img image.Image, colors int) *image.Paletted {
	src := imgutil.ToNRGBA(img)
	transparent := false
	for i := 3; i < len(src.Pix); i += 4 {
		if src.Pix[i] < 128 {
			transparent = true
			break
		}
	}
	if transparent {
		colors--
	}
	pal := imgutil.MedianCut(src, colors)
	opaque := len(pal)
	if transparent {
		pal = append(pal, color.NRGBA{})
	}
	rgb := make([][3]int32, opaque)
	for i, c := range pal[:opaque] {
		n := c.(color.NRGBA)
		rgb[i] = [3]int32{int32(n.R), int32(n.G), int32(n.B)}
	}
	cache := make([]int16, 1<<15)
	for i := range cache {
		cache[i] = -1
	}
	nearest := func(r, g, b int32) uint8 {
		key := (r>>3)<<10 | (g>>3)<<5 | b>>3
		if c := cache[key]; c >= 0 {
			return uint8(c)
		}
		best, bestD := 0, int32(1<<30)
		for i, p := range rgb {
			dr, dg, db := r-p[0], g-p[1], b-p[2]
			if d := dr*dr + dg*dg + db*db; d < bestD {
				best, bestD = i, d
			}
		}
		cache[key] = int16(best)
		return uint8(best)
	}

	w, h := src.Rect.Dx(), src.Rect.Dy()
	dst := image.NewPaletted(src.Rect, pal)
	cur := make([][3]int32, w+2)
	next := make([][3]int32, w+2)
	clamp := func(v int32) int32 { return max(0, min(255, v)) }
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := (y*w + x) * 4
			if transparent && src.Pix[i+3] < 128 {
				dst.Pix[y*dst.Stride+x] = uint8(opaque)
				continue
			}
			e := cur[x+1]
			r := clamp(int32(src.Pix[i]) + e[0]/16)
			g := clamp(int32(src.Pix[i+1]) + e[1]/16)
			b := clamp(int32(src.Pix[i+2]) + e[2]/16)
			idx := nearest(r, g, b)
			dst.Pix[y*dst.Stride+x] = idx
			p := rgb[idx]
			er := [3]int32{r - p[0], g - p[1], b - p[2]}
			for c := 0; c < 3; c++ {
				cur[x+2][c] += er[c] * 7
				next[x][c] += er[c] * 3
				next[x+1][c] += er[c] * 5
				next[x+2][c] += er[c]
			}
		}
		cur, next = next, cur
		clear(next)
	}
	return dst
}
