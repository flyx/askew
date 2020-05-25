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

type template struct {
	// HTML id. internally generated.
	id string
	// maps CSS selector to object description.
	// used to ensure selectors are unique.
	objects      []dynamicObject
	embeds       []embed
	strippedHTML *html.Node
	needsList    bool
}

// maps name to template. For components, the string key is its Go name; for
// macros, the string key is the internal name. In both cases, the name can be
// used to tbc:include the template.
type templateSet map[string]*template

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

func (t *template) process(set templateSet, n *html.Node, indexList []int) {
	if n.Type != html.ElementNode {
		return
	}
	var tbcAttrs generalAttribs
	extractTbcAttribs(n, &tbcAttrs)
	switch n.DataAtom {
	case 0:
		if n.Data == "tbc:embed" {
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
		} else {
			panic("unknown element: <" + n.Data + ">")
		}
	case atom.Input:
		if tbcAttrs.classSwitch != "" {
			panic("tbc:classSwitch not allowed on <input>")
		}
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
		case "", "text", "submit":
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
