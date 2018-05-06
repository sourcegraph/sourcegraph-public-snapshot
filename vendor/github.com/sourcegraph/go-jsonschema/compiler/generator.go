package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-jsonschema/jsonschema"
)

// generateDecls returns Go type declarations for the schemas, which are all in the same root JSON
// Schema.
func generateDecls(schemas map[*jsonschema.Schema]schemaLocation, resolutions map[*jsonschema.Schema]*jsonschema.Schema, schemaLocator schemaLocator) ([]ast.Decl, error) {
	g := generator{schemas: schemas, resolutions: resolutions, schemaLocator: schemaLocator}
	var decls []ast.Decl
	for schema := range schemas {
		decl, err := g.emit(schema)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to emit decl for schema")
		}
		if decl != nil {
			decls = append(decls, decl)
		}
	}
	return decls, nil
}

type generator struct {
	schemas       map[*jsonschema.Schema]schemaLocation     // for the current root schema only
	resolutions   map[*jsonschema.Schema]*jsonschema.Schema // for all schemas in scope
	schemaLocator schemaLocator

	decls []ast.Decl
}

var emptyInterfaceType = &ast.InterfaceType{
	Methods: &ast.FieldList{
		// Setting these to 1 makes the `interface{}` print on the same line (because it makes this
		// FieldList pass go/printer's (token.Pos).IsValid() check).
		Opening: 1,
		Closing: 1,
	},
}

// emit returns the declaration for the Go type for schema, or nil if no declaration is needed (such
// as when schema is represented by a builtin Go type).
func (g *generator) emit(schema *jsonschema.Schema) (ast.Decl, error) {
	needsNamedGoType := len(schema.Type) == 1 && schema.Type[0] == jsonschema.ObjectType && schema.Properties != nil
	if !needsNamedGoType {
		return nil, nil
	}

	// Sort properties deterministically (by name).
	names := make([]string, 0, len(*schema.Properties))
	for name := range *schema.Properties {
		names = append(names, name)
	}
	sort.Strings(names)

	// Create a field for each property.
	fields := make([]*ast.Field, len(names))
	for i, name := range names {
		prop := (*schema.Properties)[name]

		typeExpr, err := g.expr(prop)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("failed to get type expression for property %q", name))
		}

		var jsonStructTagExtra string
		if !schema.IsRequiredProperty(name) {
			// In Go, a pointer-to-{array,map}-type doesn't add (necessary) expressiveness for our use
			// case vs. just an {array,map} type.
			_, isPtrToArray := typeExpr.(*ast.ArrayType)
			_, isPtrToMap := typeExpr.(*ast.MapType)
			isGoBuiltinType := len(prop.Type) == 1 && goBuiltinType(prop.Type[0]) != ""
			if !isPtrToArray && !isPtrToMap && !isGoBuiltinType {
				typeExpr = &ast.StarExpr{X: typeExpr}
			}
			jsonStructTagExtra = ",omitempty"
		}

		fields[i] = &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(toGoName(name, "Property_"))},
			Type:  typeExpr,
			Tag: &ast.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf("`json:%q`", name+jsonStructTagExtra),
			},
		}
	}

	goName, err := goNameForSchema(schema, g.schemas[schema])
	if err != nil {
		return nil, err
	}
	typeSpec := &ast.TypeSpec{
		Name: ast.NewIdent(goName),
		Type: &ast.StructType{Fields: &ast.FieldList{List: fields}},
	}
	return &ast.GenDecl{
		Doc: &ast.CommentGroup{
			List: []*ast.Comment{{Text: "\n// " + docForSchema(schema, goName)}},
		},
		Tok:   token.TYPE,
		Specs: []ast.Spec{typeSpec},
	}, nil

}

// expr returns the Go expression AST node that refers to the Go type (builtin or named) for schema.
func (g *generator) expr(schema *jsonschema.Schema) (ast.Expr, error) {
	// Handle $ref to another schema.
	if schema.Reference != nil {
		return g.expr(g.resolutions[schema])
	}

	// Handle array types.
	if len(schema.Type) == 1 && schema.Type[0] == jsonschema.ArrayType {
		var elt ast.Expr
		if schema.Items != nil && schema.Items.Schema != nil {
			var err error
			elt, err = g.expr(schema.Items.Schema)
			if err != nil {
				return nil, err
			}
			// Prefer array-of-pointer-to-struct over array-of-struct.
			//
			// TODO(sqs): Not all $ref values point to things that are Go named types.
			if isEmittedAsGoNamedType(schema.Items.Schema) || schema.Items.Schema.Reference != nil {
				elt = &ast.StarExpr{X: elt}
			}
		} else {
			elt = emptyInterfaceType
		}
		return &ast.ArrayType{Elt: elt}, nil
	}

	// Handle object types that are emitted as Go map types (not named struct types).
	if len(schema.Type) == 1 && schema.Type[0] == jsonschema.ObjectType && schema.Properties == nil {
		typeExpr, err := g.expr(schema.AdditionalProperties)
		if err != nil {
			return nil, err
		}
		return &ast.MapType{Key: ast.NewIdent("string"), Value: typeExpr}, nil
	}

	// Handle types represented by Go builtin types or some other non-named types.
	if len(schema.Type) != 1 {
		return emptyInterfaceType, nil
	}
	if len(schema.Type) == 1 && goBuiltinType(schema.Type[0]) != "" {
		return ast.NewIdent(goBuiltinType(schema.Type[0])), nil
	}
	if schema.IsEmpty {
		return emptyInterfaceType, nil
	}
	if schema.IsNegated {
		return emptyInterfaceType, nil
	}

	// Otherwise, use a Go named type.
	_, location := g.schemaLocator.locateSchema(schema)
	if location == nil {
		return nil, errors.New("unable to locate schema")
	}
	goName, err := goNameForSchema(schema, *location)
	if err != nil {
		return nil, err
	}
	return ast.NewIdent(goName), nil
}

func docForSchema(schema *jsonschema.Schema, goName string) string {
	doc := goName
	if schema.Description != nil {
		doc += " " + *schema.Description
	}
	return doc
}
