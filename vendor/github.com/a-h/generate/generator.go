// Package generate creates Go structs from JSON schemas.
package generate

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"
	"unicode"

	"github.com/a-h/generate/jsonschema"
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

// CreateTypes creates types from the JSON schemas, keyed by the golang name.
func (g *Generator) CreateTypes() (structs map[string]Struct, aliases map[string]Field, err error) {
	schemaIDs := make([]*url.URL, len(g.schemas))
	for i, schema := range g.schemas {
		if schema.ID() != "" {
			schemaIDs[i], err = url.Parse(schema.ID())
			if err != nil {
				return
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
	aliases = make(map[string]Field)
	errs := []error{}

	for _, typeKey := range getOrderedKeyNamesFromSchemaMap(types) {
		v := types[typeKey]

		if v.TypeValue == "object" || v.TypeValue == nil {
			s, errtype := createStruct(typeKey, v, types)
			if errtype != nil {
				errs = append(errs, errtype...)
			}

			if _, ok := structs[s.Name]; ok {
				errs = append(errs, errors.New("Duplicate struct name : "+s.Name))
			}

			structs[s.Name] = s
		} else {
			a, errtype := createAlias(typeKey, v, types)
			if errtype != nil {
				errs = append(errs, errtype...)
			}

			aliases[a.Name] = a
		}
	}

	if len(errs) > 0 {
		err = errors.New(joinErrors(errs))
	}
	return
}

// createStruct creates a struct type from the JSON schema.
func createStruct(typeKey string, schema *jsonschema.Schema, types map[string]*jsonschema.Schema) (s Struct, errs []error) {
	typeKeyURI, err := url.Parse(typeKey)
	if err != nil {
		errs = append(errs, err)
	}

	fields, err := getFields(typeKeyURI, schema.Properties, types, schema.Required)
	if err != nil {
		errs = append(errs, err)
	}

	structName := getTypeName(typeKeyURI, schema, 1)
	if err != nil {
		errs = append(errs, err)
	}

	s = Struct{
		ID:          typeKey,
		Name:        structName,
		Description: schema.Description,
		Fields:      fields,
	}

	return
}

// createAlias creates a simple alias type from the JSON schema.
func createAlias(typeKey string, schema *jsonschema.Schema, types map[string]*jsonschema.Schema) (a Field, errs []error) {
	typeKeyURI, err := url.Parse(typeKey)
	if err != nil {
		errs = append(errs, err)
	}

	aliasName := getTypeName(typeKeyURI, schema, 1)
	if err != nil {
		errs = append(errs, err)
	}

	tn, err := getTypeForField(typeKeyURI, typeKey, aliasName, schema, types, true)
	if err != nil {
		errs = append(errs, err)
	}

	a = Field{
		Name:     aliasName,
		JSONName: "",
		Type:     tn,
		Required: false,
	}

	return
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

func getFields(parentTypeKey *url.URL, properties map[string]*jsonschema.Schema,
	types map[string]*jsonschema.Schema, requiredFields []string) (field map[string]Field, err error) {
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
		return fields, fmt.Errorf("missing types for %s with errors %s",
			strings.Join(missingTypes, ","), joinErrors(errors))
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

func getTypeForField(parentTypeKey *url.URL, fieldName string, fieldGoName string,
	fieldSchema *jsonschema.Schema, types map[string]*jsonschema.Schema, pointer bool) (typeName string, err error) {
	// If there's no schema, or the field can be more than one type, we have to use interface{} and allow the caller to use type assertions to determine
	// the actual underlying type.
	if fieldSchema == nil {
		return "interface{}", nil
	}

	majorType, multiple := fieldSchema.Type()
	if multiple {
		return "interface{}", nil
	}

	var subType string

	// Look up by named reference.
	if fieldSchema.Reference != "" {
		// Resolve reference URI relative to schema's ID (URI).
		ref, err := url.Parse(fieldSchema.Reference)
		if err != nil {
			return "", err
		}
		ref = parentTypeKey.ResolveReference(ref)

		if t, ok := types[ref.String()]; ok {
			sn := getTypeName(ref, t, 1)

			majorType = "object"
			subType = sn
		} else {
			return "", fmt.Errorf("failed to resolve the reference %s", ref)
		}
	}

	// Look up any embedded types.
	if subType == "" && (majorType == "object" || majorType == "") {
		if len(fieldSchema.Properties) == 0 && len(fieldSchema.AdditionalProperties) > 0 {
			if len(fieldSchema.AdditionalProperties) == 1 {
				sn, _ := getTypeForField(parentTypeKey, fieldName, fieldGoName,
					fieldSchema.AdditionalProperties[0], types, pointer)
				subType = "map[string]" + sn
				pointer = false
			} else {
				subType = "map[string]interface{}"
				pointer = false
			}
		} else {
			ref := joinURLFragmentPath(parentTypeKey, "properties/"+fieldName)

			// Root schema without properties, try array item instead
			if _, ok := types[ref.String()]; !ok && isRootSchemaKey(parentTypeKey) {
				ref = joinURLFragmentPath(parentTypeKey, "arrayitems")
			}

			if parentType, ok := types[ref.String()]; ok {
				sn := getTypeName(ref, parentType, 1)
				subType = sn
			} else {
				subType = "undefined"
			}
		}
	}

	// Find named array references.
	if majorType == "array" {
		s, _ := getTypeForField(parentTypeKey, fieldName, fieldGoName, fieldSchema.Items, types, true)
		subType = s
	}

	name, err := getPrimitiveTypeName(majorType, subType, pointer)

	if err != nil {
		return name, fmt.Errorf("failed to get the type for %s with error %s",
			fieldGoName,
			err.Error())
	}

	return name, nil
}

// isRootSchemaKey returns whether a given type key references the root schema.
func isRootSchemaKey(typeKey *url.URL) bool {
	return typeKey.Fragment == ""
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

	return "undefined", fmt.Errorf("failed to get a primitive type for schemaType %s and subtype %s",
		schemaType, subType)
}

// getTypeName makes a golang type name from an input reference in the form of #/definitions/address
// The parts refers to the number of segments from the end to take as the name.
func getTypeName(reference *url.URL, structType *jsonschema.Schema, n int) string {
	if len(structType.Title) > 0 {
		return getGolangName(structType.Title)
	}

	if isRootSchemaKey(reference) {
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

	for i, v := range splitOnAll(s, isNotAGoNameCharacter) {
		if i == 0 && strings.IndexAny(v, "0123456789") == 0 {
			// Go types are not allowed to start with a number, lets prefix with an underscore.
			buf.WriteRune('_')
		}
		buf.WriteString(capitaliseFirstLetter(v))
	}

	return buf.String()
}

func splitOnAll(s string, shouldSplit func(r rune) bool) []string {
	rv := []string{}

	buf := bytes.NewBuffer([]byte{})
	for _, c := range s {
		if shouldSplit(c) {
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

func isNotAGoNameCharacter(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return false
	}
	return true
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
	// The golang type of the field, e.g. a built-in type like "string" or the name of a struct generated
	// from the JSON schema.
	Type string
	// Required is set to true when the field is required.
	Required bool
}
