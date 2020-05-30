package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

type tbcPackage struct {
	macros     map[string]macro
	components map[string]*component
}

type symbols struct {
	pkgBasePath  string
	packages     map[string]*tbcPackage
	curPkg       string
	curComponent *component
}

func (s *symbols) split(id string) (*tbcPackage, string, string, error) {
	last := strings.LastIndexByte(id, '.')
	if last == -1 {
		return s.packages[s.curPkg], s.curPkg, id, nil
	}
	pkgName := id[0:last]
	if strings.LastIndexByte(pkgName, '.') != -1 {
		return nil, "", "", errors.New("symbol id cannot include multiple '.': " + id)
	}
	pkg, ok := s.packages[pkgName]
	if !ok {
		return nil, "", "", fmt.Errorf("unknown package '%s' in id '%s'", pkgName, id)
	}
	return pkg, pkgName, id[last+1:], nil
}

func (s *symbols) findMacro(id string) (macro, error) {
	pkg, _, name, err := s.split(id)
	if err != nil {
		return macro{}, err
	}
	ret, ok := pkg.macros[name]
	if !ok {
		return macro{}, fmt.Errorf("unknown macro: '%s'", id)
	}
	return ret, nil
}

func (s *symbols) findComponent(id string) (*component, string, string, error) {
	pkg, pkgName, name, err := s.split(id)
	if err != nil {
		return nil, "", "", err
	}
	ret, ok := pkg.components[name]
	if !ok {
		return nil, "", "", fmt.Errorf("unknown component: '%s'", id)
	}
	if pkgName != s.curPkg {
		s.curComponent.dependencies[filepath.Join(s.pkgBasePath, pkgName)] = struct{}{}
	}
	return ret, pkgName, name, nil
}
