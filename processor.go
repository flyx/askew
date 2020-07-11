package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/flyx/askew/components"
	"github.com/flyx/askew/data"
	"github.com/flyx/askew/output"
	"github.com/flyx/askew/walker"

	"golang.org/x/mod/modfile"
	"golang.org/x/net/html"
)

type processor struct {
	syms    data.Symbols
	counter int
	mod     *modfile.File
}

func (p *processor) init(base *data.BaseDir) {
	p.syms.Packages = base.Packages
}

func (p *processor) processMacros(pkgName string) error {
	p.syms.CurPkg = pkgName
	pkg := p.syms.Packages[pkgName]
	for _, file := range pkg.Files {
		var err error

		p.syms.CurFile = file
		os.Stdout.WriteString("[info] processing macros: " + file.Path + "\n")
		var dummyParent *html.Node
		if dummyParent, err = processMacros(file.Content, &p.syms); err != nil {
			return errors.New(file.Path + err.Error())
		}

		// we need to write out the nodes and parse it again since text nodes may
		// be merged and additional elements may be created now with includes
		// processed. If we don't do this, paths to access the dynamic objects will
		// be wrong.
		b := strings.Builder{}
		for cur := dummyParent.FirstChild; cur != nil; cur = cur.NextSibling {
			html.Render(&b, cur)
		}
		file.Content, err = html.ParseFragment(strings.NewReader(b.String()), &data.BodyEnv)
		if err != nil {
			return errors.New(file.Path + ": " + err.Error())
		}
	}
	return nil
}

func (p *processor) processComponents(pkgName string) error {
	p.syms.CurPkg = pkgName
	pkg := p.syms.Packages[pkgName]
	for _, file := range pkg.Files {
		p.syms.CurFile = file
		os.Stdout.WriteString("[info] processing components: " + file.Path + "\n")
		w := walker.Walker{Text: walker.WhitespaceOnly{}, Component: components.NewProcessor(&p.syms, &p.counter)}
		_, _, err := w.WalkChildren(nil, &walker.NodeSlice{Items: file.Content})
		if err != nil {
			return errors.New(file.Path + ": " + err.Error())
		}
	}
	return nil
}

func (p *processor) dump(skeleton *data.Skeleton, outputPath, initPath string) {
	htmlFile, err := os.Create(filepath.Join(outputPath, "index.html"))
	if err != nil {
		panic("unable to write HTML output: " + err.Error())
	}
	if skeleton == nil {
		for _, pkg := range p.syms.Packages {
			for _, f := range pkg.Files {
				for _, c := range f.Components {
					html.Render(htmlFile, c.Template)
				}
			}
		}
	} else {
		html.Render(htmlFile, skeleton.Root)
	}

	htmlFile.Close()

	for pkgName, pkg := range p.syms.Packages {
		w := output.PackageWriter{Syms: &p.syms, PackageName: filepath.Base(pkgName),
			PackagePath: pkg.Path}
		if err := os.MkdirAll(w.PackagePath, os.ModePerm); err != nil {
			panic("failed to create package directory '" + w.PackagePath +
				"': " + err.Error())
		}
		for _, f := range pkg.Files {
			w.WriteFile(f)
		}
	}

	if skeleton != nil {
		output.WriteSkeleton(&p.syms, initPath, skeleton)
	}
}
