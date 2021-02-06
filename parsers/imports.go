package parsers

import (
	"errors"
	"strings"

	"github.com/yhirose/go-peg"
)

var importsParser *peg.Parser

type importItem struct {
	alias string
	path  string
}

func init() {
	var err error
	importsParser, err = peg.NewParser(`
	ROOT        ← < [\n;]* > IMPORT ( < [\n;]+ > IMPORT)* < [\n;]* >
	IMPORT      ← IDENTIFIER? '"' IMPORTPATH '"'
	IMPORTPATH  ← < [a-zA-Z_0-9./]* >
	` + identifierSyntax + whitespace)
	if err != nil {
		panic(err)
	}
	g := importsParser.Grammar
	g["IDENTIFIER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Token(), nil
	}
	g["IMPORTPATH"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Token(), nil
	}
	g["IMPORT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		if len(v.Vs) == 2 {
			return importItem{alias: v.ToStr(0), path: v.ToStr(1)}, nil
		}
		ret := importItem{path: v.ToStr(0)}
		lastDot := strings.LastIndexByte(ret.path, '/')
		if lastDot == -1 {
			ret.alias = ret.path
		} else {
			ret.alias = ret.path[lastDot+1:]
		}
		return ret, nil
	}
	g["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make(map[string]string)
		for i := range v.Vs {
			item := v.Vs[i].(importItem)
			_, ok := ret[item.alias]
			if ok {
				return nil, errors.New("duplicate import alias: " + item.alias)
			}
			ret[item.alias] = item.path
		}
		return ret, nil
	}
}

// ParseImports parses the content of an a:import element.
func ParseImports(s string) (map[string]string, error) {
	ret, err := importsParser.ParseAndGetValue(s, nil)
	if err != nil {
		return nil, err
	}
	return ret.(map[string]string), nil
}
