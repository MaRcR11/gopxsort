package catalog

import (
	"fmt"
	"strings"
)

type Example struct {
	Name      string
	Title     string
	CLI       string
	Code      string
	Needs     []string
	Highlight bool
	Animated  bool
}

type Section struct {
	Slug     string
	Title    string
	Intro    string
	Examples []Example
}

func (e Example) File() string {
	if e.Animated {
		return e.Name + ".gif"
	}
	return e.Name + ".jpg"
}

func (e Example) Op() string {
	return strings.SplitN(e.Name, "__", 2)[0]
}

func sortOptions(fields ...string) string {
	if len(fields) == 0 {
		return "out := sorter.Sort(img, sorter.Options{})"
	}
	width := 0
	for _, f := range fields {
		width = max(width, len(strings.SplitN(f, ":", 2)[0])+1)
	}
	lines := make([]string, len(fields))
	for i, f := range fields {
		kv := strings.SplitN(f, ":", 2)
		lines[i] = fmt.Sprintf("\t%-*s %s,", width, kv[0]+":", strings.TrimSpace(kv[1]))
	}
	return "out := sorter.Sort(img, sorter.Options{\n" + strings.Join(lines, "\n") + "\n})"
}

func srt(name, title, flags string, fields ...string) Example {
	return Example{Name: name, Title: title, CLI: strings.TrimSpace("sort -i city.jpg -o out.jpg " + flags), Code: sortOptions(fields...)}
}

func fx(name, title, flags, code string) Example {
	op := strings.SplitN(name, "__", 2)[0]
	return Example{Name: name, Title: title, CLI: strings.TrimSpace(op + " -i city.jpg -o out.jpg " + flags), Code: code}
}

func preset(name string, random bool) Example {
	seed := ""
	if random {
		seed = " --seed 7"
	}
	return Example{
		Name:  "preset_" + name,
		Title: "@" + name,
		CLI:   "@" + name + " -i city.jpg -o out.jpg" + seed,
		Code:  fmt.Sprintf("p, err := pipeline.ParseSeed(%q, 7)\nout, err := p.Apply(img)", "@"+name),
	}
}

func pipe(name, title, spec string) Example {
	return Example{
		Name:  name,
		Title: title,
		CLI:   fmt.Sprintf("run -i city.jpg -o out.jpg %q", spec),
		Code:  fmt.Sprintf("p, err := pipeline.Parse(%q)\nout, err := p.Apply(img)", spec),
	}
}

func anim(name, title, spec string) Example {
	return Example{
		Name:     name,
		Title:    title,
		Animated: true,
		CLI:      fmt.Sprintf("run -i city.jpg -o out.gif --frames 16 --pingpong --delay 90 --seed 7 %q", spec),
		Code: fmt.Sprintf(`frames, err := pipeline.Animate([]image.Image{img},
	%q,
	pipeline.AnimateOptions{Frames: 16, PingPong: true, Seed: 7})
anim := &imageio.Animation{}
for _, f := range frames {
	anim.Frames = append(anim.Frames, f)
	anim.Delays = append(anim.Delays, 9)
}
err = imageio.SaveAnimation("out.gif", anim)`, spec),
	}
}

func highlight(e Example) Example {
	e.Highlight = true
	return e
}

func needs(e Example, names ...string) Example {
	e.Needs = names
	return e
}

func keyExamples() []Example {
	var out []Example
	for _, k := range []string{"hue", "saturation", "value", "intensity", "minimum", "chroma", "red", "green", "blue"} {
		out = append(out, srt("sort_key_"+k, "Key: "+k, "--key "+k+" --lower 0.15 --upper 0.95",
			"Key: sorter."+strings.ToUpper(k[:1])+k[1:], "Interval: sorter.Threshold(0.15, 0.95)"))
	}
	return out
}

var Sections = []Section{
	{Slug: "sorting", Title: "Sorting: direction and threshold", Intro: "Where and in which direction pixels are sorted.", Examples: []Example{
		srt("sort_default", "Default settings", ""),
		srt("sort_horizontal", "Horizontal", "--lower 0.1 --upper 0.9",
			"Interval: sorter.Threshold(0.1, 0.9)"),
		highlight(srt("sort_vertical", "Vertical (top to bottom)", "--angle 270 --lower 0.1 --upper 0.9",
			"Interval: sorter.Threshold(0.1, 0.9)", "Path: sorter.Linear(270)")),
		srt("sort_diagonal_45", "Diagonal 45°", "--angle 45 --lower 0.1",
			"Interval: sorter.Threshold(0.1, 0.8)", "Path: sorter.Linear(45)"),
		srt("sort_diagonal_135", "Diagonal 135°", "--angle 135 --lower 0.2",
			"Interval: sorter.Threshold(0.2, 0.8)", "Path: sorter.Linear(135)"),
		srt("sort_all_horizontal", "Every pixel, whole rows", "--interval none",
			"Interval: sorter.None()"),
		srt("sort_all_vertical", "Every pixel, whole columns", "--interval none --angle 270",
			"Interval: sorter.None()", "Path: sorter.Linear(270)"),
		srt("sort_classic", "Classic luminance sort (the algorithm of the very first version)", "--interval none --key luminance",
			"Key: sorter.Luminance", "Interval: sorter.None()"),
		srt("sort_reverse", "Reversed (bright to dark)", "--lower 0.2 --reverse",
			"Interval: sorter.Threshold(0.2, 0.8)", "Reverse: true"),
		srt("sort_only_bright", "Only bright pixels", "--lower 0.6 --upper 1",
			"Interval: sorter.Threshold(0.6, 1)"),
		srt("sort_only_dark", "Only dark pixels", "--lower 0 --upper 0.35",
			"Interval: sorter.Threshold(0, 0.35)"),
		srt("sort_threshold_on_hue", "Threshold on hue instead of lightness", "--threshold-key hue --lower 0.4 --upper 0.7",
			"Interval: sorter.ThresholdBy(sorter.Hue, 0.4, 0.7)"),
		srt("sort_randomness", "Skip half of the spans", "--interval none --randomness 0.5 --seed 7",
			"Interval: sorter.None()", "Randomness: 0.5", "Seed: 7"),
	}},
	{Slug: "keys", Title: "Sorting: keys", Intro: "The value pixels are ordered by.", Examples: keyExamples()},
	{Slug: "intervals", Title: "Sorting: intervals", Intro: "How each line is cut into the spans that get sorted.", Examples: []Example{
		srt("sort_interval_random", "Random intervals", "--interval random --length 80 --seed 7",
			"Interval: sorter.Random(80)", "Seed: 7"),
		highlight(srt("sort_random_vertical", "Random intervals, vertical, reversed",
			"--interval random --length 100 --angle 270 --reverse --seed 7",
			"Interval: sorter.Random(100)", "Path: sorter.Linear(270)", "Reverse: true", "Seed: 7")),
		srt("sort_interval_waves", "Wave intervals", "--interval waves --length 80 --seed 7",
			"Interval: sorter.Waves(80)", "Seed: 7"),
		srt("sort_interval_edges", "Between edges", "--interval edges --edge 0.2",
			"Interval: sorter.Edges(0.2)"),
		highlight(srt("sort_hue_edges", "By hue, between edges", "--key hue --interval edges --edge 0.15",
			"Key: sorter.Hue", "Interval: sorter.Edges(0.15)")),
	}},
	{Slug: "paths", Title: "Sorting: paths", Intro: "The lines the image is walked along.", Examples: []Example{
		srt("sort_path_wave", "Sine-wave path", "--path wave --amplitude 40 --wavelength 300 --lower 0.1",
			"Interval: sorter.Threshold(0.1, 0.8)", "Path: sorter.Wave(40, 300, false)"),
		srt("sort_path_wave_vertical", "Vertical sine-wave path",
			"--path wave --amplitude 30 --wavelength 250 --vertical --lower 0.15",
			"Interval: sorter.Threshold(0.15, 0.8)", "Path: sorter.Wave(30, 250, true)"),
		highlight(srt("sort_path_spiral", "Spiral", "--path spiral --spacing 6 --interval edges --edge 0.15",
			"Interval: sorter.Edges(0.15)", "Path: sorter.Spiral(0.5, 0.5, 6)")),
		srt("sort_path_concentric", "Concentric rings", "--path concentric --lower 0.1",
			"Interval: sorter.Threshold(0.1, 0.8)", "Path: sorter.Concentric(0.5, 0.5)"),
		srt("sort_path_radial", "Radial rays", "--path radial --interval edges --edge 0.1",
			"Interval: sorter.Edges(0.1)", "Path: sorter.Radial(0.5, 0.5)"),
		srt("sort_path_flow", "Flow field", "--path flow --lower 0.1 --seed 7",
			"Interval: sorter.Threshold(0.1, 0.8)", "Path: sorter.FlowField(150, 60, false, 7)"),
		highlight(srt("sort_path_flow_liquid", "Flow field, every pixel", "--path flow --interval none --seed 7",
			"Interval: sorter.None()", "Path: sorter.FlowField(150, 60, false, 7)")),
		srt("sort_path_flow_vertical", "Vertical flow field", "--path flow --vertical --lower 0.1 --seed 7",
			"Interval: sorter.Threshold(0.1, 0.8)", "Path: sorter.FlowField(150, 60, true, 7)"),
		srt("sort_path_contour", "Contour (lines bend around bright shapes)", "--path contour --warp 300 --contour-blur 10 --lower 0.05",
			"Interval: sorter.Threshold(0.05, 0.8)", "Path: sorter.Contour(300, 10, false)"),
		srt("sort_path_contour_vertical", "Vertical contour with random intervals",
			"--path contour --vertical --warp 200 --interval random --length 120 --seed 7",
			"Interval: sorter.Random(120)", "Path: sorter.Contour(200, 4, true)", "Seed: 7"),
		srt("sort_path_hilbert", "Hilbert curve", "--path hilbert --lower 0.1",
			"Interval: sorter.Threshold(0.1, 0.8)", "Path: sorter.Hilbert()"),
		srt("sort_path_hilbert_random", "Hilbert curve with random intervals", "--path hilbert --interval random --length 400 --seed 7",
			"Interval: sorter.Random(400)", "Path: sorter.Hilbert()", "Seed: 7"),
	}},
	{Slug: "modes", Title: "Sorting: modes", Intro: "What happens to the pixels of each span.", Examples: []Example{
		srt("sort_mode_shuffle", "Shuffle", "--mode shuffle --interval edges --edge 0.15 --angle 270 --seed 7",
			"Interval: sorter.Edges(0.15)", "Path: sorter.Linear(270)", "Mode: sorter.ModeShuffle", "Seed: 7"),
		srt("sort_mode_average", "Average", "--mode average --interval edges --edge 0.15 --angle 270",
			"Interval: sorter.Edges(0.15)", "Path: sorter.Linear(270)", "Mode: sorter.ModeAverage"),
		srt("sort_mode_gradient", "Gradient", "--mode gradient --interval random --length 60 --angle 90 --seed 7",
			"Interval: sorter.Random(60)", "Path: sorter.Linear(90)", "Mode: sorter.ModeGradient", "Seed: 7"),
		highlight(srt("sort_mode_smear", "Smear", "--mode smear --interval edges --edge 0.12 --angle 270",
			"Interval: sorter.Edges(0.12)", "Path: sorter.Linear(270)", "Mode: sorter.ModeSmear")),
		srt("sort_mode_channels", "Per-channel sort", "--mode channels --lower 0.1",
			"Interval: sorter.Threshold(0.1, 0.8)", "Mode: sorter.ModeChannels"),
		highlight(srt("sort_mode_channels_all", "Per-channel sort, whole columns", "--mode channels --interval none --angle 270",
			"Interval: sorter.None()", "Path: sorter.Linear(270)", "Mode: sorter.ModeChannels")),
		srt("sort_mode_mirror", "Mirror sort", "--mode mirror --lower 0.1 --angle 270",
			"Interval: sorter.Threshold(0.1, 0.8)", "Path: sorter.Linear(270)", "Mode: sorter.ModeMirror"),
		srt("sort_mode_mirror_all", "Mirror sort, whole rows", "--mode mirror --interval none",
			"Interval: sorter.None()", "Mode: sorter.ModeMirror"),
	}},
	{Slug: "tiles-and-masks", Title: "Sorting: tiles, masks and guide images", Intro: "Sort blocks instead of pixels, or let a second image decide where sorting happens.", Examples: []Example{
		srt("sort_tile", "Tile sort", "--tile 24 --interval none",
			"Interval: sorter.None()", "Tile: 24"),
		srt("sort_tile_vertical", "Vertical tile sort", "--tile 16 --angle 270 --lower 0.1",
			"Interval: sorter.Threshold(0.1, 0.8)", "Path: sorter.Linear(270)", "Tile: 16"),
		needs(Example{
			Name:  "sort_interval_image",
			Title: "Intervals from another image (the kaleidoscope example)",
			CLI:   "sort -i city.jpg -o out.jpg --interval edges --edge 0.2 --interval-image kaleidoscope.jpg",
			Code: `guide, err := imageio.Load("kaleidoscope.jpg")
out := sorter.Sort(img, sorter.Options{
	Interval: sorter.FromImage(guide, sorter.Edges(0.2)),
})`,
		}, "kaleidoscope"),
		needs(Example{
			Name:  "sort_mask",
			Title: "Mask (the white areas of the threshold example)",
			CLI:   "sort -i city.jpg -o out.jpg --interval none --angle 270 --mask threshold.jpg",
			Code: `mask, err := imageio.Load("threshold.jpg")
out := sorter.Sort(img, sorter.Options{
	Interval: sorter.None(),
	Path:     sorter.Linear(270),
	Mask:     mask,
})`,
		}, "threshold"),
	}},
	{Slug: "color", Title: "Color", Intro: "Tone and color adjustments.", Examples: []Example{
		fx("invert", "Invert", "", "out := effects.Invert(img)"),
		fx("grayscale", "Grayscale", "", "out := effects.Grayscale(img)"),
		fx("threshold", "Threshold", "--level 0.3", "out := effects.Threshold(img, 0.3)"),
		fx("posterize", "Posterize", "--levels 4", "out := effects.Posterize(img, 4)"),
		fx("brightness", "Brightness", "--amount 0.2", "out := effects.Brightness(img, 0.2)"),
		fx("contrast", "Contrast", "--factor 1.5", "out := effects.Contrast(img, 1.5)"),
		fx("gamma", "Gamma", "--gamma 1.8", "out := effects.Gamma(img, 1.8)"),
		fx("saturation", "Saturation", "--factor 2", "out := effects.Saturation(img, 2)"),
		fx("hue", "Hue rotation", "--degrees 180", "out := effects.HueShift(img, 180)"),
		fx("solarize", "Solarize", "--threshold 0.3", "out := effects.Solarize(img, 0.3)"),
		fx("swap", "Channel swap", "--order bgr", `out, err := effects.ChannelSwap(img, "bgr")`),
		fx("gradientmap", "Gradient map", `--colors "#120458,#ff2a6d,#ffd319"`,
			"colors, err := effects.ParsePalette(\"#120458,#ff2a6d,#ffd319\")\nout := effects.GradientMap(img, colors)"),
		fx("autolevels", "Auto levels", "--clip 0.005", "out := effects.AutoLevels(img, 0.005)"),
		fx("equalize", "Histogram equalization", "", "out := effects.Equalize(img)"),
		fx("levels", "Levels", "--black 0.05 --white 0.6 --gamma 1.2", "out := effects.Levels(img, 0.05, 0.6, 1.2)"),
		fx("curves", "Curves", `--points "0:0,0.25:0.15,0.6:0.75,1:1"`,
			"pts, err := effects.ParseCurve(\"0:0,0.25:0.15,0.6:0.75,1:1\")\nout := effects.Curves(img, pts)"),
		fx("temperature", "Warmer", "--warmth 0.5", "out := effects.Temperature(img, 0.5, 0)"),
		fx("temperature__cool", "Cooler with a green tint", "--warmth -0.5 --tint 0.2", "out := effects.Temperature(img, -0.5, 0.2)"),
		fx("sepia", "Sepia", "--amount 1", "out := effects.Sepia(img, 1)"),
		fx("channelmix", "Channel mixer", `--matrix "0.6,0.4,0,0.2,0.8,0,0,0.5,0.6"`,
			"out := effects.ChannelMix(img, [3][3]float64{\n\t{0.6, 0.4, 0},\n\t{0.2, 0.8, 0},\n\t{0, 0.5, 0.6},\n})"),
		fx("lut", "Built-in LUT: teal and orange", "--lut tealorange",
			"l, _ := effects.LUTByName(\"tealorange\")\nout := effects.ApplyLUT(img, l, 1)"),
		fx("lut__cross", "Built-in LUT: cross process", "--lut crossprocess --amount 0.8",
			"l, _ := effects.LUTByName(\"crossprocess\")\nout := effects.ApplyLUT(img, l, 0.8)"),
		fx("lut__file", "LUT from a .cube file", "--lut film.cube",
			"l, err := effects.LoadCubeLUT(\"film.cube\")\nout := effects.ApplyLUT(img, l, 1)"),
	}},
	{Slug: "blending", Title: "Blending", Intro: "Combine the current result with the original image, or with any other image, using blend modes.", Examples: []Example{
		pipe("blend_screen", "Screen a vertical sort over the original", "sort angle=270 interval=edges edge=0.15 | blend mode=screen"),
		pipe("blend_difference", "Difference with a mirrored copy", "flip | blend mode=difference"),
		pipe("blend_overlay", "Overlay a blurred copy for a soft glow", "gaussian sigma=12 | blend mode=overlay"),
		pipe("blend_multiply", "Multiply with a gradient map", "gradientmap colors=vapor | blend mode=multiply amount=0.8"),
		needs(fx("blend__with", "Soft light with another image (the kaleidoscope example)", "--mode softlight --with kaleidoscope.jpg",
			"top, err := imageio.Load(\"kaleidoscope.jpg\")\nout := effects.Blend(img, top, effects.BlendSoftLight, 1)"), "kaleidoscope"),
	}},
	{Slug: "glitch", Title: "Glitch", Intro: "Broken files, signals and channels.", Examples: []Example{
		highlight(fx("glitch", "Slice glitch", "--slices 30 --shift 0.15 --seed 7", "out := effects.Glitch(img, 30, 0.15, 7)")),
		fx("shift", "RGB split", "--r 12,0 --b -12,0",
			"out := effects.ChannelShift(img, image.Pt(12, 0), image.Pt(0, 0), image.Pt(-12, 0))"),
		fx("aberration", "Chromatic aberration", "--strength 0.015", "out := effects.ChromaticAberration(img, 0.015)"),
		fx("jpeg", "JPEG corruption", "--amount 15 --quality 40 --seed 2", "out, err := effects.JPEGCorrupt(img, 15, 40, 2)"),
		fx("blocks", "Block shuffle", "--size 48 --amount 0.15 --seed 7", "out := effects.BlockShuffle(img, 48, 0.15, 7)"),
		fx("melt", "Melt", "--shift 150 --seed 7", "out := effects.Melt(img, 150, 7)"),
		fx("wave", "Wave displacement", "--amplitude 10 --wavelength 60", "out := effects.Wave(img, 10, 60, false)"),
		fx("scanlines", "Scanlines", "--spacing 3 --intensity 0.4", "out := effects.Scanlines(img, 3, 0.4)"),
		fx("noise", "Noise", "--amount 0.1 --seed 7", "out := effects.Noise(img, 0.1, false, 7)"),
		fx("echo", "Echo / ghosting", "--copies 4 --offset 12,0 --decay 0.7", "out := effects.Echo(img, 4, image.Pt(12, 0), 0.7)"),
		fx("bend", "Data bending", "--delay 3001 --feedback 0.4", "out := effects.DataBend(img, 3001, 0.4)"),
		fx("bend__strong", "Data bending, longer delay", "--delay 12347 --feedback 0.6", "out := effects.DataBend(img, 12347, 0.6)"),
		fx("interlace", "Interlace", "--shift 8 --spacing 2 --darken 0.2", "out := effects.Interlace(img, 8, 2, 0.2)"),
		fx("scatter", "Pixel scatter", "--radius 4 --seed 7", "out := effects.Scatter(img, 4, 7)"),
		fx("tracking", "VHS tracking", "--bands 3 --shift 0.05 --noise 0.3 --seed 7", "out := effects.Tracking(img, 3, 0.05, 0.3, 7)"),
	}},
	{Slug: "distortion", Title: "Distortion", Intro: "Bend, twist and remap the image.", Examples: []Example{
		fx("displace", "Displacement by the image's own brightness", "--x 150", "out := effects.Displace(img, nil, 150, 0)"),
		needs(fx("displace__map", "Displacement by a map image (the kaleidoscope example)",
			"--x 120 --y 60 --map kaleidoscope.jpg",
			"dmap, err := imageio.Load(\"kaleidoscope.jpg\")\nout := effects.Displace(img, dmap, 120, 60)"), "kaleidoscope"),
		fx("swirl", "Swirl", "--degrees 180 --radius 0.5", "out := effects.Swirl(img, 180, 0.5)"),
		fx("ripple", "Ripple", "--amplitude 8 --wavelength 40", "out := effects.Ripple(img, 8, 40)"),
		fx("bulge", "Bulge (fisheye)", "--strength 0.6 --radius 0.5", "out := effects.Bulge(img, 0.6, 0.5)"),
		fx("bulge__pinch", "Pinch", "--strength -0.6 --radius 0.5", "out := effects.Bulge(img, -0.6, 0.5)"),
		highlight(fx("polar", "Polar coordinates", "", "out := effects.Polar(img, false)")),
		fx("polar__inverse", "Inverse polar coordinates", "--inverse", "out := effects.Polar(img, true)"),
		fx("curvature", "CRT curvature", "--strength 0.12", "out := effects.Curvature(img, 0.12)"),
		fx("slitscan", "Slit-scan stretch", "--position 0.6", "out := effects.SlitScan(img, 0.6, false)"),
		fx("slitscan__vertical", "Vertical slit-scan stretch", "--position 0.4 --vertical", "out := effects.SlitScan(img, 0.4, true)"),
	}},
	{Slug: "retro", Title: "Retro", Intro: "Dithering, palettes and big pixels.", Examples: []Example{
		fx("dither", "Dithering with a palette", "--method atkinson --palette gameboy",
			"out := effects.Dither(img, effects.DitherOptions{\n\tMethod:  effects.Atkinson,\n\tPalette: effects.Palettes[\"gameboy\"],\n})"),
		fx("dither__bayer", "Ordered dithering, 1 bit per channel", "--method bayer4 --levels 2",
			"out := effects.Dither(img, effects.DitherOptions{\n\tMethod: effects.Bayer4,\n\tLevels: 2,\n})"),
		fx("quantize", "Median-cut quantization", "--colors 8 --method floyd", "out := effects.Quantize(img, 8, effects.FloydSteinberg)"),
		fx("pixelate", "Pixelate", "--size 12", "out := effects.Pixelate(img, 12)"),
	}},
	{Slug: "artistic", Title: "Artistic", Intro: "Painterly and decorative effects.", Examples: []Example{
		fx("halftone", "Color halftone", "--size 8 --color", "out := effects.Halftone(img, 8, true)"),
		fx("halftone__mono", "Black and white halftone", "--size 8", "out := effects.Halftone(img, 8, false)"),
		fx("kaleidoscope", "Kaleidoscope", "--segments 6", "out := effects.Kaleidoscope(img, 6, 0)"),
		fx("bloom", "Bloom", "--threshold 0.5 --radius 16", "out := effects.Bloom(img, 0.5, 16, 1)"),
		fx("oilpaint", "Oil paint", "--radius 5", "out := effects.OilPaint(img, 5)"),
		fx("mirror", "Mirror (quad)", "--mode quad", "out := effects.Mirror(img, effects.MirrorQuad)"),
		fx("ascii", "ASCII art", "--cols 160", "out := effects.ASCIIImage(img, 160, effects.DefaultASCIIChars, false)"),
		fx("ascii__color", "Colored ASCII art", "--cols 120 --color", "out := effects.ASCIIImage(img, 120, effects.DefaultASCIIChars, true)"),
		fx("voronoi", "Voronoi mosaic", "--cells 2000 --seed 7", "out := effects.Voronoi(img, 2000, 7)"),
		fx("lowpoly", "Low poly", "--cells 800 --seed 7", "out := effects.LowPoly(img, 800, 7)"),
		fx("toon", "Toon", "--levels 6 --edge 0.2", "out := effects.Toon(img, 6, 0.2)"),
		fx("neon", "Neon edges", "--edge 0.25 --glow 12", "out := effects.Neon(img, 0.25, 12)"),
		fx("emboss", "Emboss", "--strength 2 --angle 45", "out := effects.Emboss(img, 2, 45)"),
		fx("tile", "Mirrored tiles", "--cols 3 --rows 3 --mirror", "out := effects.Tile(img, 3, 3, true)"),
		fx("droste", "Droste effect", "--scale 0.6 --steps 6", "out := effects.Droste(img, 0.6, 6)"),
	}},
	{Slug: "geometry", Title: "Geometry and filters", Intro: "Resize, crop, rotate, blur and friends.", Examples: []Example{
		fx("resize", "Resize", "--width 320", "out := effects.Resize(img, 320, 180, effects.Bilinear)"),
		fx("crop", "Crop", "--x 600 --y 200 --width 800 --height 450",
			"out := effects.Crop(img, image.Rect(600, 200, 1400, 650))"),
		fx("flip", "Flip", "--axis h", "out := effects.FlipHorizontal(img)"),
		fx("rotate", "Rotate", "--degrees 90", "out, err := effects.Rotate(img, 90)"),
		fx("blur", "Blur", "--radius 6", "out := effects.BoxBlur(img, 6)"),
		fx("sharpen", "Sharpen", "--amount 2 --radius 2", "out := effects.Sharpen(img, 2, 2)"),
		fx("edges", "Edge detection", "", "out := effects.EdgeDetect(img)"),
		fx("vignette", "Vignette", "--strength 0.8", "out := effects.Vignette(img, 0.8)"),
		fx("gaussian", "Gaussian blur", "--sigma 6", "out := effects.GaussianBlur(img, 6)"),
		fx("motionblur", "Motion blur", "--angle 0 --length 40", "out := effects.MotionBlur(img, 0, 40)"),
		fx("zoomblur", "Zoom blur", "--strength 0.3", "out := effects.ZoomBlur(img, 0.3, 24)"),
		fx("median", "Median filter", "--radius 4", "out := effects.Median(img, 4)"),
		fx("dilate", "Dilate", "--radius 3", "out := effects.Dilate(img, 3)"),
		fx("erode", "Erode", "--radius 3", "out := effects.Erode(img, 3)"),
	}},
	{Slug: "presets", Title: "Presets", Intro: "Ready-made looks. Each preset is a pipeline and can be combined with other operations.", Examples: []Example{
		preset("classic", false), preset("vertical", false), preset("spiral", false), preset("melt", true),
		preset("glitch", true), preset("vhs", true), preset("crt", false), highlight(preset("vaporwave", false)),
		preset("dream", false), preset("gameboy", false), preset("8bit", false), highlight(preset("chunky", false)),
		preset("newspaper", false), preset("oil", false), preset("planet", false), preset("databend", false),
		preset("cinematic", false), preset("noir", true), preset("comic", false), preset("neon", false),
	}},
	{Slug: "pipelines", Title: "Pipelines", Intro: "Chain operations with `|`.", Examples: []Example{
		pipe("pipeline_streaks", "Sort, split and scanlines",
			"sort angle=270 interval=edges edge=0.15 | shift | scanlines"),
		pipe("pipeline_blocky", "Low-res sort scaled back up",
			"resize scale=0.2 | sort angle=90 lower=0.1 | resize original filter=nearest | aberration"),
		pipe("pipeline_mix", "Half-strength effect with mix",
			"sort key=hue lower=0.1 | gradientmap colors=vapor | mix 0.5"),
		{Name: "mix", Title: "Mix an effect with the original",
			CLI:  `run -i city.jpg -o out.jpg "pixelate 24 | mix 0.5"`,
			Code: "out := effects.Mix(img, effects.Pixelate(img, 24), 0.5)"},
	}},
	{Slug: "animation", Title: "Animation", Intro: "Write `{from..to}` in any parameter to sweep it across the frames.", Examples: []Example{
		highlight(anim("anim_threshold", "Threshold and angle sweep", "resize width=480 | sort angle={0..90} lower={0.05..0.5}")),
		anim("anim_threshold_only", "Threshold sweep", "resize width=480 | sort lower={0.05..0.6}"),
		anim("anim_angle", "Angle sweep", "resize width=480 | sort angle={0..180} lower=0.2"),
		anim("anim_glitch", "Changing glitch", "resize width=480 | glitch seed={1..16} | shift"),
		anim("anim_kaleidoscope", "Rotating kaleidoscope", "resize width=480 | kaleidoscope rotation={0..60}"),
	}},
}

func All() []Example {
	var out []Example
	for _, s := range Sections {
		out = append(out, s.Examples...)
	}
	return out
}
