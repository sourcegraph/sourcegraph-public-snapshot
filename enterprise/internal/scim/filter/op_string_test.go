package filter

import (
	"fmt"
	"testing"

	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
)

func TestValidatorString(t *testing.T) {
	var (
		exp = func(op filter.CompareOperator) string {
			return fmt.Sprintf("str %s \"x\"", op)
		}
		attrs = [3]map[string]interface{}{
			{"str": "x"},
			{"str": "X"},
			{"str": "y"},
		}
	)

	for _, test := range []struct {
		op      filter.CompareOperator
		valid   [3]bool
		validCE [3]bool
	}{
		{filter.EQ, [3]bool{true, true, false}, [3]bool{true, false, false}},
		{filter.NE, [3]bool{false, false, true}, [3]bool{false, true, true}},
		{filter.CO, [3]bool{true, true, false}, [3]bool{true, false, false}},
		{filter.SW, [3]bool{true, true, false}, [3]bool{true, false, false}},
		{filter.EW, [3]bool{true, true, false}, [3]bool{true, false, false}},
		{filter.GT, [3]bool{false, false, true}, [3]bool{false, false, true}},
		{filter.LT, [3]bool{false, false, false}, [3]bool{false, true, false}},
		{filter.GE, [3]bool{true, true, true}, [3]bool{true, false, true}},
		{filter.LE, [3]bool{true, true, false}, [3]bool{true, true, false}},
	} {
		t.Run(string(test.op), func(t *testing.T) {
			f := exp(test.op)
			for i, attr := range attrs {
				validator, err := NewValidator(f, schema.Schema{
					Attributes: []schema.CoreAttribute{
						schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
							Name: "str",
						})),
					},
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := validator.PassesFilter(attr); (err == nil) != test.valid[i] {
					t.Errorf("(0.%d) %s %s | actual %v, expected %v", i, f, attr, err, test.valid[i])
				}
				validatorCE, err := NewValidator(f, schema.Schema{
					Attributes: []schema.CoreAttribute{
						schema.SimpleCoreAttribute(schema.SimpleReferenceParams(schema.ReferenceParams{
							Name: "str",
						})),
					},
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := validatorCE.PassesFilter(attr); (err == nil) != test.validCE[i] {
					t.Errorf("(1.%d) %s %s | actual %v, expected %v", i, f, attr, err, test.validCE[i])
				}
			}
		})
	}
}
