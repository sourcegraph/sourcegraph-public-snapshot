// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package golang

import (
	"fmt"
	"go/ast"
	"go/token"
)

func walkIdentList(v Visitor, scope *ast.Scope, list []*ast.Ident) {
	for _, x := range list {
		Walk(v, scope, x)
	}
}

func walkExprList(v Visitor, scope *ast.Scope, list []ast.Expr) {
	for _, x := range list {
		Walk(v, scope, x)
	}
}

func walkStmtList(v Visitor, scope *ast.Scope, list []ast.Stmt) {
	for _, x := range list {
		Walk(v, scope, x)
	}
}

func walkDeclList(v Visitor, scope *ast.Scope, list []ast.Decl) {
	for _, x := range list {
		Walk(v, scope, x)
	}
}

type Visitor func(scope *ast.Scope, node ast.Node) bool

var missing int

func Walk(v Visitor, scope *ast.Scope, node ast.Node) {
	if !v(scope, node) {
		return
	}
	switch n := node.(type) {
	// Comments and fields
	case *ast.Comment:
		// nothing to do

	case *ast.CommentGroup:
		for _, c := range n.List {
			Walk(v, scope, c)
		}

	case *ast.Field:
		// NOW: n.Names???
		if n.Doc != nil {
			Walk(v, scope, n.Doc)
		}
		walkIdentList(v, scope, n.Names)
		Walk(v, scope, n.Type)
		if n.Tag != nil {
			Walk(v, scope, n.Tag)
		}
		if n.Comment != nil {
			Walk(v, scope, n.Comment)
		}

	case *ast.FieldList:
		for _, f := range n.List {
			Walk(v, scope, f)
		}

	// Expressions
	case *ast.BadExpr, *ast.Ident, *ast.BasicLit:
		// nothing to do

	case *ast.Ellipsis:
		if n.Elt != nil {
			Walk(v, scope, n.Elt)
		}

	case *ast.FuncLit:
		Walk(v, scope, n.Type)
		Walk(v, scope, n.Body)

	case *ast.CompositeLit:
		if n.Type != nil {
			Walk(v, scope, n.Type)
		}
		walkExprList(v, scope, n.Elts)

	case *ast.ParenExpr:
		Walk(v, scope, n.X)

	case *ast.SelectorExpr:
		Walk(v, scope, n.X)
		Walk(v, scope, n.Sel)

	case *ast.IndexExpr:
		Walk(v, scope, n.X)
		Walk(v, scope, n.Index)

	case *ast.SliceExpr:
		Walk(v, scope, n.X)
		if n.Low != nil {
			Walk(v, scope, n.Low)
		}
		if n.High != nil {
			Walk(v, scope, n.High)
		}
		if n.Max != nil {
			Walk(v, scope, n.Max)
		}

	case *ast.TypeAssertExpr:
		Walk(v, scope, n.X)
		if n.Type != nil {
			Walk(v, scope, n.Type)
		}

	case *ast.CallExpr:
		Walk(v, scope, n.Fun)
		walkExprList(v, scope, n.Args)

	case *ast.StarExpr:
		Walk(v, scope, n.X)

	case *ast.UnaryExpr:
		Walk(v, scope, n.X)

	case *ast.BinaryExpr:
		Walk(v, scope, n.X)
		Walk(v, scope, n.Y)

	case *ast.KeyValueExpr:
		Walk(v, scope, n.Key)
		Walk(v, scope, n.Value)

	// Types
	case *ast.ArrayType:
		if n.Len != nil {
			Walk(v, scope, n.Len)
		}
		Walk(v, scope, n.Elt)

	case *ast.StructType:
		Walk(v, scope, n.Fields)

	case *ast.FuncType:
		if n.Params != nil {
			for _, field := range n.Params.List {
				for _, name := range field.Names {
					obj := ast.NewObj(ast.Var, name.Name)
					// TODO: populate obj.Decl, obj.Data for correctness
					scope.Insert(obj)
				}
			}
			Walk(v, scope, n.Params)
		}
		if n.Results != nil {
			for _, field := range n.Results.List {
				for _, name := range field.Names {
					obj := ast.NewObj(ast.Var, name.Name)
					// TODO: populate obj.Decl, obj.Data for correctness
					scope.Insert(obj)
				}
			}
			Walk(v, scope, n.Results)
		}

	case *ast.InterfaceType:
		Walk(v, scope, n.Methods)

	case *ast.MapType:
		Walk(v, scope, n.Key)
		Walk(v, scope, n.Value)

	case *ast.ChanType:
		Walk(v, scope, n.Value)

	// Statements
	case *ast.BadStmt:
		// nothing to do

	case *ast.DeclStmt:
		Walk(v, scope, n.Decl)

	case *ast.EmptyStmt:
		// nothing to do

	case *ast.LabeledStmt:
		obj := ast.NewObj(ast.Lbl, n.Label.Name)
		// TODO: populate obj.Decl, obj.Data for correctness
		scope.Insert(obj)

		Walk(v, scope, n.Label)
		Walk(v, scope, n.Stmt)

	case *ast.ExprStmt:
		Walk(v, scope, n.X)

	case *ast.SendStmt:
		Walk(v, scope, n.Chan)
		Walk(v, scope, n.Value)

	case *ast.IncDecStmt:
		Walk(v, scope, n.X)

	case *ast.AssignStmt:
		for _, expr := range n.Lhs {
			if v, ok := expr.(*ast.Ident); ok {
				obj := ast.NewObj(ast.Var, v.Name)
				// TODO: populate obj.Decl, obj.Data for correctness
				scope.Insert(obj)
			}
		}
		walkExprList(v, scope, n.Lhs)
		walkExprList(v, scope, n.Rhs)

	case *ast.GoStmt:
		Walk(v, scope, n.Call)

	case *ast.DeferStmt:
		Walk(v, scope, n.Call)

	case *ast.ReturnStmt:
		walkExprList(v, scope, n.Results)

	case *ast.BranchStmt:
		if n.Label != nil {
			Walk(v, scope, n.Label)
		}

	case *ast.BlockStmt:
		walkStmtList(v, ast.NewScope(scope), n.List)

	case *ast.IfStmt:
		if n.Init != nil {
			Walk(v, scope, n.Init)
		}
		Walk(v, scope, n.Cond)
		Walk(v, scope, n.Body)
		if n.Else != nil {
			Walk(v, scope, n.Else)
		}

	case *ast.CaseClause:
		walkExprList(v, scope, n.List)
		walkStmtList(v, scope, n.Body)

	case *ast.SwitchStmt:
		if n.Init != nil {
			Walk(v, scope, n.Init)
		}
		if n.Tag != nil {
			Walk(v, scope, n.Tag)
		}
		Walk(v, scope, n.Body)

	case *ast.TypeSwitchStmt:
		if n.Init != nil {
			Walk(v, scope, n.Init)
		}
		Walk(v, scope, n.Assign)
		Walk(v, scope, n.Body)

	case *ast.CommClause:
		if n.Comm != nil {
			Walk(v, scope, n.Comm)
		}
		walkStmtList(v, scope, n.Body)

	case *ast.SelectStmt:
		Walk(v, scope, n.Body)

	case *ast.ForStmt:
		if n.Init != nil {
			Walk(v, scope, n.Init)
		}
		if n.Cond != nil {
			Walk(v, scope, n.Cond)
		}
		if n.Post != nil {
			Walk(v, scope, n.Post)
		}
		Walk(v, scope, n.Body)

	case *ast.RangeStmt:
		if n.Key != nil {
			if v, ok := n.Key.(*ast.Ident); ok {
				obj := ast.NewObj(ast.Var, v.Name)
				// TODO: populate obj.Decl, obj.Data for correctness
				scope.Insert(obj)
			}
			Walk(v, scope, n.Key)
		}
		if n.Value != nil {
			if v, ok := n.Value.(*ast.Ident); ok {
				obj := ast.NewObj(ast.Var, v.Name)
				// TODO: populate obj.Decl, obj.Data for correctness
				scope.Insert(obj)
			}
			Walk(v, scope, n.Value)
		}
		Walk(v, scope, n.X)
		Walk(v, scope, n.Body)

	// Declarations
	case *ast.ImportSpec:
		if n.Doc != nil {
			Walk(v, scope, n.Doc)
		}
		if n.Name != nil {
			// intentionally not inserted into scope
			Walk(v, scope, n.Name)
		}
		Walk(v, scope, n.Path)
		if n.Comment != nil {
			Walk(v, scope, n.Comment)
		}

	case *ast.ValueSpec:
		if n.Doc != nil {
			Walk(v, scope, n.Doc)
		}
		walkIdentList(v, scope, n.Names)
		if n.Type != nil {
			Walk(v, scope, n.Type)
		}
		walkExprList(v, scope, n.Values)
		if n.Comment != nil {
			Walk(v, scope, n.Comment)
		}

	case *ast.TypeSpec:
		if n.Doc != nil {
			Walk(v, scope, n.Doc)
		}

		obj := ast.NewObj(ast.Typ, n.Name.Name)
		// TODO: populate obj.Decl, obj.Data for correctness
		scope.Insert(obj)

		Walk(v, scope, n.Name)
		Walk(v, scope, n.Type)
		if n.Comment != nil {
			Walk(v, scope, n.Comment)
		}

	case *ast.BadDecl:
		// nothing to do

	case *ast.GenDecl:
		if n.Doc != nil {
			Walk(v, scope, n.Doc)
		}
		for _, s := range n.Specs {
			var obj *ast.Object
			switch v := s.(type) {
			case *ast.ImportSpec:
				// intentionally omitted
			case *ast.ValueSpec:
				// NOW: v.Type ?
				kind := ast.Con
				if n.Tok == token.VAR {
					kind = ast.Var
				}
				for _, name := range v.Names {
					obj = ast.NewObj(kind, name.Name)
					// TODO: populate obj.Decl, obj.Data for correctness
					scope.Insert(obj)
				}
			case *ast.TypeSpec:
				obj = ast.NewObj(ast.Typ, v.Name.Name)
				// TODO: populate obj.Decl, obj.Data for correctness
				scope.Insert(obj)
			}
			Walk(v, scope, s)
		}

	case *ast.FuncDecl:
		if n.Doc != nil {
			Walk(v, scope, n.Doc)
		}
		if n.Recv != nil {
			for _, field := range n.Recv.List {
				for _, name := range field.Names {
					obj := ast.NewObj(ast.Var, name.Name)
					// TODO: populate obj.Decl, obj.Data for correctness
					scope.Insert(obj)
				}
			}
			Walk(v, scope, n.Recv)
		}

		obj := ast.NewObj(ast.Fun, n.Name.Name)
		// TODO: populate obj.Decl, obj.Data for correctness
		scope.Insert(obj)

		Walk(v, scope, n.Name)
		Walk(v, scope, n.Type)
		if n.Body != nil {
			Walk(v, scope, n.Body)
		}

	// Files and packages
	case *ast.File:
		if n.Doc != nil {
			Walk(v, n.Scope, n.Doc)
		}
		Walk(v, n.Scope, n.Name)
		walkDeclList(v, n.Scope, n.Decls)
		// don't walk n.Comments - they have been
		// visited already through the individual
		// nodes

	case *ast.Package:
		for _, f := range n.Files {
			Walk(v, f.Scope, f)
		}

	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
	}

	v(scope, nil)
}
