package attributes

import (
	"errors"
	"fmt"
	"strings"

	"github.com/flyx/askew/data"
	"github.com/flyx/askew/parsers"
	"github.com/flyx/net/html"
)

func invalidAttribute(name string) error {
	return errors.New(": element does not allow attribute `a:" + name + "`")
}

// Collector collects askew attributes (attributes with `a:` prefix).
type Collector interface {
	collect(name string, val string) error
}

// Component lists the attributes of a component
type Component struct {
	Name                       string
	Params                     []data.ComponentParam
	Init, OnInclude, OnExclude bool
}

func (t *Component) collect(name, val string) error {
	switch name {
	case "name":
		t.Name = val
		return nil
	case "params":
		var err error
		t.Params, err = parsers.ParseParameters(val)
		return err
	case "init":
		t.Init = true
		return nil
	case "oninclude":
		t.OnInclude = true
		return nil
	case "onexclude":
		t.OnExclude = true
		return nil
	}
	return invalidAttribute(name)
}

// Site lists the attributes of a site
type Site struct {
	JSFile, HTMLFile string
}

func (s *Site) collect(name, val string) error {
	switch name {
	case "a:htmlfile":
		s.HTMLFile = val
		return ErrRemoveAttribute
	case "a:jsfile":
		s.JSFile = val
		return ErrRemoveAttribute
	default:
		if strings.HasPrefix(name, "a:") {
			return fmt.Errorf(": unknown askew attribute on <a:site>: %s", name)
		}
	}
	return nil
}

// IncludeChild collects the attributes on any node that is a child of
// <a:include>.
type IncludeChild struct {
	Slot   string
	Others map[string]string
}

func (i *IncludeChild) collect(name, val string) error {
	if name == "slot" {
		i.Slot = val
	} else {
		if i.Others == nil {
			i.Others = make(map[string]string)
		}
		i.Others[name] = val
	}
	return nil
}

// Embed collects the attributes of <a:embed>.
type Embed struct {
	List, Optional bool
	T, Name        string
	Args           data.Arguments
	Control        bool
}

func (e *Embed) collect(name, val string) error {
	switch name {
	case "list":
		e.List = true
		return nil
	case "optional":
		e.Optional = true
		return nil
	case "type":
		e.T = val
		return nil
	case "name":
		e.Name = val
		return nil
	case "args":
		var err error
		e.Args, err = parsers.AnalyseArguments(val)
		return err
	case "control":
		e.Control = true
		return nil
	}
	return invalidAttribute(name)
}

// General collects attributes that may occur on any element.
type General struct {
	Bindings []data.VariableMapping
	Capture  []data.UnboundEventMapping
	If, For  *data.ControlBlock
	Assign   []data.Assignment
}

func (g *General) collect(name, val string) error {
	switch name {
	case "bindings":
		var err error
		g.Bindings, err = parsers.ParseBindings(val)
		if err != nil {
			return errors.New(": invalid bindings: " + err.Error())
		}
		for _, binding := range g.Bindings {
			if binding.Value.Kind == data.BoundEventValue {
				return errors.New(": cannot use event() in bindings")
			}
		}
	case "capture":
		var err error
		g.Capture, err = parsers.ParseCapture(val)
		if err != nil {
			return errors.New(": invalid capture: " + err.Error())
		}
	case "if":
		g.If = &data.ControlBlock{Kind: data.IfBlock, Expression: val}
	case "for":
		var err error
		g.For, err = parsers.ParseFor(val)
		if err != nil {
			return errors.New(": invalid for: " + err.Error())
		}
	case "assign":
		var err error
		g.Assign, err = parsers.ParseAssignments(val)
		if err != nil {
			return errors.New(": invalid assign: " + err.Error())
		}
	default:
		return invalidAttribute(name)
	}
	return nil
}

// ExtractAskewAttribs removes all askew attributes from the node and hands them
// to the collector.
func ExtractAskewAttribs(n *html.Node, target Collector) error {
	seen := make(map[string]struct{})

	i := 0
	for i < len(n.Attr) {
		attr := n.Attr[i]
		if len(attr.Key) < 2 || attr.Key[0:2] != "a:" {
			i++
			continue
		}

		// erase attribute from token (won't be written out)
		copy(n.Attr[i:], n.Attr[i+1:])
		n.Attr = n.Attr[:len(n.Attr)-1]

		key := attr.Key[2:]

		if _, ok := seen[key]; ok {
			panic("duplicate attribute: " + attr.Key)
		}
		seen[key] = struct{}{}
		if err := target.collect(key, attr.Val); err != nil {
			return err
		}
	}
	return nil
}

// ErrRemoveAttribute is an error that can be returned from a collector to
// remove the given attribute from the list of attributes.
var ErrRemoveAttribute = errors.New("remove attribute")

// Collect collects all attributes from the node and hands them to the
// collector. The collector can return ErrRemoveAttribute if it wants the
// attribute to be removed from the node.
func Collect(n *html.Node, target Collector) error {
	seen := make(map[string]struct{})

	for i := 0; i < len(n.Attr); {
		attr := &n.Attr[i]
		if _, ok := seen[attr.Key]; ok {
			panic("duplicate attribute: " + attr.Key)
		}
		seen[attr.Key] = struct{}{}
		if err := target.collect(attr.Key, attr.Val); err != nil {
			if err == ErrRemoveAttribute {
				n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
			} else {
				return err
			}
		} else {
			i++
		}
	}
	return nil
}

// Val retrieves the value of the attribute with the given name, or the empty
// string if no such attribute exists.
func Val(a []html.Attribute, name string) string {
	for i := range a {
		if a[i].Key == name {
			return a[i].Val
		}
	}
	return ""
}

// Exists checks whether an attribute with the given name exists in `a`
func Exists(a []html.Attribute, name string) bool {
	for i := range a {
		if a[i].Key == name {
			return true
		}
	}
	return false
}
