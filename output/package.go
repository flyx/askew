package output

import (
	"path/filepath"
	"strings"

	"github.com/flyx/askew/data"
)

// PackageWriter writes the Go code for a file into a file.
type PackageWriter struct {
	Syms        *data.Symbols
	PackageName string
	PackagePath string
}

// WriteFile writes a file of the package.
func (pw *PackageWriter) WriteFile(f *data.File) {
	b := strings.Builder{}
	if err := fileHeader.Execute(&b, struct {
		PackageName string
		Imports     map[string]string
	}{pw.PackageName, f.Imports}); err != nil {
		panic(err)
	}

	if err := file.Execute(&b, f); err != nil {
		panic(err)
	}

	writeFormatted(b.String(), filepath.Join(pw.PackagePath, f.BaseName+".go"))
}
