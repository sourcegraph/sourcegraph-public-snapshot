package scim

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/internal/patch"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
)

// unmarshal unifies the unmarshal of the requests.
func unmarshal(data []byte, v interface{}) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	return d.Decode(v)
}

// ResourceType specifies the metadata about a resource type.
type ResourceType struct {
	// ID is the resource type's server unique id. This is often the same value as the "name" attribute.
	ID optional.String
	// Name is the resource type name. This name is referenced by the "meta.resourceType" attribute in all resources.
	Name string
	// Description is the resource type's human-readable description.
	Description optional.String
	// Endpoint is the resource type's HTTP-addressable endpoint relative to the Base URL of the service provider,
	// e.g., "/Users".
	Endpoint string
	// Schema is the resource type's primary/base schema.
	Schema schema.Schema
	// SchemaExtensions is a list of the resource type's schema extensions.
	SchemaExtensions []SchemaExtension

	// Handler is the set of callback method that connect the SCIM server with a provider of the resource type.
	Handler ResourceHandler
}

func (t ResourceType) getRaw() map[string]interface{} {
	return map[string]interface{}{
		"schemas":          []string{"urn:ietf:params:scim:schemas:core:2.0:ResourceType"},
		"id":               t.ID.Value(),
		"name":             t.Name,
		"description":      t.Description.Value(),
		"endpoint":         t.Endpoint,
		"schema":           t.Schema.ID,
		"schemaExtensions": t.getRawSchemaExtensions(),
	}
}

func (t ResourceType) getRawSchemaExtensions() []map[string]interface{} {
	schemas := make([]map[string]interface{}, 0)
	for _, e := range t.SchemaExtensions {
		schemas = append(schemas, map[string]interface{}{
			"schema":   e.Schema.ID,
			"required": e.Required,
		})
	}
	return schemas
}

func (t ResourceType) getSchemaExtensions() []schema.Schema {
	var extensions []schema.Schema
	for _, e := range t.SchemaExtensions {
		extensions = append(extensions, e.Schema)
	}
	return extensions
}

func (t ResourceType) schemaWithCommon() schema.Schema {
	s := t.Schema

	externalID := schema.SimpleCoreAttribute(
		schema.SimpleStringParams(schema.StringParams{
			CaseExact:  true,
			Mutability: schema.AttributeMutabilityReadWrite(),
			Name:       schema.CommonAttributeExternalID,
			Uniqueness: schema.AttributeUniquenessNone(),
		}),
	)

	s.Attributes = append(s.Attributes, externalID)

	return s
}

func (t ResourceType) validate(raw []byte) (ResourceAttributes, *errors.ScimError) {
	var m map[string]interface{}
	if err := unmarshal(raw, &m); err != nil {
		return ResourceAttributes{}, &errors.ScimErrorInvalidSyntax
	}

	attributes, scimErr := t.schemaWithCommon().Validate(m)
	if scimErr != nil {
		return ResourceAttributes{}, scimErr
	}

	for _, extension := range t.SchemaExtensions {
		extensionField := m[extension.Schema.ID]
		if extensionField == nil {
			if extension.Required {
				return ResourceAttributes{}, &errors.ScimErrorInvalidValue
			}
			continue
		}

		extensionAttributes, scimErr := extension.Schema.Validate(extensionField)
		if scimErr != nil {
			return ResourceAttributes{}, scimErr
		}

		attributes[extension.Schema.ID] = extensionAttributes
	}

	return attributes, nil
}

// validatePatch parse and validate PATCH request.
func (t ResourceType) validatePatch(r *http.Request) ([]PatchOperation, *errors.ScimError) {
	data, err := readBody(r)
	if err != nil {
		return nil, &errors.ScimErrorInvalidSyntax
	}

	var req struct {
		Schemas    []string
		Operations []json.RawMessage
	}
	if err := unmarshal(data, &req); err != nil {
		return nil, &errors.ScimErrorInvalidSyntax
	}

	// The body of each request MUST contain the "schemas" attribute with the URI value of
	// "urn:ietf:params:scim:api:messages:2.0:PatchOp".
	if len(req.Schemas) != 1 || req.Schemas[0] != "urn:ietf:params:scim:api:messages:2.0:PatchOp" {
		return nil, &errors.ScimErrorInvalidValue
	}

	// The body of an HTTP PATCH request MUST contain the attribute "Operations",
	// whose value is an array of one or more PATCH operations.
	if len(req.Operations) < 1 {
		return nil, &errors.ScimErrorInvalidValue
	}

	// Evaluation continues until all operations are successfully applied or until an error condition is encountered.
	var operations []PatchOperation
	for _, v := range req.Operations {
		validator, err := patch.NewValidator(
			string(v),
			t.schemaWithCommon(),
			t.getSchemaExtensions()...,
		)
		if err != nil {
			return nil, &errors.ScimErrorInvalidPath
		}
		value, err := validator.Validate()
		if err != nil {
			return nil, &errors.ScimErrorInvalidValue
		}
		operations = append(operations, PatchOperation{
			Op:    string(validator.Op),
			Path:  validator.Path,
			Value: value,
		})
	}

	return operations, nil
}

// SchemaExtension is one of the resource type's schema extensions.
type SchemaExtension struct {
	// Schema is the URI of an extended schema, e.g., "urn:edu:2.0:Staff".
	Schema schema.Schema
	// Required is a boolean value that specifies whether or not the schema extension is required for the resource
	// type. If true, a resource of this type MUST include this schema extension and also include any attributes
	// declared as required in this schema extension. If false, a resource of this type MAY omit this schema
	// extension.
	Required bool
}
