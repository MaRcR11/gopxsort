package sortcery

import (
	"image"

	"github.com/MaRcR11/sortcery/imageio"
	"github.com/MaRcR11/sortcery/pipeline"
	"github.com/MaRcR11/sortcery/sorter"
)

func Sort(img image.Image, opts sorter.Options) *image.NRGBA {
	return sorter.Sort(img, opts)
}

func Apply(img image.Image, ops string) (*image.NRGBA, error) {
	p, err := pipeline.Parse(ops)
	if err != nil {
		return nil, err
	}
	return p.Apply(img)
}

func ApplyFile(input, output, ops string) error {
	img, err := imageio.Load(input)
	if err != nil {
		return err
	}
	out, err := Apply(img, ops)
	if err != nil {
		return err
	}
	return imageio.Save(output, out, imageio.Options{})
}
