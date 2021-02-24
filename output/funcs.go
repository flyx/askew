package output

import (
	"strconv"
	"strings"

	"github.com/flyx/askew/data"
	"github.com/flyx/net/html"
)

func nameForBound(b data.BoundKind) string {
	switch b {
	case data.BoundDataset:
		return "BoundDataset"
	case data.BoundProperty:
		return "BoundProperty"
	case data.BoundStyle:
		return "BoundStlye"
	case data.BoundClass:
		return "BoundClass"
	case data.BoundFormValue:
		return "BoundFormValue"
	case data.BoundEventValue:
		return "BoundEventValue"
	default:
		panic("unknown boundKind")
	}
}

func pathItems(path []int, exclude int) string {
	b := strings.Builder{}
	for i := 0; i < len(path)-exclude; i++ {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.FormatInt(int64(path[i]), 10))
	}
	return b.String()
}

func last(path []int) int {
	return path[len(path)-1]
}

func wrapperForType(t data.ParamType) string {
	switch t.Kind {
	case data.StringType:
		return "askew.StringValue"
	case data.IntType:
		return "askew.IntValue"
	case data.BoolType:
		return "askew.BoolValue"
	case data.JSValueType:
		return "askew.RawValue"
	}
	panic("no wrapper for type: " + t.String())
}

func fieldType(e data.Embed) string {
	if e.T == "" {
		switch e.Kind {
		case data.OptionalEmbed:
			return "askew.GenericOptional"
		case data.ListEmbed:
			return "askew.GenericList"
		default:
			panic("unexpected field type")
		}
	}
	var b strings.Builder
	if e.Ns != "" {
		b.WriteString(e.Ns)
		b.WriteRune('.')
	}
	if e.Kind == data.OptionalEmbed {
		b.WriteString("Optional")
	}
	b.WriteString(e.T)
	if e.Kind == data.ListEmbed {
		b.WriteString("List")
	}
	return b.String()
}

func renderTemplateHTML(n *html.Node) string {
	var w strings.Builder
	html.Render(&w, n)
	var ret strings.Builder
	for _, r := range w.String() {
		if r == '`' {
			ret.WriteString("` + \"`\" + `")
		} else {
			ret.WriteRune(r)
		}
	}
	return ret.String()
}
