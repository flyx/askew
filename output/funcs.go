package output

import (
	"strconv"
	"strings"

	"github.com/flyx/askew/data"
)

func nameForBound(b data.BoundKind) string {
	switch b {
	case data.BoundData:
		return "BoundData"
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
		return "runtime.StringValue"
	case data.IntType:
		return "runtime.IntValue"
	case data.BoolType:
		return "runtime.BoolValue"
	case data.JSValueType:
		return "runtime.RawValue"
	}
	panic("no wrapper for type: " + t.String())
}

func fieldType(e data.Embed) string {
	if e.T == "" {
		switch e.Kind {
		case data.OptionalEmbed:
			return "runtime.GenericOptional"
		case data.ListEmbed:
			return "runtime.GenericList"
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
