package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/flyx/tbc/parsers"

	"github.com/flyx/tbc/data"
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
	collectAttribs(n, &cmpAttrs)
	if len(cmpAttrs.name) == 0 {
		return false, nil, errors.New(": attribute `name` missing")
	}

	replacement = &html.Node{Type: html.ElementNode, DataAtom: atom.Template,
		Data: "template"}
	cmp := &data.Component{EmbedHost: data.EmbedHost{Dependencies: make(map[string]struct{})},
		Name: cmpAttrs.name, Template: replacement, NeedsController: cmpAttrs.controller}
	cp.syms.CurHost = &cmp.EmbedHost
	(*cp.counter)++
	cmp.ID = fmt.Sprintf("tbc-component-%d-%s", *cp.counter, strings.ToLower(cmpAttrs.name))
	replacement.Attr = []html.Attribute{html.Attribute{Key: "id", Val: cmp.ID}}

	var indexList []int
	w := walker{
		text:        allow{},
		embed:       &embedProcessor{cp.syms, &indexList},
		handler:     &handlerProcessor{cp.syms, &cmp.Handlers, &indexList},
		stdElements: &stdElementHandler{cp.syms, &indexList, cmp},
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
		return false, nil, errors.New(": must be defined as direct child of <tbc:component>")
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

func mapCaptures(c *data.Component, n *html.Node, path []int, v []data.EventMapping) error {
	if len(v) == 0 {
		return nil
	}
	for _, m := range v {
		h, ok := c.Handlers[m.Handler]
		if !ok {
			return errors.New("capture references unknown handler: " + m.Handler)
		}
		for pName := range m.ParamMappings {
			_, ok = h.Params[pName]
			if !ok {
				return errors.New("unknown param for capture mapping: " + pName)
			}
		}
		for pName := range h.Params {
			_, ok = m.ParamMappings[pName]
			if !ok {
				m.ParamMappings[pName] = data.BoundValue{Kind: data.BoundAttribute, ID: "data-" + pName}
			}
		}
	}

	c.Captures = append(c.Captures, data.Capture{
		Path: append([]int(nil), path...), Mappings: v})
	return nil
}

func processBindings(c *data.Component, path []int, arr []data.VariableMapping) {
	for _, vb := range arr {
		if vb.Variable.Type == data.AutoVar {
			if vb.Value.Kind == data.BoundClass {
				vb.Variable.Type = data.BoolVar
			} else {
				vb.Variable.Type = data.StringVar
			}
		}
		vb.Path = path
		c.Variables = append(c.Variables, vb)
	}
}

type stdElementHandler struct {
	syms      *data.Symbols
	indexList *[]int
	c         *data.Component
}

func (seh *stdElementHandler) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	var tbcAttrs generalAttribs
	extractTbcAttribs(n, &tbcAttrs)
	if n.DataAtom == atom.Input {
		if err := mapCaptures(seh.c, n, *seh.indexList, tbcAttrs.capture); err != nil {
			return false, nil, errors.New(": " + err.Error())
		}
		path := append([]int(nil), *seh.indexList...)
		if !tbcAttrs.ignore {
			htmlName := attrVal(n.Attr, "name")
			found := false
			for _, vb := range tbcAttrs.bindings {
				if vb.Variable.Name == htmlName {
					found = true
					break
				}
			}
			if !found && htmlName != "" {
				var t data.VariableType
				switch inputType := attrVal(n.Attr, "type"); inputType {
				case "number", "range":
					if strings.ContainsRune(attrVal(n.Attr, "min"), '.') ||
						strings.ContainsRune(attrVal(n.Attr, "max"), '.') ||
						strings.ContainsRune(attrVal(n.Attr, "step"), '.') {
						return false, nil, errors.New("non-integer " + inputType + " inputs not supported")
					}
					t = data.IntVar
				case "", "text":
					t = data.StringVar
				case "submit", "reset":
					break
				default:
					return false, nil, errors.New("input type not supported: " + inputType)
				}

				tbcAttrs.bindings = append(tbcAttrs.bindings, data.VariableMapping{
					Value:    data.BoundValue{Kind: data.BoundProperty, ID: "value"},
					Variable: data.Variable{Type: t, Name: htmlName}})
			}
		}
		processBindings(seh.c, path, tbcAttrs.bindings)
	} else {
		if err := mapCaptures(seh.c, n, *seh.indexList, tbcAttrs.capture); err != nil {
			return false, nil, errors.New(": " + err.Error())
		}
		processBindings(seh.c, append([]int(nil), *seh.indexList...), tbcAttrs.bindings)
		descend = true
	}
	return
}

func processComponents(syms *data.Symbols, n *html.Node, counter *int) (err error) {
	p := syms.Packages[syms.CurPkg]
	p.Components = make(map[string]*data.Component)
	w := walker{text: whitespaceOnly{}, component: &componentProcessor{syms, counter}}
	_, _, err = w.walkChildren(n, &siblings{n.FirstChild})
	return
}
