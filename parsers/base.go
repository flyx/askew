package parsers

import (
	"github.com/flyx/askew/data"
	"github.com/yhirose/go-peg"
)

var typeSyntax = `
TYPE     ← TYPENAME / ARRAY / MAP / POINTER
TYPENAME ← IDENTIFIER ('.' IDENTIFIER)?
ARRAY    ← '[]' TYPE
MAP      ← 'map' '[' IDENTIFIER ']' TYPE
POINTER  ← '*' TYPE`

var identifierSyntax = `
IDENTIFIER  ← < [a-zA-Z_][a-zA-Z_0-9]* >`

var whitespace = `
%whitespace ← [ \t]*`

func registerType(p *peg.Parser) {
	p.Grammar["TYPE"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Vs[0].(*data.ParamType), nil
	}
	p.Grammar["TYPENAME"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		name := v.ToStr(0)
		if len(v.Vs) == 1 {
			switch name {
			case "int":
				return &data.ParamType{Kind: data.IntType}, nil
			case "bool":
				return &data.ParamType{Kind: data.BoolType}, nil
			case "string":
				return &data.ParamType{Kind: data.StringType}, nil
			default:
				return &data.ParamType{Kind: data.NamedType, Name: name}, nil
			}
		}
		name += "."
		name += v.ToStr(1)
		if name == "js.Value" {
			return &data.ParamType{Kind: data.JSValueType}, nil
		}
		return &data.ParamType{Kind: data.NamedType, Name: name}, nil
	}
	p.Grammar["ARRAY"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return &data.ParamType{Kind: data.ArrayType, ValueType: v.Vs[0].(*data.ParamType)}, nil
	}
	p.Grammar["MAP"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return &data.ParamType{Kind: data.ArrayType,
			KeyType:   v.Vs[0].(*data.ParamType),
			ValueType: v.Vs[1].(*data.ParamType)}, nil
	}
	p.Grammar["POINTER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return &data.ParamType{Kind: data.PointerType, ValueType: v.Vs[0].(*data.ParamType)}, nil
	}
}
