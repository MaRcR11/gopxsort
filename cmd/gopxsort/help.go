package main

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/MaRcR11/gopxsort/pipeline"
)

const usage = `gopxsort - pixel sorting and glitch effects

USAGE:
    gopxsort <operation> -i <input> -o <output> [--param value | param=value ...]
    gopxsort run -i <input> -o <output> "<op> [params] | <op> [params] | ..."
    gopxsort @<preset> -i <input> -o <output>
    gopxsort -i <input> -o <output> [-t <0-255>]      (classic threshold sort)

COMMANDS:
    <operation>     run a single operation, e.g. sort, glitch, dither
    run             run a pipeline of operations separated by "|"
    @<preset>       run a ready-made look, e.g. @vhs (presets can also be pipeline stages)
    list            list all operations
    presets         list all presets
    help <op>       show the parameters of an operation
    version         print the version

OPTIONS:
    -i, --input <path>      input image (png, jpg, gif, bmp, tiff, webp), a folder,
                            a glob like "photos/*.jpg", or - for stdin
    -o, --output <path>     output image (extension selects the format), a folder
                            for batch input, frame%03d.png for frame files, or - for stdout
    -f, --format <ext>      output format, overriding the extension (png, jpg, gif, bmp, tiff)
    -q, --quality <1-100>   JPEG quality [default: 95]
        --seed <n>          seed for random operations (printed when omitted)
        --frames <n>        render an animation with n frames
        --delay <ms>        animation frame delay [default: 80, or the input GIF's delays]
        --pingpong          animate parameters forth and back so the animation loops
        --quiet             do not print progress
    -t, --threshold <n>     classic mode only: sort pixels brighter than n
    -h, --help              print help

ANIMATION:
    Write a range as {from..to} in any parameter; it is swept over the frames.
    Animated GIF input is processed frame by frame.

EXAMPLES:
    gopxsort sort -i city.jpg -o out.png --angle 90 --key hue
    gopxsort sort -i city.jpg -o out.png interval=edges edge=0.2 reverse
    gopxsort run -i city.jpg -o out.png "sort angle=45 lower=0.4 | shift | scanlines"
    gopxsort run -i city.jpg -o out.png "@chunky | @crt"
    gopxsort run -i city.jpg -o anim.gif --frames 30 --pingpong "sort lower={0.1..0.7}"
    gopxsort run -i "photos/*.jpg" -o sorted/ "sort angle=90"
    cat in.png | gopxsort @vhs -i - -o - > out.png
`

func printUsage(w io.Writer) {
	fmt.Fprint(w, usage)
	fmt.Fprintln(w, "\nOPERATIONS:")
	printOpList(w)
}

func printOpList(w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	for _, s := range pipeline.Specs() {
		fmt.Fprintf(tw, "    %s\t%s\n", s.Name, s.Help)
	}
	tw.Flush()
}

func printOpHelp(w io.Writer, name string) error {
	s, ok := pipeline.Lookup(name)
	if !ok {
		return fmt.Errorf("unknown operation %q, run \"gopxsort list\"", name)
	}
	fmt.Fprintf(w, "%s - %s\n\nUSAGE:\n    gopxsort %s -i <input> -o <output>", s.Name, s.Help, s.Name)
	if len(s.Params) > 0 {
		fmt.Fprint(w, " [--param value ...]")
	}
	fmt.Fprintln(w)
	if len(s.Params) == 0 {
		return nil
	}
	fmt.Fprintln(w, "\nPARAMETERS:")
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, p := range s.Params {
		def := ""
		if p.Default != "" {
			def = " [default: " + p.Default + "]"
		}
		fmt.Fprintf(tw, "    --%s <%s>\t%s%s\n", p.Name, p.Kind, p.Help, def)
		if len(p.Choices) > 0 {
			fmt.Fprintf(tw, "    \t  one of: %s\n", strings.Join(p.Choices, ", "))
		}
	}
	tw.Flush()
	fmt.Fprintf(w, "\nParameters can also be written as name=value, or as positional values in the order above.\n")
	return nil
}

func printPresets(w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	for _, name := range pipeline.PresetNames() {
		fmt.Fprintf(tw, "    @%s\t%s\n", name, pipeline.Presets[name])
	}
	tw.Flush()
}
