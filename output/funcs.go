package output

import (
	"strconv"
	"strings"

	"github.com/flyx/askew/data"
)

func nameForType(k data.VariableType) string {
	switch k {
	case data.StringVar:
		return "string"
	case data.IntVar:
		return "int"
	case data.BoolVar:
		return "bool"
	default:
		panic("unsupported type")
	}
}

func nameForBound(b data.BoundKind) string {
	switch b {
	case data.BoundData:
		return "BoundData"
	case data.BoundProperty:
		return "BoundProperty"
	case data.BoundClass:
		return "BoundClass"
	default:
		panic("unknown boundKind")
	}
}

func pathItems(path []int) string {
	b := strings.Builder{}
	for i := range path {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.FormatInt(int64(path[i]), 10))
	}
	return b.String()
}

func parentPath(path []int) string {
	return pathItems(path[:len(path)-1])
}

func last(path []int) int {
	return path[len(path)-1]
}

func wrapperForType(k data.VariableType) string {
	switch k {
	case data.StringVar:
		return "StringValue"
	case data.IntVar:
		return "IntValue"
	case data.BoolVar:
		return "BoolValue"
	default:
		panic("unsupported type")
	}
}
