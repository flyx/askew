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
	FIELD       ← FIELDNAMES TYPE ('=' EXPR)?
	FIELDNAMES  ← IDENTIFIER (',' IDENTIFIER)*` +
		typeSyntax + exprSyntax)
	if err != nil {
		panic(err)
	}
	registerType(fieldsParser)
	fieldsParser.Grammar["IDENTIFIER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Token(), nil
	}
	fieldsParser.Grammar["FIELDNAMES"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make([]string, len(v.Vs))
		for i := 0; i < len(v.Vs); i++ {
			ret[i] = v.ToStr(i)
		}
		return ret, nil
	}
	fieldsParser.Grammar["FIELD"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		names := v.Vs[0].([]string)
		ret := make([]*data.Field, len(names))
		t := v.Vs[1].(*data.ParamType)
		var defaultValue *string
		if len(v.Vs) == 3 {
			str := v.ToStr(2)
			defaultValue = &str
		}
		for index, name := range names {
			ret[index] = &data.Field{Name: name, Type: t, DefaultValue: defaultValue}
		}
		return ret, nil
	}
	fieldsParser.Grammar["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make([]*data.Field, 0, len(v.Vs)*3)
		names := make(map[string]struct{})
		for _, c := range v.Vs {
			fields := c.([]*data.Field)
			for _, field := range fields {
				if _, ok := names[field.Name]; ok {
					return nil, errors.New("duplicate field name: " + field.Name)
				}
				names[field.Name] = struct{}{}
				ret = append(ret, field)
			}
		}
		return ret, nil
	}
}

// ParseFields parses the content of a <a:data> element.
func ParseFields(s string) ([]*data.Field, error) {
	ret, err := fieldsParser.ParseAndGetValue(s, nil)
	if err != nil {
		return nil, err
	}
	return ret.([]*data.Field), nil
}
