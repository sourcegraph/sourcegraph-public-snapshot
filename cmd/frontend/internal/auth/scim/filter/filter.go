package filter

import (
	"fmt"

	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// This is almost unmodified copy-pasta from https://github.com/elimity-com/scim/
// The only changes to the package are the package names and imports in the test files,
// and some cosmetics to comply with our CI checks.
// Elimity's SCIM package has the following licensing information:
// ----------------------------------------------------------------------------
// MIT License
//
// Copyright (c) 2019 Elimity NV
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software
// and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute,
// sublicense, and/or sell copies of the Software, and to permit persons to whom the Software
// is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or
// substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
// BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE  FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
// ----------------------------------------------------------------------------

// validateAttributePath checks whether the given attribute path is a valid path within the given reference schema.
func validateAttributePath(ref schema.Schema, attrPath filter.AttributePath) (schema.CoreAttribute, error) {
	if uri := attrPath.URI(); uri != "" && uri != ref.ID {
		return schema.CoreAttribute{}, errors.Newf("the uri does not match the schema id: %s", uri)
	}

	attr, ok := ref.Attributes.ContainsAttribute(attrPath.AttributeName)
	if !ok {
		return schema.CoreAttribute{}, errors.Newf(
			"the reference schema does not have an attribute with the name: %s",
			attrPath.AttributeName,
		)
	}
	// e.g. name.givenName
	//           ^________
	if subAttrName := attrPath.SubAttributeName(); subAttrName != "" {
		if err := validateSubAttribute(attr, subAttrName); err != nil {
			return schema.CoreAttribute{}, err
		}
	}
	return attr, nil
}

// validateExpression checks whether the given expression is a valid expression within the given reference schema.
func validateExpression(ref schema.Schema, e filter.Expression) error {
	switch e := e.(type) {
	case *filter.ValuePath:
		attr, err := validateAttributePath(ref, e.AttributePath)
		if err != nil {
			return nil
		}
		if err := validateExpression(
			schema.Schema{
				ID:         ref.ID,
				Attributes: attr.SubAttributes(),
			},
			e.ValueFilter,
		); err != nil {
			return err
		}
		return nil
	case *filter.AttributeExpression:
		if _, err := validateAttributePath(ref, e.AttributePath); err != nil {
			return err
		}
		return nil
	case *filter.LogicalExpression:
		if err := validateExpression(ref, e.Left); err != nil {
			return err
		}
		if err := validateExpression(ref, e.Right); err != nil {
			return err
		}
		return nil
	case *filter.NotExpression:
		if err := validateExpression(ref, e.Expression); err != nil {
			return err
		}
		return nil
	default:
		panic(fmt.Sprintf("unknown expression type: %s", e))
	}
}

// validateSubAttribute checks whether the given attribute name is a attribute within the given reference attribute.
func validateSubAttribute(attr schema.CoreAttribute, subAttrName string) error {
	if !attr.HasSubAttributes() {
		return errors.Newf("the attribute has no sub-attributes")
	}

	if _, ok := attr.SubAttributes().ContainsAttribute(subAttrName); !ok {
		return errors.Newf("the attribute has no sub-attributes named: %s", subAttrName)
	}
	return nil
}

// Validator represents a filter validator.
type Validator struct {
	filter     filter.Expression
	schema     schema.Schema
	extensions []schema.Schema
}

// NewFilterValidator constructs a new filter validator.
func NewFilterValidator(exp filter.Expression, s schema.Schema, exts ...schema.Schema) Validator {
	return Validator{
		filter:     exp,
		schema:     s,
		extensions: exts,
	}
}

// NewValidator constructs a new filter validator.
func NewValidator(exp string, s schema.Schema, exts ...schema.Schema) (Validator, error) {
	e, err := filter.ParseFilter([]byte(exp))
	if err != nil {
		return Validator{}, err
	}
	return Validator{
		filter:     e,
		schema:     s,
		extensions: exts,
	}, nil
}

// GetFilter returns the filter contained within the validator.
func (v Validator) GetFilter() filter.Expression {
	return v.filter
}

// PassesFilter checks whether given resources passes the filter.
func (v Validator) PassesFilter(resource map[string]interface{}) error {
	switch e := v.filter.(type) {
	case *filter.ValuePath:
		ref, attr, ok := v.referenceContains(e.AttributePath)
		if !ok {
			return errors.Newf("could not find an attribute that matches the attribute path: %s", e.AttributePath)
		}
		if !attr.MultiValued() {
			return errors.Newf("value path filters can only be applied to multi-valued attributes")
		}

		value, ok := resource[attr.Name()]
		if !ok {
			// Also try with the id as prefix.
			value, ok = resource[fmt.Sprintf("%s:%s", ref.ID, attr.Name())]
			if !ok {
				return errors.Newf("the resource does contain the attribute specified in the filter")
			}
		}
		valueFilter := Validator{
			filter: e.ValueFilter,
			schema: schema.Schema{
				ID:         ref.ID,
				Attributes: attr.SubAttributes(),
			},
		}
		switch value := value.(type) {
		case []interface{}:
			for _, a := range value {
				attr, ok := a.(map[string]interface{})
				if !ok {
					return errors.Newf("the target is not a complex attribute")
				}
				if err := valueFilter.PassesFilter(attr); err == nil {
					// Found an attribute that passed the value filter.
					return nil
				}
			}
		}
		return errors.Newf("the resource does not pass the filter")
	case *filter.AttributeExpression:
		ref, attr, ok := v.referenceContains(e.AttributePath)
		if !ok {
			return errors.Newf("could not find an attribute that matches the attribute path: %s", e.AttributePath)
		}

		value, ok := resource[attr.Name()]
		if !ok {
			// Also try with the id as prefix.
			value, ok = resource[fmt.Sprintf("%s:%s", ref.ID, attr.Name())]
			if !ok {
				return errors.Newf("the resource does contain the attribute specified in the filter")
			}
		}

		var (
			// cmpAttr will be the attribute to validate the filter against.
			cmpAttr = attr

			subAttr     schema.CoreAttribute
			subAttrName = e.AttributePath.SubAttributeName()
		)

		if subAttrName != "" {
			if !attr.HasSubAttributes() {
				// The attribute has no sub-attributes.
				return errors.Newf("the specified attribute has no sub-attributes")
			}
			subAttr, ok = attr.SubAttributes().ContainsAttribute(subAttrName)
			if !ok {
				return errors.Newf("the resource has no sub-attribute named: %s", subAttrName)
			}

			attr, ok := value.(map[string]interface{})
			if !ok {
				return errors.Newf("the target is not a complex attribute")
			}
			value, ok = attr[subAttr.Name()]
			if !ok {
				return errors.Newf("the resource does contain the attribute specified in the filter")
			}

			cmpAttr = subAttr
		}

		// If the attribute has a non-empty or non-null value or if it contains a non-empty node for complex attributes, there is a match.
		if e.Operator == filter.PR {
			// We already found a value.
			return nil
		}

		cmp, err := createCompareFunction(e, cmpAttr)
		if err != nil {
			return err
		}

		if !attr.MultiValued() {
			if err := cmp(value); err != nil {
				return errors.Newf("the resource does not pass the filter: %s", err)
			}
			return nil
		}

		switch value := value.(type) {
		case []interface{}:
			var err error
			for _, v := range value {
				if err = cmp(v); err == nil {
					return nil
				}
			}
			return errors.Newf("the resource does not pass the filter: %s", err)
		default:
			panic(fmt.Sprintf("given value is not a []interface{}: %v", value))
		}
	case *filter.LogicalExpression:
		switch e.Operator {
		case filter.AND:
			leftValidator := Validator{
				e.Left,
				v.schema,
				v.extensions,
			}
			if err := leftValidator.PassesFilter(resource); err != nil {
				return err
			}
			rightValidator := Validator{
				e.Right,
				v.schema,
				v.extensions,
			}
			return rightValidator.PassesFilter(resource)
		case filter.OR:
			leftValidator := Validator{
				e.Left,
				v.schema,
				v.extensions,
			}
			if err := leftValidator.PassesFilter(resource); err == nil {
				return nil
			}
			rightValidator := Validator{
				e.Right,
				v.schema,
				v.extensions,
			}
			return rightValidator.PassesFilter(resource)
		}
		return errors.Newf("the resource does not pass the filter")
	case *filter.NotExpression:
		validator := Validator{
			e.Expression,
			v.schema,
			v.extensions,
		}
		if err := validator.PassesFilter(resource); err != nil {
			return nil
		}
		return errors.Newf("the resource does not pass the filter")
	default:
		panic(fmt.Sprintf("unknown expression type: %s", e))
	}
}

// Validate checks whether the expression is a valid path within the given reference schemas.
func (v Validator) Validate() error {
	err := validateExpression(v.schema, v.filter)
	if err == nil {
		return nil
	}
	for _, e := range v.extensions {
		if err := validateExpression(e, v.filter); err == nil {
			return nil
		}
	}
	return err
}

// referenceContains returns the schema and attribute to which the attribute path applies.
func (v Validator) referenceContains(attrPath filter.AttributePath) (schema.Schema, schema.CoreAttribute, bool) {
	for _, s := range append([]schema.Schema{v.schema}, v.extensions...) {
		if uri := attrPath.URI(); uri != "" && s.ID != uri {
			continue
		}
		if attr, ok := s.Attributes.ContainsAttribute(attrPath.AttributeName); ok {
			return s, attr, true
		}
	}
	return schema.Schema{}, schema.CoreAttribute{}, false
}
