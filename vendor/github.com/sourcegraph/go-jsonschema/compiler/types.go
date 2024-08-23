package compiler

import (
	"go/ast"

	"github.com/sourcegraph/go-jsonschema/jsonschema"
)

func isNullable(schema *jsonschema.Schema) bool {
	for _, typ := range schema.Type {
		if typ == jsonschema.NullType {
			return true
		}
	}
	return false
}

func isTypeOrNull(schema *jsonschema.Schema, typ jsonschema.PrimitiveType) bool {
	if len(schema.Type) == 1 {
		return schema.Type[0] == typ
	}
	if len(schema.Type) == 2 && isNullable(schema) {
		return schema.Type[0] == typ || schema.Type[1] == typ
	}
	return false
}

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
	return isTypeOrNull(schema, jsonschema.ObjectType)
}

func derefPtrType(x ast.Expr) *ast.Ident {
	dx, ok := x.(*ast.StarExpr)
	if ok {
		return dx.X.(*ast.Ident)
	}
	return x.(*ast.Ident)
}

func isBasicType(x ast.Expr) bool {
	t, ok := x.(*ast.Ident)
	if !ok {
		return false
	}
	return t.Name == "string" || t.Name == "bool" || t.Name == "int" || t.Name == "float64"
}

func forceGoPointer(schema *jsonschema.Schema) bool { return schema.Go != nil && schema.Go.Pointer }
