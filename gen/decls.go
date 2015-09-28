package gen

import (
	"go/ast"
	"go/token"
)

// Types returns all top-level type declarations in fileOrPkg (an
// *ast.File or *ast.Package) for which the filter func returns true.
func Types(fileOrPkg ast.Node, filter func(*ast.TypeSpec) bool) []*ast.TypeSpec {
	var types []*ast.TypeSpec
	ast.Walk(visitFn(func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.GenDecl:
			if node.Tok == token.TYPE {
				for _, spec := range node.Specs {
					tspec := spec.(*ast.TypeSpec)
					if filter(tspec) {
						types = append(types, tspec)
					}
				}
			}
			return false
		default:
			return true
		}
	}), fileOrPkg)
	return types
}

// Vars returns all top-level var declarations in fileOrPkg (an
// *ast.File or *ast.Package) for which the filter func returns true.
func Vars(fileOrPkg ast.Node, filter func(*ast.ValueSpec) bool) []*ast.ValueSpec {
	var vars []*ast.ValueSpec
	ast.Walk(visitFn(func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.GenDecl:
			if node.Tok == token.VAR {
				for _, spec := range node.Specs {
					vspec := spec.(*ast.ValueSpec)
					if filter(vspec) {
						vars = append(vars, vspec)
					}
				}
			}
			return false
		default:
			return true
		}
	}), fileOrPkg)
	return vars
}

type visitFn func(node ast.Node) (descend bool)

func (v visitFn) Visit(node ast.Node) ast.Visitor {
	descend := v(node)
	if descend {
		return v
	}
	return nil
}
