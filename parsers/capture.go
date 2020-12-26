package parsers

import (
	"errors"
	"fmt"

	"github.com/flyx/askew/data"
	peg "github.com/yhirose/go-peg"
)

type paramMapping struct {
	param    string
	supplier data.BoundValue
}

type tag struct {
	name   string
	params []string
}

var captureParser *peg.Parser

func init() {
	var err error
	captureParser, err = peg.NewParser(`
	ROOT        ← CAPTURE (',' CAPTURE)*
	CAPTURE     ← EVENTID ':' HANDLER MAPPINGS TAGS
	EVENTID     ← < [a-z]+ >
	HANDLER     ← < [a-zA-Z_][a-zA-Z_0-9]* >
	MAPPINGS    ← ('(' (MAPPING (',' MAPPING)*)? ')')?
	MAPPING     ← IDENTIFIER '=' BOUND
	TAGS        ← ('{' (TAG (',' TAG)*)? '}')?
	TAG         ← IDENTIFIER ( '(' (IDENTIFIER (',' IDENTIFIER)*)? ')' )?
	` + identifierSyntax + boundSyntax + whitespace)
	if err != nil {
		panic(err)
	}
	registerBinders(captureParser)
	g := captureParser.Grammar
	g["IDENTIFIER"].Action = strToken
	g["EVENTID"].Action = strToken
	g["HANDLER"].Action = strToken
	g["MAPPING"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return paramMapping{param: v.ToStr(0), supplier: v.Vs[1].(data.BoundValue)}, nil
	}
	g["MAPPINGS"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make(map[string]data.BoundValue)
		if len(v.Vs) == 0 {
			return ret, nil
		}
		first := v.Vs[0].(paramMapping)
		ret[first.param] = first.supplier
		for i := 1; i < len(v.Vs); i++ {
			next := v.Vs[i].(paramMapping)
			_, ok := ret[next.param]
			if ok {
				return nil, errors.New("duplicate param: " + next.param)
			}
			ret[next.param] = next.supplier
		}
		return ret, nil
	}
	g["TAG"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := tag{name: v.ToStr(0)}
		if len(v.Vs) > 1 {
			ret.params = make([]string, len(v.Vs)-1)
			for i := 1; i < len(v.Vs); i++ {
				ret.params[i-1] = v.ToStr(i)
			}
		}
		return ret, nil
	}
	g["TAGS"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		handling := data.AutoPreventDefault
		for i := range v.Vs {
			t := v.Vs[i].(tag)
			switch t.name {
			case "preventDefault":
				if handling != data.AutoPreventDefault {
					return nil, errors.New("duplicate preventDefault")
				}
				switch len(t.params) {
				case 0:
					handling = data.PreventDefault
				case 1:
					switch t.params[0] {
					case "true":
						handling = data.PreventDefault
					case "false":
						handling = data.DontPreventDefault
					case "ask":
						handling = data.AskPreventDefault
					default:
						return nil, fmt.Errorf("unsupported value for preventDefault: %s", t.params[0])
					}
				default:
					return nil, errors.New("too many parameters for preventDefault")
				}
			default:
				return nil, errors.New("unknown tag: " + t.name)
			}
		}
		return handling, nil
	}
	g["CAPTURE"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return data.UnboundEventMapping{Event: v.ToStr(0), Handler: v.ToStr(1),
			ParamMappings: v.Vs[2].(map[string]data.BoundValue),
			Handling:      v.Vs[3].(data.EventHandling)}, nil
	}
	g["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make([]data.UnboundEventMapping, v.Len())
		for i, c := range v.Vs {
			ret[i] = c.(data.UnboundEventMapping)
		}
		return ret, nil
	}
}

// ParseCapture parses the content of an a:capture attribute.
func ParseCapture(s string) ([]data.UnboundEventMapping, error) {
	ret, err := captureParser.ParseAndGetValue(s, nil)
	if err != nil {
		return nil, err
	}
	return ret.([]data.UnboundEventMapping), nil
}
