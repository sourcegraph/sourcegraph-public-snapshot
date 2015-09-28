package gen

import "go/ast"

// UserSpecExpr returns the AST expression that evaluates to the
// UserSpec, given a protobuf RPC service method argument (XxxOp,
// RepoRevSpec, UserSpec, BuildSpec, etc.).
//
// For example, if arg="x" and argType="*sourcegraph.OrgsListOp", then
// UserSpecExpr returns an AST expression equivalent to "x.Member".
func UserSpecExpr(arg ast.Expr, argType ast.Expr) ast.Expr {
	switch AstString(argType) {
	case "*sourcegraph.UserSpec", "*UserSpec":
		return &ast.StarExpr{X: arg}
	case "sourcegraph.UserSpec", "UserSpec":
		return arg
	}

	switch t := argType.(type) {
	case *ast.StarExpr:
		return UserSpecExpr(arg, t.X)
	case *ast.Ident:
		if ast.IsExported(t.Name) {
			_, spec, err := typeSpec("sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph", t.Name)
			if err != nil {
				return nil
			}
			structType, ok := spec.Type.(*ast.StructType)
			if !ok {
				return nil
			}
			for _, field := range structType.Fields.List {
				var argField ast.Expr
				if len(field.Names) > 0 {
					argField = &ast.SelectorExpr{X: arg, Sel: field.Names[0]}
				} else {
					argField = arg
				}
				x := UserSpecExpr(argField, field.Type)
				if x != nil {
					return x
				}
			}
		}
	case *ast.SelectorExpr:
		if id, ok := t.X.(*ast.Ident); ok && id.Name == "sourcegraph" {
			return UserSpecExpr(arg, t.Sel)
		}
	}
	return nil
}
