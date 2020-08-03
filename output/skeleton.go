package output

import (
	"path/filepath"
	"strings"

	"github.com/flyx/askew/data"
)

type skeletonWriter struct {
	Syms        *data.Symbols
	PackagePath string
}

// WriteSkeleton writes the the go file that initializes embeds in the skeletons.
func WriteSkeleton(syms *data.Symbols, path string, pkgName, skeletonVarName string, s *data.Skeleton) {
	b := strings.Builder{}
	if err := fileHeader.Execute(&b, struct {
		PackageName string
		Imports     map[string]string
	}{pkgName, s.Imports}); err != nil {
		panic(err)
	}

	if err := skeleton.Execute(&b, struct {
		*data.Skeleton
		VarName string
	}{Skeleton: s, VarName: skeletonVarName}); err != nil {
		panic(err)
	}

	writeFormatted(b.String(), filepath.Join(path, "init.go"))
}
