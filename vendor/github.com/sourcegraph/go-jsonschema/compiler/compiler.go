package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-jsonschema/jsonschema"
)

// Compile generates Go declarations for types that hold values described by the JSON Schemas.
//
// 1. Parse (per-schema)
// 2. Resolve references (all schemas)
// 3. Generate code (per-schema)
func Compile(schemas []*jsonschema.Schema) ([]ast.Decl, []*ast.ImportSpec, error) {
	//
	// Step 1: Parse (per-schema)
	//
	locationsByRoot := make(schemaLocationsByRoot, len(schemas))
	for _, root := range schemas {
		var err error
		locationsByRoot[root], err = parseSchema(root)
		if err != nil {
			return nil, nil, err
		}
	}

	//
	// Step 2: Resolve references (all schemas together)
	//
	resolutions, err := resolveReferences(locationsByRoot)
	if err != nil {
		return nil, nil, err
	}

	//
	// Step 3: Generate code (per-schema)
	//
	var allDecls []ast.Decl
	var allImports []*ast.ImportSpec
	for _, schemas := range locationsByRoot {
		decls, imports, err := generateDecls(schemas, resolutions, locationsByRoot)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "generating decls")
		}
		allDecls = append(allDecls, decls...)
		allImports = append(allImports, imports...)
	}
	// Sort decls.
	sort.SliceStable(allDecls, func(i, j int) bool {
		name := func(k int) string {
			switch d := allDecls[k].(type) {
			case *ast.GenDecl:
				return d.Specs[0].(*ast.TypeSpec).Name.Name
			case *ast.FuncDecl:
				return derefPtrType(d.Recv.List[0].Type).Name
			default:
				panic(fmt.Sprintf("unhandled %T", d))
			}
		}
		return name(i) < name(j)
	})

	// Imports must also be in the decl list, or else they won't be printed in the Go source by
	// go/printer.
	if len(allImports) > 0 {
		tmp := make([]ast.Decl, len(allDecls)+1)
		d := &ast.GenDecl{Tok: token.IMPORT}
		for _, imp := range allImports {
			d.Specs = append(d.Specs, imp)
		}
		if len(allImports) > 1 {
			d.Lparen = 1
			d.Rparen = 1
		}
		tmp[0] = d
		copy(tmp[1:], allDecls)
		allDecls = tmp
	}

	return allDecls, allImports, nil
}

type schemaLocator interface {
	locateSchema(schema *jsonschema.Schema) (root *jsonschema.Schema, location *schemaLocation)
}

// schemaLocationsByRoot maps root -> subschema -> location.
type schemaLocationsByRoot map[*jsonschema.Schema]map[*jsonschema.Schema]schemaLocation

// locateSchema implements schemaLocator.
func (s schemaLocationsByRoot) locateSchema(schema *jsonschema.Schema) (root *jsonschema.Schema, location *schemaLocation) {
	for root, locations := range s {
		location, ok := locations[schema]
		if ok {
			return root, &location
		}
	}
	return nil, nil
}
