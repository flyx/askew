package packages

import (
	"errors"
	"strings"

	"github.com/flyx/askew/data"
)

type sorter struct {
	packages map[string]*data.Package
	done     map[string]struct{}
	result   []string
}

func (s *sorter) process(depPath *[]string) error {
	name := (*depPath)[len(*depPath)-1]
	pkg := s.packages[name]
	for _, file := range pkg.Files {
		for _, item := range file.Imports {
			if _, ok := s.done[item]; ok {
				continue
			}
			if _, ok := s.packages[item]; !ok {
				continue
			}
			for i := range *depPath {
				if item == (*depPath)[i] {
					var b strings.Builder
					b.WriteString("circular dependency:")
					for j := i; j < len(*depPath); j++ {
						b.WriteByte(' ')
						b.WriteString((*depPath)[j])
					}
					return errors.New(b.String())
				}
			}
			*depPath = append(*depPath, item)
			if err := s.process(depPath); err != nil {
				return err
			}
			*depPath = (*depPath)[:len(*depPath)-1]
		}
		s.result = append(s.result, name)
		s.done[name] = struct{}{}
	}
	return nil
}

// Sort sorts the given list of packages by order of dependencies between them.
// raises an error in case of cyclic dependencies
func Sort(packages map[string]*data.Package) ([]string, error) {
	s := sorter{result: make([]string, 0, len(packages)),
		done: make(map[string]struct{}), packages: packages}

	path := make([]string, 0, len(packages))
	for name := range packages {
		if _, ok := s.done[name]; ok {
			continue
		}
		path = append(path, name)
		if err := s.process(&path); err != nil {
			return nil, err
		}
		path = path[:len(path)-1]
	}
	return s.result, nil
}
