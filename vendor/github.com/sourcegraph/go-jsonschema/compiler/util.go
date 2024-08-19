package compiler

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"text/template"
)

func makeMethod(f *ast.FuncDecl, recvType ast.Expr, name string) {
	f.Recv = &ast.FieldList{
		List: []*ast.Field{{
			Names: []*ast.Ident{ast.NewIdent("v")},
			Type:  recvType,
		}},
	}
	f.Name = ast.NewIdent(name)
}

func executeTemplate(tmpl *template.Template, data any) string {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		panic(err)
	}
	return buf.String()
}

func parseFuncLitToFuncDecl(funcLitExpr string) (*ast.FuncDecl, error) {
	x, err := parser.ParseExpr(funcLitExpr)
	if err != nil {
		return nil, fmt.Errorf("parsing func lit expr %q: %w", funcLitExpr, err)
	}
	funcLit, ok := x.(*ast.FuncLit)
	if !ok {
		panic("not an *ast.FuncLit")
	}
	return &ast.FuncDecl{
		Type: funcLit.Type,
		Body: funcLit.Body,
	}, nil
}
