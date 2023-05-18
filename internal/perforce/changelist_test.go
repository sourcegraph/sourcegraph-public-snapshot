package perforce

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
			input:                `[git-p4: depot-path = "//test-perms/": change = 83725]`,
			expectedChangeListID: "83725",
		},
		{
			input:                `[p4-fusion: depot-paths = "//test-perms/": change = 80972]`,
			expectedChangeListID: "80972",
		},
		{
			input:                `[p4-fusion: depot-path = "//test-perms/": change = 80972]`,
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
		result, err := GetP4ChangelistID(tc.input)
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
