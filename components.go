package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/flyx/askew/parsers"

	"github.com/flyx/askew/data"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type componentProcessor struct {
	syms    *data.Symbols
	counter *int
}

func (cp *componentProcessor) process(n *html.Node) (descend bool,
	replacement *html.Node, err error) {
	var cmpAttrs componentAttribs
	err = collectAttribs(n, &cmpAttrs)
	if err != nil {
		return
	}
	if len(cmpAttrs.name) == 0 {
		return false, nil, errors.New(": attribute `name` missing")
	}

	replacement = &html.Node{Type: html.ElementNode, DataAtom: atom.Template,
		Data: "template"}
	cmp := &data.Component{EmbedHost: data.EmbedHost{Dependencies: make(map[string]struct{})},
		Name: cmpAttrs.name, Template: replacement, NeedsController: cmpAttrs.controller,
		Parameters: cmpAttrs.params}
	cp.syms.CurHost = &cmp.EmbedHost
	(*cp.counter)++
	cmp.ID = fmt.Sprintf("askew-component-%d-%s", *cp.counter, strings.ToLower(cmpAttrs.name))
	replacement.Attr = []html.Attribute{html.Attribute{Key: "id", Val: cmp.ID}}

	var indexList []int
	w := walker{
		text: allow{}, aText: &aTextProcessor{cmp, &indexList},
		embed:       &embedProcessor{cp.syms, &indexList},
		handler:     &handlerProcessor{cp.syms, &cmp.Handlers, &indexList},
		stdElements: &stdElementHandler{cp.syms, &indexList, cmp, -1, nil},
		indexList:   &indexList}
	replacement.FirstChild, replacement.LastChild, err = w.walkChildren(
		replacement, &siblings{n.FirstChild})

	// reverse Embed list so that they get embedded in reverse order.
	// this is necessary because embedding may change the number of elements in
	// a node, rendering the path of following embeds invalid.
	tmp := make([]data.Embed, len(cmp.Embeds))
	for i, e := range cmp.Embeds {
		tmp[len(tmp)-i-1] = e
	}
	cmp.Embeds = tmp

	cp.syms.Packages[cp.syms.CurPkg].Components[cmpAttrs.name] = cmp

	return
}

type embedProcessor struct {
	syms      *data.Symbols
	indexList *[]int
}

func resolveEmbed(n *html.Node, syms *data.Symbols, indexList []int) (data.Embed, error) {
	targetType := attrVal(n.Attr, "type")
	if len(targetType) == 0 {
		return data.Embed{}, errors.New(": attribute `type` missing")
	}
	target, pkgName, typeName, err := syms.ResolveComponent(targetType)
	if err != nil {
		return data.Embed{}, errors.New(": attribute `type` invalid: %s" + err.Error())
	}
	e := data.Embed{Path: append([]int(nil), indexList...),
		List: attrExists(n.Attr, "list")}
	if e.List {
		target.NeedsList = true
	}
	if pkgName != syms.CurPkg {
		e.Pkg = pkgName
	}
	e.Field = attrVal(n.Attr, "name")
	if e.Field == "" {
		return data.Embed{}, errors.New(": attribute `name` missing")
	}
	e.T = typeName
	if n.FirstChild != nil {
		return data.Embed{}, errors.New(": illegal content")
	}
	args := attrVal(n.Attr, "args")
	if e.List {
		if args != "" {
			return data.Embed{}, errors.New(": embed with `list` cannot have `args`")
		}
	} else {
		if args != "" {
			var err error
			e.Args, err = parsers.AnalyseArguments(args)
			if err != nil {
				return data.Embed{}, errors.New(": in `args`: " + err.Error())
			}
		}
		if len(target.Parameters) != e.Args.Count {
			return data.Embed{}, fmt.Errorf(
				": target component requires %d arguments, but %d were given", len(target.Parameters), e.Args.Count)
		}
	}
	return e, nil
}

func (ep *embedProcessor) process(n *html.Node) (descend bool,
	replacement *html.Node, err error) {
	e, err := resolveEmbed(n, ep.syms, *ep.indexList)
	if err != nil {
		return false, nil, err
	}

	ep.syms.CurHost.Embeds = append(ep.syms.CurHost.Embeds, e)
	replacement = &html.Node{Type: html.CommentNode,
		Data: "embed(" + e.Field + ")"}
	return
}

type handlerProcessor struct {
	syms      *data.Symbols
	handlers  *map[string]data.Handler
	indexList *[]int
}

func (hp *handlerProcessor) process(n *html.Node) (descend bool,
	replacement *html.Node, err error) {
	if len(*hp.indexList) != 1 {
		return false, nil, errors.New(": must be defined as direct child of <a:component>")
	}
	def := n.FirstChild
	if def.Type != html.TextNode || def.NextSibling != nil {
		return false, nil, errors.New(": must have plain text as content and nothing else")
	}
	parsed, err := parsers.ParseHandler(def.Data)
	if err != nil {
		return false, nil, errors.New(": unable to parse `" + def.Data + "`: " + err.Error())
	}
	if *hp.handlers == nil {
		*hp.handlers = make(map[string]data.Handler)
	} else {
		_, ok := (*hp.handlers)[parsed.Name]
		if ok {
			return false, nil, errors.New(": duplicate handler name: " + parsed.Name)
		}
	}
	(*hp.handlers)[parsed.Name] = data.Handler{Params: parsed.Params}
	replacement = &html.Node{Type: html.CommentNode, Data: "handler: " + def.Data}
	return
}

type formValue struct {
	t     data.VariableType
	radio bool
}

type formValueDiscovery struct {
	values map[string]formValue
}

func (d *formValueDiscovery) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	var v formValue
	name := attrVal(n.Attr, "name")
	if name == "" {
		return false, nil, nil
	}
	switch n.DataAtom {
	case atom.Input:
		switch inputType := attrVal(n.Attr, "type"); inputType {
		case "radio":
			v.radio = true
			v.t = data.StringVar
		case "number", "range":
			if strings.ContainsRune(attrVal(n.Attr, "min"), '.') ||
				strings.ContainsRune(attrVal(n.Attr, "max"), '.') ||
				strings.ContainsRune(attrVal(n.Attr, "step"), '.') {
				return false, nil, errors.New(": non-integer " + inputType + " inputs not supported")
			}
			v.t = data.IntVar
		case "text", "":
			v.t = data.StringVar
		case "submit", "reset", "hidden":
			return false, nil, nil
		default:
			return false, nil, errors.New(": unsupported input type: `" + inputType + "`")
		}
	case atom.Select, atom.Textarea:
		v.t, v.radio = data.StringVar, false
	default:
		return false, nil, nil
	}
	existing, ok := d.values[name]
	if ok {
		if v.radio && existing.radio {
			return false, nil, nil
		}
		return false, nil, errors.New(": duplicate name `" + name + "` in same form")
	}
	d.values[name] = v
	return false, nil, nil
}

func discoverFormValues(form *html.Node) (map[string]formValue, error) {
	fvd := formValueDiscovery{values: make(map[string]formValue)}
	w := walker{text: allow{}, embed: dontDescend{}, handler: dontDescend{},
		aText: allow{}, stdElements: &fvd}
	var err error
	form.FirstChild, form.LastChild, err = w.walkChildren(form, &siblings{form.FirstChild})
	if err != nil {
		return nil, err
	}
	return fvd.values, nil
}

type stdElementHandler struct {
	syms       *data.Symbols
	indexList  *[]int
	c          *data.Component
	curFormPos int
	curForm    map[string]formValue
}

func (seh *stdElementHandler) mapCaptures(n *html.Node, v []data.EventMapping) error {
	if len(v) == 0 {
		return nil
	}
	formDepth := -1
	if seh.curFormPos != -1 {
		formDepth = len(*seh.indexList) - seh.curFormPos
	}
	for _, m := range v {
		h, ok := seh.c.Handlers[m.Handler]
		if !ok {
			return errors.New("capture references unknown handler: " + m.Handler)
		}
		notMapped := make(map[string]struct{})
		for pName := range m.ParamMappings {
			notMapped[pName] = struct{}{}
		}
		for _, p := range h.Params {
			bVal, ok := m.ParamMappings[p.Name]
			if !ok {
				m.ParamMappings[p.Name] = data.BoundValue{Kind: data.BoundData, ID: p.Name}
			} else {
				delete(notMapped, p.Name)
				if bVal.Kind == data.BoundFormValue {
					if formDepth == -1 {
						return errors.New(": illegal form() binding outside of <form> element")
					}
					bVal.FormDepth = formDepth
					_, ok := seh.curForm[bVal.ID]
					if !ok {
						return errors.New(": unknown form value name: `" + bVal.ID + "`")
					}
				}
			}
		}
		for unknown := range notMapped {
			return errors.New("unknown param for capture mapping: " + unknown)
		}
	}

	seh.c.Captures = append(seh.c.Captures, data.Capture{
		Path: append([]int(nil), *seh.indexList...), Mappings: v})
	return nil
}

func (seh *stdElementHandler) processBindings(arr []data.VariableMapping) error {
	formDepth := -1
	if seh.curFormPos != -1 {
		formDepth = len(*seh.indexList) - seh.curFormPos
	}

	for _, vb := range arr {
		if vb.Value.Kind == data.BoundFormValue {
			if formDepth == -1 {
				return errors.New(": illegal form() binding outside of <form> element")
			}
			vb.Value.FormDepth = formDepth
			val, ok := seh.curForm[vb.Variable.Name]
			if !ok {
				return errors.New(": unknown form value name: `" + vb.Variable.Name + "`")
			}
			if vb.Variable.Type == data.AutoVar {
				vb.Variable.Type = val.t
			}
		} else {
			if vb.Variable.Type == data.AutoVar {
				if vb.Value.Kind == data.BoundClass {
					vb.Variable.Type = data.BoolVar
				} else {
					vb.Variable.Type = data.StringVar
				}
			}
		}
		vb.Path = append([]int(nil), *seh.indexList...)
		seh.c.Variables = append(seh.c.Variables, vb)
	}
	return nil
}

func (seh *stdElementHandler) formDepth() int {
	if seh.curFormPos == -1 {
		return -1
	}
	return len(*seh.indexList) - seh.curFormPos
}

func (seh *stdElementHandler) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	if len(*seh.indexList) <= seh.curFormPos {
		seh.curFormPos = -1
		seh.curForm = nil
	}

	var attrs generalAttribs
	err = extractAskewAttribs(n, &attrs)
	if err != nil {
		return
	}
	switch n.DataAtom {
	case atom.Form:
		if seh.curFormPos != -1 {
			return false, nil, errors.New(": nested <form> not allowed")
		}
		seh.curFormPos = len(*seh.indexList)
		vals, err := discoverFormValues(n)
		if err != nil {
			return false, nil, err
		}
		seh.curForm = vals
		break
	default:
		break
	}
	if err := seh.mapCaptures(n, attrs.capture); err != nil {
		return false, nil, errors.New(": " + err.Error())
	}
	if attrs._if != "" {
		seh.c.Conditionals = append(seh.c.Conditionals, data.Conditional{
			Condition: attrs._if, Path: append([]int(nil), *seh.indexList...)})
	}
	if attrs.assign != nil {
		path := append([]int(nil), *seh.indexList...)
		for i := range attrs.assign {
			attrs.assign[i].Path = path
		}
		seh.c.Assignments = append(seh.c.Assignments, attrs.assign...)
	}
	err = seh.processBindings(attrs.bindings)
	descend = err == nil
	return
}

type aTextProcessor struct {
	c         *data.Component
	indexList *[]int
}

func (atp *aTextProcessor) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	expr := attrVal(n.Attr, "expr")
	if expr == "" {
		return false, nil, errors.New(": missing attribute `expr`")
	}
	if n.FirstChild != nil {
		return false, nil, errors.New(": node may not have child nodes")
	}
	atp.c.Assignments = append(atp.c.Assignments, data.ParamAssignment{Expression: expr, Path: append([]int(nil), *atp.indexList...)})
	return false, &html.Node{Type: html.CommentNode, Data: "a:text"}, nil
}

type packageProcessor struct {
	syms    *data.Symbols
	counter *int
}

func (pp *packageProcessor) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	pp.syms.CurPkg = attrVal(n.Attr, "name")
	p := pp.syms.Packages[pp.syms.CurPkg]
	p.Components = make(map[string]*data.Component)
	w := walker{text: whitespaceOnly{}, component: &componentProcessor{pp.syms, pp.counter}}
	n.FirstChild, n.LastChild, err = w.walkChildren(n, &siblings{n.FirstChild})
	return false, nil, err
}
