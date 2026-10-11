package imgutil

import (
	"image"
	"image/color"
	"slices"
)

type box struct {
	px   []color.NRGBA
	axis int
	span int
}

func newBox(px []color.NRGBA) box {
	lo := [3]uint8{255, 255, 255}
	hi := [3]uint8{}
	for _, c := range px {
		ch := [3]uint8{c.R, c.G, c.B}
		for i, v := range ch {
			lo[i] = min(lo[i], v)
			hi[i] = max(hi[i], v)
		}
	}
	b := box{px: px}
	for i := range lo {
		if s := int(hi[i]) - int(lo[i]); s > b.span {
			b.axis, b.span = i, s
		}
	}
	return b
}

func (b box) mean() color.NRGBA {
	var r, g, bl int
	for _, c := range b.px {
		r, g, bl = r+int(c.R), g+int(c.G), bl+int(c.B)
	}
	n := len(b.px)
	return color.NRGBA{uint8((r + n/2) / n), uint8((g + n/2) / n), uint8((bl + n/2) / n), 255}
}

func channel(c color.NRGBA, axis int) uint8 {
	switch axis {
	case 0:
		return c.R
	case 1:
		return c.G
	}
	return c.B
}

func MedianCut(img *image.NRGBA, n int) color.Palette {
	n = max(n, 1)
	total := img.Rect.Dx() * img.Rect.Dy()
	if total == 0 {
		return color.Palette{color.NRGBA{A: 255}}
	}
	step := max(total/65536, 1)
	px := make([]color.NRGBA, 0, total/step+1)
	for i := 0; i < total; i += step {
		px = append(px, At(img.Pix, i))
	}
	boxes := []box{newBox(px)}
	for len(boxes) < n {
		best, bestScore := -1, 0
		for i, b := range boxes {
			if score := b.span * len(b.px); len(b.px) > 1 && b.span > 0 && score > bestScore {
				best, bestScore = i, score
			}
		}
		if best < 0 {
			break
		}
		b := boxes[best]
		slices.SortFunc(b.px, func(x, y color.NRGBA) int {
			return int(channel(x, b.axis)) - int(channel(y, b.axis))
		})
		mid := len(b.px) / 2
		boxes[best] = newBox(b.px[:mid])
		boxes = append(boxes, newBox(b.px[mid:]))
	}
	pal := make(color.Palette, len(boxes))
	for i, b := range boxes {
		pal[i] = b.mean()
	}
	return pal
}
