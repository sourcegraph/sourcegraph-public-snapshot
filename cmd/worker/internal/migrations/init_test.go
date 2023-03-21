package migrations

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

func TestParseVersion(t *testing.T) {
	testCases := []struct {
		input    string
		expected oobmigration.Version
	}{
		{"v4.5.3", oobmigration.NewVersion(4, 5)},
		{"201149_2023-02-23_4.5-dc8d16268c06", oobmigration.NewVersion(4, 5)},
		{"201149_2023-02-23_4.5-dc8d16268c06_patch", oobmigration.NewVersion(4, 5)},
		{"main-dry-run-ef-un-revert_201149_2023-02-23_4.5-dc8d16268c06", oobmigration.NewVersion(4, 5)},
		{"main-dry-run-ef-un-revert_201149_2023-02-23_4.5-dc8d16268c06_patch", oobmigration.NewVersion(4, 5)},
	}

	for _, testCase := range testCases {
		version, ok := parseVersion(testCase.input)
		if !ok {
			t.Errorf("unexpected unparseable version %q", testCase.input)
		} else if version.String() != testCase.expected.String() {
			t.Errorf("unexpected version for %q. want=%q have=%q", testCase.input, testCase.expected, version)
		}
	}
}
