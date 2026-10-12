package effects

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/MaRcR11/sortmancy/internal/imgutil"
)

func Voronoi(img image.Image, cells int, seed int64) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	step := math.Max(2, math.Sqrt(float64(w*h)/float64(max(cells, 1))))
	gw, gh := int(math.Ceil(float64(w)/step)), int(math.Ceil(float64(h)/step))
	rng := rand.New(rand.NewSource(seed))
	sites := make([][2]float64, gw*gh)
	for gy := 0; gy < gh; gy++ {
		for gx := 0; gx < gw; gx++ {
			sites[gy*gw+gx] = [2]float64{(float64(gx) + rng.Float64()) * step, (float64(gy) + rng.Float64()) * step}
		}
	}
	owner := make([]int32, w*h)
	imgutil.Parallel(h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			gy := int(float64(y) / step)
			for x := 0; x < w; x++ {
				gx := int(float64(x) / step)
				best, bestD := 0, math.Inf(1)
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						cx, cy := gx+dx, gy+dy
						if cx < 0 || cy < 0 || cx >= gw || cy >= gh {
							continue
						}
						s := sites[cy*gw+cx]
						ddx, ddy := float64(x)+0.5-s[0], float64(y)+0.5-s[1]
						if d := ddx*ddx + ddy*ddy; d < bestD {
							best, bestD = cy*gw+cx, d
						}
					}
				}
				owner[y*w+x] = int32(best)
			}
		}
	})
	return fillRegions(src, owner, len(sites))
}

func fillRegions(src *image.NRGBA, owner []int32, regions int) *image.NRGBA {
	sum := make([][5]int, regions)
	for i, o := range owner {
		if o < 0 {
			continue
		}
		p := src.Pix[i*4 : i*4+4]
		s := &sum[o]
		s[0], s[1], s[2], s[3], s[4] = s[0]+int(p[0]), s[1]+int(p[1]), s[2]+int(p[2]), s[3]+int(p[3]), s[4]+1
	}
	dst := imgutil.ToNRGBA(src)
	for i, o := range owner {
		if o < 0 || sum[o][4] == 0 {
			continue
		}
		s, n := sum[o], sum[o][4]
		imgutil.Set(dst.Pix, i, color.NRGBA{uint8(s[0] / n), uint8(s[1] / n), uint8(s[2] / n), uint8(s[3] / n)})
	}
	return dst
}

func LowPoly(img image.Image, cells int, seed int64) *image.NRGBA {
	src := imgutil.ToNRGBA(img)
	w, h := src.Rect.Dx(), src.Rect.Dy()
	step := math.Max(4, math.Sqrt(float64(w*h)/float64(max(cells, 1))))
	gw, gh := int(math.Ceil(float64(w)/step)), int(math.Ceil(float64(h)/step))
	rng := rand.New(rand.NewSource(seed))
	pts := make([][2]float64, (gw+1)*(gh+1))
	for gy := 0; gy <= gh; gy++ {
		for gx := 0; gx <= gw; gx++ {
			x := math.Min(float64(gx)*step, float64(w))
			y := math.Min(float64(gy)*step, float64(h))
			if gx > 0 && gx < gw {
				x += (rng.Float64() - 0.5) * step * 0.8
			}
			if gy > 0 && gy < gh {
				y += (rng.Float64() - 0.5) * step * 0.8
			}
			pts[gy*(gw+1)+gx] = [2]float64{x, y}
		}
	}
	var tris [][3][2]float64
	for gy := 0; gy < gh; gy++ {
		for gx := 0; gx < gw; gx++ {
			a, b := pts[gy*(gw+1)+gx], pts[gy*(gw+1)+gx+1]
			c, d := pts[(gy+1)*(gw+1)+gx], pts[(gy+1)*(gw+1)+gx+1]
			if rng.Intn(2) == 0 {
				tris = append(tris, [3][2]float64{a, b, d}, [3][2]float64{a, d, c})
			} else {
				tris = append(tris, [3][2]float64{a, b, c}, [3][2]float64{b, d, c})
			}
		}
	}
	owner := make([]int32, w*h)
	for i := range owner {
		owner[i] = -1
	}
	for ti, t := range tris {
		minX := int(math.Floor(math.Min(t[0][0], math.Min(t[1][0], t[2][0]))))
		maxX := int(math.Ceil(math.Max(t[0][0], math.Max(t[1][0], t[2][0]))))
		minY := int(math.Floor(math.Min(t[0][1], math.Min(t[1][1], t[2][1]))))
		maxY := int(math.Ceil(math.Max(t[0][1], math.Max(t[1][1], t[2][1]))))
		edge := func(a, b [2]float64, px, py float64) float64 {
			return (b[0]-a[0])*(py-a[1]) - (b[1]-a[1])*(px-a[0])
		}
		area := edge(t[0], t[1], t[2][0], t[2][1])
		if area == 0 {
			continue
		}
		for y := max(minY, 0); y < min(maxY, h); y++ {
			for x := max(minX, 0); x < min(maxX, w); x++ {
				px, py := float64(x)+0.5, float64(y)+0.5
				w0 := edge(t[1], t[2], px, py) / area
				w1 := edge(t[2], t[0], px, py) / area
				w2 := edge(t[0], t[1], px, py) / area
				if w0 >= -1e-9 && w1 >= -1e-9 && w2 >= -1e-9 && owner[y*w+x] < 0 {
					owner[y*w+x] = int32(ti)
				}
			}
		}
	}
	return fillRegions(src, owner, len(tris))
}
