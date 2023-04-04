package filter

import (
	"fmt"
	"strings"

	"github.com/scim2/filter-parser/v2"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func cmpBool(ref bool, cmp func(v, ref bool) error) func(interface{}) error {
	return func(i interface{}) error {
		value, ok := i.(bool)
		if !ok {
			panic(fmt.Sprintf("given value is not a boolean: %v", i))
		}
		return cmp(value, ref)
	}
}

func cmpBoolStr(ref bool, cmp func(v, ref string) error) (func(interface{}) error, error) {
	return func(i interface{}) error {
		if _, ok := i.(bool); !ok {
			panic(fmt.Sprintf("given value is not a boolean: %v", i))
		}
		return cmp(fmt.Sprintf("%t", i), fmt.Sprintf("%t", ref))
	}, nil
}

// cmpBoolean returns a compare function that compares a given value to the reference boolean based on the given
// attribute expression and string/reference attribute. The filter operators gt, lt, ge and le are not supported on
// boolean attributes.
//
// Expects a boolean attribute. Will panic on unknown filter operator.
// Known operators: eq, ne, co, sw, ew, gt, lt, ge and le.
func cmpBoolean(e *filter.AttributeExpression, ref bool) (func(interface{}) error, error) {
	switch op := e.Operator; op {
	case filter.EQ:
		return cmpBool(ref, func(v, ref bool) error {
			if v != ref {
				return errors.Newf("%t is not equal to %t", v, ref)
			}
			return nil
		}), nil
	case filter.NE:
		return cmpBool(ref, func(v, ref bool) error {
			if v == ref {
				return errors.Newf("%t is equal to %t", v, ref)
			}
			return nil
		}), nil
	case filter.CO:
		return cmpBoolStr(ref, func(v, ref string) error {
			if !strings.Contains(v, ref) {
				return errors.Newf("%s does not contain %s", v, ref)
			}
			return nil
		})
	case filter.SW:
		return cmpBoolStr(ref, func(v, ref string) error {
			if !strings.HasPrefix(v, ref) {
				return errors.Newf("%s does not start with %s", v, ref)
			}
			return nil
		})
	case filter.EW:
		return cmpBoolStr(ref, func(v, ref string) error {
			if !strings.HasSuffix(v, ref) {
				return errors.Newf("%s does not end with %s", v, ref)
			}
			return nil
		})
	case filter.GT, filter.LT, filter.GE, filter.LE:
		return nil, errors.Newf("can not use op %q on boolean values", op)
	default:
		panic(fmt.Sprintf("unknown operator in expression: %s", e))
	}
}
