package filter

import (
	"fmt"
	"strings"

	"github.com/scim2/filter-parser/v2"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// cmpDecimal returns a compare function that compares a given value to the reference float based on the given attribute
// expression and decimal attribute.
//
// Expects a decimal attribute. Will panic on unknown filter operator.
// Known operators: eq, ne, co, sw, ew, gt, lt, ge and le.
func cmpDecimal(e *filter.AttributeExpression, ref float64) (func(interface{}) error, error) {
	switch op := e.Operator; op {
	case filter.EQ:
		return cmpFloat(ref, func(v, ref float64) error {
			if v != ref {
				return errors.Newf("%f is not equal to %f", v, ref)
			}
			return nil
		}), nil
	case filter.NE:
		return cmpFloat(ref, func(v, ref float64) error {
			if v == ref {
				return errors.Newf("%f is equal to %f", v, ref)
			}
			return nil
		}), nil
	case filter.CO:
		return cmpFloatStr(ref, func(v, ref string) error {
			if !strings.Contains(v, ref) {
				return errors.Newf("%s does not contain %s", v, ref)
			}
			return nil
		})
	case filter.SW:
		return cmpFloatStr(ref, func(v, ref string) error {
			if !strings.HasPrefix(v, ref) {
				return errors.Newf("%s does not start with %s", v, ref)
			}
			return nil
		})
	case filter.EW:
		return cmpFloatStr(ref, func(v, ref string) error {
			if !strings.HasSuffix(v, ref) {
				return errors.Newf("%s does not end with %s", v, ref)
			}
			return nil
		})
	case filter.GT:
		return cmpFloat(ref, func(v, ref float64) error {
			if v <= ref {
				return errors.Newf("%f is not greater than %f", v, ref)
			}
			return nil
		}), nil
	case filter.LT:
		return cmpFloat(ref, func(v, ref float64) error {
			if v >= ref {
				return errors.Newf("%f is not less than %f", v, ref)
			}
			return nil
		}), nil
	case filter.GE:
		return cmpFloat(ref, func(v, ref float64) error {
			if v < ref {
				return errors.Newf("%f is not greater or equal to %f", v, ref)
			}
			return nil
		}), nil
	case filter.LE:
		return cmpFloat(ref, func(v, ref float64) error {
			if v > ref {
				return errors.Newf("%f is not less or equal to %f", v, ref)
			}
			return nil
		}), nil
	default:
		panic(fmt.Sprintf("unknown operator in expression: %s", e))
	}
}

func cmpFloat(ref float64, cmp func(v, ref float64) error) func(interface{}) error {
	return func(i interface{}) error {
		f, ok := i.(float64)
		if !ok {
			panic(fmt.Sprintf("given value is not a float: %v", i))
		}
		return cmp(f, ref)
	}
}

func cmpFloatStr(ref float64, cmp func(v, ref string) error) (func(interface{}) error, error) {
	return func(i interface{}) error {
		if _, ok := i.(float64); !ok {
			panic(fmt.Sprintf("given value is not a float: %v", i))
		}
		// fmt.Sprintf("%f") would give them both the same precision.
		return cmp(fmt.Sprint(i), fmt.Sprint(ref))
	}, nil
}
