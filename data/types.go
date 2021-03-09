package data

import "strings"

// TypeKind is the kind of a type.
type TypeKind int

const (
	// IntType is an int
	IntType TypeKind = iota
	// StringType is a string
	StringType
	// BoolType is a bool
	BoolType
	// JSValueType is js.Value
	JSValueType
	// NamedType is any named type that is not an int, a string or a bool.
	NamedType
	// ArrayType is an array
	ArrayType
	// MapType is a map
	MapType
	// ChanType is a chan
	ChanType
	// FuncType is a func
	FuncType
	// PointerType is a pointer
	PointerType
)

// ParamType is the type of a handler or controller method parameter.
type ParamType struct {
	Kind TypeKind
	// used when Kind == NamedType
	Name string
	// used when Kind == MapType
	KeyType *ParamType
	// used when Kind in [ArrayType, MapType, PointerType, FuncType (return type)]
	ValueType *ParamType
	// used when Kind == FuncType
	Params []Param
}

func (pt ParamType) String() string {
	switch pt.Kind {
	case IntType:
		return "int"
	case StringType:
		return "string"
	case BoolType:
		return "bool"
	case JSValueType:
		return "js.Value"
	case NamedType:
		return pt.Name
	case ArrayType:
		return "[]" + pt.ValueType.String()
	case MapType:
		return "map[" + pt.KeyType.String() + "]" + pt.ValueType.String()
	case ChanType:
		return "chan " + pt.ValueType.String()
	case FuncType:
		var sb strings.Builder
		sb.WriteString("func(")
		for _, p := range pt.Params {
			sb.WriteString(p.Name)
			sb.WriteByte(' ')
			sb.WriteString(p.Type.String())
			sb.WriteByte(',')
		}
		sb.WriteString(") ")
		if pt.ValueType != nil {
			sb.WriteString(pt.ValueType.String())
		}
		return sb.String()
	case PointerType:
		return "*" + pt.ValueType.String()
	default:
		panic("unexpected type kind")
	}
}

// Param is a parameter of a handler or controller method.
type Param struct {
	Name string
	Type *ParamType
}

func (p *Param) String() string {
	return p.Name + " " + p.Type.String()
}

// Field is a declared field of a component.
type Field struct {
	Name         string
	Type         *ParamType
	DefaultValue *string
}
