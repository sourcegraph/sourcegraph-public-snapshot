package graphqlbackend

import (
	"testing"
)

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
