package filter

import (
	"fmt"
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
	"strings"
)

func cmpStr(ref string, caseExact bool, cmp func(v, ref string) error) (func(interface{}) error, error) {
	if caseExact {
		return func(i interface{}) error {
			value, ok := i.(string)
			if !ok {
				panic(fmt.Sprintf("given value is not a string: %v", i))
			}
			return cmp(value, ref)
		}, nil
	}
	return func(i interface{}) error {
		value, ok := i.(string)
		if !ok {
			panic(fmt.Sprintf("given value is not a string: %v", i))
		}
		return cmp(strings.ToLower(value), strings.ToLower(ref))
	}, nil
}

// cmpString returns a compare function that compares a given value to the reference string based on the given attribute
// expression and string/reference attribute.
//
// Expects a string/reference attribute. Will panic on unknown filter operator.
// Known operators: eq, ne, co, sw, ew, gt, lt, ge and le.
func cmpString(e *filter.AttributeExpression, attr schema.CoreAttribute, ref string) (func(interface{}) error, error) {
	switch op := e.Operator; op {
	case filter.EQ:
		return cmpStr(ref, attr.CaseExact(), func(v, ref string) error {
			if v != ref {
				return fmt.Errorf("%s is not equal to %s", v, ref)
			}
			return nil
		})
	case filter.NE:
		return cmpStr(ref, attr.CaseExact(), func(v, ref string) error {
			if v == ref {
				return fmt.Errorf("%s is equal to %s", v, ref)
			}
			return nil
		})
	case filter.CO:
		return cmpStr(ref, attr.CaseExact(), func(v, ref string) error {
			if !strings.Contains(v, ref) {
				return fmt.Errorf("%s does not contain %s", v, ref)
			}
			return nil
		})
	case filter.SW:
		return cmpStr(ref, attr.CaseExact(), func(v, ref string) error {
			if !strings.HasPrefix(v, ref) {
				return fmt.Errorf("%s does not start with %s", v, ref)
			}
			return nil
		})
	case filter.EW:
		return cmpStr(ref, attr.CaseExact(), func(v, ref string) error {
			if !strings.HasSuffix(v, ref) {
				return fmt.Errorf("%s does not end with %s", v, ref)
			}
			return nil
		})
	case filter.GT:
		return cmpStr(ref, attr.CaseExact(), func(v, ref string) error {
			if strings.Compare(v, ref) <= 0 {
				return fmt.Errorf("%s is not lexicographically greater than %s", v, ref)
			}
			return nil
		})
	case filter.LT:
		return cmpStr(ref, attr.CaseExact(), func(v, ref string) error {
			if strings.Compare(v, ref) >= 0 {
				return fmt.Errorf("%s is not lexicographically less than %s", v, ref)
			}
			return nil
		})
	case filter.GE:
		return cmpStr(ref, attr.CaseExact(), func(v, ref string) error {
			if strings.Compare(v, ref) < 0 {
				return fmt.Errorf("%s is not lexicographically greater or equal to %s", v, ref)
			}
			return nil
		})
	case filter.LE:
		return cmpStr(ref, attr.CaseExact(), func(v, ref string) error {
			if strings.Compare(v, ref) > 0 {
				return fmt.Errorf("%s is not lexicographically less or equal to %s", v, ref)
			}
			return nil
		})
	default:
		panic(fmt.Sprintf("unknown operator in expression: %s", e))
	}
}
