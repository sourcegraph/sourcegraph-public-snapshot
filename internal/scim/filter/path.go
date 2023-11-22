package filter

import (
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
)

// MultiValuedFilterAttributes returns the attributes of the given attribute on which can be filtered. In the case of a
// complex attribute, the sub-attributes get returned. Otherwise if the given attribute is not complex, a "value" sub-
// attribute gets created to filter against.
func MultiValuedFilterAttributes(attr schema.CoreAttribute) schema.Attributes {
	switch attr.AttributeType() {
	case "complex":
		return attr.SubAttributes()
	default:
		return schema.Attributes{schema.SimpleCoreAttribute(getSimpleParams(attr))}
	}
}

// getSimpleParams returns simple params based on the type of the attribute. The simple params only have their name
// assigned to "value", everything else is left out (i.e. default values). Can not be used for complex attributes.
func getSimpleParams(attr schema.CoreAttribute) schema.SimpleParams {
	switch attr.AttributeType() {
	case "decimal":
		return schema.SimpleNumberParams(schema.NumberParams{
			Name: "value",
			Type: schema.AttributeTypeDecimal(),
		})
	case "integer":
		return schema.SimpleNumberParams(schema.NumberParams{
			Name: "value",
			Type: schema.AttributeTypeInteger(),
		})
	case "binary":
		return schema.SimpleBinaryParams(schema.BinaryParams{Name: "value"})
	case "boolean":
		return schema.SimpleBooleanParams(schema.BooleanParams{Name: "value"})
	case "dateTime":
		return schema.SimpleDateTimeParams(schema.DateTimeParams{Name: "value"})
	case "reference":
		return schema.SimpleReferenceParams(schema.ReferenceParams{Name: "value"})
	default:
		return schema.SimpleStringParams(schema.StringParams{Name: "value"})
	}
}

// PathValidator represents a path validator.
type PathValidator struct {
	path       filter.Path
	schema     schema.Schema
	extensions []schema.Schema
}

// NewPathValidator constructs a new path validator.
func NewPathValidator(pathFilter string, s schema.Schema, exts ...schema.Schema) (PathValidator, error) {
	f, err := filter.ParsePath([]byte(pathFilter))
	if err != nil {
		return PathValidator{}, err
	}
	return PathValidator{
		path:       f,
		schema:     s,
		extensions: exts,
	}, nil
}

func (v PathValidator) Path() filter.Path {
	return v.path
}

// Validate checks whether the path is a valid path within the given reference schemas.
func (v PathValidator) Validate() error {
	err := v.validatePath(v.schema)
	if err == nil {
		return nil
	}
	for _, e := range v.extensions {
		if err := v.validatePath(e); err == nil {
			return nil
		}
	}
	return err
}

// validatePath tries to validate the path against the given schema.
func (v PathValidator) validatePath(ref schema.Schema) error {
	// e.g. members
	//      ^______
	attr, err := validateAttributePath(ref, v.path.AttributePath)
	if err != nil {
		return err
	}

	// e.g. members[value eq "0"]
	//             ^_____________
	if v.path.ValueExpression != nil {
		if err := validateExpression(
			schema.Schema{
				ID:         ref.ID,
				Attributes: MultiValuedFilterAttributes(attr),
			},
			v.path.ValueExpression,
		); err != nil {
			return err
		}
	}

	// e.g. members[value eq "0"].displayName
	//                            ^__________
	if subAttrName := v.path.SubAttributeName(); subAttrName != "" {
		if err := validateSubAttribute(attr, subAttrName); err != nil {
			return err
		}
	}
	return nil
}
