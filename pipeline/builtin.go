package pipeline

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"

	"github.com/MaRcR11/sortmancy/effects"
	"github.com/MaRcR11/sortmancy/imageio"
	"github.com/MaRcR11/sortmancy/sorter"
)

func str(name, def, help string, choices ...string) Param {
	return Param{Name: name, Kind: KindString, Default: def, Help: help, Choices: choices}
}
func integer(name, def, help string) Param {
	return Param{Name: name, Kind: KindInt, Default: def, Help: help}
}
func float(name, def, help string) Param {
	return Param{Name: name, Kind: KindFloat, Default: def, Help: help}
}
func boolean(name, help string) Param {
	return Param{Name: name, Kind: KindBool, Default: "false", Help: help}
}

var seedParam = Param{Name: "seed", Kind: KindInt, Help: "random seed (default: derived from the global seed)"}

func simple(fn func(img *image.NRGBA) *image.NRGBA) ApplyFunc {
	return func(img *image.NRGBA) (*image.NRGBA, error) { return fn(img), nil }
}

func choice[T any](a *Args, name string, lookup func(string) (T, bool), names []string) T {
	v, ok := lookup(a.String(name))
	if !ok {
		a.Fail(fmt.Errorf("%s: unknown %s %q (choose from %s)", a.spec.Name, name, a.String(name), strings.Join(names, ", ")))
	}
	return v
}

func init() {
	registerSort()
	registerColor()
	registerGlitch()
	registerFilters()
	registerExtra()
	registerDistort()
	registerColorMore()
	registerArtisticMore()
}

var intervalNames = []string{"threshold", "random", "waves", "edges", "none"}
var pathNames = []string{"linear", "wave", "concentric", "radial", "spiral", "flow", "contour", "hilbert"}

func registerSort() {
	Register(Spec{
		Name: "sort",
		Help: "Pixel sort: walk the image along a path, cut each line into intervals and sort them.",
		Params: []Param{
			str("key", "lightness", "what pixels are sorted by", sorter.KeyNames()...),
			str("interval", "threshold", "how lines are split into sorted spans", intervalNames...),
			float("lower", "0.25", "threshold interval: lower bound (0..1)"),
			float("upper", "0.8", "threshold interval: upper bound (0..1)"),
			str("threshold-key", "lightness", "threshold interval: value compared with lower/upper", sorter.KeyNames()...),
			integer("length", "50", "random/waves interval: typical span length in pixels"),
			float("edge", "0.25", "edges interval: edge strength (0..1) that stops a span"),
			str("path", "linear", "how the image is traversed", pathNames...),
			float("angle", "0", "linear path: direction in degrees (0 = left→right, 90 = bottom→top)"),
			float("amplitude", "20", "wave path: amplitude in pixels"),
			float("wavelength", "200", "wave path: wavelength in pixels"),
			boolean("vertical", "wave/flow/contour path: run the lines vertically"),
			float("cx", "0.5", "concentric/radial/spiral path: center x (0..1)"),
			float("cy", "0.5", "concentric/radial/spiral path: center y (0..1)"),
			float("spacing", "4", "spiral path: distance between spiral turns in pixels"),
			float("flow-scale", "150", "flow path: size of the swirls in pixels"),
			float("warp", "60", "flow/contour path: how far lines bend, in pixels"),
			integer("contour-blur", "4", "contour path: smoothing of the brightness that bends the lines"),
			str("mode", "sort", "what happens to each span", sorter.ModeNames()...),
			boolean("reverse", "sort from high to low"),
			float("randomness", "0", "probability (0..1) that a span is skipped"),
			integer("tile", "1", "sort square tiles of this size instead of single pixels"),
			str("interval-image", "", "image file whose pixels decide the intervals instead of the input"),
			str("mask", "", "image file; only its bright areas are sorted"),
			boolean("invert-mask", "sort the dark areas of the mask instead"),
			seedParam,
		},
		Build: func(a *Args) (ApplyFunc, error) {
			o := sorter.Options{
				Key:        choice(a, "key", sorter.KeyByName, sorter.KeyNames()),
				Mode:       choice(a, "mode", sorter.ModeByName, sorter.ModeNames()),
				Reverse:    a.Bool("reverse"),
				Randomness: a.Float("randomness"),
				InvertMask: a.Bool("invert-mask"),
				Seed:       a.Seed(),
				Tile:       a.Int("tile"),
			}
			switch a.String("interval") {
			case "threshold":
				tk := choice(a, "threshold-key", sorter.KeyByName, sorter.KeyNames())
				o.Interval = sorter.ThresholdBy(tk, a.Float("lower"), a.Float("upper"))
			case "random":
				o.Interval = sorter.Random(a.Int("length"))
			case "waves":
				o.Interval = sorter.Waves(a.Int("length"))
			case "edges":
				o.Interval = sorter.Edges(a.Float("edge"))
			case "none":
				o.Interval = sorter.None()
			default:
				return nil, fmt.Errorf("sort: unknown interval %q (choose from %s)", a.String("interval"), strings.Join(intervalNames, ", "))
			}
			switch a.String("path") {
			case "linear":
				o.Path = sorter.Linear(a.Float("angle"))
			case "wave":
				o.Path = sorter.Wave(a.Float("amplitude"), a.Float("wavelength"), a.Bool("vertical"))
			case "concentric":
				o.Path = sorter.Concentric(a.Float("cx"), a.Float("cy"))
			case "radial":
				o.Path = sorter.Radial(a.Float("cx"), a.Float("cy"))
			case "spiral":
				o.Path = sorter.Spiral(a.Float("cx"), a.Float("cy"), a.Float("spacing"))
			case "flow":
				o.Path = sorter.FlowField(a.Float("flow-scale"), a.Float("warp"), a.Bool("vertical"), a.Seed())
			case "contour":
				o.Path = sorter.Contour(a.Float("warp"), a.Int("contour-blur"), a.Bool("vertical"))
			case "hilbert":
				o.Path = sorter.Hilbert()
			default:
				return nil, fmt.Errorf("sort: unknown path %q (choose from %s)", a.String("path"), strings.Join(pathNames, ", "))
			}
			if f := a.String("interval-image"); f != "" {
				src, err := imageio.Load(f)
				if err != nil {
					return nil, fmt.Errorf("sort: interval-image: %w", err)
				}
				o.Interval = sorter.FromImage(src, o.Interval)
			}
			if m := a.String("mask"); m != "" {
				mask, err := imageio.Load(m)
				if err != nil {
					return nil, fmt.Errorf("sort: mask: %w", err)
				}
				o.Mask = mask
			}
			return func(img *image.NRGBA) (*image.NRGBA, error) {
				sorter.SortInPlace(img, o)
				return img, nil
			}, nil
		},
	})
}

func registerColor() {
	Register(Spec{Name: "invert", Help: "Invert the colors.",
		Build: func(a *Args) (ApplyFunc, error) {
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Invert(img) }), nil
		}})
	Register(Spec{Name: "grayscale", Help: "Convert to grayscale (Rec. 709 luminance).",
		Build: func(a *Args) (ApplyFunc, error) {
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Grayscale(img) }), nil
		}})
	Register(Spec{Name: "threshold", Help: "Black and white: pixels brighter than level become white.",
		Params: []Param{float("level", "0.5", "luminance cut-off (0..1)")},
		Build: func(a *Args) (ApplyFunc, error) {
			l := a.Float("level")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Threshold(img, l) }), nil
		}})
	Register(Spec{Name: "posterize", Help: "Reduce every channel to a few levels.",
		Params: []Param{integer("levels", "4", "levels per channel (>= 2)")},
		Build: func(a *Args) (ApplyFunc, error) {
			n := a.Int("levels")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Posterize(img, n) }), nil
		}})
	Register(Spec{Name: "brightness", Help: "Brighten or darken.",
		Params: []Param{float("amount", "0.1", "-1..1")},
		Build: func(a *Args) (ApplyFunc, error) {
			v := a.Float("amount")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Brightness(img, v) }), nil
		}})
	Register(Spec{Name: "contrast", Help: "Scale contrast around mid-gray.",
		Params: []Param{float("factor", "1.3", "1 = unchanged, 0 = flat gray")},
		Build: func(a *Args) (ApplyFunc, error) {
			v := a.Float("factor")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Contrast(img, v) }), nil
		}})
	Register(Spec{Name: "gamma", Help: "Gamma correction.",
		Params: []Param{float("gamma", "1.2", "> 1 brightens midtones, < 1 darkens")},
		Build: func(a *Args) (ApplyFunc, error) {
			v := a.Float("gamma")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Gamma(img, v) }), nil
		}})
	Register(Spec{Name: "saturation", Help: "Scale color saturation.",
		Params: []Param{float("factor", "1.5", "0 = grayscale, 1 = unchanged")},
		Build: func(a *Args) (ApplyFunc, error) {
			v := a.Float("factor")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Saturation(img, v) }), nil
		}})
	Register(Spec{Name: "hue", Help: "Rotate all hues.",
		Params: []Param{float("degrees", "180", "rotation in degrees")},
		Build: func(a *Args) (ApplyFunc, error) {
			v := a.Float("degrees")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.HueShift(img, v) }), nil
		}})
	Register(Spec{Name: "solarize", Help: "Invert channel values above a threshold.",
		Params: []Param{float("threshold", "0.5", "0..1")},
		Build: func(a *Args) (ApplyFunc, error) {
			v := a.Float("threshold")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Solarize(img, v) }), nil
		}})
	Register(Spec{Name: "swap", Help: "Rearrange the color channels.",
		Params: []Param{str("order", "bgr", "three letters from r, g, b, e.g. bgr or grb")},
		Build: func(a *Args) (ApplyFunc, error) {
			order := a.String("order")
			if _, err := effects.ChannelSwap(image.NewNRGBA(image.Rect(0, 0, 1, 1)), order); err != nil {
				return nil, fmt.Errorf("swap: %w", err)
			}
			return func(img *image.NRGBA) (*image.NRGBA, error) { return effects.ChannelSwap(img, order) }, nil
		}})
	Register(Spec{Name: "gradientmap", Help: "Map brightness onto a color gradient (duotone, heat map, ...).",
		Params: []Param{str("colors", "#120458,#ff2a6d,#ffd319", "comma-separated colors from dark to bright, or a palette name")},
		Build: func(a *Args) (ApplyFunc, error) {
			p, err := effects.ParsePalette(a.String("colors"))
			if err != nil {
				return nil, fmt.Errorf("gradientmap: %w", err)
			}
			stops := []color.Color(p)
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.GradientMap(img, stops) }), nil
		}})
	Register(Spec{Name: "dither", Help: "Reduce colors with error diffusion or ordered dithering.",
		Params: []Param{
			str("method", "floyd", "dithering algorithm", effects.DitherNames()...),
			str("palette", "", "palette name or comma-separated colors; empty uses levels", effects.PaletteNames()...),
			integer("levels", "2", "levels per channel when no palette is given"),
		},
		Build: func(a *Args) (ApplyFunc, error) {
			o := effects.DitherOptions{
				Method: choice(a, "method", effects.DitherByName, effects.DitherNames()),
				Levels: a.Int("levels"),
			}
			if s := a.String("palette"); s != "" {
				p, err := effects.ParsePalette(s)
				if err != nil {
					return nil, fmt.Errorf("dither: %w", err)
				}
				o.Palette = p
			}
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Dither(img, o) }), nil
		}})
}

func registerGlitch() {
	Register(Spec{Name: "shift", Help: "Offset the red, green and blue channels (RGB split).",
		Params: []Param{
			str("r", "8,0", "red offset as dx,dy"),
			str("g", "0,0", "green offset as dx,dy"),
			str("b", "-8,0", "blue offset as dx,dy"),
		},
		Build: func(a *Args) (ApplyFunc, error) {
			r, g, b := point(a, "r"), point(a, "g"), point(a, "b")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.ChannelShift(img, r, g, b) }), nil
		}})
	Register(Spec{Name: "aberration", Help: "Radial chromatic aberration.",
		Params: []Param{float("strength", "0.01", "fraction of the distance to the center")},
		Build: func(a *Args) (ApplyFunc, error) {
			v := a.Float("strength")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.ChromaticAberration(img, v) }), nil
		}})
	Register(Spec{Name: "scanlines", Help: "Darken every n-th row like a CRT.",
		Params: []Param{integer("spacing", "3", "distance between lines in pixels"), float("intensity", "0.4", "0..1")},
		Build: func(a *Args) (ApplyFunc, error) {
			s, i := a.Int("spacing"), a.Float("intensity")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Scanlines(img, s, i) }), nil
		}})
	Register(Spec{Name: "noise", Help: "Add random noise or film grain.",
		Params: []Param{float("amount", "0.1", "0..1"), boolean("mono", "same noise on all channels (grain)"), seedParam},
		Build: func(a *Args) (ApplyFunc, error) {
			v, mono, seed := a.Float("amount"), a.Bool("mono"), a.Seed()
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Noise(img, v, mono, seed) }), nil
		}})
	Register(Spec{Name: "glitch", Help: "Displace random horizontal slices.",
		Params: []Param{integer("slices", "20", "number of slices"), float("shift", "0.1", "max shift as a fraction of the width"), seedParam},
		Build: func(a *Args) (ApplyFunc, error) {
			n, s, seed := a.Int("slices"), a.Float("shift"), a.Seed()
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Glitch(img, n, s, seed) }), nil
		}})
	Register(Spec{Name: "wave", Help: "Displace rows (or columns) along a sine wave.",
		Params: []Param{float("amplitude", "10", "pixels"), float("wavelength", "60", "pixels"), boolean("vertical", "displace columns instead of rows")},
		Build: func(a *Args) (ApplyFunc, error) {
			amp, wl, v := a.Float("amplitude"), a.Float("wavelength"), a.Bool("vertical")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Wave(img, amp, wl, v) }), nil
		}})
	Register(Spec{Name: "melt", Help: "Drag columns downwards so the image drips.",
		Params: []Param{integer("shift", "80", "maximum drip in pixels"), seedParam},
		Build: func(a *Args) (ApplyFunc, error) {
			s, seed := a.Int("shift"), a.Seed()
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Melt(img, s, seed) }), nil
		}})
	Register(Spec{Name: "jpeg", Help: "Corrupt the compressed JPEG data for broken-file artifacts.",
		Params: []Param{integer("amount", "10", "bytes to corrupt"), integer("quality", "50", "JPEG quality 1..100"), seedParam},
		Build: func(a *Args) (ApplyFunc, error) {
			n, q, seed := a.Int("amount"), a.Int("quality"), a.Seed()
			return func(img *image.NRGBA) (*image.NRGBA, error) { return effects.JPEGCorrupt(img, n, q, seed) }, nil
		}})
}

func registerFilters() {
	Register(Spec{Name: "pixelate", Help: "Replace blocks with their average color.",
		Params: []Param{integer("size", "8", "block size in pixels")},
		Build: func(a *Args) (ApplyFunc, error) {
			s := a.Int("size")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Pixelate(img, s) }), nil
		}})
	Register(Spec{Name: "mirror", Help: "Reflect one half of the image onto the other.",
		Params: []Param{str("mode", "left", "half that is kept", effects.MirrorNames()...)},
		Build: func(a *Args) (ApplyFunc, error) {
			m := choice(a, "mode", effects.MirrorByName, effects.MirrorNames())
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Mirror(img, m) }), nil
		}})
	Register(Spec{Name: "flip", Help: "Flip horizontally or vertically.",
		Params: []Param{str("axis", "h", "h (left↔right) or v (top↔bottom)", "h", "v")},
		Build: func(a *Args) (ApplyFunc, error) {
			switch a.String("axis") {
			case "h", "horizontal":
				return simple(func(img *image.NRGBA) *image.NRGBA { return effects.FlipHorizontal(img) }), nil
			case "v", "vertical":
				return simple(func(img *image.NRGBA) *image.NRGBA { return effects.FlipVertical(img) }), nil
			}
			return nil, fmt.Errorf("flip: axis must be h or v, got %q", a.String("axis"))
		}})
	Register(Spec{Name: "rotate", Help: "Rotate clockwise by a multiple of 90 degrees.",
		Params: []Param{integer("degrees", "90", "90, 180 or 270")},
		Build: func(a *Args) (ApplyFunc, error) {
			d := a.Int("degrees")
			if d%90 != 0 {
				return nil, fmt.Errorf("rotate: degrees must be a multiple of 90, got %d", d)
			}
			return func(img *image.NRGBA) (*image.NRGBA, error) { return effects.Rotate(img, d) }, nil
		}})
	Register(Spec{Name: "blur", Help: "Box blur.",
		Params: []Param{integer("radius", "2", "pixels")},
		Build: func(a *Args) (ApplyFunc, error) {
			r := a.Int("radius")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.BoxBlur(img, r) }), nil
		}})
	Register(Spec{Name: "sharpen", Help: "Unsharp mask.",
		Params: []Param{float("amount", "1", "strength"), integer("radius", "1", "blur radius in pixels")},
		Build: func(a *Args) (ApplyFunc, error) {
			v, r := a.Float("amount"), a.Int("radius")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Sharpen(img, v, r) }), nil
		}})
	Register(Spec{Name: "edges", Help: "Sobel edge detection.",
		Build: func(a *Args) (ApplyFunc, error) {
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.EdgeDetect(img) }), nil
		}})
	Register(Spec{Name: "vignette", Help: "Darken towards the corners.",
		Params: []Param{float("strength", "0.5", "0..1")},
		Build: func(a *Args) (ApplyFunc, error) {
			v := a.Float("strength")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Vignette(img, v) }), nil
		}})
}

func point(a *Args, name string) image.Point {
	s := a.String(name)
	xs, ys, ok := strings.Cut(s, ",")
	x, errX := strconv.Atoi(strings.TrimSpace(xs))
	y, errY := strconv.Atoi(strings.TrimSpace(ys))
	if !ok || errX != nil || errY != nil {
		a.Fail(fmt.Errorf("%s: %s must be dx,dy, got %q", a.spec.Name, name, s))
	}
	return image.Pt(x, y)
}
