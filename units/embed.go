package units

import (
	"errors"
	"fmt"
	"unicode"

	"github.com/flyx/askew/attributes"
	"github.com/flyx/askew/data"
	"github.com/flyx/askew/walker"
	"github.com/flyx/net/html"
)

type embedProcessor struct {
	syms      *data.Symbols
	indexList *[]int
}

func resolveEmbed(n *html.Node, syms *data.Symbols,
	indexList []int) (data.Embed, *data.Component, error) {
	var attrs attributes.Embed
	if err := attributes.Collect(n, &attrs); err != nil {
		return data.Embed{}, nil, err
	}

	e := data.Embed{Kind: data.DirectEmbed, Path: append([]int(nil), indexList...),
		Field: attrs.Name, Control: attrs.Control}
	if e.Field == "" {
		return data.Embed{}, nil, errors.New(": attribute `name` missing")
	}
	if attrs.List {
		e.Kind = data.ListEmbed
	}
	if attrs.Optional {
		if e.Kind != data.DirectEmbed {
			return data.Embed{}, nil, errors.New(": cannot mix `list` and `optional`")
		}
		e.Kind = data.OptionalEmbed
	}
	if attrs.T == "" {
		if e.Kind == data.DirectEmbed {
			return data.Embed{}, nil, errors.New(": attribute `type` missing (may only be omitted for optional or list embeds)")
		}
		if attrs.Args.Count != 0 {
			return data.Embed{}, nil, errors.New(": embed with `list` or `optional` cannot have `args`")
		}
		return e, nil, nil
	}
	target, typeName, aliasName, err := syms.ResolveComponent(attrs.T)
	if err != nil {
		return data.Embed{}, nil, errors.New(": attribute `type` invalid: " + err.Error())
	}
	switch e.Kind {
	case data.ListEmbed:
		target.NeedsList = true
	case data.OptionalEmbed:
		target.NeedsOptional = true
	}
	e.T = typeName
	e.Ns = aliasName
	if e.Kind != data.DirectEmbed {
		if attrs.Args.Count != 0 {
			return data.Embed{}, nil, errors.New(": embed with `list` or `optional` cannot have `args`")
		}
	} else {
		e.Args = attrs.Args
		if len(target.Parameters) != e.Args.Count {
			return data.Embed{}, nil, fmt.Errorf(
				": target component requires %d arguments, but %d were given", len(target.Parameters), e.Args.Count)
		}
	}
	runes := []rune(typeName)
	if unicode.IsUpper(runes[0]) {
		e.ConstructorName = "New" + typeName
	} else {
		e.ConstructorName = "new" + string(unicode.ToUpper(runes[0])) + string(runes[1:])
	}
	return e, target, nil
}

// Process implements Walker.NodeHandler.
func (ep *embedProcessor) Process(n *html.Node) (descend bool,
	replacement *html.Node, err error) {
	e, target, err := resolveEmbed(n, ep.syms, *ep.indexList)
	if err != nil {
		return false, nil, err
	}

	w := walker.Walker{TextNode: &walker.WhitespaceOnly{},
		Construct: &constructProcessor{&e, target}}
	_, _, err = w.WalkChildren(n, &walker.Siblings{Cur: n.FirstChild})
	if e.Kind == data.OptionalEmbed && len(e.ConstructorCalls) > 1 {
		return false, nil, errors.New(": too many <a:construct> for optional embed")
	}
	ep.syms.CurUnit.Embeds = append(ep.syms.CurUnit.Embeds, e)
	replacement = &html.Node{Type: html.CommentNode,
		Data: "embed(" + e.Field + ")"}
	return
}
