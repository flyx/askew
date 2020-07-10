package data

import (
	"golang.org/x/net/html"
)

// File describes an .askew file.
type File struct {
	// map of alias -> relative package path.
	// package path is relative to module.
	// alias is by default the part of the package path after the last '.'
	Imports    map[string]string
	Macros     map[string]Macro
	Components map[string]*Component
	Content    []*html.Node
	BaseName   string
	Path       string
}

// Package describes a Go package.
type Package struct {
	Files []*File
	Path  string
}

// BaseDir describes the directory on which askew is executed
type BaseDir struct {
	Packages   map[string]*Package
	ImportPath string
}
