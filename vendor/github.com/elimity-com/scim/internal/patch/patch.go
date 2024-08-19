package patch

import (
	"encoding/json"
	"fmt"
	f "github.com/elimity-com/scim/internal/filter"
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
	"strings"
)

// Op represents the possible value the operation is to perform.
// Possible values are one of "add", "remove", or "replace".
type Op string

const (
	// OperationAdd is used to add a new attribute value to an existing resource.
	OperationAdd Op = "add"
	// OperationRemove removes the value at the target location specified by the required attribute "path".
	OperationRemove Op = "remove"
	// OperationReplace replaces the value at the target location specified by the "path".
	OperationReplace Op = "replace"
)

// OperationValidator represents a validator to validate PATCH requests.
type OperationValidator struct {
	Op    Op
	Path  *filter.Path
	value interface{}

	schema  schema.Schema
	schemas map[string]schema.Schema
}

// NewValidator creates an OperationValidator based on the given JSON string and reference schemas.
// Returns an error if patchReq is not valid.
func NewValidator(patchReq string, s schema.Schema, extensions ...schema.Schema) (OperationValidator, error) {
	var operation struct {
		Op    string
		Path  string
		Value interface{}
	}
	if err := json.Unmarshal([]byte(patchReq), &operation); err != nil {
		return OperationValidator{}, err
	}

	operation.Op = strings.ToLower(operation.Op)

	switch v := operation.Value.(type) {
	// Okta also send the ID on PATCH requests.
	// See: internal/idp_test/testdata/okta/update_group_name.json
	// https://developer.okta.com/docs/reference/scim/scim-20/#update-a-specific-group-name
	case map[string]interface{}:
		var key string
		var found bool
		for k := range v {
			if strings.ToLower(k) == "id" {
				if found {
					return OperationValidator{}, fmt.Errorf("duplicate attributes: %s and %s", k, key)
				}
				found = true
				key = k
			}
		}
		delete(v, key)
	}

	var path *filter.Path
	if operation.Path != "" {
		validator, err := f.NewPathValidator(operation.Path, s, extensions...)
		if err != nil {
			return OperationValidator{}, err
		}
		if err := validator.Validate(); err != nil {
			return OperationValidator{}, err
		}
		p := validator.Path()
		path = &p
	}

	schemas := map[string]schema.Schema{
		s.ID: s,
	}
	for _, e := range extensions {
		schemas[e.ID] = e
	}
	return OperationValidator{
		Op:    Op(operation.Op),
		Path:  path,
		value: operation.Value,

		schema:  s,
		schemas: schemas,
	}, nil
}

// Validate validates the PATCH operation. Unknown attributes in complex values are ignored. The returned interface
// contains a (sanitised) version of given value based on the attribute it targets. Multi-valued attributes will always
// be returned wrapped in a slice, even if it is just one value that was defined within the operation.
func (v OperationValidator) Validate() (interface{}, error) {
	switch v.Op {
	case OperationAdd, OperationReplace:
		return v.validateUpdate()
	case OperationRemove:
		return nil, v.validateRemove()
	default:
		return nil, fmt.Errorf("invalid operation Op: %s", v.Op)
	}
}

// getRefAttribute returns the corresponding attribute based on the given attribute path.
//
// e.g.
//  - `userName` would return the `userName` attribute.
//	- `name.givenName` would return the `givenName` attribute.
//  - `ext:employeeNumber` would return the `employeeNumber` attribute from the extension.
func (v OperationValidator) getRefAttribute(attrPath filter.AttributePath) (*schema.CoreAttribute, error) {
	// Get the corresponding schema, this can be the main schema or an extension.
	var refSchema = v.schema
	if uri := attrPath.URI(); uri != "" {
		// It can also be an extension if it has a uri prefix.
		var ok bool
		if refSchema, ok = v.schemas[uri]; !ok {
			return nil, fmt.Errorf("invalid uri prefix: %s", uri)
		}
	}

	// Get the correct attribute corresponding to the given attribute path.
	var (
		refAttr  *schema.CoreAttribute
		attrName = attrPath.AttributeName
	)
	for _, attr := range refSchema.Attributes {
		if strings.EqualFold(attr.Name(), attrName) {
			refAttr = &attr
			break
		}
	}
	if refAttr == nil {
		return nil, fmt.Errorf("could not find attribute %s", v.Path)
	}
	if subAttrName := attrPath.SubAttributeName(); subAttrName != "" {
		refSubAttr, err := v.getRefSubAttribute(refAttr, subAttrName)
		if err != nil {
			return nil, err
		}
		refAttr = refSubAttr
	}
	return refAttr, nil
}

// getRefSubAttribute returns the sub-attribute of the reference attribute that matches the given subAttrName, if none
// are found it will return an error.
func (v OperationValidator) getRefSubAttribute(refAttr *schema.CoreAttribute, subAttrName string) (*schema.CoreAttribute, error) {
	if !refAttr.HasSubAttributes() {
		return nil, fmt.Errorf("the referred attribute has no sub-attributes: %s", v.Path)
	}
	var refSubAttr *schema.CoreAttribute
	for _, attr := range refAttr.SubAttributes() {
		if strings.EqualFold(attr.Name(), subAttrName) {
			refSubAttr = &attr
			break
		}
	}
	if refSubAttr == nil {
		return nil, fmt.Errorf("could not find attribute %s", v.Path)
	}
	return refSubAttr, nil
}

// validateEmptyPath validates paths that don't have a "path" value. In this case the target location is assumed to be
// the resource itself. The "value" parameter contains a set of attributes to be added to the resource.
func (v OperationValidator) validateEmptyPath() (interface{}, error) {
	attributes, ok := v.value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("the given value should be a complex attribute if path is empty")
	}

	rootValue := map[string]interface{}{}
	for p, value := range attributes {
		path, err := filter.ParsePath([]byte(p))
		if err != nil {
			return nil, fmt.Errorf("invalid attribute path: %s", p)
		}
		validator := OperationValidator{
			Op:      v.Op,
			Path:    &path,
			value:   value,
			schema:  v.schema,
			schemas: v.schemas,
		}
		v, err := validator.Validate()
		if err != nil {
			return nil, err
		}
		rootValue[p] = v
	}
	return rootValue, nil
}
