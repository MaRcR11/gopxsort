package sorter

import (
	"image/color"
	"sort"
	"strings"

	"github.com/MaRcR11/sortmancy/internal/imgutil"
)

type Key func(c color.NRGBA) float64

var (
	Lightness Key = func(c color.NRGBA) float64 {
		return (float64(max(c.R, c.G, c.B)) + float64(min(c.R, c.G, c.B))) / 510
	}

	Luminance Key = imgutil.Luma

	Hue Key = func(c color.NRGBA) float64 {
		h, _, _ := imgutil.RGBToHSL(c.R, c.G, c.B)
		return h
	}

	Saturation Key = func(c color.NRGBA) float64 {
		_, s, _ := imgutil.RGBToHSL(c.R, c.G, c.B)
		return s
	}

	Value     Key = func(c color.NRGBA) float64 { return float64(max(c.R, c.G, c.B)) / 255 }
	Intensity Key = func(c color.NRGBA) float64 { return (float64(c.R) + float64(c.G) + float64(c.B)) / 765 }
	Minimum   Key = func(c color.NRGBA) float64 { return float64(min(c.R, c.G, c.B)) / 255 }
	Chroma    Key = func(c color.NRGBA) float64 {
		return (float64(max(c.R, c.G, c.B)) - float64(min(c.R, c.G, c.B))) / 255
	}

	Red   Key = func(c color.NRGBA) float64 { return float64(c.R) / 255 }
	Green Key = func(c color.NRGBA) float64 { return float64(c.G) / 255 }
	Blue  Key = func(c color.NRGBA) float64 { return float64(c.B) / 255 }
	Alpha Key = func(c color.NRGBA) float64 { return float64(c.A) / 255 }
)

var keys = map[string]Key{
	"lightness":  Lightness,
	"luminance":  Luminance,
	"hue":        Hue,
	"saturation": Saturation,
	"value":      Value,
	"brightness": Value,
	"intensity":  Intensity,
	"minimum":    Minimum,
	"chroma":     Chroma,
	"red":        Red,
	"green":      Green,
	"blue":       Blue,
	"alpha":      Alpha,
}

func KeyByName(name string) (Key, bool) {
	k, ok := keys[strings.ToLower(name)]
	return k, ok
}

func KeyNames() []string {
	names := make([]string, 0, len(keys))
	for n := range keys {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
