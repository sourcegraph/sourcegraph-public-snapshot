package compiler

import "github.com/sourcegraph/go-jsonschema/jsonschema"

func goBuiltinType(typ jsonschema.PrimitiveType) string {
	switch typ {
	case jsonschema.NullType:
		return "nil"
	case jsonschema.BooleanType:
		return "bool"
	case jsonschema.NumberType:
		return "float64"
	case jsonschema.StringType:
		return "string"
	case jsonschema.IntegerType:
		return "int"
	default:
		return ""
	}
}

func isEmittedAsGoNamedType(schema *jsonschema.Schema) bool {
	return len(schema.Type) == 1 && schema.Type[0] == jsonschema.ObjectType
}
