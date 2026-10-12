package effects

import (
	"image"
	"math"
	"strings"

	"github.com/MaRcR11/sortmancy/internal/imgutil"
)

type BlendMode int

const (
	BlendNormal BlendMode = iota
	BlendMultiply
	BlendScreen
	BlendOverlay
	BlendDifference
	BlendAdd
	BlendSubtract
	BlendDarken
	BlendLighten
	BlendSoftLight
	BlendHardLight
	BlendColorDodge
	BlendColorBurn
	BlendExclusion
)

var blendNames = map[string]BlendMode{
	"normal":     BlendNormal,
	"multiply":   BlendMultiply,
	"screen":     BlendScreen,
	"overlay":    BlendOverlay,
	"difference": BlendDifference,
	"add":        BlendAdd,
	"subtract":   BlendSubtract,
	"darken":     BlendDarken,
	"lighten":    BlendLighten,
	"softlight":  BlendSoftLight,
	"hardlight":  BlendHardLight,
	"colordodge": BlendColorDodge,
	"colorburn":  BlendColorBurn,
	"exclusion":  BlendExclusion,
}

func BlendByName(name string) (BlendMode, bool) {
	m, ok := blendNames[strings.ToLower(name)]
	return m, ok
}

func BlendNames() []string { return sortedKeys(blendNames) }

func blendChannel(mode BlendMode, a, b float64) float64 {
	switch mode {
	case BlendMultiply:
		return a * b
	case BlendScreen:
		return 1 - (1-a)*(1-b)
	case BlendOverlay:
		if a < 0.5 {
			return 2 * a * b
		}
		return 1 - 2*(1-a)*(1-b)
	case BlendDifference:
		return math.Abs(a - b)
	case BlendAdd:
		return math.Min(1, a+b)
	case BlendSubtract:
		return math.Max(0, a-b)
	case BlendDarken:
		return math.Min(a, b)
	case BlendLighten:
		return math.Max(a, b)
	case BlendSoftLight:
		return (1-2*b)*a*a + 2*b*a
	case BlendHardLight:
		return blendChannel(BlendOverlay, b, a)
	case BlendColorDodge:
		if b >= 1 {
			return 1
		}
		return math.Min(1, a/(1-b))
	case BlendColorBurn:
		if b <= 0 {
			return 0
		}
		return 1 - math.Min(1, (1-a)/b)
	case BlendExclusion:
		return a + b - 2*a*b
	}
	return b
}

func Blend(base, top image.Image, mode BlendMode, amount float64) *image.NRGBA {
	dst := imgutil.ToNRGBA(base)
	w, h := dst.Rect.Dx(), dst.Rect.Dy()
	t := imgutil.ToNRGBA(top)
	if t.Rect.Dx() != w || t.Rect.Dy() != h {
		t = Resize(t, w, h, Bilinear)
	}
	imgutil.Parallel(h, func(y0, y1 int) {
		for i := y0 * w; i < y1*w; i++ {
			p := dst.Pix[i*4 : i*4+4 : i*4+4]
			q := t.Pix[i*4 : i*4+4 : i*4+4]
			k := amount * float64(q[3]) / 255
			for c := 0; c < 3; c++ {
				a, b := float64(p[c])/255, float64(q[c])/255
				v := blendChannel(mode, a, b)
				p[c] = imgutil.Clamp8((a + (v-a)*k) * 255)
			}
		}
	})
	return dst
}
