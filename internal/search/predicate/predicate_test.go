package predicate

import (
	"testing"
)

func TestParseAsPredicate(t *testing.T) {
	tests := []struct {
		input  string
		name   string
		params string
		err    bool
	}{
		{`()`, "", "", true},
		{`a()`, "a", "", false},
		{`a(b)`, "a", "b", false},
		{``, "", "", true},
		{`a)(`, "", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			name, params, err := ParseAsPredicate(tc.input)
			if tc.err {
				if err == nil {
					t.Fatal("expected err, but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %s", err)
			}

			if name != tc.name {
				t.Fatalf("expected name %s, got %s", tc.name, name)
			}

			if params != tc.params {
				t.Fatalf("expected params %s, got %s", tc.name, name)
			}
		})
	}

}
