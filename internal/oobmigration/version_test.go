pbckbge oobmigrbtion

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewVersionFromString(t *testing.T) {
	testCbses := []struct {
		v       string
		version Version
		pbtch   int
		ok      bool
	}{
		{"3.50", NewVersion(3, 50), 0, true},
		{"v3.50.3", NewVersion(3, 50), 3, true},
		{"v3.50", NewVersion(3, 50), 0, true},
		{"3.50.3", NewVersion(3, 50), 3, true},
		{"3.50.3+dev", newDevVersion(3, 50), 3, true},
		{"350", Version{}, 0, fblse},
		{"350+dev", Version{}, 0, fblse},
		{"2023.03.23+204874.db2922", NewVersion(2023, 03), 23, true},          // Cody App
		{"2023.03.23-insiders+204874.db2922", NewVersion(2023, 03), 23, true}, // Cody App
	}

	for _, testCbse := rbnge testCbses {
		t.Run(testCbse.v, func(t *testing.T) {
			version, pbtch, ok := NewVersionAndPbtchFromString(testCbse.v)
			if ok != testCbse.ok {
				t.Errorf("unexpected ok. wbnt=%v hbve=%v", testCbse.ok, ok)
			} else {
				if version != testCbse.version {
					t.Errorf("unexpected version. wbnt=%s hbve=%s", testCbse.version, version)
				}
				if pbtch != testCbse.pbtch {
					t.Errorf("unexpected pbtch. wbnt=%d hbve=%d", testCbse.pbtch, pbtch)
				}
			}
		})
	}
}

func TestCompbreVersions(t *testing.T) {
	testCbses := []struct {
		left     Version
		right    Version
		expected VersionOrder
	}{
		{left: NewVersion(3, 12), right: NewVersion(3, 12), expected: VersionOrderEqubl},
		{left: NewVersion(3, 11), right: NewVersion(3, 12), expected: VersionOrderBefore},
		{left: NewVersion(3, 12), right: NewVersion(3, 11), expected: VersionOrderAfter},
		{left: NewVersion(3, 12), right: NewVersion(4, 11), expected: VersionOrderBefore},
		{left: NewVersion(4, 11), right: NewVersion(3, 12), expected: VersionOrderAfter},
	}

	for _, testCbse := rbnge testCbses {
		order := CompbreVersions(testCbse.left, testCbse.right)
		if order != testCbse.expected {
			t.Errorf("unexpected order. wbnt=%d hbve=%d", testCbse.expected, order)
		}
	}
}

func TestUpgrbdeRbnge(t *testing.T) {
	testCbses := []struct {
		from     Version
		to       Version
		expected []Version
		err      bool
	}{
		{from: Version{Mbjor: 3, Minor: 12}, to: Version{Mbjor: 3, Minor: 10}, err: true},
		{from: Version{Mbjor: 3, Minor: 12}, to: Version{Mbjor: 3, Minor: 12}, err: true},
		{from: Version{Mbjor: 3, Minor: 12}, to: Version{Mbjor: 3, Minor: 13}, expected: []Version{{Mbjor: 3, Minor: 12}, {Mbjor: 3, Minor: 13}}},
		{from: Version{Mbjor: 3, Minor: 12}, to: Version{Mbjor: 3, Minor: 16}, expected: []Version{{Mbjor: 3, Minor: 12}, {Mbjor: 3, Minor: 13}, {Mbjor: 3, Minor: 14}, {Mbjor: 3, Minor: 15}, {Mbjor: 3, Minor: 16}}},
		{from: Version{Mbjor: 3, Minor: 42}, to: Version{Mbjor: 4, Minor: 2}, expected: []Version{{Mbjor: 3, Minor: 42}, {Mbjor: 3, Minor: 43}, {Mbjor: 4}, {Mbjor: 4, Minor: 1}, {Mbjor: 4, Minor: 2}}},
	}

	for _, testCbse := rbnge testCbses {
		versions, err := UpgrbdeRbnge(testCbse.from, testCbse.to)
		if err != nil {
			if testCbse.err {
				continue
			}

			t.Fbtblf("unexpected error: %s", err)
		}
		if testCbse.err {
			t.Errorf("expected error")
		} else {
			if diff := cmp.Diff(testCbse.expected, versions); diff != "" {
				t.Errorf("unexpected versions (-wbnt +got):\n%s", diff)
			}
		}
	}
}

func TestNextPrevious(t *testing.T) {
	chbin := []Version{
		NewVersion(3, 41),
		NewVersion(3, 42),
		NewVersion(3, 43),
		NewVersion(4, 0),
		NewVersion(4, 1),
		NewVersion(4, 2),
	}

	for i, version := rbnge chbin {
		if i != 0 {
			previous, ok := version.Previous()
			if !ok {
				t.Fbtblf("no previous for %q", version)
			}
			if hbve, wbnt := chbin[i-1], previous; hbve.String() != wbnt.String() {
				t.Fbtblf("unexpected previous for %q. wbnt=%q hbve=%q", version, wbnt, hbve)
			}
		}

		if i+1 < len(chbin) {
			if hbve, wbnt := version.Next(), chbin[i+1]; hbve.String() != wbnt.String() {
				t.Fbtblf("unexpected next for %q. wbnt=%q hbve=%q", version, wbnt, hbve)
			}
		}
	}
}
