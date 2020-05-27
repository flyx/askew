package main

import (
	"reflect"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type dynamicObjectKind int

const (
	textContent dynamicObjectKind = iota
	inputValue
	classSwitch
)

type dynamicObject struct {
	kind      dynamicObjectKind
	path      []int
	goType    reflect.Kind
	goName    string
	className string
}

type embed struct {
	path              []int
	fieldName, goName string
	list              bool
}

type handler struct {
	params map[string]reflect.Kind
}

type captureSource struct {
	path     []int
	captures []capture
}

type component struct {
	// HTML id. internally generated.
	id string
	// maps CSS selector to object description.
	// used to ensure selectors are unique.
	objects         []dynamicObject
	embeds          []embed
	handlers        map[string]handler
	captureSources  []captureSource
	processedHTML   *html.Node
	needsController bool
	needsList       bool
}

// maps name to template. For components, the string key is its Go name; for
// macros, the string key is the internal name. In both cases, the name can be
// used to tbc:include the template.
type componentSet map[string]*component

func attrVal(a []html.Attribute, name string) string {
	for i := range a {
		if a[i].Key == name {
			return a[i].Val
		}
	}
	return ""
}

func attrExists(a []html.Attribute, name string) bool {
	for i := range a {
		if a[i].Key == name {
			return true
		}
	}
	return false
}

var isValidIdentifier = regexp.MustCompile(`^[\pL_][\pL0-9]*$`).MatchString

func parseHandler(input string) (name string, h handler) {
	tmp := strings.Split(input, "(")
	if len(tmp) != 2 || tmp[1][len(tmp[1])-1] != ')' {
		panic("invalid handler: " + input)
	}
	name = tmp[0]
	if len(tmp[1]) > 1 {
		h.params = make(map[string]reflect.Kind)
		params := strings.Split(tmp[1][:len(tmp[1])-1], ",")
		for i := range params {
			param := strings.TrimSpace(params[i])
			tmp = strings.Split(param, " ")
			if len(tmp) != 2 {
				panic("invalid parameter def in handler: " + param)
			}
			pName := strings.TrimSpace(tmp[0])
			if !isValidIdentifier(pName) {
				panic("invalid parameter name: " + pName)
			}
			_, ok := h.params[pName]
			if ok {
				panic("duplicate parameter name: " + pName)
			}
			switch t := strings.TrimSpace(tmp[1]); t {
			case "string":
				h.params[pName] = reflect.String
			case "int":
				h.params[pName] = reflect.Int
			default:
				panic("unsupported parameter type: " + t)
			}
		}
	}
	return
}

func (t *component) mapCaptures(n *html.Node, path []int, v []capture) {
	if len(v) == 0 {
		return
	}
	for i := range v {
		item := v[i]
		h, ok := t.handlers[item.handler]
		if !ok {
			panic("capture references unknown handler: " + item.handler)
		}
		for pName := range item.paramMappings {
			_, ok = h.params[pName]
			if !ok {
				panic("unknown param for capture mapping: " + pName)
			}
		}
		for pName := range h.params {
			_, ok = item.paramMappings[pName]
			if !ok {
				item.paramMappings[pName] = paramSupplier{kind: attrSupplier, id: "data-" + pName}
			}
		}
	}
	t.captureSources = append(t.captureSources, captureSource{
		path: append([]int(nil), path...), captures: v})
}

func (t *component) process(set componentSet, n *html.Node, indexList []int) {
	if n.Type != html.ElementNode {
		return
	}
	var tbcAttrs generalAttribs
	extractTbcAttribs(n, &tbcAttrs)
	switch n.DataAtom {
	case 0:
		switch n.Data {
		case "tbc:embed":
			targetType := attrVal(n.Attr, "type")
			if len(targetType) == 0 {
				panic("tbc:embed misses `type` attribute")
			}
			tmpl, ok := set[targetType]
			if !ok {
				panic("tbc:embed references unknown type `" + targetType + "`")
			}
			e := embed{path: append([]int(nil), indexList...),
				list: attrExists(n.Attr, "list")}
			if e.list {
				tmpl.needsList = true
			}
			e.fieldName = attrVal(n.Attr, "name")
			if len(e.fieldName) == 0 {
				panic("tbc:embed must give a `name` attribute!")
			}
			e.goName = targetType
			if n.FirstChild != nil {
				panic("tbc:embed may not have content")
			}
			t.embeds = append(t.embeds, e)
			n.Type = html.CommentNode
			n.Data = "embed(" + e.fieldName + "=" + e.goName + ")"
			n.Attr = nil
		case "tbc:handler":
			if len(indexList) != 1 {
				panic("tbc:handler must be defined as direct child of <template>")
			}
			def := n.FirstChild
			if def.Type != html.TextNode || def.NextSibling != nil {
				panic("tbc:handler must have plain text as content and nothing else")
			}
			name, h := parseHandler(def.Data)
			if t.handlers == nil {
				t.handlers = make(map[string]handler)
			} else {
				_, ok := t.handlers[name]
				if ok {
					panic("duplicate handler name: " + name)
				}
			}
			t.handlers[name] = h
			n.Type = html.CommentNode
			n.Data = "handler: " + def.Data
			n.Attr = nil
		default:
			panic("unknown element: <" + n.Data + ">")
		}
	case atom.Input:
		if tbcAttrs.classSwitch != "" {
			panic("tbc:classSwitch not allowed on <input>")
		}
		t.mapCaptures(n, indexList, tbcAttrs.captures)
		if tbcAttrs.interactive == inactive {
			return
		}
		var goType reflect.Kind
		switch inputType := attrVal(n.Attr, "type"); inputType {
		case "number", "range":
			if strings.ContainsRune(attrVal(n.Attr, "min"), '.') ||
				strings.ContainsRune(attrVal(n.Attr, "max"), '.') ||
				strings.ContainsRune(attrVal(n.Attr, "step"), '.') {
				panic("non-integer " + inputType + " inputs not supported")
			}
			goType = reflect.Int
		case "", "text":
			goType = reflect.String
		case "submit", "reset":
			if tbcAttrs.interactive != forceActive {
				return
			}
			goType = reflect.String
		default:
			panic("input type not supported: " + inputType)
		}
		htmlName := attrVal(n.Attr, "name")
		if len(htmlName) == 0 {
			panic("<input> misses a name!")
		}
		goName := htmlName
		if len(tbcAttrs.name) > 0 {
			goName = tbcAttrs.name
		}
		if !isValidIdentifier(goName) {
			panic("not a valid identifier: " + goName)
		}

		t.objects = append(t.objects, dynamicObject{
			kind: inputValue, path: append([]int(nil), indexList...),
			goType: goType, goName: goName})
	default:
		t.mapCaptures(n, indexList, tbcAttrs.captures)
		if tbcAttrs.interactive != forceActive {
			if tbcAttrs.classSwitch != "" {
				if tbcAttrs.name == "" {
					panic("tbc:classSwitch requires tbc:name!")
				}
				t.objects = append(t.objects, dynamicObject{
					kind: classSwitch, path: append([]int(nil), indexList...),
					goType: reflect.Bool, goName: tbcAttrs.name,
					className: tbcAttrs.classSwitch})
			}
			indexList = append(indexList, 0)
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				t.process(set, c, indexList)
				indexList[len(indexList)-1]++
			}
			return
		}
		if n.FirstChild != nil && (n.FirstChild.Type != html.TextNode ||
			n.FirstChild.NextSibling != nil) {
			panic("tbc:dynamic on a node with child nodes")
		}
		if len(tbcAttrs.name) == 0 {
			panic("tbc:dynamic on a node without tbc:name")
		}
		if !isValidIdentifier(tbcAttrs.name) {
			panic("not a valid identifier: " + tbcAttrs.name)
		}
		t.objects = append(t.objects, dynamicObject{
			kind: textContent, path: append([]int(nil), indexList...),
			goType: reflect.String, goName: tbcAttrs.name})
	}
}
