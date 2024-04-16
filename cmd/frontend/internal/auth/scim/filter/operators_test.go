package filter

import (
	"testing"

	"github.com/elimity-com/scim/schema"
)

// TestValidatorInvalidResourceTypes contains all the cases where an *errors.ScimError gets returned.
func TestValidatorInvalidResourceTypes(t *testing.T) {
	for _, test := range []struct {
		name     string
		filter   string
		attr     schema.CoreAttribute
		resource map[string]interface{}
	}{
		{
			"string", `attr eq "value"`,
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name: "attr",
			})),
			map[string]interface{}{
				"attr": 1, // expects a string
			},
		},
		{
			"stringMv", `attr eq "value"`,
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:        "attr",
				MultiValued: true,
			})),
			map[string]interface{}{
				"attr": []interface{}{1}, // expects a []interface{string}
			},
		},
		{
			"stringMv",
			`attr eq "value"`,
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:        "attr",
				MultiValued: true,
			})),
			map[string]interface{}{
				"attr": []string{"value"}, // expects a []interface{}
			},
		},
		{
			"dateTime", `attr eq "2006-01-02T15:04:05"`,
			schema.SimpleCoreAttribute(schema.SimpleDateTimeParams(schema.DateTimeParams{
				Name: "attr",
			})),
			map[string]interface{}{
				"attr": 1, // expects a string
			},
		},
		{
			"dateTime", `attr eq "2006-01-02T15:04:05"`,
			schema.SimpleCoreAttribute(schema.SimpleDateTimeParams(schema.DateTimeParams{
				Name: "attr",
			})),
			map[string]interface{}{
				"attr": "2006-01-02T", // expects a valid dateTime
			},
		},
		{
			"boolean", `attr eq true`,
			schema.SimpleCoreAttribute(schema.SimpleBooleanParams(schema.BooleanParams{
				Name: "attr",
			})),
			map[string]interface{}{
				"attr": 1, // expects a boolean
			},
		},
		{
			"decimal", `attr eq 0.0`,
			schema.SimpleCoreAttribute(schema.SimpleNumberParams(schema.NumberParams{
				Name: "attr",
				Type: schema.AttributeTypeDecimal(),
			})),
			map[string]interface{}{
				"attr": "0", // expects a decimal value
			},
		},
		{
			"integer", `attr eq 0`,
			schema.SimpleCoreAttribute(schema.SimpleNumberParams(schema.NumberParams{
				Name: "attr",
				Type: schema.AttributeTypeInteger(),
			})),
			map[string]interface{}{
				"attr": 0.0, // expects an integer
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			validator, err := NewValidator(test.filter, schema.Schema{
				Attributes: []schema.CoreAttribute{test.attr},
			})
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := recover(); err == nil {
					t.Fatal(test)
				}
			}()
			if err := validator.PassesFilter(test.resource); err != nil {
				t.Error(err)
			}
		})
	}
}
