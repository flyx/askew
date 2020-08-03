package output

import (
	"strings"

	"github.com/flyx/askew/data"
)

type skeletonWriter struct {
	Syms        *data.Symbols
	PackagePath string
}

// WriteSkeleton writes the the go file that initializes embeds in the skeletons.
func WriteSkeleton(syms *data.Symbols, path string, pkgName string, s *data.Skeleton) {
	b := strings.Builder{}
	if err := fileHeader.Execute(&b, struct {
		PackageName string
		Imports     map[string]string
	}{pkgName, s.Imports}); err != nil {
		panic(err)
	}

	if err := skeleton.Execute(&b, s); err != nil {
		panic(err)
	}

	writeFormatted(b.String(), path)
}
