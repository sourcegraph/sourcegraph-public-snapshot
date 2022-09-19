package promql

import (
	"testing"
)

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
			name:       "valid variable expression",
			expression: `foobar{foo="$var"}`, // "$variable" is valid promql
			wantErr:    false,
		},
		{
			name:       "invalid variable expression",
			expression: `foobar[$time]`, // not valid promql
			wantErr:    true,
		},
		{
			name:       "invalid expression fixed by vars",
			expression: `foobar[$time]`, // not valid promql
			vars:       VariableApplier{"time": "1m"},
			wantErr:    false,
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
