package data

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// Unit is a top-level named unit (either the main site or <a:component>)
// that may have embedded components.
type Unit struct {
	Block

	Variables []VariableMapping
	Embeds    []Embed
}

// Symbols is the context of the procesor. It stores all seen symbols along
// with the packages they are declared in.
type Symbols struct {
	BaseDir
	CurPkg       string
	curAskewFile *AskewFile
	curAsiteFile *ASiteFile
	CurUnit      *Unit
}

// SetAskewFile sets the currently processed file to be the given .askew file.
func (s *Symbols) SetAskewFile(f *AskewFile) {
	s.curAskewFile = f
	s.curAsiteFile = nil
}

// SetASiteFile sets the currently processed file to be the given .asite file.
func (s *Symbols) SetASiteFile(f *ASiteFile) {
	s.curAskewFile = nil
	s.curAsiteFile = f
}

// CurFile returns the currently processed file.
func (s *Symbols) CurFile() *File {
	if s.curAskewFile == nil {
		return &s.curAsiteFile.File
	}
	return &s.curAskewFile.File
}

// CurAskewFile returns the currently processed .askew file.
func (s *Symbols) CurAskewFile() *AskewFile {
	return s.curAskewFile
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
	pkgPath, ok := s.CurFile().Imports[aliasName]
	if !ok {
		return nil, "", "", fmt.Errorf("unknown namespace '%s' in id '%s'", aliasName, id)
	}
	relPath, err := filepath.Rel(s.ImportPath, pkgPath)
	if err != nil {
		return nil, "", "", fmt.Errorf("cannot use controls from import path '%s' which is outside current module", pkgPath)
	}
	retPkg, ok := s.Packages[relPath]
	if !ok {
		return nil, "", "", fmt.Errorf("cannot use controls from import path '%s' which is unknown (has it been excluded?)", pkgPath)
	}
	return retPkg, id[last+1:], aliasName, nil
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
