package graphqlbackend

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetP4ChangelistID(t *testing.T) {
	testCases := []struct {
		input                string
		expectedChangeListID string
	}{
		{
			input:                `[git-p4: depot-paths = "//test-perms/": change = 83725]`,
			expectedChangeListID: "83725",
		},
		{
			input:                `[p4-fusion: depot-paths = "//test-perms/": change = 80972]`,
			expectedChangeListID: "80972",
		},
		{
			input:                "invalid string",
			expectedChangeListID: "",
		},
		{
			input:                "",
			expectedChangeListID: "",
		},
	}

	for _, tc := range testCases {
		result, err := getP4ChangelistID(tc.input)
		if tc.expectedChangeListID != "" {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}

		if !reflect.DeepEqual(result, tc.expectedChangeListID) {
			t.Errorf("getP4ChangelistID failed (%q) => got %q, want %q", tc.input, result, tc.expectedChangeListID)
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
			expectedErr:     `failed to parse commit subject "abc1234 - updating config" for commit converted by p4-fusion`,
		},
		{
			input:           "- fixing bug",
			expectedSubject: "",
			expectedErr:     `failed to parse commit subject "- fixing bug" for commit converted by p4-fusion`,
		},
		{
			input:           "fixing bug",
			expectedSubject: "",
			expectedErr:     `failed to parse commit subject "fixing bug" for commit converted by p4-fusion`,
		},
	}

	for _, tc := range testCases {
		subject, err := parseP4FusionCommitSubject(tc.input)
		if err != nil && err.Error() != tc.expectedErr {
			t.Errorf("Expected error %q, got %q", err.Error(), tc.expectedErr)
		}

		if subject != tc.expectedSubject {
			t.Errorf("Expected subject %q, got %q", tc.expectedSubject, subject)
		}
	}
}
