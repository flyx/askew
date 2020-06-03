package data

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// Package describes as <a:package>.
type Package struct {
	Macros     map[string]Macro
	Components map[string]*Component
}

// EmbedHost is an entity that allows embedding components.
type EmbedHost struct {
	Dependencies map[string]struct{}
	Embeds       []Embed
}

// Symbols is the context of the procesor. It stores all seen symbols along
// with the packages they are declared in.
type Symbols struct {
	PkgBasePath string
	Packages    map[string]*Package
	CurPkg      string
	CurHost     *EmbedHost
}

// split takes an identifier consisting of a symbol name optionally
// prefixed by a package name. It returns the package name, the corresponding
// package, and the remaining symbol name.
func (s *Symbols) split(id string) (pkg *Package, pkgName string, symName string, err error) {
	last := strings.LastIndexByte(id, '.')
	if last == -1 {
		return s.Packages[s.CurPkg], s.CurPkg, id, nil
	}
	pkgName = id[0:last]
	if strings.LastIndexByte(pkgName, '.') != -1 {
		return nil, "", "", errors.New("symbol id cannot include multiple '.': " + id)
	}
	pkg, ok := s.Packages[pkgName]
	if !ok {
		return nil, "", "", fmt.Errorf("unknown package '%s' in id '%s'", pkgName, id)
	}
	return pkg, pkgName, id[last+1:], nil
}

// ResolveMacro resolves the given identifier to a Macro.
func (s *Symbols) ResolveMacro(id string) (Macro, error) {
	pkg, _, name, err := s.split(id)
	if err != nil {
		return Macro{}, err
	}
	ret, ok := pkg.Macros[name]
	if !ok {
		return Macro{}, fmt.Errorf("unknown macro: '%s'", id)
	}
	return ret, nil
}

// ResolveComponent resolves the given identifier to a Component.
func (s *Symbols) ResolveComponent(id string) (*Component, string, string, error) {
	pkg, pkgName, name, err := s.split(id)
	if err != nil {
		return nil, "", "", err
	}
	ret, ok := pkg.Components[name]
	if !ok {
		return nil, "", "", fmt.Errorf("unknown component: '%s'", id)
	}
	if pkgName != s.CurPkg {
		s.CurHost.Dependencies[filepath.Join(s.PkgBasePath, pkgName)] = struct{}{}
	}
	return ret, pkgName, name, nil
}
