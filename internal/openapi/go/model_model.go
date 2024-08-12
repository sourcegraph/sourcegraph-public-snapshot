// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

/*
 * Cody Service
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 0.0.0
 */

package models




// Model - Describes an OpenAI model offering that can be used with the API.
type Model struct {

	// The model identifier, which can be referenced in the API endpoints.
	Id string `json:"id"`

	// The object type, which is always \"model\".
	Object string `json:"object"`

	// The Unix timestamp (in seconds) when the model was created.
	Created int64 `json:"created"`

	// The organization that owns the model.
	OwnedBy string `json:"owned_by"`
}

// AssertModelRequired checks if the required fields are not zero-ed
func AssertModelRequired(obj Model) error {
	elements := map[string]interface{}{
		"id": obj.Id,
		"object": obj.Object,
		"created": obj.Created,
		"owned_by": obj.OwnedBy,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertModelConstraints checks if the values respects the defined constraints
func AssertModelConstraints(obj Model) error {
	return nil
}
