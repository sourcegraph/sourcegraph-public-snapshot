package filter

import (
	"fmt"
	"testing"

	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
)

func TestValidatorInteger(t *testing.T) {
	var (
		exp = func(op filter.CompareOperator) string {
			return fmt.Sprintf("int %s 0", op)
		}
		ref = schema.Schema{
			Attributes: []schema.CoreAttribute{
				schema.SimpleCoreAttribute(schema.SimpleNumberParams(schema.NumberParams{
					Name: "int",
					Type: schema.AttributeTypeInteger(),
				})),
			},
		}
		attrs = [3]map[string]interface{}{
			{"int": -1}, // less
			{"int": 0},  // equal
			{"int": 10}, // greater
		}
	)

	for _, test := range []struct {
		op    filter.CompareOperator
		valid [3]bool
	}{
		{filter.EQ, [3]bool{false, true, false}},
		{filter.NE, [3]bool{true, false, true}},
		{filter.CO, [3]bool{false, true, true}},
		{filter.SW, [3]bool{false, true, false}},
		{filter.EW, [3]bool{false, true, true}},
		{filter.GT, [3]bool{false, false, true}},
		{filter.LT, [3]bool{true, false, false}},
		{filter.GE, [3]bool{false, true, true}},
		{filter.LE, [3]bool{true, true, false}},
	} {
		t.Run(string(test.op), func(t *testing.T) {
			f := exp(test.op)
			validator, err := NewValidator(f, ref)
			if err != nil {
				t.Fatal(err)
			}
			for i, attr := range attrs {
				if err := validator.PassesFilter(attr); (err == nil) != test.valid[i] {
					t.Errorf("(%d) %s %v | actual %v, expected %v", i, f, attr, err, test.valid[i])
				}
			}
		})
	}
}
