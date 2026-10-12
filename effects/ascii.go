package effects

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/MaRcR11/sortmancy/internal/imgutil"
)

const DefaultASCIIChars = " .:-=+*#%@"

const (
	glyphW = 7
	glyphH = 13
)

type asciiCell struct {
	ch    byte
	color color.NRGBA
}

func asciiGrid(img image.Image, cols int, charset string) ([][]asciiCell, int) {
	src := imgutil.ToNRGBA(img)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	if charset == "" {
		charset = DefaultASCIIChars
	}
	cols = max(1, min(cols, w))
	cellW := float64(w) / float64(cols)
	cellH := cellW * glyphH / glyphW
	rows := max(1, int(math.Round(float64(h)/cellH)))
	grid := make([][]asciiCell, rows)
	imgutil.Parallel(rows, func(r0, r1 int) {
		for r := r0; r < r1; r++ {
			grid[r] = make([]asciiCell, cols)
			y0, y1 := int(float64(r)*cellH), min(int(float64(r+1)*cellH), h)
			for c := 0; c < cols; c++ {
				x0, x1 := int(float64(c)*cellW), min(int(float64(c+1)*cellW), w)
				var sum [3]int
				n := 0
				for y := y0; y < max(y1, y0+1) && y < h; y++ {
					for x := x0; x < max(x1, x0+1) && x < w; x++ {
						i := (y*w + x) * 4
						sum[0], sum[1], sum[2] = sum[0]+int(src.Pix[i]), sum[1]+int(src.Pix[i+1]), sum[2]+int(src.Pix[i+2])
						n++
					}
				}
				if n == 0 {
					n = 1
				}
				avg := color.NRGBA{uint8(sum[0] / n), uint8(sum[1] / n), uint8(sum[2] / n), 255}
				idx := min(int(imgutil.Luma(avg)*float64(len(charset))), len(charset)-1)
				grid[r][c] = asciiCell{charset[idx], avg}
			}
		}
	})
	return grid, rows
}

func ASCII(img image.Image, cols int, charset string) string {
	grid, _ := asciiGrid(img, cols, charset)
	var b strings.Builder
	for _, row := range grid {
		for _, cell := range row {
			b.WriteByte(cell.ch)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func ASCIIImage(img image.Image, cols int, charset string, colored bool) *image.NRGBA {
	grid, rows := asciiGrid(img, cols, charset)
	cols = len(grid[0])
	dst := image.NewNRGBA(image.Rect(0, 0, cols*glyphW, rows*glyphH))
	draw.Draw(dst, dst.Rect, image.NewUniform(color.NRGBA{0, 0, 0, 255}), image.Point{}, draw.Src)
	d := font.Drawer{Dst: dst, Face: basicfont.Face7x13}
	white := image.NewUniform(color.NRGBA{235, 235, 235, 255})
	for r, row := range grid {
		for c, cell := range row {
			if cell.ch == ' ' {
				continue
			}
			d.Src = white
			if colored {
				d.Src = image.NewUniform(brighten(cell.color))
			}
			d.Dot = fixed.P(c*glyphW, r*glyphH+basicfont.Face7x13.Ascent)
			d.DrawString(string(cell.ch))
		}
	}
	return dst
}

func brighten(c color.NRGBA) color.NRGBA {
	m := max(c.R, c.G, c.B)
	if m == 0 {
		return color.NRGBA{235, 235, 235, 255}
	}
	f := 235 / float64(m)
	return color.NRGBA{imgutil.Clamp8(float64(c.R) * f), imgutil.Clamp8(float64(c.G) * f), imgutil.Clamp8(float64(c.B) * f), 255}
}
