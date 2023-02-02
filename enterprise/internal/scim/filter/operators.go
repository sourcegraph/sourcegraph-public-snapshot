package filter

import (
	"fmt"
	datetime "github.com/di-wu/xsd-datetime"
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
)

// createCompareFunction returns a compare function based on the attribute expression and attribute.
// e.g. `userName eq "john"` will return a string comparator that checks whether the passed value is equal to "john".
func createCompareFunction(e *filter.AttributeExpression, attr schema.CoreAttribute) (func(interface{}) error, error) {
	switch typ := attr.AttributeType(); typ {
	case "binary":
		ref, ok := e.CompareValue.(string)
		if !ok {
			return nil, fmt.Errorf("a binary attribute needs to be compared to a string")
		}
		return cmpBinary(e, ref)
	case "dateTime":
		date, ok := e.CompareValue.(string)
		if !ok {
			return nil, fmt.Errorf("a dateTime attribute needs to be compared to a string")
		}
		ref, err := datetime.Parse(date)
		if err != nil {
			return nil, fmt.Errorf("a dateTime attribute needs to be compared to a dateTime")
		}
		return cmpDateTime(e, date, ref)
	case "reference", "string":
		ref, ok := e.CompareValue.(string)
		if !ok {
			return nil, fmt.Errorf("a %s attribute needs to be compared to a string", typ)
		}
		return cmpString(e, attr, ref)
	case "boolean":
		ref, ok := e.CompareValue.(bool)
		if !ok {
			return nil, fmt.Errorf("a boolean attribute needs to be compared to a boolean")
		}
		return cmpBoolean(e, attr, ref)
	case "decimal":
		ref, ok := e.CompareValue.(float64)
		if !ok {
			return nil, fmt.Errorf("a decimal attribute needs to be compared to a float/int")
		}
		return cmpDecimal(e, ref)
	case "integer":
		ref, ok := e.CompareValue.(int)
		if !ok {
			return nil, fmt.Errorf("a integer attribute needs to be compared to a int")
		}
		return cmpInteger(e, ref)
	default:
		panic(fmt.Sprintf("unknown attribute type: %s", typ))
	}
}
