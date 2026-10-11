package pipeline

import (
	"fmt"
	"image"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var rangePattern = regexp.MustCompile(`\{\s*(-?[0-9]*\.?[0-9]+)\s*\.\.\s*(-?[0-9]*\.?[0-9]+)\s*\}`)

func HasRanges(template string) bool { return rangePattern.MatchString(template) }

func Expand(template string, t float64) string {
	return rangePattern.ReplaceAllStringFunc(template, func(m string) string {
		parts := rangePattern.FindStringSubmatch(m)
		from, _ := strconv.ParseFloat(parts[1], 64)
		to, _ := strconv.ParseFloat(parts[2], 64)
		v := from + (to-from)*t
		if !strings.Contains(parts[1], ".") && !strings.Contains(parts[2], ".") {
			return strconv.Itoa(int(math.Round(v)))
		}
		return strconv.FormatFloat(math.Round(v*1e4)/1e4, 'f', -1, 64)
	})
}

type AnimateOptions struct {
	Frames   int
	Seed     int64
	PingPong bool
	Progress func(frame, total int)
}

func FrameTime(i, n int, pingpong bool) float64 {
	if n <= 1 {
		return 0
	}
	t := float64(i) / float64(n-1)
	if pingpong {
		t = 1 - math.Abs(2*float64(i)/float64(n)-1)
	}
	return t
}

func Animate(frames []image.Image, template string, opts AnimateOptions) ([]*image.NRGBA, error) {
	if len(frames) == 0 {
		return nil, fmt.Errorf("animate: no input frames")
	}
	n := max(opts.Frames, len(frames))
	out := make([]*image.NRGBA, n)
	for i := 0; i < n; i++ {
		p, err := ParseSeed(Expand(template, FrameTime(i, n, opts.PingPong)), opts.Seed)
		if err != nil {
			return nil, err
		}
		if out[i], err = p.Apply(frames[i%len(frames)]); err != nil {
			return nil, fmt.Errorf("frame %d: %w", i, err)
		}
		if opts.Progress != nil {
			opts.Progress(i+1, n)
		}
	}
	return out, nil
}
