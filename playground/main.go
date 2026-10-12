//go:build js && wasm

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"syscall/js"
	"time"

	"github.com/MaRcR11/sortmancy/effects"
	"github.com/MaRcR11/sortmancy/imageio"
	"github.com/MaRcR11/sortmancy/pipeline"
)

var fileParams = map[string]bool{"mask": true, "interval-image": true, "map": true, "with": true}

type paramInfo struct {
	Name    string   `json:"name"`
	Kind    string   `json:"kind"`
	Default string   `json:"default"`
	Help    string   `json:"help"`
	Choices []string `json:"choices,omitempty"`
}

type opInfo struct {
	Name   string      `json:"name"`
	Help   string      `json:"help"`
	Params []paramInfo `json:"params"`
}

type catalogInfo struct {
	Ops     []opInfo          `json:"ops"`
	Common  []paramInfo       `json:"common"`
	Presets map[string]string `json:"presets"`
}

func toInfo(params []pipeline.Param) []paramInfo {
	hasFile := false
	for _, p := range params {
		hasFile = hasFile || p.Name == "mask"
	}
	out := []paramInfo{}
	for _, p := range params {
		if fileParams[p.Name] || (hasFile && p.Name == "invert-mask") {
			continue
		}
		out = append(out, paramInfo{p.Name, p.Kind.String(), p.Default, p.Help, p.Choices})
	}
	return out
}

func info(js.Value, []js.Value) any {
	c := catalogInfo{Common: toInfo(pipeline.CommonParams()), Presets: pipeline.Presets}
	for _, s := range pipeline.Specs() {
		c.Ops = append(c.Ops, opInfo{s.Name, s.Help, toInfo(s.Params)})
	}
	data, _ := json.Marshal(c)
	return string(data)
}

var source image.Image

func setImage(_ js.Value, args []js.Value) any {
	data := make([]byte, args[0].Get("length").Int())
	js.CopyBytesToGo(data, args[0])
	maxSide := args[1].Int()
	anim, err := imageio.DecodeAnimation(bytes.NewReader(data))
	if err != nil {
		return map[string]any{"error": "could not read the image: " + err.Error()}
	}
	img := anim.Frames[0]
	b := img.Bounds()
	if s := max(b.Dx(), b.Dy()); maxSide > 0 && s > maxSide {
		img = effects.Resize(img, b.Dx()*maxSide/s, b.Dy()*maxSide/s, effects.Bilinear)
	}
	source = img
	nb := img.Bounds()
	return map[string]any{"width": nb.Dx(), "height": nb.Dy(), "originalWidth": b.Dx(), "originalHeight": b.Dy()}
}

type request struct {
	Op       string   `json:"op"`
	Args     []string `json:"args"`
	Pipeline string   `json:"pipeline"`
	Seed     int64    `json:"seed"`
}

func run(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if r := recover(); r != nil {
			result = map[string]any{"error": fmt.Sprint(r)}
		}
	}()
	if source == nil {
		return map[string]any{"error": "no image loaded"}
	}
	var req request
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return map[string]any{"error": err.Error()}
	}
	var p pipeline.Pipeline
	if req.Pipeline != "" {
		var err error
		if p, err = pipeline.ParseSeed(req.Pipeline, req.Seed); err != nil {
			return map[string]any{"error": err.Error()}
		}
	} else {
		op, err := pipeline.New(req.Op, req.Args, req.Seed)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		p = pipeline.Pipeline{op}
	}
	start := time.Now()
	out, err := p.Apply(source)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	elapsed := time.Since(start)
	arr := js.Global().Get("Uint8ClampedArray").New(len(out.Pix))
	js.CopyBytesToJS(arr, out.Pix)
	return map[string]any{
		"pixels": arr,
		"width":  out.Bounds().Dx(),
		"height": out.Bounds().Dy(),
		"ms":     elapsed.Milliseconds(),
	}
}

func main() {
	js.Global().Set("sortmancy", js.ValueOf(map[string]any{
		"info":     js.FuncOf(info),
		"setImage": js.FuncOf(setImage),
		"run":      js.FuncOf(run),
	}))
	if ready := js.Global().Get("sortmancyReady"); ready.Type() == js.TypeFunction {
		ready.Invoke()
	}
	select {}
}
