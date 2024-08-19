package schema

import (
	"encoding/json"
	"strings"

	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
)

const (
	// UserSchema is the URI for the User resource.
	UserSchema = "urn:ietf:params:scim:schemas:core:2.0:User"

	// GroupSchema is the URI for the Group resource.
	GroupSchema = "urn:ietf:params:scim:schemas:core:2.0:Group"
)

func cannotBePatched(op string, attr CoreAttribute) bool {
	return isImmutable(op, attr) || isReadOnly(attr)
}

func isImmutable(op string, attr CoreAttribute) bool {
	return attr.mutability == attributeMutabilityImmutable && (op == "replace" || op == "remove")
}

func isReadOnly(attr CoreAttribute) bool {
	return attr.mutability == attributeMutabilityReadOnly
}

// Attributes represent a list of Core Attributes.
type Attributes []CoreAttribute

// ContainsAttribute checks whether the list of Core Attributes contains an attribute with the given name.
func (as Attributes) ContainsAttribute(name string) (CoreAttribute, bool) {
	for _, a := range as {
		if strings.EqualFold(name, a.name) {
			return a, true
		}
	}
	return CoreAttribute{}, false
}

// Schema is a collection of attribute definitions that describe the contents of an entire or partial resource.
type Schema struct {
	Attributes  Attributes
	Description optional.String
	ID          string
	Name        optional.String
}

// MarshalJSON converts the schema struct to its corresponding json representation.
func (s Schema) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.ToMap())
}

// ToMap returns the map representation of a schema.
func (s Schema) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":          s.ID,
		"name":        s.Name.Value(),
		"description": s.Description.Value(),
		"attributes":  s.getRawAttributes(),
		"schemas":     []string{"urn:ietf:params:scim:schemas:core:2.0:Schema"},
	}
}

// Validate validates given resource based on the schema. Does NOT validate mutability.
// NOTE: only used in POST and PUT requests where attributes MAY be (re)defined.
func (s Schema) Validate(resource interface{}) (map[string]interface{}, *errors.ScimError) {
	return s.validate(resource, false)
}

// ValidateMutability validates given resource based on the schema, including strict immutability checks.
func (s Schema) ValidateMutability(resource interface{}) (map[string]interface{}, *errors.ScimError) {
	return s.validate(resource, true)
}

// ValidatePatchOperation validates an individual operation and its related value.
func (s Schema) ValidatePatchOperation(operation string, operationValue map[string]interface{}, isExtension bool) *errors.ScimError {
	for k, v := range operationValue {
		var attr *CoreAttribute
		var scimErr *errors.ScimError

		for _, attribute := range s.Attributes {
			if strings.EqualFold(attribute.name, k) {
				attr = &attribute
				break
			}
			if isExtension && strings.EqualFold(s.ID+":"+attribute.name, k) {
				attr = &attribute
				break
			}
		}

		// Attribute does not exist in the schema, thus it is an invalid request.
		// Immutable attrs can only be added and Readonly attrs cannot be patched
		if attr == nil || cannotBePatched(operation, *attr) {
			return &errors.ScimErrorInvalidValue
		}

		// "remove" operations simply have to exist
		if operation != "remove" {
			_, scimErr = attr.validate(v)
		}

		if scimErr != nil {
			return scimErr
		}
	}

	return nil
}

// ValidatePatchOperationValue validates an individual operation and its related value.
func (s Schema) ValidatePatchOperationValue(operation string, operationValue map[string]interface{}) *errors.ScimError {
	return s.ValidatePatchOperation(operation, operationValue, false)
}

func (s Schema) getRawAttributes() []map[string]interface{} {
	attributes := make([]map[string]interface{}, len(s.Attributes))

	for i, a := range s.Attributes {
		attributes[i] = a.getRawAttributes()
	}

	return attributes
}

func (s Schema) validate(resource interface{}, checkMutability bool) (map[string]interface{}, *errors.ScimError) {
	core, ok := resource.(map[string]interface{})
	if !ok {
		return nil, &errors.ScimErrorInvalidSyntax
	}

	attributes := make(map[string]interface{})
	for _, attribute := range s.Attributes {
		var hit interface{}
		var found bool
		for k, v := range core {
			if strings.EqualFold(attribute.name, k) {
				// duplicate found
				if found {
					return nil, &errors.ScimErrorInvalidSyntax
				}
				found = true
				hit = v
			}
		}

		// An immutable attribute SHALL NOT be updated.
		if found && checkMutability &&
			attribute.mutability == attributeMutabilityImmutable {
			return nil, &errors.ScimErrorMutability
		}

		attr, scimErr := attribute.validate(hit)
		if scimErr != nil {
			return nil, scimErr
		}
		if attr != nil {
			attributes[attribute.name] = attr
		}
	}
	return attributes, nil
}
