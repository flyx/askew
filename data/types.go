package data

// TypeKind is the kind of a type.
type TypeKind int

const (
	// IntType is an int
	IntType TypeKind = iota
	// StringType is a string
	StringType
	// BoolType is a bool
	BoolType
	// ObjectType is js.Object
	ObjectType
	// NamedType is any named type that is not an int, a string or a bool.
	NamedType
	// ArrayType is an array
	ArrayType
	// MapType is a map
	MapType
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
	// used when Kind in [ArrayType, MapType, PointerType]
	ValueType *ParamType
}

func (pt ParamType) String() string {
	switch pt.Kind {
	case IntType:
		return "int"
	case StringType:
		return "string"
	case BoolType:
		return "bool"
	case ObjectType:
		return "js.Object"
	case NamedType:
		return pt.Name
	case ArrayType:
		return "[]" + pt.ValueType.String()
	case MapType:
		return "map[" + pt.KeyType.String() + "]" + pt.ValueType.String()
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
