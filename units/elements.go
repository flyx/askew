package units

import (
	"errors"
	"strings"

	"github.com/flyx/askew/attributes"
	"github.com/flyx/askew/data"
	"github.com/flyx/askew/walker"
	"github.com/flyx/net/html"
	"github.com/flyx/net/html/atom"
)

type formValue struct {
	t     *data.ParamType
	radio bool
}

type formValueDiscovery struct {
	values map[string]formValue
}

func (d *formValueDiscovery) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	var v formValue
	name := attributes.Val(n.Attr, "name")
	if name == "" {
		return true, nil, nil
	}
	switch n.DataAtom {
	case atom.Input:
		switch inputType := attributes.Val(n.Attr, "type"); inputType {
		case "radio":
			v.radio = true
			v.t = &data.ParamType{Kind: data.StringType}
		case "number", "range":
			if strings.ContainsRune(attributes.Val(n.Attr, "min"), '.') ||
				strings.ContainsRune(attributes.Val(n.Attr, "max"), '.') ||
				strings.ContainsRune(attributes.Val(n.Attr, "step"), '.') {
				return false, nil, errors.New(": non-integer " + inputType + " inputs not supported")
			}
			v.t = &data.ParamType{Kind: data.IntType}
		case "text", "":
			v.t = &data.ParamType{Kind: data.StringType}
		case "submit", "reset", "hidden":
			return false, nil, nil
		default:
			return false, nil, errors.New(": unsupported input type: `" + inputType + "`")
		}
	case atom.Select, atom.Textarea:
		v.t, v.radio = &data.ParamType{Kind: data.StringType}, false
	default:
		return true, nil, nil
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
	w := walker.Walker{TextNode: walker.Allow{}, Embed: walker.DontDescend{},
		Handlers: walker.DontDescend{},
		Text:     walker.Allow{}, StdElements: &fvd}
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

type elementHandler struct {
	stdElementHandler
	cmp *data.Component
}

func (eh *elementHandler) mapCaptures(n *html.Node, v []data.UnboundEventMapping) error {
	if len(v) == 0 {
		return nil
	}
	formDepth := -1
	if eh.curFormPos != -1 {
		formDepth = len(*eh.indexList) - eh.curFormPos
	}
	ret := make([]data.EventMapping, 0, len(v))
	for _, unmapped := range v {
		fromController := false
		var h data.Handler
		ok := false
		if eh.cmp.Handlers != nil {
			h, ok = eh.cmp.Handlers[unmapped.Handler]
		}
		if !ok {
			if eh.cmp.Controller != nil {
				var c data.ControllerMethod
				c, ok = eh.cmp.Controller[unmapped.Handler]
				if ok {
					if c.CaptureError != nil {
						return errors.New("cannot use method " + unmapped.Handler + " for capturing: " + c.CaptureError.Error())
					}
					h = c.Handler
					fromController = true
				}
			}
			if !ok {
				return errors.New("capture references unknown handler: " + unmapped.Handler)
			}
		}
		notMapped := make(map[string]struct{})
		for pName := range unmapped.ParamMappings {
			notMapped[pName] = struct{}{}
		}
		mapped := make([]data.BoundParam, 0, len(h.Params))
		for _, p := range h.Params {
			bVal, ok := unmapped.ParamMappings[p.Name]
			if !ok {
				mapped = append(mapped, data.BoundParam{
					Param: p.Name, Value: data.BoundValue{
						Kind: data.BoundDataset, IDs: []string{p.Name}}})
			} else {
				delete(notMapped, p.Name)
				if bVal.Kind == data.BoundFormValue {
					if formDepth == -1 {
						return errors.New(": illegal form() binding outside of <form> element")
					}
					bVal.FormDepth = formDepth
					_, ok := eh.curForm[bVal.ID()]
					if !ok {
						return errors.New(": unknown form value name: `" + bVal.ID() + "`")
					}
				}
				mapped = append(mapped, data.BoundParam{Param: p.Name, Value: bVal})
			}
		}
		for unknown := range notMapped {
			return errors.New("unknown param for capture mapping: " + unknown)
		}
		handling := unmapped.Handling
		if handling == data.AutoPreventDefault {
			if h.Returns != nil && h.Returns.Kind == data.BoolType {
				handling = data.AskPreventDefault
			} else {
				handling = data.DontPreventDefault
			}
		}
		ret = append(ret, data.EventMapping{
			Event: unmapped.Event, Handler: unmapped.Handler, ParamMappings: mapped,
			Handling: handling, FromController: fromController})
	}

	eh.cmp.Captures = append(eh.cmp.Captures, data.Capture{
		Path: append([]int(nil), *eh.indexList...), Mappings: ret})
	return nil
}

func (eh *elementHandler) processBindings(arr []data.VariableMapping) error {
	formDepth := -1
	if eh.curFormPos != -1 {
		formDepth = len(*eh.indexList) - eh.curFormPos
	}
	path := append([]int(nil), *eh.indexList...)

	for _, vb := range arr {
		if vb.Value.Kind == data.BoundFormValue {
			if formDepth == -1 {
				return errors.New(": illegal form() binding outside of <form> element")
			}
			vb.Value.FormDepth = formDepth
			val, ok := eh.curForm[vb.Value.ID()]
			if !ok {
				return errors.New(": unknown form value name: `" + vb.Value.ID() + "`")
			}
			if vb.Variable.Type == nil {
				vb.Variable.Type = val.t
			}
		} else {
			if vb.Variable.Type == nil {
				switch vb.Value.Kind {
				case data.BoundClass:
					if len(vb.Value.IDs) > 1 {
						vb.Variable.Type = &data.ParamType{Kind: data.IntType}
					} else {
						vb.Variable.Type = &data.ParamType{Kind: data.BoolType}
					}
				case data.BoundSelf:
					vb.Variable.Type = &data.ParamType{Kind: data.JSValueType}
				default:
					vb.Variable.Type = &data.ParamType{Kind: data.StringType}
				}
			}
		}
		vb.Path = path
		eh.cmp.Variables = append(eh.cmp.Variables, vb)
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
			_, ok := seh.curForm[a.Target.IDs[0]]
			if !ok {
				return errors.New(": unknown form value name: `" + a.Target.ID() + "`")
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
			TextNode: walker.Allow{}, Text: &aTextProcessor{&block.Block, &indexList},
			Embed:       &embedProcessor{seh.syms, &indexList, block.Path},
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

func (eh *elementHandler) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	if err = eh.updateCurForm(n); err != nil {
		return
	}

	var attrs attributes.General
	if err = attributes.ExtractAskewAttribs(n, &attrs); err != nil {
		return
	}
	descend, err = eh.handleControlBlocksAndAssignments(n, attrs)
	if descend {
		if err := eh.mapCaptures(n, attrs.Capture); err != nil {
			return false, nil, errors.New(": " + err.Error())
		}

		if err = eh.processBindings(attrs.Bindings); err != nil {
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
