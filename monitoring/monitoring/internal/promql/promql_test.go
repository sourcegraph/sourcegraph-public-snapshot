package promql

import (
	"strings"
	"testing"
)

type mapVariableApplier map[string]string

var _ VariableApplier = &mapVariableApplier{}

func (m mapVariableApplier) ApplyDefaults(expression string) string {
	for k, def := range m {
		expression = strings.ReplaceAll(expression, k, def)
	}
	return expression
}

func TestValidate(t *testing.T) {
	for _, tc := range []struct {
		name       string
		expression string
		vars       VariableApplier

		wantErr bool
	}{
		{
			name:       "valid expression",
			expression: "foobar",
			wantErr:    false,
		},
		{
			name:       "invalid expression",
			expression: "foobar{foo=$var}", // $variable is not valid promql
			wantErr:    true,
		},
		{
			name:       "invalid expression fixed by vars",
			expression: "foobar{foo=$var}", // $variable is not valid promql
			vars: mapVariableApplier{
				"$var": "'bar'",
			},
			wantErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.expression, tc.vars)
			if (err != nil) != tc.wantErr {
				t.Errorf("unexpected result '%+v'", err)
			}
		})
	}
}
