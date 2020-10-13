package units

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/flyx/askew/attributes"
	"github.com/flyx/askew/data"
	"github.com/flyx/askew/walker"
)

// ProcessFile processes a file containing units (*.askew)
func ProcessFile(file *data.AskewFile, syms *data.Symbols, counter *int) error {
	syms.SetAskewFile(file)
	os.Stdout.WriteString("[info] processing units: " + file.Path + "\n")
	w := walker.Walker{TextNode: walker.WhitespaceOnly{},
		Component: &componentProcessor{unitProcessor{syms}, counter},
	}
	_, _, err := w.WalkChildren(nil, &walker.NodeSlice{Items: file.Content})
	if err != nil {
		return errors.New(file.Path + ": " + err.Error())
	}
	return err
}

func processSiteDescriptor(site *data.ASiteFile) error {
	var siteAttrs attributes.Site
	err := attributes.Collect(site.Descriptor, &siteAttrs)
	if err != nil {
		return err
	}
	if siteAttrs.HTMLFile == "" {
		site.HTMLFile = "index.html"
	} else {
		site.HTMLFile = siteAttrs.HTMLFile
	}
	if siteAttrs.JSFile == "" {
		site.JSFile = filepath.Base(site.BaseName) + ".js"
	} else {
		site.JSFile = siteAttrs.JSFile
	}
	return nil
}

// ProcessSite processes a file containing a site skeleton (*.asite)
func ProcessSite(file *data.ASiteFile, syms *data.Symbols) error {
	syms.SetASiteFile(file)
	os.Stdout.WriteString("[info] processing site: " + file.Path + "\n")
	if err := processSiteDescriptor(file); err != nil {
		return err
	}

	p := unitProcessor{syms}

	return p.processUnitContent(file.RootNode(), &file.Unit, nil, file.RootNode(), false)
}
