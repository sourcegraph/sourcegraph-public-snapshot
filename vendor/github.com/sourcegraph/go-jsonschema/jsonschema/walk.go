package jsonschema

// A Visitor's Visit method is invoked for each schema (with the relative reference tokens
// identifying it) to encountered by Walk.  If the result visitor w is not nil, Walk visits each of
// the subschemas of schema with the visitor w, followed by a call of w.Visit(nil).
type Visitor interface {
	Visit(schema *Schema, rel []ReferenceToken) (w Visitor)
}

// Walk traverses a JSON Schema in depth-first order. It starts by calling v.Visit(schema); schema
// must not be nil. If the visitor w returned by v.Visit(schema) is not nil, Walk is invoked
// recursively with visitor w for each of the non-nil children of schema, followed by a call of
// w.Visit(nil).
func Walk(v Visitor, schema *Schema) {
	walk(v, schema, nil)
}

func walk(v Visitor, schema *Schema, rel []ReferenceToken) {
	if v = v.Visit(schema, rel); v == nil {
		return
	}

	// TODO(sqs): When unmarshaling fields that can be singular or an array in Go, we don't record
	// the input type (singular or array). This means that the index ReferenceTokens we use below
	// might be incorrect when i==0 and the list only has a single item.

	// Walk children. The order of the fields matches the their order in the Schema struct type
	// definition.
	if schema.AdditionalItems != nil {
		walk(v, schema.AdditionalItems, []ReferenceToken{{Name: "additionalItems"}})
	}
	if schema.AdditionalProperties != nil {
		walk(v, schema.AdditionalProperties, []ReferenceToken{{Name: "additionalProperties"}})
	}
	for i, s := range schema.AllOf {
		walk(v, s, []ReferenceToken{{Name: "allOf", Keyword: true}, {Index: i}})
	}
	for i, s := range schema.AnyOf {
		walk(v, s, []ReferenceToken{{Name: "anyOf", Keyword: true}, {Index: i}})
	}
	if schema.Contains != nil {
		walk(v, schema.Contains, []ReferenceToken{{Name: "contains", Keyword: true}})
	}
	if schema.Definitions != nil {
		for name, s := range *schema.Definitions {
			walk(v, s, []ReferenceToken{{Name: "definitions", Keyword: true}, {Name: name}})
		}
	}
	if schema.Dependencies != nil {
		for name, s := range *schema.Dependencies {
			if s != nil && s.Schema != nil {
				walk(v, s.Schema, []ReferenceToken{{Name: "dependencies", Keyword: true}, {Name: name}})
			}
		}
	}
	if schema.Else != nil {
		walk(v, schema.Else, []ReferenceToken{{Name: "else", Keyword: true}})
	}
	if schema.If != nil {
		walk(v, schema.If, []ReferenceToken{{Name: "if", Keyword: true}})
	}
	if schema.Items != nil {
		if schema.Items.Schema != nil {
			walk(v, schema.Items.Schema, []ReferenceToken{{Name: "items", Keyword: true}})
		}
		for i, s := range schema.Items.Schemas {
			walk(v, s, []ReferenceToken{{Name: "items", Keyword: true}, {Index: i}})
		}
	}
	for i, s := range schema.OneOf {
		walk(v, s, []ReferenceToken{{Name: "oneOf", Keyword: true}, {Index: i}})
	}
	if schema.PatternProperties != nil {
		for name, s := range *schema.PatternProperties {
			walk(v, s, []ReferenceToken{{Name: "patternProperties", Keyword: true}, {Name: name}})
		}
	}
	if schema.Properties != nil {
		for name, s := range *schema.Properties {
			walk(v, s, []ReferenceToken{{Name: "properties", Keyword: true}, {Name: name}})
		}
	}
	if schema.PropertyNames != nil {
		walk(v, schema.PropertyNames, []ReferenceToken{{Name: "propertyNames", Keyword: true}})
	}
	if schema.Then != nil {
		walk(v, schema.Then, []ReferenceToken{{Name: "then", Keyword: true}})
	}

	v.Visit(nil, rel)
}
