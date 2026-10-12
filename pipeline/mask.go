package pipeline

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/MaRcR11/sortmancy/effects"
	"github.com/MaRcR11/sortmancy/imageio"
	"github.com/MaRcR11/sortmancy/internal/imgutil"
)

var commonParams = []Param{
	str("mask", "", "image file; the operation only changes the white areas"),
	boolean("invert-mask", "change the black areas of the mask (or everything outside the region) instead"),
	str("region", "", "only change this rectangle: x,y,width,height in pixels"),
	integer("feather", "0", "soften the edge of the mask or region by this many pixels"),
}

func CommonParams() []Param { return append([]Param(nil), commonParams...) }

func (s *Spec) own(name string) bool {
	for _, p := range s.Params {
		if p.Name == name {
			return true
		}
	}
	return false
}

func commonValues(spec *Spec, values map[string]string) map[string]string {
	out := map[string]string{}
	for _, p := range commonParams {
		if v, ok := values[p.Name]; ok && !spec.own(p.Name) {
			out[p.Name] = v
		}
	}
	return out
}

func maskWrapper(spec *Spec, a *Args, inner SourceFunc) (SourceFunc, error) {
	c := commonValues(spec, a.values)
	if c["mask"] == "" && c["region"] == "" {
		for _, k := range []string{"invert-mask", "feather"} {
			if _, ok := c[k]; ok {
				return nil, fmt.Errorf("%s: %s needs mask or region", spec.Name, k)
			}
		}
		return inner, nil
	}
	invert, feather := a.Bool("invert-mask"), a.Int("feather")
	var file image.Image
	if f := c["mask"]; f != "" {
		var err error
		if file, err = imageio.Load(f); err != nil {
			return nil, fmt.Errorf("%s: mask: %w", spec.Name, err)
		}
	}
	var region image.Rectangle
	hasRegion := c["region"] != ""
	if hasRegion {
		var err error
		if region, err = effects.ParseRegion(c["region"]); err != nil {
			return nil, fmt.Errorf("%s: %w", spec.Name, err)
		}
	}
	return func(img, src *image.NRGBA) (*image.NRGBA, error) {
		before := imgutil.ToNRGBA(img)
		after, err := inner(img, src)
		if err != nil {
			return nil, err
		}
		w, h := before.Rect.Dx(), before.Rect.Dy()
		mask := image.NewGray(before.Rect)
		draw.Draw(mask, mask.Rect, image.White, image.Point{}, draw.Src)
		if file != nil {
			mask = imgutil.ToGray(imgutil.ResizeNearest(file, w, h))
		}
		if hasRegion {
			r := effects.RegionMask(w, h, region)
			for i := range mask.Pix {
				mask.Pix[i] = uint8(int(mask.Pix[i]) * int(r.Pix[i]) / 255)
			}
		}
		return effects.ApplyMask(before, after, mask, invert, feather)
	}, nil
}
