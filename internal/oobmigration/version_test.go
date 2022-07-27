package oobmigration

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCompareVersions(t *testing.T) {
	testCases := []struct {
		left     Version
		right    Version
		expected VersionOrder
	}{
		{left: NewVersion(3, 12), right: NewVersion(3, 12), expected: VersionOrderEqual},
		{left: NewVersion(3, 11), right: NewVersion(3, 12), expected: VersionOrderBefore},
		{left: NewVersion(3, 12), right: NewVersion(3, 11), expected: VersionOrderAfter},
		{left: NewVersion(3, 12), right: NewVersion(4, 11), expected: VersionOrderBefore},
		{left: NewVersion(4, 11), right: NewVersion(3, 12), expected: VersionOrderAfter},
	}

	for _, testCase := range testCases {
		order := CompareVersions(testCase.left, testCase.right)
		if order != testCase.expected {
			t.Errorf("unexpected order. want=%d have=%d", testCase.expected, order)
		}
	}
}

func TestUpgradeRange(t *testing.T) {
	testCases := []struct {
		from     Version
		to       Version
		expected []Version
		err      bool
	}{
		{from: Version{3, 12}, to: Version{3, 10}, err: true},
		{from: Version{3, 12}, to: Version{3, 12}, err: true},
		{from: Version{3, 12}, to: Version{3, 13}, expected: []Version{{3, 12}, {3, 13}}},
		{from: Version{3, 12}, to: Version{3, 16}, expected: []Version{{3, 12}, {3, 13}, {3, 14}, {3, 15}, {3, 16}}},
		{from: Version{3, 42}, to: Version{4, 2}, expected: []Version{{3, 42}, {3, 43}, {4, 0}, {4, 1}, {4, 2}}},
	}

	for _, testCase := range testCases {
		versions, err := UpgradeRange(testCase.from, testCase.to)
		if err != nil {
			if testCase.err {
				continue
			}

			t.Fatalf("unexpected error: %s", err)
		}
		if testCase.err {
			t.Errorf("expected error")
		} else {
			if diff := cmp.Diff(testCase.expected, versions); diff != "" {
				t.Errorf("unexpected versions (-want +got):\n%s", diff)
			}
		}
	}
}
