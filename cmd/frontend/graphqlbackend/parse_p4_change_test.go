package graphqlbackend

import (
	"reflect"
	"testing"
)

func TestParseP4Change(t *testing.T) {
	testCases := []struct {
		input    string
		expected *P4Change
	}{
		{
			input: `[git-p4: depot-paths = "//test-perms/": change = 83725]`,
			expected: &P4Change{
				DepotPaths: "//test-perms/",
				Change:     "83725",
			},
		},
		{
			input: `[git-p4: depot-paths = "//depot/": change = 12345]`,
			expected: &P4Change{
				DepotPaths: "//depot/",
				Change:     "12345",
			},
		},
		{
			input: `[git-p4: depot-paths = "//src/": change = 1]`,
			expected: &P4Change{
				DepotPaths: "//src/",
				Change:     "1",
			},
		},
		{
			input:    "invalid string",
			expected: nil,
		},
		{
			input:    "",
			expected: nil,
		},
	}

	for _, tc := range testCases {
		result := ParseP4Change(tc.input)
		if !reflect.DeepEqual(result, tc.expected) {
			t.Errorf("ParseP4Change(%q) => %v, want %v", tc.input, result, tc.expected)
		}
	}
}
