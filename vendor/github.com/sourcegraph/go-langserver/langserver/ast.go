package langserver

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"golang.org/x/tools/go/loader"

	"github.com/sourcegraph/go-langserver/langserver/util"
)

func offsetForPosition(contents []byte, p lsp.Position) (offset int, valid bool, whyInvalid string) {
	line := 0
	col := 0
	// TODO(sqs): count chars, not bytes, per LSP. does that mean we
	// need to maintain 2 separate counters since we still need to
	// return the offset as bytes?
	for _, b := range contents {
		if line == p.Line && col == p.Character {
			return offset, true, ""
		}
		if (line == p.Line && col > p.Character) || line > p.Line {
			return 0, false, fmt.Sprintf("character %d is beyond line %d boundary", p.Character, p.Line)
		}
		offset++
		if b == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	if line == 0 {
		return 0, false, fmt.Sprintf("character %d is beyond first line boundary", p.Character)
	}
	return 0, false, fmt.Sprintf("file only has %d lines", line+1)
}

func rangeForNode(fset *token.FileSet, node ast.Node) lsp.Range {
	start := fset.Position(node.Pos())
	end := fset.Position(node.End()) // node.End is exclusive, and so is the LSP spec
	return lsp.Range{
		Start: lsp.Position{Line: start.Line - 1, Character: start.Column - 1},
		End:   lsp.Position{Line: end.Line - 1, Character: end.Column - 1},
	}
}

type fakeNode struct{ p, e token.Pos }

func (n fakeNode) Pos() token.Pos { return n.p }
func (n fakeNode) End() token.Pos { return n.e }

func goRangesToLSPLocations(fset *token.FileSet, nodes []*ast.Ident) []lsp.Location {
	locs := make([]lsp.Location, len(nodes))
	for i, node := range nodes {
		locs[i] = goRangeToLSPLocation(fset, node.Pos(), node.End())
	}
	return locs
}

// goRangeToLSPLocation converts a token.Pos range into a lsp.Location. end is
// exclusive.
func goRangeToLSPLocation(fset *token.FileSet, pos token.Pos, end token.Pos) lsp.Location {
	return lsp.Location{
		URI:   util.PathToURI(fset.Position(pos).Filename),
		Range: rangeForNode(fset, fakeNode{p: pos, e: end}),
	}

}

type action int

const (
	actionUnknown action = iota // None of the below
	actionExpr                  // FuncDecl, true Expr or Ident(types.{Const,Var})
	actionType                  // type Expr or Ident(types.TypeName).
	actionStmt                  // Stmt or Ident(types.Label)
	actionPackage               // Ident(types.Package) or ImportSpec
)

// findInterestingNode classifies the syntax node denoted by path as one of:
//    - an expression, part of an expression or a reference to a constant
//      or variable;
//    - a type, part of a type, or a reference to a named type;
//    - a statement, part of a statement, or a label referring to a statement;
//    - part of a package declaration or import spec.
//    - none of the above.
// and returns the most "interesting" associated node, which may be
// the same node, an ancestor or a descendent.
//
// Adapted from golang.org/x/tools/cmd/guru (Copyright (c) 2013 The Go Authors). All rights
// reserved. See NOTICE for full license.
func findInterestingNode(pkginfo *loader.PackageInfo, path []ast.Node) ([]ast.Node, action) {
	// TODO(adonovan): integrate with go/types/stdlib_test.go and
	// apply this to every AST node we can find to make sure it
	// doesn't crash.

	// TODO(adonovan): audit for ParenExpr safety, esp. since we
	// traverse up and down.

	// TODO(adonovan): if the users selects the "." in
	// "fmt.Fprintf()", they'll get an ambiguous selection error;
	// we won't even reach here.  Can we do better?

	// TODO(adonovan): describing a field within 'type T struct {...}'
	// describes the (anonymous) struct type and concludes "no methods".
	// We should ascend to the enclosing type decl, if any.

	for len(path) > 0 {
		switch n := path[0].(type) {
		case *ast.GenDecl:
			if len(n.Specs) == 1 {
				// Descend to sole {Import,Type,Value}Spec child.
				path = append([]ast.Node{n.Specs[0]}, path...)
				continue
			}
			return path, actionUnknown // uninteresting

		case *ast.FuncDecl:
			// Descend to function name.
			path = append([]ast.Node{n.Name}, path...)
			continue

		case *ast.ImportSpec:
			return path, actionPackage

		case *ast.ValueSpec:
			if len(n.Names) == 1 {
				// Descend to sole Ident child.
				path = append([]ast.Node{n.Names[0]}, path...)
				continue
			}
			return path, actionUnknown // uninteresting

		case *ast.TypeSpec:
			// Descend to type name.
			path = append([]ast.Node{n.Name}, path...)
			continue

		case ast.Stmt:
			return path, actionStmt

		case *ast.ArrayType,
			*ast.StructType,
			*ast.FuncType,
			*ast.InterfaceType,
			*ast.MapType,
			*ast.ChanType:
			return path, actionType

		case *ast.Comment, *ast.CommentGroup, *ast.File, *ast.KeyValueExpr, *ast.CommClause:
			return path, actionUnknown // uninteresting

		case *ast.Ellipsis:
			// Continue to enclosing node.
			// e.g. [...]T in ArrayType
			//      f(x...) in CallExpr
			//      f(x...T) in FuncType

		case *ast.Field:
			// TODO(adonovan): this needs more thought,
			// since fields can be so many things.
			if len(n.Names) == 1 {
				// Descend to sole Ident child.
				path = append([]ast.Node{n.Names[0]}, path...)
				continue
			}
			// Zero names (e.g. anon field in struct)
			// or multiple field or param names:
			// continue to enclosing field list.

		case *ast.FieldList:
			// Continue to enclosing node:
			// {Struct,Func,Interface}Type or FuncDecl.

		case *ast.BasicLit:
			if _, ok := path[1].(*ast.ImportSpec); ok {
				return path[1:], actionPackage
			}
			return path, actionExpr

		case *ast.SelectorExpr:
			// TODO(adonovan): use Selections info directly.
			if pkginfo.Uses[n.Sel] == nil {
				// TODO(adonovan): is this reachable?
				return path, actionUnknown
			}
			// Descend to .Sel child.
			path = append([]ast.Node{n.Sel}, path...)
			continue

		case *ast.Ident:
			switch pkginfo.ObjectOf(n).(type) {
			case *types.PkgName:
				return path, actionPackage

			case *types.Const:
				return path, actionExpr

			case *types.Label:
				return path, actionStmt

			case *types.TypeName:
				return path, actionType

			case *types.Var:
				// For x in 'struct {x T}', return struct type, for now.
				if _, ok := path[1].(*ast.Field); ok {
					_ = path[2].(*ast.FieldList) // assertion
					if _, ok := path[3].(*ast.StructType); ok {
						return path[3:], actionType
					}
				}
				return path, actionExpr

			case *types.Func:
				return path, actionExpr

			case *types.Builtin:
				// For reference to built-in function, return enclosing call.
				path = path[1:] // ascend to enclosing function call
				continue

			case *types.Nil:
				return path, actionExpr
			}

			// No object.
			switch path[1].(type) {
			case *ast.SelectorExpr:
				// Return enclosing selector expression.
				return path[1:], actionExpr

			case *ast.Field:
				// TODO(adonovan): test this.
				// e.g. all f in:
				//  struct { f, g int }
				//  interface { f() }
				//  func (f T) method(f, g int) (f, g bool)
				//
				// switch path[3].(type) {
				// case *ast.FuncDecl:
				// case *ast.StructType:
				// case *ast.InterfaceType:
				// }
				//
				// return path[1:], actionExpr
				//
				// Unclear what to do with these.
				// Struct.Fields             -- field
				// Interface.Methods         -- field
				// FuncType.{Params.Results} -- actionExpr
				// FuncDecl.Recv             -- actionExpr

			case *ast.File:
				// 'package foo'
				return path, actionPackage

			case *ast.ImportSpec:
				return path[1:], actionPackage

			default:
				// e.g. blank identifier
				// or y in "switch y := x.(type)"
				// or code in a _test.go file that's not part of the package.
				return path, actionUnknown
			}

		case *ast.StarExpr:
			if pkginfo.Types[n].IsType() {
				return path, actionType
			}
			return path, actionExpr

		case ast.Expr:
			// All Expr but {BasicLit,Ident,StarExpr} are
			// "true" expressions that evaluate to a value.
			return path, actionExpr
		}

		// Ascend to parent.
		path = path[1:]
	}

	return nil, actionUnknown // unreachable
}
