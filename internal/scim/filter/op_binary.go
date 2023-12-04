package filter

import (
	"fmt"
	"strings"

	"github.com/scim2/filter-parser/v2"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// cmpBinary returns a compare function that compares a given value to the reference string based on the given attribute
// expression and binary attribute. The filter operators gt, lt, ge and le are not supported on binary attributes.
//
// Expects a binary attribute. Will panic on unknown filter operator.
// Known operators: eq, ne, co, sw, ew, gt, lt, ge and le.
func cmpBinary(e *filter.AttributeExpression, ref string) (func(interface{}) error, error) {
	switch op := e.Operator; op {
	case filter.EQ:
		return cmpStr(ref, true, func(v, ref string) error {
			if v != ref {
				return errors.Newf("%s is not equal to %s", v, ref)
			}
			return nil
		})
	case filter.NE:
		return cmpStr(ref, true, func(v, ref string) error {
			if v == ref {
				return errors.Newf("%s is equal to %s", v, ref)
			}
			return nil
		})
	case filter.CO:
		return cmpStr(ref, true, func(v, ref string) error {
			if !strings.Contains(v, ref) {
				return errors.Newf("%s does not contain %s", v, ref)
			}
			return nil
		})
	case filter.SW:
		return cmpStr(ref, true, func(v, ref string) error {
			if !strings.HasPrefix(v, ref) {
				return errors.Newf("%s does not start with %s", v, ref)
			}
			return nil
		})
	case filter.EW:
		return cmpStr(ref, true, func(v, ref string) error {
			if !strings.HasSuffix(v, ref) {
				return errors.Newf("%s does not end with %s", v, ref)
			}
			return nil
		})
	case filter.GT, filter.LT, filter.GE, filter.LE:
		return nil, errors.Newf("can not use op %q on binary values", op)
	default:
		panic(fmt.Sprintf("unknown operator in expression: %s", e))
	}
}
