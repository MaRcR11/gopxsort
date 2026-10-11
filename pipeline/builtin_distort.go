package pipeline

import (
	"fmt"
	"image"

	"github.com/MaRcR11/sortcery/effects"
	"github.com/MaRcR11/sortcery/imageio"
)

func registerDistort() {
	Register(Spec{Name: "displace", Help: "Shift pixels by brightness (of the image itself or of a map image).",
		Params: []Param{
			float("x", "100", "horizontal displacement in pixels"),
			float("y", "0", "vertical displacement in pixels"),
			str("map", "", "image file used as displacement map; empty uses the image itself"),
		},
		Build: func(a *Args) (ApplyFunc, error) {
			dx, dy := a.Float("x"), a.Float("y")
			var m image.Image
			if f := a.String("map"); f != "" {
				var err error
				if m, err = imageio.Load(f); err != nil {
					return nil, fmt.Errorf("displace: map: %w", err)
				}
			}
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Displace(img, m, dx, dy) }), nil
		}})
	Register(Spec{Name: "echo", Help: "Ghosting: fading, shifted copies of the image.",
		Params: []Param{
			integer("copies", "4", "number of echoes"),
			str("offset", "12,0", "shift between echoes as dx,dy"),
			float("decay", "0.7", "strength of each further echo (0..1)"),
		},
		Build: func(a *Args) (ApplyFunc, error) {
			n, off, d := a.Int("copies"), point(a, "offset"), a.Float("decay")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Echo(img, n, off, d) }), nil
		}})
	Register(Spec{Name: "bend", Help: "Data bending: an audio-style echo on the raw pixel bytes.",
		Params: []Param{
			integer("delay", "3001", "echo delay in bytes; values that are not a multiple of 3 shift colors"),
			float("feedback", "0.4", "echo strength (0..1)"),
		},
		Build: func(a *Args) (ApplyFunc, error) {
			d, f := a.Int("delay"), a.Float("feedback")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.DataBend(img, d, f) }), nil
		}})
	Register(Spec{Name: "swirl", Help: "Twist the image around its center.",
		Params: []Param{float("degrees", "180", "twist at the center"), float("radius", "0.5", "affected area (0..1)")},
		Build: func(a *Args) (ApplyFunc, error) {
			d, r := a.Float("degrees"), a.Float("radius")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Swirl(img, d, r) }), nil
		}})
	Register(Spec{Name: "ripple", Help: "Circular waves around the center.",
		Params: []Param{float("amplitude", "8", "pixels"), float("wavelength", "40", "pixels")},
		Build: func(a *Args) (ApplyFunc, error) {
			amp, wl := a.Float("amplitude"), a.Float("wavelength")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Ripple(img, amp, wl) }), nil
		}})
	Register(Spec{Name: "bulge", Help: "Fisheye bulge; negative strength pinches.",
		Params: []Param{float("strength", "0.6", "-1..1"), float("radius", "0.5", "affected area (0..1)")},
		Build: func(a *Args) (ApplyFunc, error) {
			s, r := a.Float("strength"), a.Float("radius")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Bulge(img, s, r) }), nil
		}})
	Register(Spec{Name: "polar", Help: "Convert to polar coordinates (or back with inverse).",
		Params: []Param{boolean("inverse", "convert from polar back to rectangular")},
		Build: func(a *Args) (ApplyFunc, error) {
			inv := a.Bool("inverse")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Polar(img, inv) }), nil
		}})
	Register(Spec{Name: "curvature", Help: "Curved CRT screen.",
		Params: []Param{float("strength", "0.12", "amount of curvature")},
		Build: func(a *Args) (ApplyFunc, error) {
			s := a.Float("strength")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Curvature(img, s) }), nil
		}})
	Register(Spec{Name: "slitscan", Help: "Stretch one column (or row) across the rest of the image.",
		Params: []Param{float("position", "0.6", "slit position (0..1)"), boolean("vertical", "stretch a row downwards instead")},
		Build: func(a *Args) (ApplyFunc, error) {
			p, v := a.Float("position"), a.Bool("vertical")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.SlitScan(img, p, v) }), nil
		}})
	Register(Spec{Name: "interlace", Help: "Shift and darken every n-th line like a broken video signal.",
		Params: []Param{integer("shift", "8", "pixels"), integer("spacing", "2", "every n-th line"), float("darken", "0.2", "0..1")},
		Build: func(a *Args) (ApplyFunc, error) {
			s, sp, d := a.Int("shift"), a.Int("spacing"), a.Float("darken")
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Interlace(img, s, sp, d) }), nil
		}})
	Register(Spec{Name: "scatter", Help: "Move every pixel by a small random amount.",
		Params: []Param{integer("radius", "4", "pixels"), seedParam},
		Build: func(a *Args) (ApplyFunc, error) {
			r, seed := a.Int("radius"), a.Seed()
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Scatter(img, r, seed) }), nil
		}})
	Register(Spec{Name: "tracking", Help: "VHS tracking errors: bright, noisy, shifted bands.",
		Params: []Param{
			integer("bands", "3", "number of bands"),
			float("shift", "0.05", "maximum shift as a fraction of the width"),
			float("noise", "0.3", "amount of white noise in the bands (0..1)"),
			seedParam,
		},
		Build: func(a *Args) (ApplyFunc, error) {
			b, s, n, seed := a.Int("bands"), a.Float("shift"), a.Float("noise"), a.Seed()
			return simple(func(img *image.NRGBA) *image.NRGBA { return effects.Tracking(img, b, s, n, seed) }), nil
		}})
}
