package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/MaRcR11/sortmancy/effects"
	"github.com/MaRcR11/sortmancy/imageio"
	"github.com/MaRcR11/sortmancy/pipeline"
	"github.com/MaRcR11/sortmancy/tools/gallery/catalog"
	"github.com/MaRcR11/sortmancy/tools/gallery/snippets"
)

const source = "img/city/city.jpg"

func main() {
	site := flag.String("site", "site", "Hugo site directory to write the gallery into; empty skips it")
	readme := flag.String("readme", "", "directory to write the README highlight images into; empty skips it")
	withExamples := flag.Bool("examples", true, "render and check all gallery examples")
	withPlayground := flag.Bool("playground", true, "build the WebAssembly playground into the site")
	flag.Parse()

	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		log.Fatal("run this from the repository root")
	}
	abs := func(p string) string {
		if p == "" || filepath.IsAbs(p) {
			return p
		}
		return filepath.Join(root, p)
	}
	*site, *readme = abs(*site), abs(*readme)

	if *withPlayground && *site != "" {
		if err := buildPlayground(root, filepath.Join(*site, "static", "playground")); err != nil {
			log.Fatalf("playground: %v", err)
		}
		fmt.Println("built the WebAssembly playground")
	}
	if !*withExamples {
		return
	}

	work, err := os.MkdirTemp("", "sortmancy-gallery-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(work)

	bin := filepath.Join(work, "sortmancy")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	build := exec.Command("go", "build", "-o", bin, "./cmd/sortmancy")
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		log.Fatalf("build sortmancy: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, source))
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(work, "city.jpg"), data, 0o644); err != nil {
		log.Fatal(err)
	}
	if err := os.Chdir(work); err != nil {
		log.Fatal(err)
	}
	if err := writeSampleLUT("film.cube"); err != nil {
		log.Fatal(err)
	}

	img, err := imageio.Load("city.jpg")
	if err != nil {
		log.Fatal(err)
	}
	examples := ordered(catalog.All())
	start := time.Now()
	var failed []string
	for i, e := range examples {
		fmt.Printf("[%3d/%d] %s\n", i+1, len(examples), e.Name)
		if err := check(bin, img, e); err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", e.Name, err))
		}
	}
	if len(failed) > 0 {
		log.Fatalf("%d examples failed:\n  %s", len(failed), strings.Join(failed, "\n  "))
	}
	fmt.Printf("all %d examples match between CLI and library (%s)\n", len(examples), time.Since(start).Round(time.Second))

	if *site != "" {
		if err := writeSite(*site, examples); err != nil {
			log.Fatal(err)
		}
		fmt.Println("wrote gallery into", *site)
	}
	if *readme != "" {
		if err := writeReadme(*readme, examples); err != nil {
			log.Fatal(err)
		}
		fmt.Println("wrote README highlights into", *readme)
	}
}

func buildPlayground(root, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	build := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", filepath.Join(dir, "sortmancy.wasm"), "./playground")
	build.Dir = root
	build.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		return err
	}
	out, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return err
	}
	goroot := strings.TrimSpace(string(out))
	loader := filepath.Join(goroot, "lib", "wasm", "wasm_exec.js")
	if _, err := os.Stat(loader); err != nil {
		loader = filepath.Join(goroot, "misc", "wasm", "wasm_exec.js")
	}
	copies := map[string]string{loader: "wasm_exec.js", filepath.Join(root, source): "city.jpg"}
	for from, to := range copies {
		data, err := os.ReadFile(from)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, to), data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func writeSampleLUT(path string) error {
	l, ok := effects.LUTByName("bleach")
	if !ok {
		return fmt.Errorf("built-in LUT bleach is missing")
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := l.WriteCube(f, "film"); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func ordered(all []catalog.Example) []catalog.Example {
	out := append([]catalog.Example(nil), all...)
	sort.SliceStable(out, func(i, j int) bool { return len(out[i].Needs) == 0 && len(out[j].Needs) > 0 })
	return out
}

func check(bin string, img image.Image, e catalog.Example) error {
	ext := ".png"
	if e.Animated {
		ext = ".gif"
	}
	cliOut, libOut := e.Name+".cli"+ext, e.Name+".lib"+ext

	args := splitArgs(e.CLI)
	for i, a := range args {
		if a == "out.jpg" || a == "out.gif" {
			args[i] = cliOut
		}
	}
	cmd := exec.Command(bin, append(args, "--quiet")...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cli: %v: %s", err, out)
	}

	fn, ok := snippets.Funcs[e.Name]
	if !ok {
		return fmt.Errorf("no generated snippet, run go run ./tools/gallery/gensnippets")
	}
	out, err := fn(img, libOut)
	if err != nil {
		return fmt.Errorf("library: %v", err)
	}
	if !e.Animated {
		if err := imageio.Save(libOut, out, imageio.Options{}); err != nil {
			return err
		}
	}
	a, err := os.ReadFile(cliOut)
	if err != nil {
		return err
	}
	b, err := os.ReadFile(libOut)
	if err != nil {
		return err
	}
	if !bytes.Equal(a, b) {
		return fmt.Errorf("CLI and library output differ")
	}
	if !e.Animated {
		return imageio.Save(e.Name+".jpg", out, imageio.Options{Quality: 92})
	}
	return nil
}

func splitArgs(s string) []string {
	var args []string
	var cur strings.Builder
	inQuote, has := false, false
	for _, r := range s {
		switch {
		case r == '"':
			inQuote, has = !inQuote, true
		case r == ' ' && !inQuote:
			if has {
				args = append(args, cur.String())
				cur.Reset()
				has = false
			}
		default:
			cur.WriteRune(r)
			has = true
		}
	}
	if has {
		args = append(args, cur.String())
	}
	return args
}

func saveScaled(dst, src string, maxWidth, quality int) error {
	if strings.HasSuffix(src, ".gif") {
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0o644)
	}
	img, err := imageio.Load(src)
	if err != nil {
		return err
	}
	if b := img.Bounds(); b.Dx() > maxWidth {
		img = effects.Resize(img, maxWidth, b.Dy()*maxWidth/b.Dx(), effects.Bilinear)
	}
	return imageio.Save(dst, img, imageio.Options{Quality: quality})
}

func output(e catalog.Example) string {
	if e.Animated {
		return e.Name + ".cli.gif"
	}
	return e.Name + ".cli.png"
}

func writeReadme(dir string, examples []catalog.Example) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	keep := map[string]bool{}
	for _, e := range examples {
		if !e.Highlight {
			continue
		}
		keep[e.File()] = true
		if err := saveScaled(filepath.Join(dir, e.File()), output(e), 720, 80); err != nil {
			return err
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range entries {
		ext := filepath.Ext(f.Name())
		if !f.IsDir() && (ext == ".jpg" || ext == ".gif") && !keep[f.Name()] {
			if err := os.Remove(filepath.Join(dir, f.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeSite(site string, examples []catalog.Example) error {
	assets := filepath.Join(site, "assets", "gallery")
	pages := filepath.Join(site, "content", "docs", "gallery")
	for _, d := range []string{assets, pages} {
		if err := os.RemoveAll(d); err != nil {
			return err
		}
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	for _, e := range examples {
		if err := saveScaled(filepath.Join(assets, e.File()), output(e), 1280, 85); err != nil {
			return err
		}
	}
	if err := saveScaled(filepath.Join(assets, "original.jpg"), "city.jpg", 1280, 85); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(pages, "_index.md"), []byte(galleryIndex()), 0o644); err != nil {
		return err
	}
	for i, s := range catalog.Sections {
		if err := os.WriteFile(filepath.Join(pages, s.Slug+".md"), []byte(sectionPage(s, i+1)), 0o644); err != nil {
			return err
		}
	}
	return os.WriteFile(filepath.Join(site, "content", "docs", "reference.md"), []byte(referencePage()), 0o644)
}

func cover(s catalog.Section) catalog.Example {
	for _, e := range s.Examples {
		if e.Highlight {
			return e
		}
	}
	return s.Examples[0]
}

func galleryIndex() string {
	var b strings.Builder
	b.WriteString("---\ntitle: Gallery\nweight: 6\n---\n\n")
	fmt.Fprintf(&b, "Every example starts from the same photo and shows the command line call and the equivalent Go code. "+
		"Both are checked to produce exactly the same image every time the site is built.\n\n")
	b.WriteString("![The original photo](/gallery/original.jpg)\n\n")
	b.WriteString("{{< cards cols=\"3\" >}}\n")
	for _, s := range catalog.Sections {
		fmt.Fprintf(&b, "  {{< card link=\"%s/\" title=%q subtitle=%q image=\"gallery/%s\" >}}\n", s.Slug, s.Title, s.Intro, cover(s).File())
	}
	b.WriteString("{{< /cards >}}\n")
	return b.String()
}

func sectionPage(s catalog.Section, weight int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "---\ntitle: %q\nweight: %d\n---\n\n%s\n\n", s.Title, weight, s.Intro)
	b.WriteString("{{< gallery cols=\"3\" >}}\n")
	for _, e := range s.Examples {
		fmt.Fprintf(&b, "  {{< gallery-item src=\"gallery/%s\" caption=%q link=\"#%s\" >}}\n", e.File(), e.Title, e.Name)
	}
	b.WriteString("{{< /gallery >}}\n\n")
	for _, e := range s.Examples {
		fmt.Fprintf(&b, "## %s {#%s}\n\n![%s](/gallery/%s)\n\n", e.Title, e.Name, e.Title, e.File())
		b.WriteString("{{< tabs >}}\n")
		fmt.Fprintf(&b, "{{< tab name=\"CLI\" >}}\n```bash\nsortmancy %s\n```\n{{< /tab >}}\n", e.CLI)
		fmt.Fprintf(&b, "{{< tab name=\"Go\" >}}\n```go\n%s\n```\n{{< /tab >}}\n", e.Code)
		b.WriteString("{{< /tabs >}}\n\n")
	}
	return b.String()
}

func exampleFor(op string) (catalog.Section, catalog.Example, bool) {
	for _, s := range catalog.Sections {
		for _, e := range s.Examples {
			if e.Op() == op {
				return s, e, true
			}
		}
	}
	return catalog.Section{}, catalog.Example{}, false
}

func escape(s string) string {
	return strings.NewReplacer("|", `\|`, "\n", " ").Replace(s)
}

func referencePage() string {
	var b strings.Builder
	b.WriteString("---\ntitle: Operations reference\nweight: 5\n---\n\n")
	fmt.Fprintf(&b, "All %d operations with their parameters. This page is generated from the code, so it always matches the current version. "+
		"Run `sortmancy help <operation>` for the same information in the terminal.\n\n", len(pipeline.Specs()))
	b.WriteString("## Common parameters {#common}\n\n" +
		"Every operation also accepts these, to change only part of the image. " +
		"`sort` has its own `mask` and `invert-mask`, where masked pixels also split the sorted spans; `region` and `feather` work for it too. " +
		"[See the examples in the gallery](../gallery/masks/).\n\n" +
		"| Parameter | Type | Default | Description |\n|-|-|-|-|\n")
	for _, p := range pipeline.CommonParams() {
		def := p.Default
		if def != "" {
			def = "`" + def + "`"
		}
		fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", p.Name, p.Kind, escape(def), escape(p.Help))
	}
	b.WriteString("\n")
	for _, s := range pipeline.Specs() {
		fmt.Fprintf(&b, "## %s {#%s}\n\n%s\n\n", s.Name, s.Name, s.Help)
		if sec, e, ok := exampleFor(s.Name); ok {
			fmt.Fprintf(&b, "[See the example in the gallery](../gallery/%s/#%s)\n\n", sec.Slug, e.Name)
		}
		if len(s.Params) == 0 {
			b.WriteString("This operation has no parameters.\n\n")
			continue
		}
		b.WriteString("| Parameter | Type | Default | Description |\n|-|-|-|-|\n")
		for _, p := range s.Params {
			desc := p.Help
			if len(p.Choices) > 0 {
				desc += " One of: `" + strings.Join(p.Choices, "`, `") + "`."
			}
			def := p.Default
			if def != "" {
				def = "`" + def + "`"
			}
			fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", p.Name, p.Kind, escape(def), escape(desc))
		}
		b.WriteString("\n")
	}
	b.WriteString("## Presets {#presets}\n\n| Preset | Pipeline |\n|-|-|\n")
	for _, name := range pipeline.PresetNames() {
		fmt.Fprintf(&b, "| `@%s` | `%s` |\n", name, escape(pipeline.Presets[name]))
	}
	return b.String()
}
