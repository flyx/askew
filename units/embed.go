package units

import (
	"errors"
	"fmt"

	"github.com/flyx/askew/attributes"
	"github.com/flyx/askew/data"
	"github.com/flyx/askew/walker"
	"github.com/flyx/net/html"
)

type embedProcessor struct {
	syms            *data.Symbols
	indexList       *[]int
	parentIndexList []int
}

// resolves the type of an embedded component.
// the target is only set when a component in the same module is referenced.
func resolveEmbed(n *html.Node, syms *data.Symbols,
	indexList []int) (e data.Embed, target *data.Component, newName string, err error) {
	var attrs attributes.Embed
	if err := attributes.Collect(n, &attrs); err != nil {
		return data.Embed{}, nil, "", err
	}

	e = data.Embed{Kind: data.DirectEmbed, Path: indexList,
		Field: attrs.Name, Control: attrs.Control}
	if e.Field == "" {
		return data.Embed{}, nil, "", errors.New(": attribute `name` missing")
	}
	if attrs.List {
		e.Kind = data.ListEmbed
	}
	if attrs.Optional {
		if e.Kind != data.DirectEmbed {
			return data.Embed{}, nil, "", errors.New(": cannot mix `list` and `optional`")
		}
		e.Kind = data.OptionalEmbed
	}
	if attrs.T == "" {
		if e.Kind == data.DirectEmbed {
			return data.Embed{}, nil, "", errors.New(": attribute `type` missing (may only be omitted for optional or list embeds)")
		}
		if attrs.Args.Count != 0 {
			return data.Embed{}, nil, "", errors.New(": embed with `list` or `optional` cannot have `args`")
		}
		return e, nil, "", nil
	}
	target, e.T, e.Ns, err = syms.ResolveComponent(attrs.T)
	canCheckArgNumber := false
	if err != nil {
		if _, ok := err.(data.OutsideModuleErr); ok {
			// components outside modules are fine, we can't check for correct number
			// of parameters though
			err = nil
		} else {
			return data.Embed{}, nil, "", errors.New(": attribute `type` invalid: " + err.Error())
		}
	} else {
		// only when askew generates the new and init funcs for the component can
		// we check whether the correct number of arguments have been provided.
		canCheckArgNumber = target.GenNewInit
	}

	if e.Kind != data.DirectEmbed {
		if attrs.Args.Count != 0 {
			return data.Embed{}, nil, "", errors.New(": embed with `list` or `optional` cannot have `args`")
		}
	} else {
		e.Args = attrs.Args
		if canCheckArgNumber {
			if len(target.Parameters) != e.Args.Count {
				return data.Embed{}, nil, "", fmt.Errorf(
					": target component requires %d arguments, but %d were given", len(target.Parameters), e.Args.Count)
			}
		}
	}
	if e.Ns != "" {
		newName = e.Ns + "."
	}
	if target != nil {
		newName += target.NewName()
	} else {
		newName += "New" + e.T
	}
	return e, target, newName, nil
}

// Process implements Walker.NodeHandler.
func (ep *embedProcessor) Process(n *html.Node) (descend bool,
	replacement *html.Node, err error) {
	path := make([]int, 0, len(ep.parentIndexList)+len(*ep.indexList))
	path = append(path, ep.parentIndexList...)
	path = append(path, *ep.indexList...)
	e, target, newName, err := resolveEmbed(n, ep.syms, path)
	if err != nil {
		return false, nil, err
	}

	cp := constructProcessor{ep.syms, &e, constructParent{newName: newName}}
	if target != nil {
		cp.parentType.numParams = len(target.Parameters)
	} else {
		cp.parentType.numParams = -1
	}
	w := walker.Walker{TextNode: &walker.WhitespaceOnly{},
		Construct: &cp}
	_, _, err = w.WalkChildren(n, &walker.Siblings{Cur: n.FirstChild})
	if e.Kind == data.OptionalEmbed && len(e.ConstructorCalls) > 1 {
		return false, nil, errors.New(": too many <a:construct> for optional embed")
	}
	ep.syms.CurUnit.Embeds = append(ep.syms.CurUnit.Embeds, e)
	replacement = &html.Node{Type: html.CommentNode,
		Data: "embed(" + e.Field + ")"}
	return
}
