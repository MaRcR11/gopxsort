package pipeline

import (
	"fmt"
	"image"

	"github.com/MaRcR11/sortcery/effects"
)

func registerExtra() {
	Register(Spec{Name: "quantize", Help: "Reduce to the N most representative colors (median cut).",
		Params: []Param{
			integer("colors", "16", "number of colors"),
			str("method", "none", "dithering algorithm", effects.DitherNames()...),
		},
		Build: func(a *Args) (ApplyFunc, error) {
			n := a.Int("colors")
			m := choice(a, "method", effects.DitherByName, effects.DitherNames())
			if n < 1 {
				return nil, fmt.Errorf("quantize: colors must be at least 1")
			}
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Quantize(img, n, m) }), nil
		}})
	Register(Spec{Name: "resize", Help: "Scale the image; \"resize original\" restores the input size.",
		Params: []Param{
			float("scale", "0.5", "scale factor, used when width and height are not given"),
			integer("width", "0", "target width; 0 keeps the aspect ratio"),
			integer("height", "0", "target height; 0 keeps the aspect ratio"),
			boolean("original", "resize back to the size of the pipeline input"),
			str("filter", "bilinear", "sampling filter", effects.FilterNames()...),
		},
		BuildSource: func(a *Args) (SourceFunc, error) {
			scale, w, h, orig := a.Float("scale"), a.Int("width"), a.Int("height"), a.Bool("original")
			f := choice(a, "filter", effects.FilterByName, effects.FilterNames())
			if !orig && w <= 0 && h <= 0 && scale <= 0 {
				return nil, fmt.Errorf("resize: scale must be positive")
			}
			return func(img, src *image.NRGBA) (*image.NRGBA, error) {
				cw, ch := img.Rect.Dx(), img.Rect.Dy()
				tw, th := w, h
				switch {
				case orig:
					tw, th = src.Rect.Dx(), src.Rect.Dy()
				case tw > 0 && th <= 0:
					th = max(tw*ch/max(cw, 1), 1)
				case th > 0 && tw <= 0:
					tw = max(th*cw/max(ch, 1), 1)
				case tw <= 0 && th <= 0:
					return effects.Scale(img, scale, f), nil
				}
				return effects.Resize(img, tw, th, f), nil
			}, nil
		}})
	Register(Spec{Name: "crop", Help: "Cut out a rectangle.",
		Params: []Param{
			integer("x", "0", "left edge"),
			integer("y", "0", "top edge"),
			integer("width", "0", "width; 0 extends to the right edge"),
			integer("height", "0", "height; 0 extends to the bottom edge"),
		},
		Build: func(a *Args) (ApplyFunc, error) {
			x, y, w, h := a.Int("x"), a.Int("y"), a.Int("width"), a.Int("height")
			return func(img *image.NRGBA) (*image.NRGBA, error) {
				r := img.Rect
				x1, y1 := r.Max.X, r.Max.Y
				if w > 0 {
					x1 = x + w
				}
				if h > 0 {
					y1 = y + h
				}
				out := effects.Crop(img, image.Rect(x, y, x1, y1))
				if out.Rect.Empty() {
					return nil, fmt.Errorf("crop: rectangle lies outside the %dx%d image", r.Dx(), r.Dy())
				}
				return out, nil
			}, nil
		}})
	Register(Spec{Name: "mix", Help: "Blend the current result with the pipeline input.",
		Params: []Param{float("amount", "0.5", "0 = only the input, 1 = only the result")},
		BuildSource: func(a *Args) (SourceFunc, error) {
			v := a.Float("amount")
			return func(img, src *image.NRGBA) (*image.NRGBA, error) {
				return effects.Mix(src, img, v), nil
			}, nil
		}})
	Register(Spec{Name: "halftone", Help: "Print-style dots sized by brightness.",
		Params: []Param{integer("size", "8", "dot cell size in pixels"), boolean("color", "colored dots on black instead of black dots on white")},
		Build: func(a *Args) (ApplyFunc, error) {
			s, c := a.Int("size"), a.Bool("color")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Halftone(img, s, c) }), nil
		}})
	Register(Spec{Name: "bloom", Help: "Glow around bright areas.",
		Params: []Param{
			float("threshold", "0.6", "brightness (0..1) where glow starts"),
			integer("radius", "8", "glow radius in pixels"),
			float("intensity", "1", "glow strength"),
		},
		Build: func(a *Args) (ApplyFunc, error) {
			t, r, i := a.Float("threshold"), a.Int("radius"), a.Float("intensity")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Bloom(img, t, r, i) }), nil
		}})
	Register(Spec{Name: "blocks", Help: "Replace random blocks with copies of other blocks.",
		Params: []Param{integer("size", "32", "block size in pixels"), float("amount", "0.1", "fraction of blocks replaced (0..1)"), seedParam},
		Build: func(a *Args) (ApplyFunc, error) {
			s, v, seed := a.Int("size"), a.Float("amount"), a.Seed()
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.BlockShuffle(img, s, v, seed) }), nil
		}})
	Register(Spec{Name: "kaleidoscope", Help: "Mirror the image into rotating segments around the center.",
		Params: []Param{integer("segments", "6", "number of mirrored segments"), float("rotation", "0", "rotation in degrees")},
		Build: func(a *Args) (ApplyFunc, error) {
			n, r := a.Int("segments"), a.Float("rotation")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Kaleidoscope(img, n, r) }), nil
		}})
	Register(Spec{Name: "oilpaint", Help: "Painterly smoothing that keeps edges (Kuwahara filter).",
		Params: []Param{integer("radius", "4", "brush size in pixels")},
		Build: func(a *Args) (ApplyFunc, error) {
			r := a.Int("radius")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.OilPaint(img, r) }), nil
		}})
}
