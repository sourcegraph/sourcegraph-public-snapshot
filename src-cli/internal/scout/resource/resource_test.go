package resource

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGetMemUnits(t *testing.T) {
	cases := []struct {
		name      string
		param     int64
		wantUnit  string
		wantValue int64
		wantError error
	}{
		{
			name:      "convert bytes below a million to KB",
			param:     999999,
			wantUnit:  "KB",
			wantValue: 999999,
			wantError: nil,
		},
		{
			name:      "convert bytes below a billion to MB",
			param:     999999999,
			wantUnit:  "MB",
			wantValue: 999,
			wantError: nil,
		},
		{
			name:      "convert bytes above a billion to GB",
			param:     12999999900,
			wantUnit:  "GB",
			wantValue: 12,
			wantError: nil,
		},
		{
			name:      "return error for a negative number",
			param:     -300,
			wantUnit:  "",
			wantValue: -300,
			wantError: errors.Newf("invalid memory value: %d", -300),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gotUnit, gotValue, gotError := getMemUnits(tc.param)

			if gotUnit != tc.wantUnit {
				t.Errorf("got %s want %s", gotUnit, tc.wantUnit)
			}

			if gotValue != tc.wantValue {
				t.Errorf("got %v want %v", gotValue, tc.wantValue)
			}

			if gotError == nil && tc.wantError != nil {
				t.Error("got nil want error")
			}

			if gotError != nil && tc.wantError == nil {
				t.Error("got error want nil")
			}
		})
	}
}
