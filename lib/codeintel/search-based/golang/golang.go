package golang

import (
	"context"

	goGramar "github.com/smacker/go-tree-sitter/golang"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/search-based/api"
)

type Indexer struct{}

var _ api.Indexer = Indexer{}

func (g Indexer) Name() string {
	return "golang"
}

func (g Indexer) FileExtensions() []string {
	return []string{".go"}
}

func (g Indexer) Index(ctx context.Context, input *api.Input) (*lsif_typed.Document, error) {
	return api.Index(ctx, input, goGramar.GetLanguage(), api.LocalIntelGrammar{
		Identifiers: map[string]struct{}{"identifier": {}},
		Fingerprints: []api.DefinitionFingerprint{
			{
				ParentTypes:      []string{"identifier", "expression_list", "short_var_declaration", "block"},
				ParentFieldNames: []string{"", "left"},
			},
			{
				ParentTypes:      []string{"identifier", "expression_list", "short_var_declaration", "for_clause", "for_statement"},
				ParentFieldNames: []string{"", "left", "initializer"},
			},
			{
				ParentTypes:      []string{"identifier", "expression_list", "range_clause", "for_statement"},
				ParentFieldNames: []string{"", "left"},
			},
			{
				ParentTypes:      []string{"identifier", "parameter_declaration", "parameter_list", "func_literal"},
				ParentFieldNames: []string{"name", "", "parameters"},
			},
			{
				ParentTypes:      []string{"identifier", "parameter_declaration", "parameter_list", "function_declaration"},
				ParentFieldNames: []string{"name", "", "parameters"},
			},
		},
	})
}

//func visit(v *api.Builder, n *sitter.Node, scope *api.Scope) {
//	if n == nil {
//		return
//	}
//	recurseFunc := func(child *sitter.Node) {
//		visit(v, child, scope.NewInnerScope())
//	}
//	switch n.Type() {
//	case "identifier":
//		sym := scope.Lookup(api.NewSimpleName(v.Input.Substring(n)))
//		if sym != nil {
//			v.EmitOccurrence(sym, n, lsif_typed.MonikerOccurrence_ROLE_REFERENCE)
//		}
//	case "block":
//		visitShortDeclaration(v, n, scope, recurseFunc)
//	case "for_statement":
//		visitForStatement(v, n, scope, recurseFunc)
//	case "function_declaration", "func_literal":
//		visitFunctionDeclaration(v, n, scope, recurseFunc)
//	default:
//		api.ForeachChild(n, func(_ int, child *sitter.Node) {
//			recurseFunc(child)
//		})
//	}
//}
//
//func visitForStatement(v *api.Builder, n *sitter.Node, scope *api.Scope, recurseFunc api.RecurseFunc) {
//	api.AssertType(n, "for_statement")
//	api.ForeachChild(n, func(_ int, child *sitter.Node) {
//		switch child.Type() {
//		case "for_clause":
//			visitShortDeclaration(v, child, scope, recurseFunc)
//		case "range_clause":
//			visitExpressionList(v, child, scope, recurseFunc)
//		default:
//			recurseFunc(child)
//		}
//	})
//}
//func visitFunctionDeclaration(v *api.Builder, n *sitter.Node, scope *api.Scope, recurseFunc api.RecurseFunc) {
//	api.ForeachChild(n, func(_ int, child *sitter.Node) {
//		if child.Type() == "parameter_list" {
//			api.ForeachChild(child, func(_ int, child2 *sitter.Node) {
//				if child2.Type() == "parameter_declaration" {
//					api.ForeachChild(child2, func(_ int, child3 *sitter.Node) {
//						if child3.Type() == "identifier" {
//							v.EmitLocalOccurrence(child3, scope, lsif_typed.MonikerOccurrence_ROLE_DEFINITION)
//						} else {
//							recurseFunc(child3)
//						}
//					})
//				} else {
//					recurseFunc(child2)
//				}
//			})
//		} else {
//			recurseFunc(child)
//		}
//	})
//}
//
//func visitExpressionList(v *api.Builder, n *sitter.Node, scope *api.Scope, recurseFunc api.RecurseFunc) {
//	api.ForeachChild(n, func(i int, child *sitter.Node) {
//		if i == 0 && child.Type() == "expression_list" {
//			api.ForeachChild(child, func(_ int, identifier *sitter.Node) {
//				if identifier.Type() == "identifier" {
//					v.EmitLocalOccurrence(identifier, scope, lsif_typed.MonikerOccurrence_ROLE_DEFINITION)
//				}
//			})
//		} else {
//			recurseFunc(child)
//		}
//	})
//}
//
//func visitShortDeclaration(v *api.Builder, n *sitter.Node, scope *api.Scope, recurseFunc api.RecurseFunc) {
//	api.ForeachChild(n, func(_ int, child *sitter.Node) {
//		if child.Type() == "short_var_declaration" {
//			visitExpressionList(v, child, scope, recurseFunc)
//		} else {
//			recurseFunc(child)
//		}
//	})
//}
