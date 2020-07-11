package components

import (
	"errors"
	"fmt"
	"strings"

	"github.com/flyx/askew/attributes"

	"github.com/flyx/askew/parsers"
	"github.com/flyx/askew/walker"

	"github.com/flyx/askew/data"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Processor reads all components.
type Processor struct {
	syms    *data.Symbols
	counter *int
}

// NewProcessor creates a new processor.
func NewProcessor(syms *data.Symbols, counter *int) *Processor {
	return &Processor{syms: syms, counter: counter}
}

// Process reads the given component element.
func (p *Processor) Process(n *html.Node) (descend bool,
	replacement *html.Node, err error) {
	var cmpAttrs attributes.Component
	err = attributes.Collect(n, &cmpAttrs)
	if err != nil {
		return
	}
	if len(cmpAttrs.Name) == 0 {
		return false, nil, errors.New(": attribute `name` missing")
	}

	replacement = &html.Node{Type: html.ElementNode, DataAtom: atom.Template,
		Data: "template"}
	cmp := &data.Component{EmbedHost: data.EmbedHost{},
		Name: cmpAttrs.Name, Template: replacement, NeedsController: cmpAttrs.Controller,
		Parameters: cmpAttrs.Params}
	p.syms.CurHost = &cmp.EmbedHost
	(*p.counter)++
	cmp.ID = fmt.Sprintf("askew-component-%d-%s", *p.counter, strings.ToLower(cmpAttrs.Name))
	replacement.Attr = []html.Attribute{html.Attribute{Key: "id", Val: cmp.ID}}

	var indexList []int
	w := walker.Walker{
		Text: walker.Allow{}, AText: &aTextProcessor{&cmp.Block, &indexList},
		Embed:       &EmbedProcessor{p.syms, &indexList},
		Handler:     &handlerProcessor{p.syms, &cmp.Handlers, &indexList},
		StdElements: &componentElementHandler{stdElementHandler{p.syms, &indexList, &cmp.Block, -1, nil}, cmp},
		IndexList:   &indexList}
	replacement.FirstChild, replacement.LastChild, err = w.WalkChildren(
		replacement, &walker.Siblings{Cur: n.FirstChild})

	{
		// reverse Embed list so that they get embedded in reverse order.
		// this is necessary because embedding may change the number of elements in
		// a node, rendering the path of following embeds invalid.
		tmp := make([]data.Embed, len(cmp.Embeds))
		for i, e := range cmp.Embeds {
			tmp[len(tmp)-i-1] = e
		}
		cmp.Embeds = tmp
	}

	{
		// reverse contained control blocks so that they are processed back to front,
		// ensuring that their paths are correct.
		tmp := make([]*data.ControlBlock, len(cmp.Controlled))
		for i, e := range cmp.Controlled {
			tmp[len(tmp)-i-1] = e
		}
		cmp.Controlled = tmp
	}

	if p.syms.CurFile.Components == nil {
		p.syms.CurFile.Components = make(map[string]*data.Component)
	}
	p.syms.CurFile.Components[cmpAttrs.Name] = cmp

	return
}

// EmbedProcessor processes <a:embed> elements.
type EmbedProcessor struct {
	syms      *data.Symbols
	indexList *[]int
}

// NewEmbedProcessor creates a new EmbedProcessor
func NewEmbedProcessor(syms *data.Symbols, indexList *[]int) *EmbedProcessor {
	return &EmbedProcessor{syms: syms, indexList: indexList}
}

func resolveEmbed(n *html.Node, syms *data.Symbols, indexList []int) (data.Embed, error) {
	if n.FirstChild != nil {
		return data.Embed{}, errors.New(": illegal content")
	}
	var attrs attributes.Embed
	if err := attributes.Collect(n, &attrs); err != nil {
		return data.Embed{}, err
	}
	e := data.Embed{Kind: data.DirectEmbed, Path: append([]int(nil), indexList...),
		Field: attrs.Name}
	if e.Field == "" {
		return data.Embed{}, errors.New(": attribute `name` missing")
	}
	if attrs.List {
		e.Kind = data.ListEmbed
	}
	if attrs.Optional {
		if e.Kind != data.DirectEmbed {
			return data.Embed{}, errors.New(": cannot mix `list` and `optional`")
		}
		e.Kind = data.OptionalEmbed
	}
	if attrs.T == "" {
		if e.Kind == data.DirectEmbed {
			return data.Embed{}, errors.New(": attribute `type` missing (may only be omitted for optional or list embeds)")
		}
		if attrs.Args.Count != 0 {
			return data.Embed{}, errors.New(": embed with `list` or `optional` cannot have `args`")
		}
	} else {
		target, typeName, aliasName, err := syms.ResolveComponent(attrs.T)
		if err != nil {
			return data.Embed{}, errors.New(": attribute `type` invalid: " + err.Error())
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
				return data.Embed{}, errors.New(": embed with `list` or `optional` cannot have `args`")
			}
		} else {
			e.Args = attrs.Args
			if len(target.Parameters) != e.Args.Count {
				return data.Embed{}, fmt.Errorf(
					": target component requires %d arguments, but %d were given", len(target.Parameters), e.Args.Count)
			}
		}
	}
	return e, nil
}

// Process implements Walker.NodeHandler.
func (ep *EmbedProcessor) Process(n *html.Node) (descend bool,
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

func (hp *handlerProcessor) Process(n *html.Node) (descend bool,
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

func (d *formValueDiscovery) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	var v formValue
	name := attributes.Val(n.Attr, "name")
	if name == "" {
		return false, nil, nil
	}
	switch n.DataAtom {
	case atom.Input:
		switch inputType := attributes.Val(n.Attr, "type"); inputType {
		case "radio":
			v.radio = true
			v.t = data.StringVar
		case "number", "range":
			if strings.ContainsRune(attributes.Val(n.Attr, "min"), '.') ||
				strings.ContainsRune(attributes.Val(n.Attr, "max"), '.') ||
				strings.ContainsRune(attributes.Val(n.Attr, "step"), '.') {
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
	w := walker.Walker{Text: walker.Allow{}, Embed: walker.DontDescend{},
		Handler: walker.DontDescend{},
		AText:   walker.Allow{}, StdElements: &fvd}
	var err error
	form.FirstChild, form.LastChild, err = w.WalkChildren(form, &walker.Siblings{Cur: form.FirstChild})
	if err != nil {
		return nil, err
	}
	return fvd.values, nil
}

type stdElementHandler struct {
	syms       *data.Symbols
	indexList  *[]int
	b          *data.Block
	curFormPos int
	curForm    map[string]formValue
}

type componentElementHandler struct {
	stdElementHandler
	c *data.Component
}

func (ceh *componentElementHandler) mapCaptures(n *html.Node, v []data.UnboundEventMapping) error {
	if len(v) == 0 {
		return nil
	}
	formDepth := -1
	if ceh.curFormPos != -1 {
		formDepth = len(*ceh.indexList) - ceh.curFormPos
	}
	ret := make([]data.EventMapping, 0, len(v))
	for _, unmapped := range v {
		h, ok := ceh.c.Handlers[unmapped.Handler]
		if !ok {
			return errors.New("capture references unknown handler: " + unmapped.Handler)
		}
		notMapped := make(map[string]struct{})
		for pName := range unmapped.ParamMappings {
			notMapped[pName] = struct{}{}
		}
		mapped := make([]data.BoundParam, 0, len(h.Params))
		for _, p := range h.Params {
			bVal, ok := unmapped.ParamMappings[p.Name]
			if !ok {
				mapped = append(mapped, data.BoundParam{Param: p.Name, Value: data.BoundValue{Kind: data.BoundData, ID: p.Name}})
			} else {
				delete(notMapped, p.Name)
				if bVal.Kind == data.BoundFormValue {
					if formDepth == -1 {
						return errors.New(": illegal form() binding outside of <form> element")
					}
					bVal.FormDepth = formDepth
					_, ok := ceh.curForm[bVal.ID]
					if !ok {
						return errors.New(": unknown form value name: `" + bVal.ID + "`")
					}
				}
				mapped = append(mapped, data.BoundParam{Param: p.Name, Value: bVal})
			}
		}
		for unknown := range notMapped {
			return errors.New("unknown param for capture mapping: " + unknown)
		}
		ret = append(ret, data.EventMapping{Event: unmapped.Event, Handler: unmapped.Handler, ParamMappings: mapped})
	}

	ceh.c.Captures = append(ceh.c.Captures, data.Capture{
		Path: append([]int(nil), *ceh.indexList...), Mappings: ret})
	return nil
}

func (ceh *componentElementHandler) processBindings(arr []data.VariableMapping) error {
	formDepth := -1
	if ceh.curFormPos != -1 {
		formDepth = len(*ceh.indexList) - ceh.curFormPos
	}
	path := append([]int(nil), *ceh.indexList...)

	for _, vb := range arr {
		if vb.Value.Kind == data.BoundFormValue {
			if formDepth == -1 {
				return errors.New(": illegal form() binding outside of <form> element")
			}
			vb.Value.FormDepth = formDepth
			val, ok := ceh.curForm[vb.Value.ID]
			if !ok {
				return errors.New(": unknown form value name: `" + vb.Value.ID + "`")
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
		vb.Path = path
		ceh.c.Variables = append(ceh.c.Variables, vb)
	}
	return nil
}

func (seh *stdElementHandler) processAssignments(arr []data.Assignment, path []int) error {
	formDepth := -1
	if seh.curFormPos != -1 {
		formDepth = len(*seh.indexList) - seh.curFormPos
	}

	for _, a := range arr {
		if a.Target.Kind == data.BoundFormValue {
			if formDepth == -1 {
				return errors.New(": illegal form() binding outside of <form> element")
			}
			a.Target.FormDepth = formDepth
			_, ok := seh.curForm[a.Target.ID]
			if !ok {
				return errors.New(": unknown form value name: `" + a.Target.ID + "`")
			}
		}
		a.Path = path
		seh.b.Assignments = append(seh.b.Assignments, a)
	}
	return nil
}

func (seh *stdElementHandler) formDepth() int {
	if seh.curFormPos == -1 {
		return -1
	}
	return len(*seh.indexList) - seh.curFormPos
}

func (seh *stdElementHandler) updateCurForm(n *html.Node) error {
	if len(*seh.indexList) <= seh.curFormPos {
		seh.curFormPos = -1
		seh.curForm = nil
	}
	if n.DataAtom == atom.Form {
		if seh.curFormPos != -1 {
			return errors.New(": nested <form> not allowed")
		}
		seh.curFormPos = len(*seh.indexList)
		vals, err := discoverFormValues(n)
		if err != nil {
			return err
		}
		seh.curForm = vals
	}
	return nil
}

func (seh *stdElementHandler) handleControlBlocksAndAssignments(n *html.Node, attrs attributes.General) (descend bool, err error) {
	var block *data.ControlBlock

	if attrs.If != nil {
		block = attrs.If
	}
	if attrs.For != nil {
		if block != nil {
			return false, errors.New(": cannot have a:if and a:for on same element")
		}

		block = attrs.For
	}
	if block != nil {
		block.Path = append([]int(nil), *seh.indexList...)
		var indexList []int
		cp := &ctrlBlockElementProcessor{stdElementHandler{seh.syms, &indexList, &block.Block, seh.curFormPos, seh.curForm}}
		cp.processAssignments(attrs.Assign, []int{})

		w := walker.Walker{
			Text: walker.Allow{}, AText: &aTextProcessor{&block.Block, &indexList},
			Embed:       &EmbedProcessor{seh.syms, &indexList},
			Handler:     nil,
			StdElements: cp,
			IndexList:   &indexList}
		n.FirstChild, n.LastChild, err = w.WalkChildren(n, &walker.Siblings{Cur: n.FirstChild})

		// reverse contained control blocks so that they are processed back to front,
		// ensuring that their paths are correct.
		tmp := make([]*data.ControlBlock, len(block.Controlled))
		for i, e := range block.Controlled {
			tmp[len(tmp)-i-1] = e
		}
		block.Controlled = tmp

		seh.b.Controlled = append(seh.b.Controlled, block)
		return false, nil
	}
	err = seh.processAssignments(attrs.Assign, append([]int(nil), *seh.indexList...))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (ceh *componentElementHandler) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	if err = ceh.updateCurForm(n); err != nil {
		return
	}

	var attrs attributes.General
	if err = attributes.ExtractAskewAttribs(n, &attrs); err != nil {
		return
	}
	descend, err = ceh.handleControlBlocksAndAssignments(n, attrs)
	if descend {
		if err := ceh.mapCaptures(n, attrs.Capture); err != nil {
			return false, nil, errors.New(": " + err.Error())
		}

		if err = ceh.processBindings(attrs.Bindings); err != nil {
			return false, nil, err
		}
	} else {
		if len(attrs.Capture) > 0 {
			return false, nil, errors.New(": cannot capture inside a:if or a:for")
		}
		if len(attrs.Bindings) > 0 {
			return false, nil, errors.New(": cannot bind inside a:if or a:for")
		}
	}

	return
}

type ctrlBlockElementProcessor struct {
	stdElementHandler
}

func (cbeh *ctrlBlockElementProcessor) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	if err = cbeh.updateCurForm(n); err != nil {
		return
	}

	var attrs attributes.General
	if err = attributes.ExtractAskewAttribs(n, &attrs); err != nil {
		return
	}

	if len(attrs.Capture) > 0 {
		return false, nil, errors.New(": cannot capture inside a:if or a:for")
	}
	if len(attrs.Bindings) > 0 {
		return false, nil, errors.New(": cannot bind inside a:if or a:for")
	}
	descend, err = cbeh.handleControlBlocksAndAssignments(n, attrs)
	return
}

type aTextProcessor struct {
	b         *data.Block
	indexList *[]int
}

func (atp *aTextProcessor) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	expr := attributes.Val(n.Attr, "expr")
	if expr == "" {
		return false, nil, errors.New(": missing attribute `expr`")
	}
	if n.FirstChild != nil {
		return false, nil, errors.New(": node may not have child nodes")
	}
	atp.b.Assignments = append(atp.b.Assignments, data.Assignment{
		Expression: expr, Path: append([]int(nil), *atp.indexList...), Target: data.BoundValue{Kind: data.BoundSelf}})
	return false, &html.Node{Type: html.CommentNode, Data: "a:text"}, nil
}
