package astrewrite

import (
	"fmt"
	"go/ast"
)

// WalkFunc describes a function to be called for each node during a Walk. The
// returned node can be used to rewrite the AST. Walking stops if the returned
// bool is false.
type WalkFunc func(ast.Node) (ast.Node, bool)

// Walk traverses an AST in depth-first order: It starts by calling
// fn(node); node must not be nil. It returns the rewritten node. If fn returns
// true, Walk invokes fn recursively for each of the non-nil children of node,
// followed by a call of fn(nil). The returned node of fn can be used to
// rewrite the passed node to fn. Panics if the returned type is not the same
// type as the original one.
func Walk(node ast.Node, fn WalkFunc) ast.Node {
	rewritten, ok := fn(node)
	if !ok {
		return rewritten
	}

	// walk children
	// (the order of the cases matches the order
	// of the corresponding node types in ast.go)
	switch n := node.(type) {
	// Comments and fields
	case *ast.Comment:
		// nothing to do

	case *ast.CommentGroup:
		for i, c := range n.List {
			n.List[i] = Walk(c, fn).(*ast.Comment)
		}

	case *ast.Field:
		if n.Doc != nil {
			n.Doc = Walk(n.Doc, fn).(*ast.CommentGroup)
		}
		walkIdentList(n.Names, fn)
		n.Type = Walk(n.Type, fn).(ast.Expr)
		if n.Tag != nil {
			n.Tag = Walk(n.Tag, fn).(*ast.BasicLit)
		}
		if n.Comment != nil {
			n.Comment = Walk(n.Comment, fn).(*ast.CommentGroup)
		}

	case *ast.FieldList:
		for i, f := range n.List {
			n.List[i] = Walk(f, fn).(*ast.Field)
		}

	// Expressions
	case *ast.BadExpr, *ast.Ident, *ast.BasicLit:
		// nothing to do

	case *ast.Ellipsis:
		if n.Elt != nil {
			n.Elt = Walk(n.Elt, fn).(ast.Expr)
		}

	case *ast.FuncLit:
		n.Type = Walk(n.Type, fn).(*ast.FuncType)
		n.Body = Walk(n.Body, fn).(*ast.BlockStmt)

	case *ast.CompositeLit:
		if n.Type != nil {
			n.Type = Walk(n.Type, fn).(ast.Expr)
		}
		walkExprList(n.Elts, fn)

	case *ast.ParenExpr:
		n.X = Walk(n.X, fn).(ast.Expr)

	case *ast.SelectorExpr:
		n.X = Walk(n.X, fn).(ast.Expr)
		n.Sel = Walk(n.Sel, fn).(*ast.Ident)

	case *ast.IndexExpr:
		n.X = Walk(n.X, fn).(ast.Expr)
		n.Index = Walk(n.Index, fn).(ast.Expr)

	case *ast.SliceExpr:
		n.X = Walk(n.X, fn).(ast.Expr)
		if n.Low != nil {
			n.Low = Walk(n.Low, fn).(ast.Expr)
		}
		if n.High != nil {
			n.High = Walk(n.High, fn).(ast.Expr)
		}
		if n.Max != nil {
			n.Max = Walk(n.Max, fn).(ast.Expr)
		}

	case *ast.TypeAssertExpr:
		n.X = Walk(n.X, fn).(ast.Expr)
		if n.Type != nil {
			n.Type = Walk(n.Type, fn).(ast.Expr)
		}

	case *ast.CallExpr:
		n.Fun = Walk(n.Fun, fn).(ast.Expr)
		walkExprList(n.Args, fn)

	case *ast.StarExpr:
		n.X = Walk(n.X, fn).(ast.Expr)

	case *ast.UnaryExpr:
		n.X = Walk(n.X, fn).(ast.Expr)

	case *ast.BinaryExpr:
		n.X = Walk(n.X, fn).(ast.Expr)
		n.Y = Walk(n.Y, fn).(ast.Expr)

	case *ast.KeyValueExpr:
		n.Key = Walk(n.Key, fn).(ast.Expr)
		n.Value = Walk(n.Value, fn).(ast.Expr)

	// Types
	case *ast.ArrayType:
		if n.Len != nil {
			n.Len = Walk(n.Len, fn).(ast.Expr)
		}
		n.Elt = Walk(n.Elt, fn).(ast.Expr)

	case *ast.StructType:
		n.Fields = Walk(n.Fields, fn).(*ast.FieldList)

	case *ast.FuncType:
		if n.Params != nil {
			n.Params = Walk(n.Params, fn).(*ast.FieldList)
		}
		if n.Results != nil {
			n.Results = Walk(n.Results, fn).(*ast.FieldList)
		}

	case *ast.InterfaceType:
		n.Methods = Walk(n.Methods, fn).(*ast.FieldList)

	case *ast.MapType:
		n.Key = Walk(n.Key, fn).(ast.Expr)
		n.Value = Walk(n.Value, fn).(ast.Expr)

	case *ast.ChanType:
		n.Value = Walk(n.Value, fn).(ast.Expr)

	// Statements
	case *ast.BadStmt:
		// nothing to do

	case *ast.DeclStmt:
		n.Decl = Walk(n.Decl, fn).(ast.Decl)

	case *ast.EmptyStmt:
		// nothing to do

	case *ast.LabeledStmt:
		n.Label = Walk(n.Label, fn).(*ast.Ident)
		n.Stmt = Walk(n.Stmt, fn).(ast.Stmt)

	case *ast.ExprStmt:
		n.X = Walk(n.X, fn).(ast.Expr)

	case *ast.SendStmt:
		n.Chan = Walk(n.Chan, fn).(ast.Expr)
		n.Value = Walk(n.Value, fn).(ast.Expr)

	case *ast.IncDecStmt:
		n.X = Walk(n.X, fn).(ast.Expr)

	case *ast.AssignStmt:
		walkExprList(n.Lhs, fn)
		walkExprList(n.Rhs, fn)

	case *ast.GoStmt:
		n.Call = Walk(n.Call, fn).(*ast.CallExpr)

	case *ast.DeferStmt:
		n.Call = Walk(n.Call, fn).(*ast.CallExpr)

	case *ast.ReturnStmt:
		walkExprList(n.Results, fn)

	case *ast.BranchStmt:
		if n.Label != nil {
			n.Label = Walk(n.Label, fn).(*ast.Ident)
		}

	case *ast.BlockStmt:
		walkStmtList(n.List, fn)

	case *ast.IfStmt:
		if n.Init != nil {
			n.Init = Walk(n.Init, fn).(ast.Stmt)
		}
		n.Cond = Walk(n.Cond, fn).(ast.Expr)
		n.Body = Walk(n.Body, fn).(*ast.BlockStmt)
		if n.Else != nil {
			n.Else = Walk(n.Else, fn).(ast.Stmt)
		}

	case *ast.CaseClause:
		walkExprList(n.List, fn)
		walkStmtList(n.Body, fn)

	case *ast.SwitchStmt:
		if n.Init != nil {
			n.Init = Walk(n.Init, fn).(ast.Stmt)
		}
		if n.Tag != nil {
			n.Tag = Walk(n.Tag, fn).(ast.Expr)
		}
		n.Body = Walk(n.Body, fn).(*ast.BlockStmt)

	case *ast.TypeSwitchStmt:
		if n.Init != nil {
			n.Init = Walk(n.Init, fn).(ast.Stmt)
		}
		n.Assign = Walk(n.Assign, fn).(ast.Stmt)
		n.Body = Walk(n.Body, fn).(*ast.BlockStmt)

	case *ast.CommClause:
		if n.Comm != nil {
			n.Comm = Walk(n.Comm, fn).(ast.Stmt)
		}
		walkStmtList(n.Body, fn)

	case *ast.SelectStmt:
		n.Body = Walk(n.Body, fn).(*ast.BlockStmt)

	case *ast.ForStmt:
		if n.Init != nil {
			n.Init = Walk(n.Init, fn).(ast.Stmt)
		}
		if n.Cond != nil {
			n.Cond = Walk(n.Cond, fn).(ast.Expr)
		}
		if n.Post != nil {
			n.Post = Walk(n.Post, fn).(ast.Stmt)
		}
		n.Body = Walk(n.Body, fn).(*ast.BlockStmt)

	case *ast.RangeStmt:
		if n.Key != nil {
			n.Key = Walk(n.Key, fn).(ast.Expr)
		}
		if n.Value != nil {
			n.Value = Walk(n.Value, fn).(ast.Expr)
		}
		n.X = Walk(n.X, fn).(ast.Expr)
		n.Body = Walk(n.Body, fn).(*ast.BlockStmt)

	// Declarations
	case *ast.ImportSpec:
		if n.Doc != nil {
			n.Doc = Walk(n.Doc, fn).(*ast.CommentGroup)
		}
		if n.Name != nil {
			n.Name = Walk(n.Name, fn).(*ast.Ident)
		}
		n.Path = Walk(n.Path, fn).(*ast.BasicLit)
		if n.Comment != nil {
			n.Comment = Walk(n.Comment, fn).(*ast.CommentGroup)
		}

	case *ast.ValueSpec:
		if n.Doc != nil {
			n.Doc = Walk(n.Doc, fn).(*ast.CommentGroup)
		}
		walkIdentList(n.Names, fn)
		if n.Type != nil {
			n.Type = Walk(n.Type, fn).(ast.Expr)
		}
		walkExprList(n.Values, fn)
		if n.Comment != nil {
			n.Comment = Walk(n.Comment, fn).(*ast.CommentGroup)
		}

	case *ast.TypeSpec:
		if n.Doc != nil {
			n.Doc = Walk(n.Doc, fn).(*ast.CommentGroup)
		}
		Walk(n.Name, fn)
		Walk(n.Type, fn)
		if n.Comment != nil {
			n.Comment = Walk(n.Comment, fn).(*ast.CommentGroup)
		}

	case *ast.BadDecl:
		// nothing to do

	case *ast.GenDecl:
		if n.Doc != nil {
			n.Doc = Walk(n.Doc, fn).(*ast.CommentGroup)
		}
		for i, s := range n.Specs {
			n.Specs[i] = Walk(s, fn).(ast.Spec)
		}

	case *ast.FuncDecl:
		if n.Doc != nil {
			n.Doc = Walk(n.Doc, fn).(*ast.CommentGroup)
		}
		if n.Recv != nil {
			n.Recv = Walk(n.Recv, fn).(*ast.FieldList)
		}
		n.Name = Walk(n.Name, fn).(*ast.Ident)
		n.Type = Walk(n.Type, fn).(*ast.FuncType)
		if n.Body != nil {
			n.Body = Walk(n.Body, fn).(*ast.BlockStmt)
		}

	// Files and packages
	case *ast.File:
		if n.Doc != nil {
			n.Doc = Walk(n.Doc, fn).(*ast.CommentGroup)
		}
		n.Name = Walk(n.Name, fn).(*ast.Ident)
		walkDeclList(n.Decls, fn)
		// don't walk n.Comments - they have been
		// visited already through the individual
		// nodes

	case *ast.Package:
		for i, f := range n.Files {
			n.Files[i] = Walk(f, fn).(*ast.File)
		}

	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
	}

	fn(nil)
	return rewritten
}

func walkIdentList(list []*ast.Ident, fn WalkFunc) {
	for i, x := range list {
		list[i] = Walk(x, fn).(*ast.Ident)
	}
}

func walkExprList(list []ast.Expr, fn WalkFunc) {
	for i, x := range list {
		list[i] = Walk(x, fn).(ast.Expr)
	}
}

func walkStmtList(list []ast.Stmt, fn WalkFunc) {
	for i, x := range list {
		list[i] = Walk(x, fn).(ast.Stmt)
	}
}

func walkDeclList(list []ast.Decl, fn WalkFunc) {
	for i, x := range list {
		list[i] = Walk(x, fn).(ast.Decl)
	}
}
