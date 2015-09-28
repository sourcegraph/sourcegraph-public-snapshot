package spec

import (
	"regexp"
	"testing"
)

func TestHost(t *testing.T) {
	pat, err := regexp.Compile("^" + host + "$")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input     string
		wantMatch bool
	}{
		{"x", true},
		{"xx", true},
		{"x.y", true},
		{"xx.yy", true},
		{"x.y.z", true},
		{"x-y", true},
		{"w-x.y-z", true},

		{"", false},
		{".", false},
		{"x.", false},
		{".x", false},
	}
	for _, test := range tests {
		match := pat.MatchString(test.input)
		if match != test.wantMatch {
			t.Errorf("%q: got match == %v, want %v", test.input, match, test.wantMatch)
		}
	}
}
