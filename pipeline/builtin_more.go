package pipeline

import (
	"fmt"
	"image"
	"strings"

	"github.com/MaRcR11/sortmancy/effects"
	"github.com/MaRcR11/sortmancy/imageio"
)

func loadLUT(name string) (*effects.LUT, error) {
	if l, ok := effects.LUTByName(name); ok {
		return l, nil
	}
	if strings.HasSuffix(strings.ToLower(name), ".cube") {
		return effects.LoadCubeLUT(name)
	}
	return nil, fmt.Errorf("unknown LUT %q (use a .cube file or one of %s)", name, strings.Join(effects.LUTNames(), ", "))
}

func registerColorMore() {
	Register(Spec{Name: "blend", Help: "Blend with the pipeline input or another image using a blend mode.",
		Params: []Param{
			str("mode", "screen", "blend mode", effects.BlendNames()...),
			float("amount", "1", "strength (0..1)"),
			str("with", "", "image file to blend with; empty uses the pipeline input"),
		},
		BuildSource: func(a *Args) (SourceFunc, error) {
			mode := choice(a, "mode", effects.BlendByName, effects.BlendNames())
			amount := a.Float("amount")
			var other image.Image
			if f := a.String("with"); f != "" {
				var err error
				if other, err = imageio.Load(f); err != nil {
					return nil, fmt.Errorf("blend: %w", err)
				}
			}
			return func(img, src *image.NRGBA) (*image.NRGBA, error) {
				top := image.Image(src)
				if other != nil {
					top = other
				}
				return effects.Blend(img, top, mode, amount), nil
			}, nil
		}})
	Register(Spec{Name: "autolevels", Help: "Stretch every channel to the full range.",
		Params: []Param{float("clip", "0.005", "fraction of darkest and brightest pixels to ignore")},
		Build: func(a *Args) (ApplyFunc, error) {
			c := a.Float("clip")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.AutoLevels(img, c) }), nil
		}})
	Register(Spec{Name: "equalize", Help: "Histogram equalization of the brightness.",
		Build: func(a *Args) (ApplyFunc, error) {
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Equalize(img) }), nil
		}})
	Register(Spec{Name: "levels", Help: "Set the black point, white point and midtone gamma.",
		Params: []Param{
			float("black", "0.05", "input black point (0..1)"),
			float("white", "0.95", "input white point (0..1)"),
			float("gamma", "1", "midtones; > 1 brightens"),
		},
		Build: func(a *Args) (ApplyFunc, error) {
			b, w, g := a.Float("black"), a.Float("white"), a.Float("gamma")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Levels(img, b, w, g) }), nil
		}})
	Register(Spec{Name: "curves", Help: "Smooth tone curve through control points.",
		Params: []Param{str("points", "0:0,0.25:0.18,0.75:0.85,1:1", "comma-separated x:y points (0..1)")},
		Build: func(a *Args) (ApplyFunc, error) {
			pts, err := effects.ParseCurve(a.String("points"))
			if err != nil {
				return nil, fmt.Errorf("curves: %w", err)
			}
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Curves(img, pts) }), nil
		}})
	Register(Spec{Name: "temperature", Help: "Warm up or cool down the image.",
		Params: []Param{float("warmth", "0.3", "-1 (cool) .. 1 (warm)"), float("tint", "0", "-1 (magenta) .. 1 (green)")},
		Build: func(a *Args) (ApplyFunc, error) {
			w, t := a.Float("warmth"), a.Float("tint")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Temperature(img, w, t) }), nil
		}})
	Register(Spec{Name: "sepia", Help: "Old photo tone.",
		Params: []Param{float("amount", "1", "0..1")},
		Build: func(a *Args) (ApplyFunc, error) {
			v := a.Float("amount")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Sepia(img, v) }), nil
		}})
	Register(Spec{Name: "channelmix", Help: "Mix the color channels with a 3×3 matrix.",
		Params: []Param{str("matrix", "0.6,0.4,0,0.2,0.8,0,0,0.5,0.6", "9 numbers: rows for red, green and blue")},
		Build: func(a *Args) (ApplyFunc, error) {
			m, err := effects.ParseMatrix(a.String("matrix"))
			if err != nil {
				return nil, fmt.Errorf("channelmix: %w", err)
			}
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.ChannelMix(img, m) }), nil
		}})
	Register(Spec{Name: "lut", Help: "Apply a 3D color lookup table (.cube file or built-in look).",
		Params: []Param{
			str("lut", "tealorange", "built-in name or path to a .cube file", effects.LUTNames()...),
			float("amount", "1", "strength (0..1)"),
		},
		Build: func(a *Args) (ApplyFunc, error) {
			l, err := loadLUT(a.String("lut"))
			if err != nil {
				return nil, fmt.Errorf("lut: %w", err)
			}
			v := a.Float("amount")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.ApplyLUT(img, l, v) }), nil
		}})
}

func registerArtisticMore() {
	Register(Spec{Name: "ascii", Help: "Render the image as ASCII characters.",
		Params: []Param{
			integer("cols", "160", "characters per line"),
			str("chars", effects.DefaultASCIIChars, "characters from dark to bright"),
			boolean("color", "color every character with its area of the image"),
		},
		Build: func(a *Args) (ApplyFunc, error) {
			cols, chars, colored := a.Int("cols"), a.String("chars"), a.Bool("color")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.ASCIIImage(img, cols, chars, colored) }), nil
		}})
	Register(Spec{Name: "voronoi", Help: "Mosaic of Voronoi cells filled with their average color.",
		Params: []Param{integer("cells", "2000", "approximate number of cells"), seedParam},
		Build: func(a *Args) (ApplyFunc, error) {
			n, seed := a.Int("cells"), a.Seed()
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Voronoi(img, n, seed) }), nil
		}})
	Register(Spec{Name: "lowpoly", Help: "Low-poly look made of flat triangles.",
		Params: []Param{integer("cells", "800", "approximate number of grid cells (two triangles each)"), seedParam},
		Build: func(a *Args) (ApplyFunc, error) {
			n, seed := a.Int("cells"), a.Seed()
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.LowPoly(img, n, seed) }), nil
		}})
	Register(Spec{Name: "toon", Help: "Cartoon look: flat colors with black outlines.",
		Params: []Param{integer("levels", "6", "color levels per channel"), float("edge", "0.2", "edge strength (0..1) that becomes an outline")},
		Build: func(a *Args) (ApplyFunc, error) {
			l, e := a.Int("levels"), a.Float("edge")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Toon(img, l, e) }), nil
		}})
	Register(Spec{Name: "neon", Help: "Glowing colored edges on black.",
		Params: []Param{float("edge", "0.25", "edge strength (0..1) for full brightness"), integer("glow", "12", "glow radius in pixels")},
		Build: func(a *Args) (ApplyFunc, error) {
			e, g := a.Float("edge"), a.Int("glow")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Neon(img, e, g) }), nil
		}})
	Register(Spec{Name: "emboss", Help: "Raised relief look.",
		Params: []Param{float("strength", "2", "relief strength"), float("angle", "45", "light direction in degrees")},
		Build: func(a *Args) (ApplyFunc, error) {
			s, d := a.Float("strength"), a.Float("angle")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Emboss(img, s, d) }), nil
		}})
	Register(Spec{Name: "gaussian", Help: "Gaussian blur.",
		Params: []Param{float("sigma", "4", "blur strength in pixels")},
		Build: func(a *Args) (ApplyFunc, error) {
			s := a.Float("sigma")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.GaussianBlur(img, s) }), nil
		}})
	Register(Spec{Name: "motionblur", Help: "Blur along a direction, like fast movement.",
		Params: []Param{float("angle", "0", "direction in degrees"), integer("length", "30", "blur length in pixels")},
		Build: func(a *Args) (ApplyFunc, error) {
			d, l := a.Float("angle"), a.Int("length")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.MotionBlur(img, d, l) }), nil
		}})
	Register(Spec{Name: "zoomblur", Help: "Radial blur towards the center, like zooming during the exposure.",
		Params: []Param{float("strength", "0.3", "0..1"), integer("samples", "24", "quality")},
		Build: func(a *Args) (ApplyFunc, error) {
			s, n := a.Float("strength"), a.Int("samples")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.ZoomBlur(img, s, n) }), nil
		}})
	Register(Spec{Name: "median", Help: "Median filter: removes noise and flattens details.",
		Params: []Param{integer("radius", "3", "pixels")},
		Build: func(a *Args) (ApplyFunc, error) {
			r := a.Int("radius")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Median(img, r) }), nil
		}})
	Register(Spec{Name: "dilate", Help: "Grow bright areas.",
		Params: []Param{integer("radius", "3", "pixels")},
		Build: func(a *Args) (ApplyFunc, error) {
			r := a.Int("radius")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Dilate(img, r) }), nil
		}})
	Register(Spec{Name: "erode", Help: "Grow dark areas.",
		Params: []Param{integer("radius", "3", "pixels")},
		Build: func(a *Args) (ApplyFunc, error) {
			r := a.Int("radius")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Erode(img, r) }), nil
		}})
	Register(Spec{Name: "tile", Help: "Repeat the image in a grid.",
		Params: []Param{integer("cols", "3", "columns"), integer("rows", "3", "rows"), boolean("mirror", "mirror every other tile")},
		Build: func(a *Args) (ApplyFunc, error) {
			c, r, m := a.Int("cols"), a.Int("rows"), a.Bool("mirror")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Tile(img, c, r, m) }), nil
		}})
	Register(Spec{Name: "droste", Help: "Picture in a picture in a picture.",
		Params: []Param{float("scale", "0.6", "size of each nested copy (0..1)"), integer("steps", "6", "number of nested copies")},
		Build: func(a *Args) (ApplyFunc, error) {
			s, n := a.Float("scale"), a.Int("steps")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Droste(img, s, n) }), nil
		}})
}
