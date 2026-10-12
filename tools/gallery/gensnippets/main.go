package main

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"strings"

	"github.com/MaRcR11/sortmancy/tools/gallery/catalog"
)

const output = "tools/gallery/snippets/snippets_gen.go"

func body(e catalog.Example) string {
	code := strings.ReplaceAll(e.Code, "out, err :=", "out, err =")
	code = strings.ReplaceAll(code, "out :=", "out =")
	code = strings.ReplaceAll(code, `"out.gif"`, "outGIF")
	return code
}

func main() {
	var b bytes.Buffer
	b.WriteString(`package snippets

import (
	"image"

	"github.com/MaRcR11/sortmancy/effects"
	"github.com/MaRcR11/sortmancy/imageio"
	"github.com/MaRcR11/sortmancy/pipeline"
	"github.com/MaRcR11/sortmancy/sorter"
)

var (
	_ = effects.Invert
	_ = imageio.Load
	_ = pipeline.Parse
	_ = sorter.Sort
)

type Func func(img image.Image, outGIF string) (out image.Image, err error)

var Funcs = map[string]Func{
`)
	for _, e := range catalog.All() {
		fmt.Fprintf(&b, "\t%q: func(img image.Image, outGIF string) (out image.Image, err error) {\n%s\n\t\treturn\n\t},\n", e.Name, body(e))
	}
	b.WriteString("}\n")
	src, err := format.Source(b.Bytes())
	if err != nil {
		log.Fatalf("format generated code: %v\n%s", err, b.String())
	}
	if err := os.MkdirAll("tools/gallery/snippets", 0o755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(output, src, 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Println("wrote", output, len(catalog.All()), "snippets")
}
