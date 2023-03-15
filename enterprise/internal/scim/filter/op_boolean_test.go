package filter

import (
	"fmt"
	"testing"

	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
)

func TestValidatorBoolean(t *testing.T) {
	var (
		exp = func(op filter.CompareOperator) string {
			return fmt.Sprintf("bool %s true", op)
		}
		ref = schema.Schema{
			Attributes: []schema.CoreAttribute{
				schema.SimpleCoreAttribute(schema.SimpleBooleanParams(schema.BooleanParams{
					Name: "bool",
				})),
			},
		}
		attr = map[string]interface{}{
			"bool": true,
		}
	)

	for _, test := range []struct {
		op    filter.CompareOperator
		valid bool // Whether the filter is valid.
	}{
		{filter.EQ, true},
		{filter.NE, false},
		{filter.CO, true},
		{filter.SW, true},
		{filter.EW, true},
		{filter.GT, false},
		{filter.LT, false},
		{filter.GE, false},
		{filter.LE, false},
	} {
		t.Run(string(test.op), func(t *testing.T) {
			f := exp(test.op)
			validator, err := NewValidator(f, ref)
			if err != nil {
				t.Fatal(err)
			}
			if err := validator.PassesFilter(attr); (err == nil) != test.valid {
				t.Errorf("%s %v | actual %v, expected %v", f, attr, err, test.valid)
			}
		})
	}
}
