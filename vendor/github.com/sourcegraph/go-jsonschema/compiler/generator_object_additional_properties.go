package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"text/template"

	"github.com/sourcegraph/go-jsonschema/jsonschema"
)

func (g *generator) emitStructAdditionalField(schema *jsonschema.Schema, goName string, fields []field) (*ast.Field, []ast.Decl, []*ast.ImportSpec, error) {
	additionalField := &ast.Field{
		Names: []*ast.Ident{ast.NewIdent("Additional")},
		Type:  &ast.MapType{Key: ast.NewIdent("string"), Value: anyType},
		Tag: &ast.BasicLit{
			Kind:  token.STRING,
			Value: fmt.Sprintf("`json:%q`", "-"),
		},
		Comment: &ast.CommentGroup{
			List: []*ast.Comment{{Text: " // additionalProperties not explicitly defined in the schema"}},
		},
	}

	// Generate MarshalJSON and UnmarshalJSON methods on the Go union type.
	templateData := map[string]any{
		"fields": fields,
		"goName": goName,
	}
	marshalJSONDecl, err := parseFuncLitToFuncDecl(executeTemplate(structAdditionalFieldMarshalJSONTemplate, templateData))
	if err != nil {
		return nil, nil, nil, err
	}
	unmarshalJSONDecl, err := parseFuncLitToFuncDecl(executeTemplate(structAdditionalFieldUnmarshalJSONTemplate, templateData))
	if err != nil {
		return nil, nil, nil, err
	}
	makeMethod(marshalJSONDecl, ast.NewIdent(goName), "MarshalJSON")
	makeMethod(unmarshalJSONDecl, &ast.StarExpr{X: ast.NewIdent(goName)}, "UnmarshalJSON")

	return additionalField, []ast.Decl{marshalJSONDecl, unmarshalJSONDecl},
		importSpecs("encoding/json"),
		nil
}

var (
	structAdditionalFieldMarshalJSONTemplate = template.Must(template.New("").Parse(`
func() ([]byte, error) {
	m := make(map[string]any, len(v.Additional))
	for k, v := range v.Additional {
		m[k] = v
	}

	type wrapper {{.goName}}
	b, err := json.Marshal(wrapper(v))
	if err != nil {
		return nil, err
	}
	var m2 map[string]any
	if err := json.Unmarshal(b, &m2); err != nil {
		return nil, err
	}

	for k, v := range m2 {
		m[k] = v
	}

	return json.Marshal(m)
}
`))
	structAdditionalFieldUnmarshalJSONTemplate = template.Must(template.New("").Parse(`
func(data []byte) error {
	type wrapper {{.goName}}
	var s wrapper
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*v = {{.goName}}(s)

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	{{- range .fields}}
	delete(m, {{printf "%q" .JSONName}})
	{{- end}}

	if len(m) > 0 {
		v.Additional = make(map[string]any, len(m))
	}
	for k, vv := range m {
		v.Additional[k] = vv
	}
	return nil
}
`))
)
