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

// OutsideModuleErr is an error that is returned when trying to resolve a
// package path that is outside of the current module.
type OutsideModuleErr struct {
	Path string
}

func (e OutsideModuleErr) Error() string {
	return "cannot use controls from import path " + e.Path + " which is outside current module"
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
			panic("unexpected empty CurPkg")
		}
		return s.Packages[s.CurPkg], id, "", nil
	}
	aliasName = id[0:last]
	symName = id[last+1:]
	if strings.LastIndexByte(aliasName, '.') != -1 {
		return nil, "", "", errors.New("symbol id cannot include multiple '.': " + id)
	}
	pkgPath, ok := s.CurFile().Imports[aliasName]
	if !ok {
		err = fmt.Errorf("unknown namespace '%s' in id '%s'", aliasName, id)
		return
	}
	relPath, err := filepath.Rel(s.ImportPath, pkgPath)
	if err != nil {
		err = OutsideModuleErr{pkgPath}
		return
	}
	retPkg, ok := s.Packages[relPath]
	if !ok {
		err = fmt.Errorf("cannot use controls from import path '%s' which is unknown (has it been excluded?)", pkgPath)
		return
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
func (s *Symbols) ResolveComponent(id string) (
	c *Component, symName string, aliasName string, err error) {
	pkg, symName, aliasName, err := s.split(id)
	if err != nil {
		return nil, symName, aliasName, err
	}
	for _, file := range pkg.Files {
		ret, ok := file.Components[symName]
		if ok {
			return ret, symName, aliasName, nil
		}
	}

	return nil, symName, aliasName, fmt.Errorf("unknown component: '%s'", id)
}
