package data

import (
	"errors"
	"fmt"
	"strings"
)

// EmbedHost is an entity that allows embedding components.
type EmbedHost struct {
	Embeds []Embed
}

// Symbols is the context of the procesor. It stores all seen symbols along
// with the packages they are declared in.
type Symbols struct {
	BaseDir
	CurPkg  string
	CurFile *File
	CurHost *EmbedHost
}

// split takes an identifier consisting of a symbol name optionally
// prefixed by a package name. It returns the package name, the corresponding
// package, and the remaining symbol name.
func (s *Symbols) split(id string) (pkg *Package, symName string, aliasName string, err error) {
	last := strings.LastIndexByte(id, '.')
	if last == -1 {
		if s.CurPkg == "" {
			return nil, "", "", errors.New("identifier in skeleton requires package prefix")
		}
		return s.Packages[s.CurPkg], id, "", nil
	}
	aliasName = id[0:last]
	if strings.LastIndexByte(aliasName, '.') != -1 {
		return nil, "", "", errors.New("symbol id cannot include multiple '.': " + id)
	}
	pkgPath, ok := s.CurFile.Imports[aliasName]
	if !ok {
		return nil, "", "", fmt.Errorf("unknown namespace '%s' in id '%s'", aliasName, id)
	}
	return s.Packages[pkgPath], id[last+1:], aliasName, nil
}

// ResolveMacro resolves the given identifier to a Macro.
func (s *Symbols) ResolveMacro(id string) (Macro, error) {
	pkg, name, _, err := s.split(id)
	if err != nil {
		return Macro{}, err
	}
	for _, file := range pkg.Files {
		ret, ok := file.Macros[name]
		if ok {
			return ret, nil
		}
	}

	return Macro{}, fmt.Errorf("unknown macro: '%s'", id)
}

// ResolveComponent resolves the given identifier to a Component.
func (s *Symbols) ResolveComponent(id string) (*Component, string, string, error) {
	pkg, name, aliasName, err := s.split(id)
	if err != nil {
		return nil, "", "", err
	}
	for _, file := range pkg.Files {
		ret, ok := file.Components[name]
		if ok {
			return ret, name, aliasName, nil
		}
	}

	return nil, "", "", fmt.Errorf("unknown component: '%s'", id)
}
