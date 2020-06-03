package output

import (
	"path/filepath"
	"strings"

	"github.com/flyx/askew/data"
)

// PackageWriter writes the Go code for a component into a file.
type PackageWriter struct {
	Syms        *data.Symbols
	PackageName string
	PackagePath string
}

// WriteComponent writes a component of the package to a file.
func (pw *PackageWriter) WriteComponent(name string, c *data.Component) {
	b := strings.Builder{}
	if err := fileHeader.Execute(&b, struct {
		PackageName string
		Deps        map[string]struct{}
	}{pw.PackageName, c.Dependencies}); err != nil {
		panic(err)
	}

	if err := component.Execute(&b, c); err != nil {
		panic(err)
	}

	writeFormatted(b.String(), filepath.Join(pw.PackagePath, strings.ToLower(name)+".go"))
}
