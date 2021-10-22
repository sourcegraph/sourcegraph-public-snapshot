package apidocs

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestReverse(t *testing.T) {
	testCases := []struct {
		input string
		want  autogold.Value
	}{
		{"", autogold.Want("empty string", "")},
		{"h", autogold.Want("one character", "h")},
		{"asdf", autogold.Want("ascii", "fdsa")},
		{"as⃝df̅", autogold.Want("unicode with combining characters", "f̅ds⃝a")},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got := Reverse(tc.input)
			tc.want.Equal(t, got)
		})
	}
}

// TODO(apidocs): tests: Truncate
