// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

/*
 * Cody Service
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 0.0.0
 */

package models

type Error struct {
	Code int32 `json:"code"`

	Message string `json:"message"`
}

// AssertErrorRequired checks if the required fields are not zero-ed
func AssertErrorRequired(obj Error) error {
	elements := map[string]interface{}{
		"code":    obj.Code,
		"message": obj.Message,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertErrorConstraints checks if the values respects the defined constraints
func AssertErrorConstraints(obj Error) error {
	return nil
}
