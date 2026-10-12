package snippets

import (
	"image"

	"github.com/MaRcR11/sortmancy/effects"
	"github.com/MaRcR11/sortmancy/imageio"
	"github.com/MaRcR11/sortmancy/pipeline"
	"github.com/MaRcR11/sortmancy/sorter"
)

var (
	_ = effects.Invert
	_ = imageio.Load
	_ = pipeline.Parse
	_ = sorter.Sort
)

type Func func(img image.Image, outGIF string) (out image.Image, err error)

var Funcs = map[string]Func{
	"sort_default": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{})
		return
	},
	"sort_horizontal": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Threshold(0.1, 0.9),
		})
		return
	},
	"sort_vertical": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Threshold(0.1, 0.9),
			Path:     sorter.Linear(270),
		})
		return
	},
	"sort_diagonal_45": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Threshold(0.1, 0.8),
			Path:     sorter.Linear(45),
		})
		return
	},
	"sort_diagonal_135": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Threshold(0.2, 0.8),
			Path:     sorter.Linear(135),
		})
		return
	},
	"sort_all_horizontal": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.None(),
		})
		return
	},
	"sort_all_vertical": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.None(),
			Path:     sorter.Linear(270),
		})
		return
	},
	"sort_classic": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Key:      sorter.Luminance,
			Interval: sorter.None(),
		})
		return
	},
	"sort_reverse": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Threshold(0.2, 0.8),
			Reverse:  true,
		})
		return
	},
	"sort_only_bright": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Threshold(0.6, 1),
		})
		return
	},
	"sort_only_dark": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Threshold(0, 0.35),
		})
		return
	},
	"sort_threshold_on_hue": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.ThresholdBy(sorter.Hue, 0.4, 0.7),
		})
		return
	},
	"sort_randomness": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval:   sorter.None(),
			Randomness: 0.5,
			Seed:       7,
		})
		return
	},
	"sort_key_hue": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Key:      sorter.Hue,
			Interval: sorter.Threshold(0.15, 0.95),
		})
		return
	},
	"sort_key_saturation": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Key:      sorter.Saturation,
			Interval: sorter.Threshold(0.15, 0.95),
		})
		return
	},
	"sort_key_value": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Key:      sorter.Value,
			Interval: sorter.Threshold(0.15, 0.95),
		})
		return
	},
	"sort_key_intensity": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Key:      sorter.Intensity,
			Interval: sorter.Threshold(0.15, 0.95),
		})
		return
	},
	"sort_key_minimum": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Key:      sorter.Minimum,
			Interval: sorter.Threshold(0.15, 0.95),
		})
		return
	},
	"sort_key_chroma": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Key:      sorter.Chroma,
			Interval: sorter.Threshold(0.15, 0.95),
		})
		return
	},
	"sort_key_red": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Key:      sorter.Red,
			Interval: sorter.Threshold(0.15, 0.95),
		})
		return
	},
	"sort_key_green": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Key:      sorter.Green,
			Interval: sorter.Threshold(0.15, 0.95),
		})
		return
	},
	"sort_key_blue": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Key:      sorter.Blue,
			Interval: sorter.Threshold(0.15, 0.95),
		})
		return
	},
	"sort_interval_random": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Random(80),
			Seed:     7,
		})
		return
	},
	"sort_random_vertical": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Random(100),
			Path:     sorter.Linear(270),
			Reverse:  true,
			Seed:     7,
		})
		return
	},
	"sort_interval_waves": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Waves(80),
			Seed:     7,
		})
		return
	},
	"sort_interval_edges": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Edges(0.2),
		})
		return
	},
	"sort_hue_edges": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Key:      sorter.Hue,
			Interval: sorter.Edges(0.15),
		})
		return
	},
	"sort_path_wave": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Threshold(0.1, 0.8),
			Path:     sorter.Wave(40, 300, false),
		})
		return
	},
	"sort_path_wave_vertical": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Threshold(0.15, 0.8),
			Path:     sorter.Wave(30, 250, true),
		})
		return
	},
	"sort_path_spiral": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Edges(0.15),
			Path:     sorter.Spiral(0.5, 0.5, 6),
		})
		return
	},
	"sort_path_concentric": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Threshold(0.1, 0.8),
			Path:     sorter.Concentric(0.5, 0.5),
		})
		return
	},
	"sort_path_radial": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Edges(0.1),
			Path:     sorter.Radial(0.5, 0.5),
		})
		return
	},
	"sort_path_flow": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Threshold(0.1, 0.8),
			Path:     sorter.FlowField(150, 60, false, 7),
		})
		return
	},
	"sort_path_flow_liquid": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.None(),
			Path:     sorter.FlowField(150, 60, false, 7),
		})
		return
	},
	"sort_path_flow_vertical": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Threshold(0.1, 0.8),
			Path:     sorter.FlowField(150, 60, true, 7),
		})
		return
	},
	"sort_path_contour": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Threshold(0.05, 0.8),
			Path:     sorter.Contour(300, 10, false),
		})
		return
	},
	"sort_path_contour_vertical": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Random(120),
			Path:     sorter.Contour(200, 4, true),
			Seed:     7,
		})
		return
	},
	"sort_path_hilbert": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Threshold(0.1, 0.8),
			Path:     sorter.Hilbert(),
		})
		return
	},
	"sort_path_hilbert_random": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Random(400),
			Path:     sorter.Hilbert(),
			Seed:     7,
		})
		return
	},
	"sort_mode_shuffle": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Edges(0.15),
			Path:     sorter.Linear(270),
			Mode:     sorter.ModeShuffle,
			Seed:     7,
		})
		return
	},
	"sort_mode_average": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Edges(0.15),
			Path:     sorter.Linear(270),
			Mode:     sorter.ModeAverage,
		})
		return
	},
	"sort_mode_gradient": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Random(60),
			Path:     sorter.Linear(90),
			Mode:     sorter.ModeGradient,
			Seed:     7,
		})
		return
	},
	"sort_mode_smear": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Edges(0.12),
			Path:     sorter.Linear(270),
			Mode:     sorter.ModeSmear,
		})
		return
	},
	"sort_mode_channels": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Threshold(0.1, 0.8),
			Mode:     sorter.ModeChannels,
		})
		return
	},
	"sort_mode_channels_all": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.None(),
			Path:     sorter.Linear(270),
			Mode:     sorter.ModeChannels,
		})
		return
	},
	"sort_mode_mirror": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Threshold(0.1, 0.8),
			Path:     sorter.Linear(270),
			Mode:     sorter.ModeMirror,
		})
		return
	},
	"sort_mode_mirror_all": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.None(),
			Mode:     sorter.ModeMirror,
		})
		return
	},
	"sort_tile": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.None(),
			Tile:     24,
		})
		return
	},
	"sort_tile_vertical": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.Threshold(0.1, 0.8),
			Path:     sorter.Linear(270),
			Tile:     16,
		})
		return
	},
	"sort_interval_image": func(img image.Image, outGIF string) (out image.Image, err error) {
		guide, err := imageio.Load("kaleidoscope.jpg")
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.FromImage(guide, sorter.Edges(0.2)),
		})
		return
	},
	"sort_mask": func(img image.Image, outGIF string) (out image.Image, err error) {
		mask, err := imageio.Load("threshold.jpg")
		out = sorter.Sort(img, sorter.Options{
			Interval: sorter.None(),
			Path:     sorter.Linear(270),
			Mask:     mask,
		})
		return
	},
	"invert": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Invert(img)
		return
	},
	"grayscale": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Grayscale(img)
		return
	},
	"threshold": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Threshold(img, 0.3)
		return
	},
	"posterize": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Posterize(img, 4)
		return
	},
	"brightness": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Brightness(img, 0.2)
		return
	},
	"contrast": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Contrast(img, 1.5)
		return
	},
	"gamma": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Gamma(img, 1.8)
		return
	},
	"saturation": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Saturation(img, 2)
		return
	},
	"hue": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.HueShift(img, 180)
		return
	},
	"solarize": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Solarize(img, 0.3)
		return
	},
	"swap": func(img image.Image, outGIF string) (out image.Image, err error) {
		out, err = effects.ChannelSwap(img, "bgr")
		return
	},
	"gradientmap": func(img image.Image, outGIF string) (out image.Image, err error) {
		colors, err := effects.ParsePalette("#120458,#ff2a6d,#ffd319")
		out = effects.GradientMap(img, colors)
		return
	},
	"glitch": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Glitch(img, 30, 0.15, 7)
		return
	},
	"shift": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.ChannelShift(img, image.Pt(12, 0), image.Pt(0, 0), image.Pt(-12, 0))
		return
	},
	"aberration": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.ChromaticAberration(img, 0.015)
		return
	},
	"jpeg": func(img image.Image, outGIF string) (out image.Image, err error) {
		out, err = effects.JPEGCorrupt(img, 15, 40, 2)
		return
	},
	"blocks": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.BlockShuffle(img, 48, 0.15, 7)
		return
	},
	"melt": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Melt(img, 150, 7)
		return
	},
	"wave": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Wave(img, 10, 60, false)
		return
	},
	"scanlines": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Scanlines(img, 3, 0.4)
		return
	},
	"noise": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Noise(img, 0.1, false, 7)
		return
	},
	"echo": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Echo(img, 4, image.Pt(12, 0), 0.7)
		return
	},
	"bend": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.DataBend(img, 3001, 0.4)
		return
	},
	"bend__strong": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.DataBend(img, 12347, 0.6)
		return
	},
	"interlace": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Interlace(img, 8, 2, 0.2)
		return
	},
	"scatter": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Scatter(img, 4, 7)
		return
	},
	"tracking": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Tracking(img, 3, 0.05, 0.3, 7)
		return
	},
	"displace": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Displace(img, nil, 150, 0)
		return
	},
	"displace__map": func(img image.Image, outGIF string) (out image.Image, err error) {
		dmap, err := imageio.Load("kaleidoscope.jpg")
		out = effects.Displace(img, dmap, 120, 60)
		return
	},
	"swirl": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Swirl(img, 180, 0.5)
		return
	},
	"ripple": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Ripple(img, 8, 40)
		return
	},
	"bulge": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Bulge(img, 0.6, 0.5)
		return
	},
	"bulge__pinch": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Bulge(img, -0.6, 0.5)
		return
	},
	"polar": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Polar(img, false)
		return
	},
	"polar__inverse": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Polar(img, true)
		return
	},
	"curvature": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Curvature(img, 0.12)
		return
	},
	"slitscan": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.SlitScan(img, 0.6, false)
		return
	},
	"slitscan__vertical": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.SlitScan(img, 0.4, true)
		return
	},
	"dither": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Dither(img, effects.DitherOptions{
			Method:  effects.Atkinson,
			Palette: effects.Palettes["gameboy"],
		})
		return
	},
	"dither__bayer": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Dither(img, effects.DitherOptions{
			Method: effects.Bayer4,
			Levels: 2,
		})
		return
	},
	"quantize": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Quantize(img, 8, effects.FloydSteinberg)
		return
	},
	"pixelate": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Pixelate(img, 12)
		return
	},
	"halftone": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Halftone(img, 8, true)
		return
	},
	"halftone__mono": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Halftone(img, 8, false)
		return
	},
	"kaleidoscope": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Kaleidoscope(img, 6, 0)
		return
	},
	"bloom": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Bloom(img, 0.5, 16, 1)
		return
	},
	"oilpaint": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.OilPaint(img, 5)
		return
	},
	"mirror": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Mirror(img, effects.MirrorQuad)
		return
	},
	"resize": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Resize(img, 320, 180, effects.Bilinear)
		return
	},
	"crop": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Crop(img, image.Rect(600, 200, 1400, 650))
		return
	},
	"flip": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.FlipHorizontal(img)
		return
	},
	"rotate": func(img image.Image, outGIF string) (out image.Image, err error) {
		out, err = effects.Rotate(img, 90)
		return
	},
	"blur": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.BoxBlur(img, 6)
		return
	},
	"sharpen": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Sharpen(img, 2, 2)
		return
	},
	"edges": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.EdgeDetect(img)
		return
	},
	"vignette": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Vignette(img, 0.8)
		return
	},
	"preset_classic": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.ParseSeed("@classic", 7)
		out, err = p.Apply(img)
		return
	},
	"preset_vertical": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.ParseSeed("@vertical", 7)
		out, err = p.Apply(img)
		return
	},
	"preset_spiral": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.ParseSeed("@spiral", 7)
		out, err = p.Apply(img)
		return
	},
	"preset_melt": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.ParseSeed("@melt", 7)
		out, err = p.Apply(img)
		return
	},
	"preset_glitch": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.ParseSeed("@glitch", 7)
		out, err = p.Apply(img)
		return
	},
	"preset_vhs": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.ParseSeed("@vhs", 7)
		out, err = p.Apply(img)
		return
	},
	"preset_crt": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.ParseSeed("@crt", 7)
		out, err = p.Apply(img)
		return
	},
	"preset_vaporwave": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.ParseSeed("@vaporwave", 7)
		out, err = p.Apply(img)
		return
	},
	"preset_dream": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.ParseSeed("@dream", 7)
		out, err = p.Apply(img)
		return
	},
	"preset_gameboy": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.ParseSeed("@gameboy", 7)
		out, err = p.Apply(img)
		return
	},
	"preset_8bit": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.ParseSeed("@8bit", 7)
		out, err = p.Apply(img)
		return
	},
	"preset_chunky": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.ParseSeed("@chunky", 7)
		out, err = p.Apply(img)
		return
	},
	"preset_newspaper": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.ParseSeed("@newspaper", 7)
		out, err = p.Apply(img)
		return
	},
	"preset_oil": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.ParseSeed("@oil", 7)
		out, err = p.Apply(img)
		return
	},
	"preset_planet": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.ParseSeed("@planet", 7)
		out, err = p.Apply(img)
		return
	},
	"preset_databend": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.ParseSeed("@databend", 7)
		out, err = p.Apply(img)
		return
	},
	"pipeline_streaks": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.Parse("sort angle=270 interval=edges edge=0.15 | shift | scanlines")
		out, err = p.Apply(img)
		return
	},
	"pipeline_blocky": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.Parse("resize scale=0.2 | sort angle=90 lower=0.1 | resize original filter=nearest | aberration")
		out, err = p.Apply(img)
		return
	},
	"pipeline_mix": func(img image.Image, outGIF string) (out image.Image, err error) {
		p, err := pipeline.Parse("sort key=hue lower=0.1 | gradientmap colors=vapor | mix 0.5")
		out, err = p.Apply(img)
		return
	},
	"mix": func(img image.Image, outGIF string) (out image.Image, err error) {
		out = effects.Mix(img, effects.Pixelate(img, 24), 0.5)
		return
	},
	"anim_threshold": func(img image.Image, outGIF string) (out image.Image, err error) {
		frames, err := pipeline.Animate([]image.Image{img},
			"resize width=480 | sort angle={0..90} lower={0.05..0.5}",
			pipeline.AnimateOptions{Frames: 16, PingPong: true, Seed: 7})
		anim := &imageio.Animation{}
		for _, f := range frames {
			anim.Frames = append(anim.Frames, f)
			anim.Delays = append(anim.Delays, 9)
		}
		err = imageio.SaveAnimation(outGIF, anim)
		return
	},
	"anim_threshold_only": func(img image.Image, outGIF string) (out image.Image, err error) {
		frames, err := pipeline.Animate([]image.Image{img},
			"resize width=480 | sort lower={0.05..0.6}",
			pipeline.AnimateOptions{Frames: 16, PingPong: true, Seed: 7})
		anim := &imageio.Animation{}
		for _, f := range frames {
			anim.Frames = append(anim.Frames, f)
			anim.Delays = append(anim.Delays, 9)
		}
		err = imageio.SaveAnimation(outGIF, anim)
		return
	},
	"anim_angle": func(img image.Image, outGIF string) (out image.Image, err error) {
		frames, err := pipeline.Animate([]image.Image{img},
			"resize width=480 | sort angle={0..180} lower=0.2",
			pipeline.AnimateOptions{Frames: 16, PingPong: true, Seed: 7})
		anim := &imageio.Animation{}
		for _, f := range frames {
			anim.Frames = append(anim.Frames, f)
			anim.Delays = append(anim.Delays, 9)
		}
		err = imageio.SaveAnimation(outGIF, anim)
		return
	},
	"anim_glitch": func(img image.Image, outGIF string) (out image.Image, err error) {
		frames, err := pipeline.Animate([]image.Image{img},
			"resize width=480 | glitch seed={1..16} | shift",
			pipeline.AnimateOptions{Frames: 16, PingPong: true, Seed: 7})
		anim := &imageio.Animation{}
		for _, f := range frames {
			anim.Frames = append(anim.Frames, f)
			anim.Delays = append(anim.Delays, 9)
		}
		err = imageio.SaveAnimation(outGIF, anim)
		return
	},
	"anim_kaleidoscope": func(img image.Image, outGIF string) (out image.Image, err error) {
		frames, err := pipeline.Animate([]image.Image{img},
			"resize width=480 | kaleidoscope rotation={0..60}",
			pipeline.AnimateOptions{Frames: 16, PingPong: true, Seed: 7})
		anim := &imageio.Animation{}
		for _, f := range frames {
			anim.Frames = append(anim.Frames, f)
			anim.Delays = append(anim.Delays, 9)
		}
		err = imageio.SaveAnimation(outGIF, anim)
		return
	},
}
