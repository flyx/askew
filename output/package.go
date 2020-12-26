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

// PackageWriter writes the Go code for a package into files.
type PackageWriter struct {
	Syms        *data.Symbols
	PackageName string
	PackagePath string
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

	if err := file.Execute(&b, f); err != nil {
		return err
	}

	writeFormatted(b.String(), filepath.Join(pw.PackagePath, f.BaseName+".go"))
	return nil
}

// WriteSite writes a file init.go in the site's package, and the HTML file
// of the site.
func (pw *PackageWriter) WriteSite(f *data.ASiteFile, outputPath string) error {
	// init.go file
	b := strings.Builder{}
	if err := fileHeader.Execute(&b, struct {
		PackageName string
		Imports     map[string]string
	}{"main", f.Imports}); err != nil {
		return err
	}

	if err := skeleton.Execute(&b, f); err != nil {
		return err
	}

	writeFormatted(b.String(), filepath.Join(pw.PackagePath, "init.go"))

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
	for _, pkg := range pw.Syms.Packages {
		for _, file := range pkg.Files {
			for _, cmp := range file.Components {
				node.NextSibling = cmp.Template
				cmp.Template.PrevSibling = node
				cmp.Template.Parent = node.Parent
				node = cmp.Template
			}
		}
	}
	node.NextSibling = &html.Node{
		Type:        html.ElementNode,
		Data:        "script",
		DataAtom:    atom.Script,
		Attr:        []html.Attribute{{Key: "src", Val: f.JSFile}},
		PrevSibling: node,
		Parent:      node.Parent,
		NextSibling: nil,
	}

	htmlFile, err := os.Create(filepath.Join(outputPath, "index.html"))
	if err != nil {
		return err
	}
	html.Render(htmlFile, f.Document)
	htmlFile.Close()

	return nil
}
