package oobmigration

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewVersionFromString(t *testing.T) {
	testCases := []struct {
		v       string
		version Version
		patch   int
		ok      bool
	}{
		{"3.50", NewVersion(3, 50), 0, true},
		{"v3.50.3", NewVersion(3, 50), 3, true},
		{"v3.50", NewVersion(3, 50), 0, true},
		{"3.50.3", NewVersion(3, 50), 3, true},
		{"3.50.3+dev", newDevVersion(3, 50), 3, true},
		{"350", Version{}, 0, false},
		{"350+dev", Version{}, 0, false},
		{"2023.03.23+204874.db2922", NewVersion(2023, 03), 23, true},          // Cody App
		{"2023.03.23-insiders+204874.db2922", NewVersion(2023, 03), 23, true}, // Cody App
	}

	for _, testCase := range testCases {
		t.Run(testCase.v, func(t *testing.T) {
			version, patch, ok := NewVersionAndPatchFromString(testCase.v)
			if ok != testCase.ok {
				t.Errorf("unexpected ok. want=%v have=%v", testCase.ok, ok)
			} else {
				if version != testCase.version {
					t.Errorf("unexpected version. want=%s have=%s", testCase.version, version)
				}
				if patch != testCase.patch {
					t.Errorf("unexpected patch. want=%d have=%d", testCase.patch, patch)
				}
			}
		})
	}
}

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
		{from: Version{Major: 3, Minor: 12}, to: Version{Major: 3, Minor: 10}, err: true},
		{from: Version{Major: 3, Minor: 12}, to: Version{Major: 3, Minor: 12}, err: true},
		{from: Version{Major: 3, Minor: 12}, to: Version{Major: 3, Minor: 13}, expected: []Version{{Major: 3, Minor: 12}, {Major: 3, Minor: 13}}},
		{from: Version{Major: 3, Minor: 12}, to: Version{Major: 3, Minor: 16}, expected: []Version{{Major: 3, Minor: 12}, {Major: 3, Minor: 13}, {Major: 3, Minor: 14}, {Major: 3, Minor: 15}, {Major: 3, Minor: 16}}},
		{from: Version{Major: 3, Minor: 42}, to: Version{Major: 4, Minor: 2}, expected: []Version{{Major: 3, Minor: 42}, {Major: 3, Minor: 43}, {Major: 4}, {Major: 4, Minor: 1}, {Major: 4, Minor: 2}}},
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

func TestNextPrevious(t *testing.T) {
	chain := []Version{
		NewVersion(3, 41),
		NewVersion(3, 42),
		NewVersion(3, 43),
		NewVersion(4, 0),
		NewVersion(4, 1),
		NewVersion(4, 2),
	}

	for i, version := range chain {
		if i != 0 {
			previous, ok := version.Previous()
			if !ok {
				t.Fatalf("no previous for %q", version)
			}
			if have, want := chain[i-1], previous; have.String() != want.String() {
				t.Fatalf("unexpected previous for %q. want=%q have=%q", version, want, have)
			}
		}

		if i+1 < len(chain) {
			if have, want := version.Next(), chain[i+1]; have.String() != want.String() {
				t.Fatalf("unexpected next for %q. want=%q have=%q", version, want, have)
			}
		}
	}
}
