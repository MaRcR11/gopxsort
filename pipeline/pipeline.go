package pipeline

import (
	"fmt"
	"image"
	"sort"
	"strconv"
	"strings"

	"github.com/MaRcR11/sortcery/internal/imgutil"
)

type ApplyFunc func(img *image.NRGBA) (*image.NRGBA, error)

type Op interface {
	Name() string
	Apply(img *image.NRGBA) (*image.NRGBA, error)
}

type SourceFunc func(img, src *image.NRGBA) (*image.NRGBA, error)

type SourceOp interface {
	Op
	ApplySource(img, src *image.NRGBA) (*image.NRGBA, error)
}

func Func(name string, fn ApplyFunc) Op { return &op{name: name, desc: name, fn: fn} }

func SourceFuncOp(name string, fn SourceFunc) SourceOp {
	return &op{name: name, desc: name, srcFn: fn}
}

type op struct {
	name  string
	desc  string
	fn    ApplyFunc
	srcFn SourceFunc
}

func (o *op) Name() string   { return o.name }
func (o *op) String() string { return o.desc }

func (o *op) Apply(img *image.NRGBA) (*image.NRGBA, error) {
	if o.srcFn != nil {
		return o.srcFn(img, img)
	}
	return o.fn(img)
}

func (o *op) ApplySource(img, src *image.NRGBA) (*image.NRGBA, error) {
	if o.srcFn != nil {
		return o.srcFn(img, src)
	}
	return o.fn(img)
}

type Kind int

const (
	KindString Kind = iota
	KindInt
	KindFloat
	KindBool
)

func (k Kind) String() string {
	return [...]string{"string", "int", "float", "bool"}[k]
}

type Param struct {
	Name    string
	Kind    Kind
	Default string
	Help    string
	Choices []string
}

type Spec struct {
	Name        string
	Help        string
	Params      []Param
	Build       func(a *Args) (ApplyFunc, error)
	BuildSource func(a *Args) (SourceFunc, error)
}

func (s *Spec) Param(name string) (Param, bool) {
	for _, p := range s.Params {
		if p.Name == name {
			return p, true
		}
	}
	return Param{}, false
}

var registry = map[string]*Spec{}

func Register(s Spec) {
	if _, dup := registry[s.Name]; dup {
		panic("pipeline: duplicate operation " + s.Name)
	}
	registry[s.Name] = &s
}

func Lookup(name string) (*Spec, bool) {
	s, ok := registry[strings.ToLower(name)]
	return s, ok
}

func Specs() []*Spec {
	out := make([]*Spec, 0, len(registry))
	for _, s := range registry {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

type Args struct {
	spec   *Spec
	values map[string]string
	seed   int64
	err    error
}

func (a *Args) raw(name string) string {
	if v, ok := a.values[name]; ok {
		return v
	}
	p, ok := a.spec.Param(name)
	if !ok {
		panic(fmt.Sprintf("pipeline: %s reads undeclared parameter %q", a.spec.Name, name))
	}
	return p.Default
}

func (a *Args) Has(name string) bool {
	_, ok := a.values[name]
	return ok
}

func (a *Args) String(name string) string { return a.raw(name) }

func (a *Args) Int(name string) int {
	v, err := strconv.Atoi(a.raw(name))
	if err != nil {
		a.Fail(fmt.Errorf("%s: %s must be an integer, got %q", a.spec.Name, name, a.raw(name)))
	}
	return v
}

func (a *Args) Float(name string) float64 {
	v, err := strconv.ParseFloat(a.raw(name), 64)
	if err != nil {
		a.Fail(fmt.Errorf("%s: %s must be a number, got %q", a.spec.Name, name, a.raw(name)))
	}
	return v
}

func (a *Args) Bool(name string) bool {
	v, err := strconv.ParseBool(a.raw(name))
	if err != nil {
		a.Fail(fmt.Errorf("%s: %s must be true or false, got %q", a.spec.Name, name, a.raw(name)))
	}
	return v
}

func (a *Args) Seed() int64 {
	if s, ok := a.values["seed"]; ok {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			a.Fail(fmt.Errorf("%s: seed must be an integer, got %q", a.spec.Name, s))
		}
		return v
	}
	return a.seed
}

func (a *Args) Fail(err error) {
	if a.err == nil {
		a.err = err
	}
}

func New(name string, args []string, seed int64) (Op, error) {
	spec, ok := Lookup(name)
	if !ok {
		return nil, fmt.Errorf("unknown operation %q (see the list of operations)", name)
	}
	values := map[string]string{}
	pos := 0
	for _, arg := range args {
		k, v, hasEq := strings.Cut(arg, "=")
		k = strings.ToLower(strings.TrimLeft(k, "-"))
		if !hasEq {
			if p, ok := spec.Param(k); ok && p.Kind == KindBool {
				values[k] = "true"
				continue
			}
			if pos >= len(spec.Params) {
				return nil, fmt.Errorf("%s: too many positional values at %q", spec.Name, arg)
			}
			values[spec.Params[pos].Name] = arg
			pos++
			continue
		}
		if _, ok := spec.Param(k); !ok {
			return nil, fmt.Errorf("%s: unknown parameter %q", spec.Name, k)
		}
		values[k] = v
	}
	a := &Args{spec: spec, values: values, seed: seed}
	o := &op{name: spec.Name, desc: describe(spec.Name, args)}
	var err error
	if spec.BuildSource != nil {
		o.srcFn, err = spec.BuildSource(a)
	} else {
		o.fn, err = spec.Build(a)
	}
	if err == nil {
		err = a.err
	}
	if err != nil {
		return nil, err
	}
	return o, nil
}

func describe(name string, args []string) string {
	if len(args) == 0 {
		return name
	}
	return name + " " + strings.Join(args, " ")
}

type Pipeline []Op

func Parse(s string) (Pipeline, error) { return ParseSeed(s, 0) }

func ParseSeed(s string, seed int64) (Pipeline, error) {
	s, err := expandPresets(s, 0)
	if err != nil {
		return nil, err
	}
	var p Pipeline
	for i, stage := range strings.Split(s, "|") {
		fields := strings.Fields(stage)
		if len(fields) == 0 {
			return nil, fmt.Errorf("empty stage %d in pipeline %q", i+1, s)
		}
		o, err := New(fields[0], fields[1:], seed+int64(i))
		if err != nil {
			return nil, err
		}
		p = append(p, o)
	}
	return p, nil
}

func (p Pipeline) Apply(img image.Image) (*image.NRGBA, error) {
	src := imgutil.ToNRGBA(img)
	cur := imgutil.ToNRGBA(src)
	for _, o := range p {
		var next *image.NRGBA
		var err error
		if so, ok := o.(SourceOp); ok {
			next, err = so.ApplySource(cur, src)
		} else {
			next, err = o.Apply(cur)
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", o.Name(), err)
		}
		cur = next
	}
	return cur, nil
}

func (p Pipeline) String() string {
	parts := make([]string, len(p))
	for i, o := range p {
		if s, ok := o.(fmt.Stringer); ok {
			parts[i] = s.String()
		} else {
			parts[i] = o.Name()
		}
	}
	return strings.Join(parts, " | ")
}
