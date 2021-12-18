package golang

import (
	"context"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/search-based/api"
)

type GoIndexer struct{}

var _ api.Indexer = GoIndexer{}

func (g GoIndexer) Name() string {
	return "golang"
}

func (g GoIndexer) FileExtensions() []string {
	return []string{".go"}
}

func (g GoIndexer) Index(_ context.Context, input *api.Input, _ *api.IndexingOptions) (*lsif_typed.Document, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, input.Bytes)
	if err != nil {
		return nil, err
	}
	doc := &lsif_typed.Document{
		Uri:         input.Uri(),
		Occurrences: nil,
	}
	visitor := &api.Builder{
		Document: doc,
		Input:    input,
	}
	visit(visitor, tree.RootNode(), api.NewScope())
	return doc, nil
}

func visit(v *api.Builder, n *sitter.Node, scope *api.Scope) {
	if n == nil {
		return
	}
	recurseFunc := func(child *sitter.Node) {
		visit(v, child, scope.NewInnerScope())
	}
	switch n.Type() {
	case "identifier":
		sym := scope.Lookup(api.NewSimpleName(v.Input.Substring(n)))
		if sym != nil {
			v.EmitOccurrence(sym, n, lsif_typed.MonikerOccurrence_ROLE_REFERENCE)
		}
	case "block":
		visitShortDeclaration(v, n, scope, recurseFunc)
	case "for_statement":
		visitForStatement(v, n, scope, recurseFunc)
	case "function_declaration", "func_literal":
		visitFunctionDeclaration(v, n, scope, recurseFunc)
	default:
		api.ForeachChild(n, func(_ int, child *sitter.Node) {
			recurseFunc(child)
		})
	}
}

func visitForStatement(v *api.Builder, n *sitter.Node, scope *api.Scope, recurseFunc api.RecurseFunc) {
	api.AssertType(n, "for_statement")
	api.ForeachChild(n, func(_ int, child *sitter.Node) {
		switch child.Type() {
		case "for_clause":
			visitShortDeclaration(v, child, scope, recurseFunc)
		case "range_clause":
			visitExpressionList(v, child, scope, recurseFunc)
		default:
			recurseFunc(child)
		}
	})
}
func visitFunctionDeclaration(v *api.Builder, n *sitter.Node, scope *api.Scope, recurseFunc api.RecurseFunc) {
	api.ForeachChild(n, func(_ int, child *sitter.Node) {
		if child.Type() == "parameter_list" {
			api.ForeachChild(child, func(_ int, child2 *sitter.Node) {
				if child2.Type() == "parameter_declaration" {
					api.ForeachChild(child2, func(_ int, child3 *sitter.Node) {
						if child3.Type() == "identifier" {
							v.EmitLocalOccurrence(child3, scope, lsif_typed.MonikerOccurrence_ROLE_DEFINITION)
						} else {
							recurseFunc(child3)
						}
					})
				} else {
					recurseFunc(child2)
				}
			})
		} else {
			recurseFunc(child)
		}
	})
}

func visitExpressionList(v *api.Builder, n *sitter.Node, scope *api.Scope, recurseFunc api.RecurseFunc) {
	api.ForeachChild(n, func(i int, child *sitter.Node) {
		if i == 0 && child.Type() == "expression_list" {
			api.ForeachChild(child, func(_ int, identifier *sitter.Node) {
				if identifier.Type() == "identifier" {
					v.EmitLocalOccurrence(identifier, scope, lsif_typed.MonikerOccurrence_ROLE_DEFINITION)
				}
			})
		} else {
			recurseFunc(child)
		}
	})
}

func visitShortDeclaration(v *api.Builder, n *sitter.Node, scope *api.Scope, recurseFunc api.RecurseFunc) {
	api.ForeachChild(n, func(_ int, child *sitter.Node) {
		if child.Type() == "short_var_declaration" {
			visitExpressionList(v, child, scope, recurseFunc)
		} else {
			recurseFunc(child)
		}
	})
}
