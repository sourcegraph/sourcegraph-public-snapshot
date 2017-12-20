package generate

import (
	"bytes"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"

	"errors"

	"github.com/sqs/generate/jsonschema"
)

// Generator will produce structs from the JSON schema.
type Generator struct {
	schemas []*jsonschema.Schema
}

// New creates an instance of a generator which will produce structs.
func New(schemas ...*jsonschema.Schema) *Generator {
	return &Generator{
		schemas: schemas,
	}
}

// CreateStructs creates structs from the JSON schemas, keyed by the golang name.
func (g *Generator) CreateStructs() (structs map[string]Struct, err error) {
	schemaIDs := make([]*url.URL, len(g.schemas))
	for i, schema := range g.schemas {
		if schema.ID != "" {
			schemaIDs[i], err = url.Parse(schema.ID)
			if err != nil {
				return nil, err
			}
		}
	}

	// Extract nested and complex types from the JSON schemas.
	types := map[string]*jsonschema.Schema{}
	for i, schema := range g.schemas {
		for name, typ := range schema.ExtractTypes() {
			if schemaIDs[i] != nil {
				name = schemaIDs[i].ResolveReference(&url.URL{Fragment: name[1:]}).String()
			}
			if typ.Reference == "" {
				types[name] = typ
			}
		}
	}

	structs = make(map[string]Struct)
	errs := []error{}

	for _, typeKey := range getOrderedKeyNamesFromSchemaMap(types) {
		v := types[typeKey]

		typeKeyURI, err := url.Parse(typeKey)
		if err != nil {
			errs = append(errs, err)
		}
		fields, err := getFields(typeKeyURI, v.Properties, types, v.Required)

		if err != nil {
			errs = append(errs, err)
		}

		structName := getStructName(typeKeyURI, v, 1)

		if err != nil {
			errs = append(errs, err)
		}

		s := Struct{
			ID:          typeKey,
			Name:        structName,
			Description: v.Description,
			Fields:      fields,
		}

		if _, ok := structs[s.Name]; ok {
			errs = append(errs, errors.New("Duplicate struct name : "+s.Name))
		}

		structs[s.Name] = s
	}

	if len(errs) > 0 {
		return structs, errors.New(joinErrors(errs))
	}

	return structs, nil
}

func joinErrors(errs []error) string {
	var buffer bytes.Buffer

	for idx, err := range errs {
		buffer.WriteString(err.Error())

		if idx+1 < len(errs) {
			buffer.WriteString(", ")
		}
	}

	return buffer.String()
}

func getOrderedKeyNamesFromSchemaMap(m map[string]*jsonschema.Schema) []string {
	keys := make([]string, len(m))
	idx := 0
	for k := range m {
		keys[idx] = k
		idx++
	}
	sort.Strings(keys)
	return keys
}

func getFields(parentTypeKey *url.URL, properties map[string]*jsonschema.Schema, types map[string]*jsonschema.Schema, requiredFields []string) (field map[string]Field, err error) {
	fields := map[string]Field{}

	missingTypes := []string{}
	errors := []error{}

	for _, fieldName := range getOrderedKeyNamesFromSchemaMap(properties) {
		v := properties[fieldName]

		golangName := getGolangName(fieldName)
		tn, err := getTypeForField(parentTypeKey, fieldName, golangName, v, types, true)

		if err != nil {
			missingTypes = append(missingTypes, golangName)
			errors = append(errors, err)
		}

		f := Field{
			Name:     golangName,
			JSONName: fieldName,
			// Look up the types, try references first, then drop to the built-in types.
			Type:     tn,
			Required: contains(requiredFields, fieldName),
		}

		fields[f.Name] = f
	}

	if len(missingTypes) > 0 {
		return fields, fmt.Errorf("missing types for %s with errors %s", strings.Join(missingTypes, ","), joinErrors(errors))
	}

	return fields, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func getTypeForField(parentTypeKey *url.URL, fieldName string, fieldGoName string, fieldSchema *jsonschema.Schema, types map[string]*jsonschema.Schema, pointer bool) (typeName string, err error) {
	if fieldSchema == nil {
		return "interface{}", nil
	}

	majorType := fieldSchema.Type
	subType := ""

	// Look up by named reference.
	if fieldSchema.Reference != "" {
		// Resolve reference URI relative to schema's ID (URI).
		ref, err := url.Parse(fieldSchema.Reference)
		if err != nil {
			return "", err
		}
		ref = parentTypeKey.ResolveReference(ref)

		if t, ok := types[ref.String()]; ok {
			sn := getStructName(ref, t, 1)

			majorType = "object"
			subType = sn
		} else {
			return "", fmt.Errorf("Failed to resolve the reference %s", ref)
		}
	}

	// Look up any embedded types.
	if subType == "" && majorType == "object" {
		if len(fieldSchema.Properties) == 0 && len(fieldSchema.AdditionalProperties) > 0 {
			if len(fieldSchema.AdditionalProperties) == 1 {
				sn, _ := getTypeForField(parentTypeKey, fieldName, fieldGoName, fieldSchema.AdditionalProperties[0], types, pointer)
				subType = "map[string]" + sn
				pointer = false
			} else {
				subType = "map[string]interface{}"
				pointer = false
			}
		} else {
			ref := joinURLFragmentPath(parentTypeKey, "properties/"+fieldName)
			if parentType, ok := types[ref.String()]; ok {
				sn := getStructName(ref, parentType, 1)
				subType = sn
			} else {
				subType = "undefined"
			}
		}
	}

	// Find named array references.
	if majorType == "array" {
		s, _ := getTypeForField(parentTypeKey, fieldName, fieldGoName, fieldSchema.Items, types, false)
		subType = s
	}

	name, err := getPrimitiveTypeName(majorType, subType, pointer)

	if err != nil {
		return name, fmt.Errorf("Failed to get the type for %s with error %s",
			fieldGoName,
			err.Error())
	}

	return name, nil
}

// joinURLFragmentPath joins elem onto u.Fragment, adding a separating slash.
func joinURLFragmentPath(base *url.URL, elem string) *url.URL {
	url := *base
	if url.Fragment == "" {
		url.Fragment = "/"
	}
	url.Fragment = path.Join(url.Fragment, elem)
	return &url
}

func getPrimitiveTypeName(schemaType string, subType string, pointer bool) (name string, err error) {
	switch schemaType {
	case "array":
		if subType == "" {
			return "error_creating_array", errors.New("can't create an array of an empty subtype")
		}
		return "[]" + subType, nil
	case "boolean":
		return "bool", nil
	case "integer":
		return "int", nil
	case "number":
		return "float64", nil
	case "null":
		return "nil", nil
	case "object":
		if pointer {
			return "*" + subType, nil
		}

		return subType, nil
	case "string":
		return "string", nil
	}

	return "undefined", fmt.Errorf("failed to get a primitive type for schemaType %s and subtype %s", schemaType, subType)
}

// getStructName makes a golang struct name from an input reference in the form of #/definitions/address
// The parts refers to the number of segments from the end to take as the name.
func getStructName(reference *url.URL, structType *jsonschema.Schema, n int) string {
	if reference.Fragment == "" {
		rootName := structType.Title

		if rootName == "" {
			rootName = structType.Description
		}

		if rootName == "" {
			rootName = "Root"
		}

		return getGolangName(rootName)
	}

	parts := strings.Split(reference.Fragment, "/")
	partsToUse := parts[len(parts)-n:]

	sb := bytes.Buffer{}

	for _, p := range partsToUse {
		sb.WriteString(getGolangName(p))
	}

	result := sb.String()

	if result == "" {
		return "Root"
	}

	if structType.NameCount > 1 {
		result = fmt.Sprintf("%v%v", result, structType.NameCount)
	}

	return result
}

// getGolangName strips invalid characters out of golang struct or field names.
func getGolangName(s string) string {
	buf := bytes.NewBuffer([]byte{})

	for _, v := range splitOnAll(s, '_', ' ', '.', '-', ':') {
		buf.WriteString(capitaliseFirstLetter(v))
	}

	return buf.String()
}

func splitOnAll(s string, splitItems ...rune) []string {
	rv := []string{}

	buf := bytes.NewBuffer([]byte{})
	for _, c := range s {
		if matches(c, splitItems) {
			rv = append(rv, buf.String())
			buf.Reset()
		} else {
			buf.WriteRune(c)
		}
	}
	if buf.Len() > 0 {
		rv = append(rv, buf.String())
	}

	return rv
}

func matches(c rune, any []rune) bool {
	for _, a := range any {
		if a == c {
			return true
		}
	}
	return false
}

func capitaliseFirstLetter(s string) string {
	if s == "" {
		return s
	}

	prefix := s[0:1]
	suffix := s[1:]
	return strings.ToUpper(prefix) + suffix
}

// Struct defines the data required to generate a struct in Go.
type Struct struct {
	// The ID within the JSON schema, e.g. #/definitions/address
	ID string
	// The golang name, e.g. "Address"
	Name string
	// Description of the struct
	Description string
	Fields      map[string]Field
}

// Field defines the data required to generate a field in Go.
type Field struct {
	// The golang name, e.g. "Address1"
	Name string
	// The JSON name, e.g. "address1"
	JSONName string
	// The golang type of the field, e.g. a built-in type like "string" or the name of a struct generated from the JSON schema.
	Type string
	// Required is set to true when the field is required.
	Required bool
}
