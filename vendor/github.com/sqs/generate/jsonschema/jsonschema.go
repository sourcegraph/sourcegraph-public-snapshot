package jsonschema

import (
	"encoding/json"
	"errors"
	"strings"
)

type AdditionalProperties []*Schema

// Schema represents JSON schema.
type Schema struct {
	SchemaType           string `json:"$schema"`
	Title                string `json:"title"`
	ID                   string `json:"id"`
	Type                 Type   `json:"type"`
	Description          string `json:"description"`
	Definitions          map[string]*Schema
	Properties           map[string]*Schema
	AdditionalProperties AdditionalProperties
	Reference            string `json:"$ref"`
	// Items represents the types that are permitted in the array.
	Items     *Schema  `json:"items"`
	Required  []string `json:"required"`
	NameCount int      `json:"-" `
}

// Parse parses a JSON schema from a string.
func Parse(schema string) (*Schema, error) {
	s := &Schema{}
	err := json.Unmarshal([]byte(schema), s)

	if err != nil {
		return s, err
	}

	if s.SchemaType == "" {
		return s, errors.New("JSON schema must have a $schema key")
	}

	return s, err
}

// Type represents a JSON Schema type, which can be a string or an array of strings.
type Type []string

var _ json.Unmarshaler = (*Type)(nil)

func (t Type) MarshalJSON() ([]byte, error) {
	if len(t) == 0 {
		return nil, errors.New("Cannot marshal empty type")
	}
	if len(t) == 1 {
		return json.Marshal(t[0])
	}
	return json.Marshal([]string(t))
}

func (t *Type) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("Cannot unmarshal JSON schema type from empty string")
	}
	if data[0] == '[' {
		return json.Unmarshal(data, (*[]string)(t))
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*t = Type{s}
	return nil
}

// Is returns true if the JSON Schema type exactly matches the argument.
func (t Type) Is(v interface{}) bool {
	switch v := v.(type) {
	case string:
		return len(t) == 1 && t[0] == v
	case Type:
		if len(v) != len(t) {
			return false
		}
		vm := make(map[string]struct{})
		for _, e := range v {
			vm[e] = struct{}{}
		}
		if len(vm) != len(t) {
			return false
		}
		for _, e := range t {
			if _, in := vm[e]; !in {
				return false
			}
		}
		return true
	case []string:
		if len(v) != len(t) {
			return false
		}
		vm := make(map[string]struct{})
		for _, e := range v {
			vm[e] = struct{}{}
		}
		if len(vm) != len(t) {
			return false
		}
		for _, e := range t {
			if _, in := vm[e]; !in {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// Set sets the value of the JSON Schema type
func (t *Type) Set(s ...string) {
	*t = s
}

// Has returns true if the JSON Schema type includes the type specified by the argument.
// For example, if the JSON Schema type is ["string", "null"] and the argument is "string",
// this function returns true.
func (t Type) Has(s string) bool {
	for _, u := range t {
		if u == s {
			return true
		}
	}
	return false
}

func (ap *AdditionalProperties) UnmarshalJSON(data []byte) error {
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		return nil
	}

	// support anyOf, allOf, oneOf
	a := map[string][]*Schema{}
	if err := json.Unmarshal(data, &a); err == nil {
		for k, v := range a {
			if k == "oneOf" || k == "allOf" || k == "anyOf" {
				*ap = append(*ap, v...)
			}
		}
		return nil
	}

	s := Schema{}
	err := json.Unmarshal(data, &s)
	if err == nil {
		*ap = append(*ap, &s)
	}
	return err
}

// ExtractTypes creates a map of defined types within the schema.
func (s *Schema) ExtractTypes() map[string]*Schema {
	types := make(map[string]*Schema)

	addTypeAndChildrenToMap("#", "", s, types)

	counts := make(map[string]int)
	for path, t := range types {
		parts := strings.Split(path, "/")
		name := parts[len(parts)-1]
		counts[name] = counts[name] + 1
		t.NameCount = counts[name]
	}

	return types
}

func addTypeAndChildrenToMap(path string, name string, s *Schema, types map[string]*Schema) {
	if s.Type.Is("array") {
		arrayTypeName := s.ID

		// If there's no ID, try the title instead.
		if arrayTypeName == "" {
			if s.Items != nil {
				arrayTypeName = s.Items.Title
			}
		}

		// If there's no title, use the property name to name the type we're creating.
		if arrayTypeName == "" {
			arrayTypeName = name
		}

		if s.Items != nil {
			addTypeAndChildrenToMap(path, arrayTypeName, s.Items, types)
		}
		return
	}

	namePrefix := "/" + name
	// Don't add the name into the root, or we end up with an extra slash.
	if path == "#" && name == "" {
		namePrefix = ""
	}

	if len(s.Properties) == 0 && len(s.AdditionalProperties) > 0 {
		// if we have more than one valid type in additionalProperties, we can disregard them
		// as we will render as a weakly-typed map i.e map[string]interface{}
		if len(s.AdditionalProperties) == 1 {
			addTypeAndChildrenToMap(path, name, s.AdditionalProperties[0], types)
		}
		return
	}

	if len(s.Properties) > 0 || s.Type.Is("object") {
		types[path+namePrefix] = s
	}

	if s.Definitions != nil {
		for k, d := range s.Definitions {
			addTypeAndChildrenToMap(path+namePrefix+"/definitions", k, d, types)
		}
	}

	if s.Properties != nil {
		for k, d := range s.Properties {
			// Only add the children as their own type if they have properties at all.
			addTypeAndChildrenToMap(path+namePrefix+"/properties", k, d, types)
		}
	}

}

// ListReferences lists all of the references in a schema.
func (s *Schema) ListReferences() map[string]bool {
	m := make(map[string]bool)
	addReferencesToMap(s, m)
	return m
}

func addReferencesToMap(s *Schema, m map[string]bool) {
	if s.Reference != "" {
		m[s.Reference] = true
	}

	if s.Definitions != nil {
		for _, d := range s.Definitions {
			addReferencesToMap(d, m)
		}
	}

	if s.Properties != nil {
		for _, p := range s.Properties {
			addReferencesToMap(p, m)
		}
	}

	if s.Items != nil {
		addReferencesToMap(s.Items, m)
	}
}
