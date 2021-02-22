package output

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/flyx/askew/data"
	"github.com/flyx/net/html"
	"github.com/flyx/net/html/atom"
)

// Backend defines for which kind of backend the HTML should be set up.
type Backend int

const (
	// GopherJSBackend assumes the Go code will be compiled with GopherJS
	GopherJSBackend Backend = iota
	// WasmBackend assumes the Go code will be compiled with Go's WASM backend.
	WasmBackend
)

// PackageWriter writes the Go code for a package into files.
type PackageWriter struct {
	Syms        *data.Symbols
	PackageName string
	RelPath     string
}

// WriteFile writes a file of the package.
func (pw *PackageWriter) WriteFile(f *data.AskewFile) error {
	b := strings.Builder{}
	if err := fileHeader.Execute(&b, struct {
		PackageName string
		Imports     map[string]string
	}{pw.PackageName, f.Imports}); err != nil {
		return err
	}

	if err := component.Execute(&b, f); err != nil {
		return err
	}
	if err := list.Execute(&b, f); err != nil {
		return err
	}
	if err := optional.Execute(&b, f); err != nil {
		return err
	}

	writeFormatted(b.String(), filepath.Join(pw.RelPath, f.BaseName+".askew.go"))
	return nil
}

// WriteSite writes a file init.go in the site's package, and the HTML file
// of the site.
func (pw *PackageWriter) WriteSite(f *data.ASiteFile, outputPath string,
	backend Backend) error {
	// init.go file
	b := strings.Builder{}
	if err := fileHeader.Execute(&b, struct {
		PackageName string
		Imports     map[string]string
	}{pw.PackageName, f.Imports}); err != nil {
		return err
	}

	if err := site.Execute(&b, f); err != nil {
		return err
	}

	writeFormatted(b.String(), filepath.Join(pw.RelPath, f.BaseName+".asite.go"))

	// HTML file
	node := f.RootNode()
	for node = node.FirstChild; node != nil; node = node.NextSibling {
		if node.Type == html.ElementNode && node.DataAtom == atom.Body {
			break
		}
	}
	if node == nil {
		return errors.New("site misses <body> node")
	}

	switch backend {
	case GopherJSBackend:
		node.NextSibling = &html.Node{
			Type:        html.ElementNode,
			Data:        "script",
			DataAtom:    atom.Script,
			Attr:        []html.Attribute{{Key: "src", Val: f.JSPath}, {Key: "charset", Val: "UTF-8"}},
			PrevSibling: node,
			Parent:      node.Parent,
			NextSibling: nil,
		}
	case WasmBackend:
		node.NextSibling = &html.Node{
			Type:        html.ElementNode,
			Data:        "script",
			DataAtom:    atom.Script,
			Attr:        []html.Attribute{{Key: "src", Val: f.WASMExecPath}, {Key: "charset", Val: "UTF-8"}},
			PrevSibling: node,
			Parent:      node.Parent,
		}

		var b strings.Builder
		wasmInit.Execute(&b, f.WASMPath)

		wasmExec := node.NextSibling
		wasmExec.NextSibling = &html.Node{
			Type:        html.ElementNode,
			Data:        "script",
			DataAtom:    atom.Script,
			PrevSibling: wasmExec,
			Parent:      node.Parent,
		}
		wasm := wasmExec.NextSibling
		wasm.FirstChild = &html.Node{
			Type:   html.TextNode,
			Data:   b.String(),
			Parent: wasm,
		}
		wasm.LastChild = wasm.FirstChild
	}

	htmlFile, err := os.Create(filepath.Join(outputPath, f.HTMLFile))
	if err != nil {
		return err
	}
	html.Render(htmlFile, f.Document)
	htmlFile.Close()

	return nil
}
