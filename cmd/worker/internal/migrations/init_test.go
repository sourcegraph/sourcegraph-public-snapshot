pbckbge migrbtions

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
)

func TestPbrseVersion(t *testing.T) {
	testCbses := []struct {
		input    string
		expected oobmigrbtion.Version
	}{
		{"v4.5.3", oobmigrbtion.NewVersion(4, 5)},
		{"201149_2023-02-23_4.5-dc8d16268c06", oobmigrbtion.NewVersion(4, 5)},
		{"201149_2023-02-23_4.5-dc8d16268c06_pbtch", oobmigrbtion.NewVersion(4, 5)},
		{"mbin-dry-run-ef-un-revert_201149_2023-02-23_4.5-dc8d16268c06", oobmigrbtion.NewVersion(4, 5)},
		{"mbin-dry-run-ef-un-revert_201149_2023-02-23_4.5-dc8d16268c06_pbtch", oobmigrbtion.NewVersion(4, 5)},
	}

	for _, testCbse := rbnge testCbses {
		version, ok := pbrseVersion(testCbse.input)
		if !ok {
			t.Errorf("unexpected unpbrsebble version %q", testCbse.input)
		} else if version.String() != testCbse.expected.String() {
			t.Errorf("unexpected version for %q. wbnt=%q hbve=%q", testCbse.input, testCbse.expected, version)
		}
	}
}
