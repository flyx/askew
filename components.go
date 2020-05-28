package main

import (
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type accessor struct {
	target   boundValue
	path     []int
	variable targetVar
}

type embed struct {
	path              []int
	fieldName, goName string
	list              bool
}

type handler struct {
	params map[string]valueKind
}

type captureSource struct {
	path     []int
	captures []capture
}

type component struct {
	// HTML id. internally generated.
	id              string
	accessors       []accessor
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
		h.params = make(map[string]valueKind)
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
				h.params[pName] = stringVal
			case "int":
				h.params[pName] = intVal
			case "bool":
				h.params[pName] = boolVal
			default:
				panic("unsupported parameter type: " + t)
			}
		}
	}
	return
}

func (c *component) mapCaptures(n *html.Node, path []int, v []capture) {
	if len(v) == 0 {
		return
	}
	for i := range v {
		item := v[i]
		h, ok := c.handlers[item.handler]
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
				item.paramMappings[pName] = boundValue{kind: boundAttribute, id: "data-" + pName}
			}
		}
	}
	c.captureSources = append(c.captureSources, captureSource{
		path: append([]int(nil), path...), captures: v})
}

func (c *component) processBindings(path []int, arr []varBinding) {
	for _, vb := range arr {
		if vb.variable.kind == autoVal {
			if vb.value.kind == boundClass {
				vb.variable.kind = boolVal
			} else {
				vb.variable.kind = stringVal
			}
		}
		c.accessors = append(c.accessors, accessor{
			target: vb.value, path: path, variable: vb.variable})
	}
}

func (c *component) process(set componentSet, n *html.Node, indexList []int) {
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
			c.embeds = append(c.embeds, e)
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
			if c.handlers == nil {
				c.handlers = make(map[string]handler)
			} else {
				_, ok := c.handlers[name]
				if ok {
					panic("duplicate handler name: " + name)
				}
			}
			c.handlers[name] = h
			n.Type = html.CommentNode
			n.Data = "handler: " + def.Data
			n.Attr = nil
		default:
			panic("unknown element: <" + n.Data + ">")
		}
	case atom.Input:
		c.mapCaptures(n, indexList, tbcAttrs.captures)
		path := append([]int(nil), indexList...)
		if !tbcAttrs.ignore {
			htmlName := attrVal(n.Attr, "name")
			found := false
			for _, vb := range tbcAttrs.bindings {
				if vb.variable.name == htmlName {
					found = true
					break
				}
			}
			if !found && htmlName != "" {
				var kind valueKind
				switch inputType := attrVal(n.Attr, "type"); inputType {
				case "number", "range":
					if strings.ContainsRune(attrVal(n.Attr, "min"), '.') ||
						strings.ContainsRune(attrVal(n.Attr, "max"), '.') ||
						strings.ContainsRune(attrVal(n.Attr, "step"), '.') {
						panic("non-integer " + inputType + " inputs not supported")
					}
					kind = intVal
				case "", "text":
					kind = stringVal
				case "submit", "reset":
					break
				default:
					panic("input type not supported: " + inputType)
				}

				tbcAttrs.bindings = append(tbcAttrs.bindings, varBinding{
					value:    boundValue{kind: boundProperty, id: "value"},
					variable: targetVar{kind: kind, name: htmlName}})
			}
		}
		c.processBindings(path, tbcAttrs.bindings)
	default:
		c.mapCaptures(n, indexList, tbcAttrs.captures)
		c.processBindings(append([]int(nil), indexList...), tbcAttrs.bindings)
		indexList = append(indexList, 0)
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			c.process(set, child, indexList)
			indexList[len(indexList)-1]++
		}
	}
}
