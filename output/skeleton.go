package output

import (
	"strings"

	"github.com/flyx/tbc/data"
)

type skeletonWriter struct {
	Syms        *data.Symbols
	PackagePath string
}

// WriteSkeleton writes the the go file that initializes embeds in the skeletons.
func WriteSkeleton(syms *data.Symbols, path string, s *data.Skeleton) {
	b := strings.Builder{}
	if err := fileHeader.Execute(&b, struct {
		PackageName string
		Deps        map[string]struct{}
	}{"main", s.Dependencies}); err != nil {
		panic(err)
	}

	if err := skeleton.Execute(&b, s); err != nil {
		panic(err)
	}

	writeFormatted(b.String(), path)
}
