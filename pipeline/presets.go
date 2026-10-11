package pipeline

import (
	"fmt"
	"sort"
	"strings"
)

var Presets = map[string]string{
	"classic":   "sort key=luminance threshold-key=luminance lower=0.5 upper=1",
	"vertical":  "sort angle=270 interval=edges edge=0.2",
	"vhs":       "shift r=4,0 b=-4,0 | wave amplitude=2 wavelength=40 | noise amount=0.06 mono | scanlines spacing=2 intensity=0.25 | saturation 1.3",
	"vaporwave": "sort key=hue angle=90 lower=0.35 | gradientmap colors=#2e2157,#ff71ce,#01cdfe,#fffb96 | mix 0.6 | scanlines spacing=3 intensity=0.2",
	"gameboy":   "resize scale=0.33 | dither atkinson palette=gameboy | resize original filter=nearest",
	"glitch":    "sort lower=0.35 randomness=0.5 | glitch slices=30 | shift | jpeg amount=2",
	"melt":      "sort angle=270 interval=edges edge=0.15 mode=smear | melt shift=120",
	"newspaper": "grayscale | contrast 1.4 | halftone size=6",
	"crt":       "bloom | aberration strength=0.006 | scanlines spacing=3 intensity=0.35 | vignette 0.5",
	"dream":     "sort path=concentric lower=0.3 | bloom threshold=0.5 radius=12 intensity=0.8 | saturation 1.3",
	"chunky":    "resize scale=0.125 | sort angle=90 lower=0.2 | resize original filter=nearest",
	"oil":       "oilpaint radius=4 | saturation 1.2",
	"8bit":      "resize scale=0.25 | quantize colors=16 | resize original filter=nearest",
	"spiral":    "sort path=spiral spacing=6 interval=edges edge=0.2",
}

func PresetNames() []string {
	names := make([]string, 0, len(Presets))
	for n := range Presets {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func expandPresets(s string, depth int) (string, error) {
	if depth > 8 {
		return "", fmt.Errorf("presets nested too deeply in %q", s)
	}
	stages := strings.Split(s, "|")
	changed := false
	for i, stage := range stages {
		name, ok := strings.CutPrefix(strings.TrimSpace(stage), "@")
		if !ok {
			continue
		}
		p, ok := Presets[strings.ToLower(name)]
		if !ok {
			return "", fmt.Errorf("unknown preset %q (choose from %s)", name, strings.Join(PresetNames(), ", "))
		}
		stages[i] = p
		changed = true
	}
	out := strings.Join(stages, "|")
	if changed {
		return expandPresets(out, depth+1)
	}
	return out, nil
}
