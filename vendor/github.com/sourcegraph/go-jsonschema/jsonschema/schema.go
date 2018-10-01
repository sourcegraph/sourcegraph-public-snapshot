package jsonschema

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
)

// Schema is a JSON Schema draft-07 document (as specified in
// [draft-handrews-json-schema-01](https://tools.ietf.org/html/draft-handrews-json-schema-01)).
type Schema struct {
	Comment              *string                      `json:"$comment,omitempty"`
	ID                   *string                      `json:"$id,omitempty"`
	Reference            *string                      `json:"$ref,omitempty"`
	SchemaRef            *string                      `json:"$schema,omitempty"`
	AdditionalItems      *Schema                      `json:"additionalItems,omitempty"`
	AdditionalProperties *Schema                      `json:"additionalProperties,omitempty"`
	AllOf                []*Schema                    `json:"allOf,omitempty"`
	AnyOf                []*Schema                    `json:"anyOf,omitempty"`
	Const                *interface{}                 `json:"const,omitempty"`
	Contains             *Schema                      `json:"contains,omitempty"`
	Default              *interface{}                 `json:"default,omitempty"`
	Definitions          *map[string]*Schema          `json:"definitions,omitempty"`
	Dependencies         *map[string]*DependencyValue `json:"dependencies,omitempty"`
	Description          *string                      `json:"description,omitempty"`
	Else                 *Schema                      `json:"else,omitempty"`
	Enum                 EnumList                     `json:"enum,omitempty"`
	ExclusiveMaximum     *float64                     `json:"exclusiveMaximum,omitempty"`
	ExclusiveMinimum     *float64                     `json:"exclusiveMinimum,omitempty"`
	Format               *Format                      `json:"format,omitempty"`
	If                   *Schema                      `json:"if,omitempty"`
	Items                *SchemaOrSchemaList          `json:"items,omitempty"`
	MaxItems             *int64                       `json:"maxItems,omitempty"`
	MaxLength            *int64                       `json:"maxLength,omitempty"`
	MaxProperties        *int64                       `json:"maxProperties,omitempty"`
	Maximum              *float64                     `json:"maximum,omitempty"`
	MinItems             *int64                       `json:"minItems,omitempty"`
	MinLength            *int64                       `json:"minLength,omitempty"`
	MinProperties        *int64                       `json:"minProperties,omitempty"`
	Minimum              *float64                     `json:"minimum,omitempty"`
	MultipleOf           *float64                     `json:"multipleOf,omitempty"`
	Not                  *Schema                      `json:"not,omitempty"`
	OneOf                []*Schema                    `json:"oneOf,omitempty"`
	Pattern              *string                      `json:"pattern,omitempty"`
	PatternProperties    *map[string]*Schema          `json:"patternProperties,omitempty"`
	Properties           *map[string]*Schema          `json:"properties,omitempty"`
	PropertyNames        *Schema                      `json:"propertyNames,omitempty"`
	Required             []string                     `json:"required,omitempty"`
	Then                 *Schema                      `json:"then,omitempty"`
	Title                *string                      `json:"title,omitempty"`
	Type                 PrimitiveTypeList            `json:"type,omitempty"`
	UniqueItems          *bool                        `json:"uniqueItems,omitempty"`

	IsEmpty   bool `json:"-"` // the schema is "true"
	IsNegated bool `json:"-"` // the schema is "false"

	// Go contains Go-specific extensions that JSON Schema authors can specify.
	Go *struct {
		TaggedUnionType bool `json:"taggedUnionType,omitempty"`
		Pointer         bool `json:"pointer,omitempty"`
	} `json:"!go,omitempty"`
}

// IsRequiredProperty reports whether propertyName is a required property for instances of this
// schema.
func (s *Schema) IsRequiredProperty(propertyName string) bool {
	// TODO(sqs): This ignores complexity like dependencies, allOf, etc.
	for _, p := range s.Required {
		if p == propertyName {
			return true
		}
	}
	return false
}

var trueBytes = []byte("true")
var falseBytes = []byte("false")

// MarshalJSON implements json.Marshaler.
func (s *Schema) MarshalJSON() ([]byte, error) {
	switch {
	case s.IsNegated:
		return falseBytes, nil
	case s.IsEmpty:
		return trueBytes, nil
	}
	type schema2 Schema
	return json.Marshal((*schema2)(s))
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Schema) UnmarshalJSON(data []byte) error {
	switch {
	case bytes.Equal(data, trueBytes):
		*s = Schema{IsEmpty: true}
	case bytes.Equal(data, falseBytes):
		*s = Schema{IsNegated: true}
	default:
		type schema2 Schema
		if err := json.Unmarshal(data, (*schema2)(s)); err != nil {
			return errors.WithMessage(err, "failed to unmarshal JSON Schema")
		}
	}
	return nil
}

// SchemaOrSchemaList represents a value that can be either a valid JSON Schema or an array of valid
// JSON Schemas.
//
// Exactly 1 field (Schema or Schemas) is set.
//
// The ["items"
// keyword](https://tools.ietf.org/html/draft-handrews-json-schema-validation-01#section-6.4.1) is
// the only keyword that this is used for.
type SchemaOrSchemaList struct {
	Schema  *Schema
	Schemas []*Schema
}

// MarshalJSON implements json.Marshaler.
func (s *SchemaOrSchemaList) MarshalJSON() ([]byte, error) {
	if s.Schema != nil {
		return json.Marshal(s.Schema)
	}
	return json.Marshal(s.Schemas)
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *SchemaOrSchemaList) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '[' {
		return json.Unmarshal(data, &s.Schemas)
	}
	return json.Unmarshal(data, &s.Schema)
}

type DependencyValue struct {
	Schema             *Schema
	RequiredProperties []string
}

// MarshalJSON implements json.Marshaler.
func (v *DependencyValue) MarshalJSON() ([]byte, error) {
	if v.Schema != nil {
		return json.Marshal(v.Schema)
	}
	return json.Marshal(v.RequiredProperties)
}

// UnmarshalJSON implements json.Unmarshaler.
func (v *DependencyValue) UnmarshalJSON(data []byte) error {
	*v = DependencyValue{}
	if len(data) > 0 && data[0] == '[' {
		return json.Unmarshal(data, &v.RequiredProperties)
	}
	return json.Unmarshal(data, &v.Schema)
}
