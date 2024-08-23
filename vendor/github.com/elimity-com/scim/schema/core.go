package schema

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	datetime "github.com/di-wu/xsd-datetime"
	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
)

// CoreAttribute represents those attributes that sit at the top level of the JSON object together with the common
// attributes (such as the resource "id").
type CoreAttribute struct {
	canonicalValues []string
	caseExact       bool
	description     optional.String
	multiValued     bool
	mutability      attributeMutability
	name            string
	referenceTypes  []AttributeReferenceType
	required        bool
	returned        attributeReturned
	subAttributes   Attributes
	typ             attributeType
	uniqueness      attributeUniqueness
}

// ComplexCoreAttribute creates a complex attribute based on given parameters.
func ComplexCoreAttribute(params ComplexParams) CoreAttribute {
	checkAttributeName(params.Name)

	names := map[string]int{}
	var sa []CoreAttribute

	for i, a := range params.SubAttributes {
		name := strings.ToLower(a.name)
		if j, ok := names[name]; ok {
			panic(fmt.Errorf("duplicate name %q for sub-attributes %d and %d", name, i, j))
		}

		names[name] = i

		sa = append(sa, CoreAttribute{
			canonicalValues: a.canonicalValues,
			caseExact:       a.caseExact,
			description:     a.description,
			multiValued:     a.multiValued,
			mutability:      a.mutability,
			name:            a.name,
			referenceTypes:  a.referenceTypes,
			required:        a.required,
			returned:        a.returned,
			typ:             a.typ,
			uniqueness:      a.uniqueness,
		})
	}

	return CoreAttribute{
		description:   params.Description,
		multiValued:   params.MultiValued,
		mutability:    params.Mutability.m,
		name:          params.Name,
		required:      params.Required,
		returned:      params.Returned.r,
		subAttributes: sa,
		typ:           attributeDataTypeComplex,
		uniqueness:    params.Uniqueness.u,
	}
}

// SimpleCoreAttribute creates a non-complex attribute based on given parameters.
func SimpleCoreAttribute(params SimpleParams) CoreAttribute {
	checkAttributeName(params.name)

	return CoreAttribute{
		canonicalValues: params.canonicalValues,
		caseExact:       params.caseExact,
		description:     params.description,
		multiValued:     params.multiValued,
		mutability:      params.mutability,
		name:            params.name,
		referenceTypes:  params.referenceTypes,
		required:        params.required,
		returned:        params.returned,
		typ:             params.typ,
		uniqueness:      params.uniqueness,
	}
}

// AttributeType returns the attribute type.
func (a CoreAttribute) AttributeType() string {
	return a.typ.String()
}

// CanonicalValues returns the canonical values of the attribute.
func (a CoreAttribute) CanonicalValues() []string {
	return a.canonicalValues
}

// CaseExact returns whether the attribute is case exact.
func (a CoreAttribute) CaseExact() bool {
	return a.caseExact
}

// Description returns whether the description of the attribute.
func (a CoreAttribute) Description() string {
	return a.description.Value()
}

// HasSubAttributes returns whether the attribute is complex and has sub attributes.
func (a CoreAttribute) HasSubAttributes() bool {
	return a.typ == attributeDataTypeComplex && len(a.subAttributes) != 0
}

// MultiValued returns whether the attribute is multi valued.
func (a CoreAttribute) MultiValued() bool {
	return a.multiValued
}

// Mutability returns the mutability of the attribute.
func (a CoreAttribute) Mutability() string {
	raw, _ := a.mutability.MarshalJSON()
	return string(raw)
}

// Name returns the case insensitive name of the attribute.
func (a CoreAttribute) Name() string {
	return a.name
}

// ReferenceTypes returns the reference types of the attribute.
func (a CoreAttribute) ReferenceTypes() []AttributeReferenceType {
	return a.referenceTypes
}

// Required returns whether the attribute is required.
func (a CoreAttribute) Required() bool {
	return a.required
}

// Returned returns when the attribute need to be returned.
func (a CoreAttribute) Returned() string {
	raw, _ := a.returned.MarshalJSON()
	return string(raw)
}

// SubAttributes returns the sub attributes.
func (a CoreAttribute) SubAttributes() Attributes {
	return a.subAttributes
}

// Uniqueness returns the attributes uniqueness.
func (a CoreAttribute) Uniqueness() string {
	raw, _ := a.uniqueness.MarshalJSON()
	return string(raw)
}

// ValidateSingular checks whether the given singular value matches the attribute data type. Unknown attributes in
// given complex value are ignored. The returned interface contains a (sanitised) version of the given attribute.
func (a CoreAttribute) ValidateSingular(attribute interface{}) (interface{}, *errors.ScimError) {
	switch a.typ {
	case attributeDataTypeBinary:
		bin, ok := attribute.(string)
		if !ok {
			return nil, &errors.ScimErrorInvalidValue
		}

		match, err := regexp.MatchString(`^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$`, bin)
		if err != nil {
			panic(err)
		}

		if !match {
			return nil, &errors.ScimErrorInvalidValue
		}

		return bin, nil
	case attributeDataTypeBoolean:
		b, ok := attribute.(bool)
		if !ok {
			return nil, &errors.ScimErrorInvalidValue
		}

		return b, nil
	case attributeDataTypeComplex:
		obj, ok := attribute.(map[string]interface{})
		if !ok {
			return nil, &errors.ScimErrorInvalidValue
		}

		attributes := make(map[string]interface{})

		for _, sub := range a.subAttributes {
			var hit interface{}
			var found bool
			for k, v := range obj {
				if strings.EqualFold(sub.name, k) {
					if found {
						return nil, &errors.ScimErrorInvalidSyntax
					}
					found = true
					hit = v
				}
			}

			attr, scimErr := sub.validate(hit)
			if scimErr != nil {
				return nil, scimErr
			}
			if attr != nil {
				attributes[sub.name] = attr
			}
		}
		return attributes, nil
	case attributeDataTypeDateTime:
		date, ok := attribute.(string)
		if !ok {
			return nil, &errors.ScimErrorInvalidValue
		}
		_, err := datetime.Parse(date)
		if err != nil {
			return nil, &errors.ScimErrorInvalidValue
		}

		return date, nil
	case attributeDataTypeDecimal:
		switch n := attribute.(type) {
		case json.Number:
			f, err := n.Float64()
			if err != nil {
				return nil, &errors.ScimErrorInvalidValue
			}

			return f, nil
		case float64:
			return n, nil
		default:
			return nil, &errors.ScimErrorInvalidValue
		}
	case attributeDataTypeInteger:
		switch n := attribute.(type) {
		case json.Number:
			i, err := n.Int64()
			if err != nil {
				return nil, &errors.ScimErrorInvalidValue
			}

			return i, nil
		case int, int8, int16, int32, int64:
			return n, nil
		default:
			return nil, &errors.ScimErrorInvalidValue
		}
	case attributeDataTypeString, attributeDataTypeReference:
		s, ok := attribute.(string)
		if !ok {
			return nil, &errors.ScimErrorInvalidValue
		}

		return s, nil
	default:
		return nil, &errors.ScimErrorInvalidSyntax
	}
}

func (a *CoreAttribute) getRawAttributes() map[string]interface{} {
	attributes := map[string]interface{}{
		"description": a.description.Value(),
		"multiValued": a.multiValued,
		"mutability":  a.mutability,
		"name":        a.name,
		"required":    a.required,
		"returned":    a.returned,
		"type":        a.typ,
	}

	if a.canonicalValues != nil {
		attributes["canonicalValues"] = a.canonicalValues
	}

	if a.referenceTypes != nil {
		attributes["referenceTypes"] = a.referenceTypes
	}

	var rawSubAttributes []map[string]interface{}
	for _, subAttr := range a.subAttributes {
		rawSubAttributes = append(rawSubAttributes, subAttr.getRawAttributes())
	}

	if a.subAttributes != nil && len(a.subAttributes) != 0 {
		attributes["subAttributes"] = rawSubAttributes
	}

	if a.typ != attributeDataTypeComplex && a.typ != attributeDataTypeBoolean {
		attributes["caseExact"] = a.caseExact
		attributes["uniqueness"] = a.uniqueness
	}

	return attributes
}

func (a CoreAttribute) validate(attribute interface{}) (interface{}, *errors.ScimError) {
	// whether or not the attribute is required.
	if attribute == nil {
		if !a.required {
			return nil, nil
		}

		// the attribute is not present but required.
		return nil, &errors.ScimErrorInvalidValue
	}

	// whether the value of the attribute can be (re)defined
	// readOnly: the attribute SHALL NOT be modified.
	if a.mutability == attributeMutabilityReadOnly {
		return nil, nil
	}

	if !a.multiValued {
		return a.ValidateSingular(attribute)
	}

	switch arr := attribute.(type) {
	case map[string]interface{}:
		// return false if the multivalued attribute is empty.
		if a.required && len(arr) == 0 {
			return nil, &errors.ScimErrorInvalidValue
		}

		validMap := map[string]interface{}{}
		for k, v := range arr {
			for _, sub := range a.subAttributes {
				if !strings.EqualFold(sub.name, k) {
					continue
				}
				_, scimErr := sub.validate(v)
				if scimErr != nil {
					return nil, scimErr
				}
				validMap[sub.name] = v
			}
		}
		return validMap, nil

	case []interface{}:
		// return false if the multivalued attribute is empty.
		if a.required && len(arr) == 0 {
			return nil, &errors.ScimErrorInvalidValue
		}

		var attributes []interface{}
		for _, ele := range arr {
			attr, scimErr := a.ValidateSingular(ele)
			if scimErr != nil {
				return nil, scimErr
			}
			attributes = append(attributes, attr)
		}
		return attributes, nil

	default:
		// return false if the multivalued attribute is not a slice.
		return nil, &errors.ScimErrorInvalidSyntax
	}
}
