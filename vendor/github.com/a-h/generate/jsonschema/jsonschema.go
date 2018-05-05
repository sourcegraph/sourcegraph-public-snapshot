// Package jsonschema provides primitives for extracting data from JSON schemas.
package jsonschema

import (
	"encoding/json"
	"errors"
	"strings"
)

// Schema represents JSON schema.
type Schema struct {
	// SchemaType identifies the schema version.
	// http://json-schema.org/draft-07/json-schema-core.html#rfc.section.7
	SchemaType string `json:"$schema"`

	// ID{04,06} is the schema URI identifier.
	// http://json-schema.org/draft-07/json-schema-core.html#rfc.section.9
	ID04 string `json:"id"`  // up to draft-04
	ID06 string `json:"$id"` // from draft-06 onwards

	// Title and Description state the intent of the schema.
	Title       string
	Description string

	// TypeValue is the schema instance type.
	// http://json-schema.org/draft-07/json-schema-validation.html#rfc.section.6.1.1
	TypeValue interface{} `json:"type"`

	// Definitions are inline re-usable schemas.
	// http://json-schema.org/draft-07/json-schema-validation.html#rfc.section.9
	Definitions map[string]*Schema

	// Properties, Required and AdditionalProperties describe an object's child instances.
	// http://json-schema.org/draft-07/json-schema-validation.html#rfc.section.6.5
	Properties           map[string]*Schema
	
	// Default value is applicable to a single sub-instance
	// http://json-schema.org/draft-07/json-schema-validation.html#rfc.section.10.2
	Default			 	 interface{} `json:"default"`
	Required             []string
	AdditionalProperties AdditionalProperties

	// Reference is a URI reference to a schema.
	// http://json-schema.org/draft-07/json-schema-core.html#rfc.section.8
	Reference string `json:"$ref"`

	// Items represents the types that are permitted in the array.
	// http://json-schema.org/draft-07/json-schema-validation.html#rfc.section.6.4
	Items *Schema

	// NameCount is the number of times the instance name was encountered accross the schema.
	NameCount int `json:"-" `
}

// ID returns the schema URI id.
func (s *Schema) ID() string {
	// prefer "$id" over "id"
	if s.ID06 == "" && s.ID04 != "" {
		return s.ID04
	}
	return s.ID06
}

// Type returns the type which is permitted or an empty string if the type field is missing.
// The 'type' field in JSON schema also allows for a single string value or an array of strings.
// Examples:
//   "a" => "a", false
//   [] => "", false
//   ["a"] => "a", false
//   ["a", "b"] => "a", true
func (s *Schema) Type() (firstOrDefault string, multiple bool) {
	// We've got a single value, e.g. { "type": "object" }
	if ts, ok := s.TypeValue.(string); ok {
		firstOrDefault = ts
		multiple = false
		return
	}

	// We could have multiple types in the type value, e.g. { "type": [ "object", "array" ] }
	if a, ok := s.TypeValue.([]interface{}); ok {
		multiple = len(a) > 1
		for _, n := range a {
			if s, ok := n.(string); ok {
				firstOrDefault = s
				return
			}
		}
	}

	return "", multiple
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
	t, multiple := s.Type()
	if multiple {
		// If we have more than one possible type for this field, the result is an interface{} in the struct definition.
		return
	}

	// Add root schemas composed only of a simple type
	if !(t == "object" || t == "") && path == "#" {
		types[path] = s
	}

	if t == "array" {
		if s.Items != nil {
			if path == "#" {
				path += "/arrayitems"
			}
			addTypeAndChildrenToMap(path, name, s.Items, types)
		}
		return
	}

	namePrefix := "/" + name
	// Don't add the name into the root, or we end up with an extra slash.
	if (path == "#" || path == "#/arrayitems") && name == "" {
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

	if len(s.Properties) > 0 || t == "object" {
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

// AdditionalProperties handles additional properties present in the JSON schema.
type AdditionalProperties []*Schema

// UnmarshalJSON handles unmarshalling AdditionalProperties from JSON.
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
