package packages

import (
	"errors"
	"strings"

	"github.com/flyx/askew/data"
)

type sorter struct {
	curDepPath []string
	packages   map[string]*data.Package
	done       map[string]struct{}
	result     []string
}

func (s *sorter) walkImports(imports map[string]string) error {
	for _, item := range imports {
		if _, ok := s.done[item]; ok {
			continue
		}
		if _, ok := s.packages[item]; !ok {
			continue
		}
		for i := range s.curDepPath {
			if item == s.curDepPath[i] {
				var b strings.Builder
				b.WriteString("circular dependency:")
				for j := i; j < len(s.curDepPath); j++ {
					b.WriteByte(' ')
					b.WriteString(s.curDepPath[j])
				}
				return errors.New(b.String())
			}
		}
		if err := s.process(item); err != nil {
			return err
		}
	}
	return nil
}

func (s *sorter) process(name string) error {
	if _, ok := s.done[name]; ok {
		return nil
	}

	s.curDepPath = append(s.curDepPath, name)
	defer func() { s.curDepPath = s.curDepPath[:len(s.curDepPath)-1] }()
	pkg := s.packages[name]
	for _, file := range pkg.Files {
		if err := s.walkImports(file.Imports); err != nil {
			return err
		}
	}
	if pkg.Site != nil {
		if err := s.walkImports(pkg.Site.Imports); err != nil {
			return err
		}
	}

	s.result = append(s.result, name)
	s.done[name] = struct{}{}
	return nil
}

// Sort sorts the given list of packages by order of dependencies between them.
// raises an error in case of cyclic dependencies
func Sort(packages map[string]*data.Package) ([]string, error) {
	s := sorter{result: make([]string, 0, len(packages)),
		done: make(map[string]struct{}), packages: packages,
		curDepPath: make([]string, 0, len(packages))}

	for name := range packages {
		if err := s.process(name); err != nil {
			return nil, err
		}
	}
	return s.result, nil
}
