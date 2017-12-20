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
	Type                 string `json:"type"`
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
	if s.Type == "array" {
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

	if len(s.Properties) > 0 || s.Type == "object" {
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
