package types

import (
	"fmt"
	"go/ast"
	"go/types"
)

type visitor struct {
	importPath string
	pkgType    *types.Package
	types      map[string]*Interface
}

func newVisitor(importPath string, pkgType *types.Package) *visitor {
	return &visitor{
		importPath: importPath,
		pkgType:    pkgType,
		types:      map[string]*Interface{},
	}
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.File:
		return v

	case *ast.GenDecl:
		for _, spec := range n.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				name := typeSpec.Name.Name
				_, obj := v.pkgType.Scope().Innermost(typeSpec.Pos()).LookupParent(name, 0)

				switch t := obj.Type().Underlying().(type) {
				case *types.Interface:
					namedType, ok := obj.Type().(*types.Named)
					if !ok {
						panic(fmt.Sprintf("Unexpected type %T: expected *types.Named", obj.Type()))
					}

					if !t.IsMethodSet() {
						// Contains type constraints - we generate illegal code in this circumstance.
						// I'm not sure it makes sense to support this case, but we can revisit if we
						// get a feature request in the future or run into a case in the wild.
						continue
					}

					v.types[name] = newInterfaceFromTypeSpec(name, v.importPath, typeSpec, t, namedType.TypeParams())
				}
			}
		}
	}

	return nil
}
