package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/MaRcR11/gopxsort/imageio"
	"github.com/MaRcR11/gopxsort/pipeline"
)

var version = "dev"

const (
	infoColor  = "\033[92m%s\033[0m"
	errorColor = "\033[91m%s\033[0m"
)

var quiet bool

func printInfo(message string) {
	if quiet {
		return
	}
	fmt.Fprint(os.Stderr, colorize(infoColor, "[INFO] "))
	fmt.Fprintln(os.Stderr, message)
}

func printError(message string) {
	fmt.Fprint(os.Stderr, colorize(errorColor, "[ERROR] "))
	fmt.Fprintln(os.Stderr, message)
}

func colorize(format, s string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	return fmt.Sprintf(format, s)
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		printError(err.Error())
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return errors.New("no command given")
	}
	switch cmd := args[0]; cmd {
	case "help", "-h", "--help":
		if len(args) > 1 {
			return printOpHelp(stdout, args[1])
		}
		printUsage(stdout)
		return nil
	case "list":
		printOpList(stdout)
		return nil
	case "presets":
		printPresets(stdout)
		return nil
	case "version", "--version":
		fmt.Fprintln(stdout, "gopxsort", version)
		return nil
	case "run":
		o, err := parseOptions(args[1:], nil)
		if err != nil || o.help {
			printUsage(stdout)
			return err
		}
		return process(o, strings.Join(o.rest, " "))
	default:
		if strings.HasPrefix(cmd, "-") {
			return runLegacy(args)
		}
		if strings.HasPrefix(cmd, "@") {
			o, err := parseOptions(args[1:], nil)
			if err != nil {
				return err
			}
			return process(o, strings.Join(append([]string{cmd}, o.rest...), " "))
		}
		spec, ok := pipeline.Lookup(cmd)
		if !ok {
			return fmt.Errorf("unknown command %q, run \"gopxsort help\" for usage", cmd)
		}
		o, err := parseOptions(args[1:], spec)
		if err != nil {
			return err
		}
		if o.help {
			return printOpHelp(stdout, spec.Name)
		}
		return process(o, spec.Name+" "+strings.Join(o.rest, " "))
	}
}

type options struct {
	input, output string
	quality       int
	seed          int64
	seedSet       bool
	help          bool
	threshold     int
	format        string
	frames        int
	delay         int
	pingpong      bool
	rest          []string
}

func parseOptions(args []string, spec *pipeline.Spec) (options, error) {
	o := options{threshold: -1}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		next := func() (string, error) {
			if i+1 >= len(args) {
				return "", fmt.Errorf("%s needs a value", arg)
			}
			i++
			return args[i], nil
		}
		var err error
		name, value, hasValue := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		isFlag := strings.HasPrefix(arg, "-") && !isNumber(arg)

		switch {
		case !isFlag:
			o.rest = append(o.rest, arg)
		case name == "i" || name == "input":
			o.input, err = valueOrNext(value, hasValue, next)
		case name == "o" || name == "output":
			o.output, err = valueOrNext(value, hasValue, next)
		case name == "q" || name == "quality":
			o.quality, err = intFlag(arg, value, hasValue, next)
		case (name == "t" || name == "threshold") && spec == nil:
			o.threshold, err = intFlag(arg, value, hasValue, next)
		case name == "seed":
			var s string
			if s, err = valueOrNext(value, hasValue, next); err == nil {
				o.seed, err = strconv.ParseInt(s, 10, 64)
				o.seedSet = true
			}
		case name == "f" || name == "format":
			o.format, err = valueOrNext(value, hasValue, next)
		case name == "frames":
			o.frames, err = intFlag(arg, value, hasValue, next)
		case name == "delay":
			o.delay, err = intFlag(arg, value, hasValue, next)
		case name == "pingpong":
			o.pingpong = true
		case name == "h" || name == "help":
			o.help = true
		case name == "quiet":
			quiet = true
		case spec != nil:
			p, ok := spec.Param(name)
			if !ok {
				return o, fmt.Errorf("%s: unknown option %s, run \"gopxsort help %s\"", spec.Name, arg, spec.Name)
			}
			switch {
			case hasValue:
			case p.Kind == pipeline.KindBool:
				value = "true"
				if i+1 < len(args) {
					if _, perr := strconv.ParseBool(args[i+1]); perr == nil {
						value, _ = next()
					}
				}
			default:
				value, err = next()
			}
			o.rest = append(o.rest, name+"="+value)
		default:
			return o, fmt.Errorf("unknown option %s", arg)
		}
		if err != nil {
			return o, err
		}
	}
	return o, nil
}

func valueOrNext(value string, hasValue bool, next func() (string, error)) (string, error) {
	if hasValue {
		return value, nil
	}
	return next()
}

func intFlag(arg, value string, hasValue bool, next func() (string, error)) (int, error) {
	s, err := valueOrNext(value, hasValue, next)
	if err != nil {
		return 0, err
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer, got %q", arg, s)
	}
	return v, nil
}

func isNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func runLegacy(args []string) error {
	o, err := parseOptions(args, nil)
	if err != nil {
		return err
	}
	if o.help {
		printUsage(os.Stdout)
		return nil
	}
	if len(o.rest) > 0 {
		return fmt.Errorf("unexpected argument %q", o.rest[0])
	}
	t := max(o.threshold, 0)
	if t > 255 {
		return errors.New("luminance threshold must be in the range 0-255")
	}
	ops := fmt.Sprintf("sort key=luminance threshold-key=luminance lower=%.4f upper=1", float64(t)/255)
	return process(o, ops)
}

func process(o options, ops string) error {
	if o.input == "" || o.output == "" {
		return errors.New("input (-i) and output (-o) are required")
	}
	if strings.TrimSpace(ops) == "" {
		return errors.New("no operations given, e.g. gopxsort run -i in.jpg -o out.png \"sort angle=90 | glitch\"")
	}
	if o.format != "" {
		f, err := imageio.ParseFormat(o.format)
		if err != nil {
			return err
		}
		if !imageio.CanEncode(f) {
			return fmt.Errorf("writing %s is not supported", f)
		}
		o.format = string(f)
	}
	if !o.seedSet {
		o.seed = time.Now().UnixNano() % 1_000_000_000
	}
	if _, err := pipeline.ParseSeed(pipeline.Expand(ops, 0), o.seed); err != nil {
		return err
	}

	inputs, err := batchInputs(o.input)
	if err != nil {
		return err
	}
	if inputs == nil {
		return processOne(o, o.input, o.output, ops)
	}
	if len(inputs) == 0 {
		return fmt.Errorf("no supported images match %q", o.input)
	}
	if err := os.MkdirAll(o.output, 0o755); err != nil {
		return err
	}
	for _, in := range inputs {
		if err := processOne(o, in, batchOutput(o, in), ops); err != nil {
			return fmt.Errorf("%s: %w", in, err)
		}
	}
	printInfo(fmt.Sprintf("Processed %d images into %s", len(inputs), o.output))
	return nil
}

func batchInputs(input string) ([]string, error) {
	var matches []string
	if info, err := os.Stat(input); err == nil && info.IsDir() {
		entries, err := os.ReadDir(input)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			matches = append(matches, filepath.Join(input, e.Name()))
		}
	} else if strings.ContainsAny(input, "*?[") {
		if matches, err = filepath.Glob(input); err != nil {
			return nil, err
		}
	} else {
		return nil, nil
	}
	files := []string{}
	for _, m := range matches {
		if info, err := os.Stat(m); err == nil && !info.IsDir() && imageio.IsSupported(m) {
			files = append(files, m)
		}
	}
	return files, nil
}

func batchOutput(o options, in string) string {
	base := strings.TrimSuffix(filepath.Base(in), filepath.Ext(in))
	ext := filepath.Ext(in)
	if f, err := imageio.FormatFromPath(in); err != nil || !imageio.CanEncode(f) {
		ext = ".png"
	}
	if o.format == string(imageio.JPEG) {
		ext = ".jpg"
	} else if o.format != "" {
		ext = "." + o.format
	}
	return filepath.Join(o.output, base+ext)
}

func processOne(o options, input, output, ops string) error {
	animatedOut := strings.Contains(output, "%")
	if output != imageio.Stdio && !animatedOut && o.format == "" {
		f, err := imageio.FormatFromPath(output)
		if err != nil {
			return err
		}
		if !imageio.CanEncode(f) {
			return fmt.Errorf("writing %s is not supported", f)
		}
	}

	printInfo("Opening file: " + input)
	anim, err := imageio.LoadAnimation(input)
	if err != nil {
		return err
	}
	b := anim.Frames[0].Bounds()
	frames := max(o.frames, len(anim.Frames))
	start := time.Now()

	if frames <= 1 {
		p, err := pipeline.ParseSeed(pipeline.Expand(ops, 0), o.seed)
		if err != nil {
			return err
		}
		printInfo(fmt.Sprintf("Applying %q to %dx%d image (seed %d)", p.String(), b.Dx(), b.Dy(), o.seed))
		out, err := p.Apply(anim.Frames[0])
		if err != nil {
			return err
		}
		printInfo(fmt.Sprintf("Done in %s", time.Since(start).Round(time.Millisecond)))
		if err := imageio.Save(output, out, imageio.Options{Quality: o.quality, Format: imageio.Format(o.format)}); err != nil {
			return err
		}
		printInfo("Saved image to: " + output)
		return nil
	}

	printInfo(fmt.Sprintf("Rendering %d frames of %q at %dx%d (seed %d)", frames, ops, b.Dx(), b.Dy(), o.seed))
	results, err := pipeline.Animate(anim.Frames, ops, pipeline.AnimateOptions{
		Frames:   frames,
		Seed:     o.seed,
		PingPong: o.pingpong,
		Progress: func(i, n int) {
			if !quiet {
				fmt.Fprintf(os.Stderr, "\r  frame %d/%d", i, n)
				if i == n {
					fmt.Fprintln(os.Stderr)
				}
			}
		},
	})
	if err != nil {
		return err
	}
	printInfo(fmt.Sprintf("Done in %s", time.Since(start).Round(time.Millisecond)))

	if animatedOut {
		for i, fr := range results {
			name := fmt.Sprintf(output, i)
			if err := imageio.Save(name, fr, imageio.Options{Quality: o.quality, Format: imageio.Format(o.format)}); err != nil {
				return err
			}
		}
		printInfo(fmt.Sprintf("Saved %d frames to: %s", len(results), output))
		return nil
	}
	a := &imageio.Animation{}
	for i, fr := range results {
		a.Frames = append(a.Frames, fr)
		delay := 8
		if d := anim.Delays[i%len(anim.Delays)]; d > 0 {
			delay = d
		}
		if o.delay > 0 {
			delay = max(o.delay/10, 1)
		}
		a.Delays = append(a.Delays, delay)
	}
	if err := imageio.SaveAnimation(output, a); err != nil {
		return err
	}
	printInfo("Saved animation to: " + output)
	return nil
}
