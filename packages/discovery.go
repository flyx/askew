package packages

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/flyx/askew/parsers"

	"github.com/flyx/askew/data"
	"github.com/flyx/askew/walker"
	"github.com/flyx/net/html"
	"github.com/flyx/net/html/atom"
	"golang.org/x/mod/modfile"
)

func findBasePath() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", errors.New("while searching for go.mod: " + err.Error())
	}
	vName := filepath.VolumeName(path)
	traversed := ""
	for {
		goModPath := filepath.Join(path, "go.mod")
		info, err := os.Stat(goModPath)
		if err == nil && !info.IsDir() {
			raw, err := ioutil.ReadFile(goModPath)
			if err != nil {
				return "", fmt.Errorf("%s: %s", goModPath, err.Error())
			}
			goMod, err := modfile.Parse("go.mod", raw, nil)
			if err != nil {
				return "", fmt.Errorf("%s: %s", goModPath, err.Error())
			}
			return filepath.ToSlash(filepath.Join(goMod.Module.Mod.Path, traversed)), nil
		}
		dir, last := filepath.Split(path)
		if dir == "" {
			return "", errors.New("did not find a Go module (go.mod)")
		}
		// remove separator
		path = dir[:len(dir)-1]
		if path == vName {
			return "", errors.New("did not find a Go module (go.mod)")
		}
		traversed = filepath.Join(last, traversed)
	}
}

type suffix int

const (
	dotAskew suffix = iota
	dotAsite
	dotOther
)

func fileKind(name string) suffix {
	lower := strings.ToLower(name)

	if strings.HasSuffix(lower, ".askew") {
		return dotAskew
	}
	if strings.HasSuffix(lower, ".asite") {
		return dotAsite
	}
	return dotOther
}

func descend(start *html.Node, path []atom.Atom) (*html.Node, error) {
	cur := start
	for _, a := range path {
		found := false
		for c := cur.FirstChild; c != nil; c = c.NextSibling {
			if c.DataAtom == a {
				cur = c
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("missing HTML element: " + a.String())
		}
	}
	return cur, nil
}

// Discover searches for a go.mod in the cwd, then walks through the file system
// to discover .askew files.
// For each file, the imports are parsed.
func Discover() (*data.BaseDir, error) {
	var err error
	ret := &data.BaseDir{}
	ret.ImportPath, err = findBasePath()
	if err != nil {
		return nil, err
	}

	ret.Packages = make(map[string]*data.Package)
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		kind := fileKind(info.Name())
		if info.IsDir() || kind == dotOther {
			return nil
		}
		os.Stdout.WriteString("[info] discovered: " + path + "\n")
		relPath := filepath.Dir(path)
		pkgPath := filepath.ToSlash(filepath.Join(ret.ImportPath, relPath))
		pkg, ok := ret.Packages[pkgPath]
		if !ok {
			pkg = &data.Package{Files: make([]*data.AskewFile, 0, 32), Path: relPath}
			ret.Packages[pkgPath] = pkg
		}

		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		baseName := info.Name()[:len(info.Name())-6]
		if kind == dotAskew {
			askewFile := &data.AskewFile{File: data.File{BaseName: baseName, Path: path}}
			askewFile.Content, err = html.ParseFragmentWithOptions(
				bytes.NewReader(contents), &data.BodyEnv,
				html.ParseOptionCustomElements(walker.AskewElements))
			if err != nil {
				return fmt.Errorf("%s: %s", path, err.Error())
			}
			w := walker.Walker{
				Import:    &importHandler{file: &askewFile.File},
				Component: walker.DontDescend{},
				Macro:     walker.DontDescend{},
				TextNode:  walker.WhitespaceOnly{}}
			_, _, err = w.WalkChildren(nil, &walker.NodeSlice{Items: askewFile.Content})
			if err != nil {
				return fmt.Errorf("%s: %s", path, err.Error())
			}
			pkg.Files = append(pkg.Files, askewFile)
		} else {
			if pkg.Site != nil {
				return fmt.Errorf("%s: a package cannot contain multiple sites", path)
			}

			asiteFile := &data.ASiteFile{File: data.File{BaseName: baseName, Path: path}}
			asiteFile.Document, err = html.ParseWithOptions(bytes.NewReader(contents),
				html.ParseOptionCustomElements(walker.AskewElements))
			if err != nil {
				return fmt.Errorf("%s: %s", path, err.Error())
			}
			if asiteFile.Document.Type != html.DocumentNode ||
				asiteFile.Document.FirstChild.Type != html.DoctypeNode {
				return fmt.Errorf("%s: does not contain a complete HTML 5 document (doctype missing?)", path)
			}
			if asiteFile.Document.FirstChild.NextSibling.Type != html.ElementNode {
				return fmt.Errorf("%s: root is not a proper <html> node", path)
			}
			w := walker.Walker{
				Import:   &importHandler{file: &asiteFile.File},
				TextNode: walker.WhitespaceOnly{}}
			body, err := descend(asiteFile.Document, []atom.Atom{atom.Html, atom.Body})
			if err != nil {
				return fmt.Errorf("%s: %s", path, err.Error())
			}
			for c := body.FirstChild; c != nil; c = c.NextSibling {
				if c.Data == "a:site" {
					if asiteFile.Descriptor != nil {
						return fmt.Errorf("%s: duplicate <a:site> element", path)
					}
					_, _, err = w.WalkChildren(nil, &walker.Siblings{Cur: c.FirstChild})
					if err != nil {
						return fmt.Errorf("%s: %s", path, err.Error())
					}
					asiteFile.Descriptor = c
					repl := &html.Node{
						Type:        html.CommentNode,
						Data:        "site",
						Parent:      c.Parent,
						NextSibling: c.NextSibling,
						PrevSibling: c.PrevSibling,
					}
					if c.NextSibling != nil {
						c.NextSibling.PrevSibling = repl
					} else {
						c.Parent.LastChild = repl
					}
					if c.PrevSibling != nil {
						c.PrevSibling.NextSibling = repl
					} else {
						c.Parent.FirstChild = repl
					}
				}
			}
			if asiteFile.Descriptor == nil {
				return fmt.Errorf("%s: missing <a:site> in /html/body", path)
			}
			asiteFile.Descriptor.NextSibling = nil
			asiteFile.Descriptor.PrevSibling = nil
			asiteFile.Descriptor.Parent = nil
			pkg.Site = asiteFile
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

type importHandler struct {
	file *data.File
}

func (ih *importHandler) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	var raw string
	if n.FirstChild != nil {
		if n.LastChild != n.FirstChild || n.FirstChild.Type != html.TextNode {
			return false, nil, errors.New(": may only contain text content")
		}
		raw = n.FirstChild.Data
	}
	imports, err := parsers.ParseImports(raw)
	if err != nil {
		return false, nil, errors.New(": " + err.Error())
	}
	if ih.file.Imports != nil {
		return false, nil, errors.New(": cannot have more than one <a:import> per file")
	}
	ih.file.Imports = imports
	return false, nil, nil
}
