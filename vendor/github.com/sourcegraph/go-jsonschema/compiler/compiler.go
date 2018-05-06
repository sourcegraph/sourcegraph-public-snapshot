package compiler

import (
	"go/ast"
	"sort"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-jsonschema/jsonschema"
)

// Compile generates Go declarations for types that hold values described by the JSON Schemas.
//
// 1. Parse (per-schema)
// 2. Resolve references (all schemas)
// 3. Generate code (per-schema)
func Compile(schemas []*jsonschema.Schema) ([]ast.Decl, error) {
	//
	// Step 1: Parse (per-schema)
	//
	locationsByRoot := make(schemaLocationsByRoot, len(schemas))
	for _, root := range schemas {
		var err error
		locationsByRoot[root], err = parseSchema(root)
		if err != nil {
			return nil, err
		}
	}

	//
	// Step 2: Resolve references (all schemas together)
	//
	resolutions, err := resolveReferences(locationsByRoot)
	if err != nil {
		return nil, err
	}

	//
	// Step 3: Generate code (per-schema)
	//
	var allDecls []ast.Decl
	for _, schemas := range locationsByRoot {
		decls, err := generateDecls(schemas, resolutions, locationsByRoot)
		if err != nil {
			return nil, errors.WithMessage(err, "generating decls")
		}
		allDecls = append(allDecls, decls...)
	}
	// Sort decls.
	sort.Slice(allDecls, func(i, j int) bool {
		name := func(k int) string { return allDecls[k].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Name.Name }
		return name(i) < name(j)
	})
	return allDecls, nil
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
