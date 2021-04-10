package packages

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/flyx/askew/data"
	"github.com/flyx/askew/parsers"
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
	dotAskewTmpl
	dotAsite
	dotAsiteTmpl
	dotOther
)

func fileKind(name string) suffix {
	lower := strings.ToLower(name)

	if strings.HasSuffix(lower, ".askew") {
		return dotAskew
	}
	if strings.HasSuffix(lower, ".askew.tmpl") {
		return dotAskewTmpl
	}
	if strings.HasSuffix(lower, ".asite") {
		return dotAsite
	}
	if strings.HasSuffix(lower, ".asite.tmpl") {
		return dotAsiteTmpl
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
func Discover(excludes []string, tmplData interface{}) (*data.BaseDir, error) {
	var err error
	ret := &data.BaseDir{}
	ret.ImportPath, err = findBasePath()
	if err != nil {
		return nil, err
	}

	ret.Packages = make(map[string]*data.Package)
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			for _, exclude := range excludes {
				if matched, _ := filepath.Match(exclude, path); matched {
					return filepath.SkipDir
				}
			}
		}
		kind := fileKind(info.Name())
		if info.IsDir() || kind == dotOther {
			return nil
		}
		os.Stdout.WriteString("[info] discovered: " + path + "\n")
		relPath := filepath.Dir(path)
		assumedPkgName := filepath.Base(relPath)
		if assumedPkgName == "." {
			assumedPkgName = filepath.Base(ret.ImportPath)
		}
		pkg, ok := ret.Packages[relPath]
		if !ok {
			// the .Name of the package is only set when the first file inside that
			// package is processed.
			pkg = &data.Package{Files: make([]*data.AskewFile, 0, 32),
				ImportPath: filepath.ToSlash(filepath.Join(ret.ImportPath, relPath))}
			ret.Packages[relPath] = pkg
		}

		var contents []byte

		var baseName string
		if kind == dotAskewTmpl || kind == dotAsiteTmpl {
			var tmpl *template.Template
			tmpl, err = template.New(filepath.Base(path)).ParseFiles(path)
			if err != nil {
				return err
			}
			var writer bytes.Buffer
			if err = tmpl.Execute(&writer, tmplData); err != nil {
				return err
			}
			contents = writer.Bytes()
			kind--
			baseName = info.Name()[:len(info.Name())-11]
		} else {
			contents, err = ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			baseName = info.Name()[:len(info.Name())-6]
		}

		if kind == dotAskew {
			askewFile := &data.AskewFile{File: data.File{BaseName: baseName, Path: path}}
			askewFile.Content, err = html.ParseFragmentWithOptions(
				bytes.NewReader(contents), &data.BodyEnv,
				html.ParseOptionCustomElements(walker.AskewElements))
			if err != nil {
				return fmt.Errorf("%s: %s", path, err.Error())
			}
			pHandler := &packageHandler{pkg: pkg, seen: false}
			w := walker.Walker{
				Package:   pHandler,
				Import:    &importHandler{file: &askewFile.File},
				Component: walker.DontDescend{},
				Macro:     walker.DontDescend{},
				TextNode:  walker.WhitespaceOnly{}}
			_, _, err = w.WalkChildren(nil, &walker.NodeSlice{Items: askewFile.Content})
			if err != nil {
				return fmt.Errorf("%s: %s", path, err.Error())
			}
			if !pHandler.seen {
				if pkg.Name == "" {
					pkg.Name = assumedPkgName
				} else if pkg.Name != assumedPkgName {
					return fmt.Errorf(
						"%s: <a:package> missing, another file has already set the package name to '%s'",
						path, pkg.Name)
				}
			}
			if askewFile.File.Imports == nil {
				askewFile.File.Imports = make(map[string]string)
			}
			if url, ok := askewFile.File.Imports["askew"]; ok {
				if url != "github.com/flyx/askew/runtime" {
					return fmt.Errorf(
						"%s: if the alias `askew` is given in imports, it must link to \"github.com/flyx/askew/runtime\"",
						path)
				}
			} else {
				askewFile.File.Imports["askew"] = "github.com/flyx/askew/runtime"
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
			rootNode := asiteFile.Document.FirstChild.NextSibling
			if rootNode.Type != html.ElementNode || rootNode.Data != "a:site" {
				return fmt.Errorf("%s: root is not a <a:site> node", path)
			}
			head, err := descend(rootNode, []atom.Atom{atom.Head})
			if err != nil {
				return fmt.Errorf("%s: %s", path, err.Error())
			}
			pHandler := &packageHandler{pkg: pkg, seen: false, remove: true}
			w := walker.Walker{
				Package:     pHandler,
				Import:      &importHandler{file: &asiteFile.File, remove: true},
				TextNode:    walker.WhitespaceOnly{},
				StdElements: walker.DontDescend{}}
			_, _, err = w.WalkChildren(head, &walker.Siblings{Cur: head.FirstChild})
			if err != nil {
				return fmt.Errorf("%s: %s", path, err.Error())
			}
			if !pHandler.seen {
				if pkg.Name == "" {
					pkg.Name = assumedPkgName
				} else if pkg.Name != assumedPkgName {
					return fmt.Errorf("%s: <a:package> missing, has been set to %s in another file", path, pkg.Name)
				}
			}
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
	file   *data.File
	remove bool
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
	if ih.remove {
		return false, &html.Node{Type: html.CommentNode, Data: "imports"}, nil
	}
	return false, nil, nil
}

type packageHandler struct {
	pkg          *data.Package
	seen, remove bool
}

func (ph *packageHandler) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	if ph.seen {
		return false, nil, errors.New(": only one package directive allowed per file")
	}
	if n.FirstChild != nil {
		if n.LastChild != n.FirstChild || n.FirstChild.Type != html.TextNode {
			return false, nil, errors.New(": may only contain text content")
		}
	} else {
		return false, nil, errors.New(": must contain text content")
	}
	name := strings.TrimSpace(n.FirstChild.Data)
	if name == "" {
		return false, nil, errors.New(": package name must not be empty")
	}
	if ph.pkg.Name == "" {
		ph.pkg.Name = name
	} else if ph.pkg.Name != name {
		return false, nil, errors.New(": conflicting package name, does not equal '" + ph.pkg.Name + "'")
	}
	ph.seen = true
	if ph.remove {
		return false, &html.Node{Type: html.CommentNode, Data: "package"}, nil
	}
	return false, nil, nil
}
