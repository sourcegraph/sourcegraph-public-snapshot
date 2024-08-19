package patch

import (
	"fmt"
	f "github.com/elimity-com/scim/internal/filter"
	"github.com/elimity-com/scim/schema"
)

// validateUpdate validates the add/replace operation contained within the validator based on on Section 3.5.2.1 in
// RFC 7644. More info: https://datatracker.ietf.org/doc/html/rfc7644#section-3.5.2.1
func (v OperationValidator) validateUpdate() (interface{}, error) {
	// The operation must contain a "value" member whose content specifies the value to be added/replaces.
	if v.value == nil {
		return nil, fmt.Errorf("an add operation must contain a value member")
	}

	// If "path" is omitted, the target location is assumed to be the resource itself.
	if v.Path == nil {
		return v.validateEmptyPath()
	}

	refAttr, err := v.getRefAttribute(v.Path.AttributePath)
	if err != nil {
		return nil, err
	}
	if v.Path.ValueExpression != nil {
		if err := f.NewFilterValidator(v.Path.ValueExpression, schema.Schema{
			Attributes: refAttr.SubAttributes(),
		}).Validate(); err != nil {
			return nil, err
		}
	}
	if subAttrName := v.Path.SubAttributeName(); subAttrName != "" {
		refSubAttr, err := v.getRefSubAttribute(refAttr, subAttrName)
		if err != nil {
			return nil, err
		}
		refAttr = refSubAttr
	}

	if !refAttr.MultiValued() {
		attr, scimErr := refAttr.ValidateSingular(v.value)
		if scimErr != nil {
			return nil, scimErr
		}
		return attr, nil
	}

	if list, ok := v.value.([]interface{}); ok {
		var attrs []interface{}
		for _, value := range list {
			attr, scimErr := refAttr.ValidateSingular(value)
			if scimErr != nil {
				return nil, scimErr
			}
			attrs = append(attrs, attr)
		}
		return attrs, nil
	}

	attr, scimErr := refAttr.ValidateSingular(v.value)
	if scimErr != nil {
		return nil, scimErr
	}
	return []interface{}{attr}, nil
}
