package units

import (
	"errors"
	"fmt"

	"github.com/flyx/askew/attributes"
	"github.com/flyx/askew/data"
	"github.com/flyx/askew/parsers"
	"github.com/flyx/askew/walker"
	"github.com/flyx/net/html"
)

type constructProcessor struct {
	e      *data.Embed
	target *data.Component
}

func (cp *constructProcessor) Process(n *html.Node) (descend bool,
	replacement *html.Node, err error) {
	if cp.e.Kind == data.DirectEmbed {
		return false, nil, errors.New(": element requires list or optional embed as parent")
	}
	if cp.target == nil {
		return false, nil, errors.New(": element requires embed with explicit type as parent")
	}

	var attrs attributes.General
	if err = attributes.ExtractAskewAttribs(n, &attrs); err != nil {
		return
	}
	if attrs.Assign != nil {
		return false, nil, errors.New(": a:assign not allowed here")
	}
	if attrs.Bindings != nil {
		return false, nil, errors.New(": a:bindings not allowed here")
	}
	if attrs.Capture != nil {
		return false, nil, errors.New(": a:capture not allowed here")
	}
	if attrs.For != nil && attrs.If != nil {
		return false, nil, errors.New(": cannot have both a:if and a:for here")
	}
	args, err := parsers.AnalyseArguments(attributes.Val(n.Attr, "args"))
	if err != nil {
		return false, nil, errors.New(": in args: " + err.Error())
	}
	if args.Count != len(cp.target.Parameters) {
		return false, nil, fmt.Errorf(
			": target component requires %d arguments, but %d were given", len(cp.target.Parameters), args.Count)
	}
	if attrs.If != nil {
		cp.e.ConstructorCalls = append(cp.e.ConstructorCalls,
			data.NestedConstructorCall{ConstructorCall: data.ConstructorCall{Args: args},
				Kind: data.NestedIf, Expression: attrs.If.Expression})
	} else if attrs.For != nil {
		if cp.e.Kind == data.OptionalEmbed {
			return false, nil, errors.New(": a:for not allowed inside optional embed")
		}
		cp.e.ConstructorCalls = append(cp.e.ConstructorCalls,
			data.NestedConstructorCall{ConstructorCall: data.ConstructorCall{Args: args},
				Kind: data.NestedFor, Index: attrs.For.Index,
				Variable: attrs.For.Variable, Expression: attrs.For.Expression})
	} else {
		cp.e.ConstructorCalls = append(cp.e.ConstructorCalls,
			data.NestedConstructorCall{ConstructorCall: data.ConstructorCall{Args: args},
				Kind: data.NestedDirect})
	}
	w := walker.Walker{TextNode: walker.WhitespaceOnly{}}
	_, _, err = w.WalkChildren(n, &walker.Siblings{Cur: n.FirstChild})
	return false, nil, err
}
