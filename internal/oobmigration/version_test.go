package oobmigration

import "testing"

func TestCompareVersions(t *testing.T) {
	testCases := []struct {
		left     Version
		right    Version
		expected VersionOrder
		err      bool
	}{
		{left: NewVersion(3, 12), right: NewVersion(3, 12), expected: VersionOrderEqual},
		{left: NewVersion(3, 11), right: NewVersion(3, 12), expected: VersionOrderBefore},
		{left: NewVersion(3, 12), right: NewVersion(3, 11), expected: VersionOrderAfter},
		{left: NewVersion(3, 12), right: NewVersion(4, 11), expected: VersionOrderBefore},
		{left: NewVersion(4, 11), right: NewVersion(3, 12), expected: VersionOrderAfter},
	}

	for _, testCase := range testCases {
		order := compareVersions(testCase.left, testCase.right)
		if order != testCase.expected {
			t.Errorf("unexpected order. want=%d have=%d", testCase.expected, order)
		}
	}
}
