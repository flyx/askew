package parsers

import (
	"errors"

	"github.com/flyx/askew/data"
	"github.com/yhirose/go-peg"
)

var fieldsParser *peg.Parser

func init() {
	var err error
	fieldsParser, err = peg.NewParser(`
	ROOT        ← < [\n;]* > FIELD (< [\n;]+ > FIELD)* < [\n;]* >
	FIELD     ← IDENTIFIER (',' IDENTIFIER)* TYPE` +
		typeSyntax + identifierSyntax + whitespace)
	if err != nil {
		panic(err)
	}
	registerType(fieldsParser)
	fieldsParser.Grammar["IDENTIFIER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Token(), nil
	}
	fieldsParser.Grammar["FIELD"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make(map[string]*data.ParamType)
		t := v.Vs[len(v.Vs)-1].(*data.ParamType)
		for i := 0; i < len(v.Vs)-1; i++ {
			name := v.ToStr(i)
			if _, ok := ret[name]; ok {
				return nil, errors.New("duplicate field name: " + name)
			}
			ret[name] = t
		}
		return ret, nil
	}
	fieldsParser.Grammar["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make(map[string]*data.ParamType)
		for _, c := range v.Vs {
			for name, t := range c.(map[string]*data.ParamType) {
				if _, ok := ret[name]; ok {
					return nil, errors.New("duplicate field anem: " + name)
				}
				ret[name] = t
			}
		}
		return ret, nil
	}
}

// ParseFields parses the content of a <a:data> element.
func ParseFields(s string) (map[string]*data.ParamType, error) {
	ret, err := fieldsParser.ParseAndGetValue(s, nil)
	if err != nil {
		return nil, err
	}
	return ret.(map[string]*data.ParamType), nil
}
