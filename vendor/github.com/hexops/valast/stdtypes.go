package valast

import (
	"fmt"
	"go/ast"
	"go/token"
	"time"
)

// Register custom reprsentations of common structs from stdlib that only
// contain unexported fields.
func init() {
	// For time.Time, returns the AST expression equivalent of:
	//
	//	time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	RegisterType(func(t time.Time) ast.Expr {
		return &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "time"},
				Sel: &ast.Ident{Name: "Date"},
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", t.Year())},
				&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", t.Month())},
				&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", t.Day())},
				&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", t.Hour())},
				&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", t.Minute())},
				&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", t.Second())},
				&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", t.Nanosecond())},
				&ast.SelectorExpr{
					X:   &ast.Ident{Name: "time"},
					Sel: &ast.Ident{Name: t.Location().String()},
				},
			},
		}
	})
}
