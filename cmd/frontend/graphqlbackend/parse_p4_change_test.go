package graphqlbackend

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/log/logtest"
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
			input: `[p4-fusion: depot-paths = "//test-perms/": change = 80972]`,
			expected: &P4Change{
				DepotPaths: "//test-perms/",
				Change:     "80972",
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

func TestParseP4FusionCommitSubject(t *testing.T) {
	testCases := []struct {
		input           string
		expectedSubject string
		expectedErr     string
	}{
		{
			input:           "83732 - adding sourcegraph repos",
			expectedSubject: "adding sourcegraph repos",
		},
		{
			input:           "abc1234 - updating config",
			expectedSubject: "",
			expectedErr:     `failed to parse commit subject "abc1234 - updating config" for commit that was converted by p4-fusion`,
		},
		{
			input:           "- fixing bug",
			expectedSubject: "",
			expectedErr:     `failed to parse commit subject "- fixing bug" for commit that was converted by p4-fusion`,
		},
	}

	logger := logtest.Scoped(t)
	for _, tc := range testCases {
		subject, err := parseP4FusionCommitSubject(logger, tc.input)
		if subject != tc.expectedSubject {
			t.Errorf("Expected subject %q, got %q", tc.expectedSubject, subject)
		}

		if err != nil && err.Error() != tc.expectedErr {
			t.Errorf("Expected error %q, got %q", err.Error(), tc.expectedErr)
		}
	}
}
