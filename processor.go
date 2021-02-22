package main

import (
	"errors"
	"os"
	"strings"

	"github.com/flyx/askew/data"
	"github.com/flyx/askew/output"
	"github.com/flyx/askew/units"
	"github.com/flyx/askew/walker"

	"github.com/flyx/net/html"
	"golang.org/x/mod/modfile"
)

type processor struct {
	syms data.Symbols
	mod  *modfile.File
}

func (p *processor) init(base *data.BaseDir) {
	p.syms.BaseDir = *base
}

func (p *processor) processMacros(pkgName string) error {
	p.syms.CurPkg = pkgName
	pkg := p.syms.Packages[pkgName]
	for _, file := range pkg.Files {
		var err error

		p.syms.SetAskewFile(file)
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
		file.Content, err = html.ParseFragmentWithOptions(
			strings.NewReader(b.String()), &data.BodyEnv,
			html.ParseOptionCustomElements(walker.AskewElements))
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
		if err := units.ProcessFile(file, &p.syms); err != nil {
			return err
		}
	}
	if pkg.Site != nil {
		return units.ProcessSite(pkg.Site, &p.syms)
	}
	return nil
}

func (p *processor) dump(outputPath string, backend output.Backend) error {
	for relPath, pkg := range p.syms.Packages {
		w := output.PackageWriter{Syms: &p.syms, PackageName: pkg.Name, RelPath: relPath}
		if err := os.MkdirAll(relPath, 0755); err != nil {
			panic("failed to create package directory '" + relPath +
				"': " + err.Error())
		}
		for _, f := range pkg.Files {
			if err := w.WriteFile(f); err != nil {
				return err
			}
		}
		if pkg.Site != nil {
			if err := w.WriteSite(pkg.Site, outputPath, backend); err != nil {
				return err
			}
		}
	}
	return nil
}
