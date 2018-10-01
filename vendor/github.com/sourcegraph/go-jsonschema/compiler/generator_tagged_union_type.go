package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"text/template"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-jsonschema/jsonschema"
)

func (g *generator) emitTaggedUnionType(schema *jsonschema.Schema) ([]ast.Decl, []*ast.ImportSpec, error) {
	// Check that this schema can use the !go.taggedUnionType extension.
	if len(schema.Type) >= 2 || (len(schema.Type) == 1 && schema.Type[0] != jsonschema.ObjectType) || len(schema.OneOf) == 0 {
		return nil, nil, errors.New("invalid schema for use with !go.taggedUnionType extension")
	}

	// Next, try to find the discriminant property.
	oneOfSchemas := make([]*jsonschema.Schema, len(schema.OneOf))
	for i, s := range schema.OneOf {
		if s.Reference != nil {
			s = g.resolutions[s]
		}
		if len(s.Type) != 1 || s.Type[0] != jsonschema.ObjectType || s.Properties == nil || len(*s.Properties) == 0 {
			return nil, nil, errors.New("invalid oneOf schema for use with !go.taggedUnionType (must be an object with properties)")
		}
		oneOfSchemas[i] = s
	}
	// Find common fields. (Build up the set of the first schema's props, then intersect it with the
	// other schemas' props.)
	commonProperties := map[string]struct{}{}
	for name := range *oneOfSchemas[0].Properties {
		commonProperties[name] = struct{}{}
	}
	for _, s := range oneOfSchemas[1:] {
		for name := range commonProperties {
			if _, ok := (*s.Properties)[name]; !ok {
				delete(commonProperties, name)
			}
		}
	}
	if len(commonProperties) == 0 {
		return nil, nil, errors.New("no discriminant property found for !go.taggedUnionType extension")
	}
	if len(commonProperties) >= 2 {
		return nil, nil, fmt.Errorf("multiple discriminant properties found for !go.taggedUnionType extension: %q", commonProperties)
	}
	var discriminantPropName string
	for discriminantPropName = range commonProperties { // get first (and only) map key
	}
	// Ensure the common property can discriminate between the schemas (no value satisfies multiple
	// of the oneOf schemas).
	constValuesToSchema := make(map[string]*jsonschema.Schema, len(oneOfSchemas))
	discriminantValues := make([]string, 0, len(oneOfSchemas))
	for _, s := range oneOfSchemas {
		prop := (*s.Properties)[discriminantPropName]

		var required bool
		for _, req := range s.Required {
			if req == discriminantPropName {
				required = true
				break
			}
		}
		if !required {
			return nil, nil, fmt.Errorf("invalid oneOf schema for !go.taggedUnionType extension (discriminant property %q must be required)", discriminantPropName)
		}

		if len(prop.Type) != 1 || prop.Type[0] != jsonschema.StringType {
			return nil, nil, errors.New("invalid oneOf schema discriminant prop type for !go.taggedUnionType extension (must be string type)")
		}
		var constVal *interface{}
		if prop.Const != nil {
			constVal = prop.Const
		}
		for _, enumVal := range prop.Enum {
			if constVal == nil {
				v := (interface{}(enumVal))
				constVal = &v
			} else if !reflect.DeepEqual(*constVal, enumVal) {
				return nil, nil, fmt.Errorf("invalid oneOf schema discriminant prop enum value for !go.taggedUnionType extension (must have exactly 1 unique string value, got %q != %q)", *constVal, enumVal)
			}
		}
		if constVal == nil {
			return nil, nil, fmt.Errorf("no oneOf schema discriminant prop enum value for !go.taggedUnionType extension (must have either const or enum)")
		}
		switch ev := (*constVal).(type) {
		case string:
			if _, seen := constValuesToSchema[ev]; seen {
				return nil, nil, fmt.Errorf("invalid oneOf schema discriminant prop const value for !go.taggedUnionType extension (value %q is allowed by other type)", ev)
			}
			constValuesToSchema[ev] = s
			discriminantValues = append(discriminantValues, ev)
		default:
			return nil, nil, fmt.Errorf("invalid oneOf schema discriminant prop const value for !go.taggedUnionType extension (got %T not string)", ev)
		}
	}

	imports := importSpecs("fmt", "encoding/json", "errors")

	// Generate Go union type.
	fields := make([]*ast.Field, len(oneOfSchemas))
	fieldNames := make([]string, len(oneOfSchemas))
	fieldNameToConstValue := make(map[string]string, len(oneOfSchemas))
	for i, s := range oneOfSchemas {
		constValue := discriminantValues[i]
		fieldNames[i] = toGoName(constValue, "Const_")
		fieldNameToConstValue[fieldNames[i]] = constValue
		typeExpr, fieldImports, err := g.expr(s)
		if err != nil {
			return nil, nil, errors.WithMessage(err, fmt.Sprintf("failed to get type expression for !go.taggedUnionType union type %q", fieldNames[i]))
		}
		imports = append(imports, fieldImports...)
		fields[i] = &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(fieldNames[i])},
			Type:  &ast.StarExpr{X: typeExpr},
		}
	}
	goName, err := goNameForSchema(schema, g.schemas[schema])
	if err != nil {
		return nil, nil, err
	}
	typeDecl := &ast.GenDecl{
		Doc: docForSchema(schema, goName),
		Tok: token.TYPE,
		Specs: []ast.Spec{&ast.TypeSpec{
			Name: ast.NewIdent(goName),
			Type: &ast.StructType{Fields: &ast.FieldList{List: fields}},
		}},
	}

	// Generate MarshalJSON and UnmarshalJSON methods on the Go union type.
	templateData := map[string]interface{}{
		"fieldNames":            fieldNames,
		"discriminantPropName":  discriminantPropName,
		"discriminantValues":    discriminantValues,
		"fieldNameToConstValue": fieldNameToConstValue,
	}
	marshalJSONDecl, err := parseFuncLitToFuncDecl(executeTemplate(taggedUnionTypeMarshalJSONTemplate, templateData))
	if err != nil {
		return nil, nil, err
	}
	unmarshalJSONDecl, err := parseFuncLitToFuncDecl(executeTemplate(taggedUnionTypeUnmarshalJSONTemplate, templateData))
	if err != nil {
		return nil, nil, err
	}
	makeMethod(marshalJSONDecl, ast.NewIdent(goName), "MarshalJSON")
	makeMethod(unmarshalJSONDecl, &ast.StarExpr{X: ast.NewIdent(goName)}, "UnmarshalJSON")

	return []ast.Decl{typeDecl, marshalJSONDecl, unmarshalJSONDecl},
		imports,
		nil
}

var (
	taggedUnionTypeMarshalJSONTemplate = template.Must(template.New("").Parse(`
func() ([]byte, error) {
	{{range .fieldNames}}
	if v.{{.}} != nil {
		return json.Marshal(v.{{.}})
	}
	{{end}}
	return nil, errors.New("tagged union type must have exactly 1 non-nil field value")
}
`))
	taggedUnionTypeUnmarshalJSONTemplate = template.Must(template.New("").Parse(`
func(data []byte) error {
	var d struct {
		DiscriminantProperty string ` + "`" + `json:{{.discriminantPropName|printf "%q"}}` + "`" + `
	}
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	switch d.DiscriminantProperty {
	{{- range $fieldName, $constValue := .fieldNameToConstValue}}
	case {{$constValue|printf "%q"}}:
		return json.Unmarshal(data, &v.{{$fieldName}}){{end}}
	}
	return fmt.Errorf("tagged union type must have a %q property whose value is one of %s", {{.discriminantPropName|printf "%q"}}, {{.discriminantValues|printf "%#v"}})
}
`))
)
