package store

import (
	"reflect"
	"testing"

	"github.com/alecthomas/binary"

	"sourcegraph.com/sourcegraph/srclib/unit"
)

type mockDefIndex struct {
	Covers_ func(interface{}) int
	Defs_   func(...DefFilter) (byteOffsets, error)
}

func (m mockDefIndex) Covers(fs interface{}) int                 { return m.Covers_(fs) }
func (m mockDefIndex) Defs(fs ...DefFilter) (byteOffsets, error) { return m.Defs_(fs...) }
func (m mockDefIndex) Ready() bool                               { return true }

type mockUnitIndex struct {
	Covers_ func(interface{}) int
	Units_  func(...UnitFilter) ([]unit.ID2, error)
}

func (m mockUnitIndex) Covers(fs interface{}) int                  { return m.Covers_(fs) }
func (m mockUnitIndex) Units(fs ...UnitFilter) ([]unit.ID2, error) { return m.Units_(fs...) }
func (m mockUnitIndex) Ready() bool                                { return true }

func TestBestCoverageIndex(t *testing.T) {
	tests := map[string]struct {
		indexes      map[string]Index
		test         func(interface{}) bool
		wantBestName string
	}{
		"empty indexes": {
			indexes:      map[string]Index{},
			wantBestName: "",
		},
		"coverage 0": {
			indexes:      map[string]Index{"a": mockDefIndex{Covers_: func(interface{}) int { return 0 }}},
			wantBestName: "",
		},
		"coverage 1": {
			indexes: map[string]Index{
				"a": mockDefIndex{
					Covers_: func(interface{}) int { return 1 },
				},
			},
			wantBestName: "a",
		},
		"choose index with highest coverage": {
			indexes: map[string]Index{
				"a": mockDefIndex{
					Covers_: func(interface{}) int { return 1 },
				},
				"b": mockDefIndex{
					Covers_: func(interface{}) int { return 2 },
				},
			},
			wantBestName: "b",
		},
		"choose index of specified type": {
			indexes: map[string]Index{
				"a": mockDefIndex{
					Covers_: func(interface{}) int { return 2 },
				},
				"b": mockUnitIndex{
					Covers_: func(interface{}) int { return 1 },
				},
			},
			test:         func(x interface{}) bool { _, ok := x.(unitIndex); return ok },
			wantBestName: "b",
		},
	}
	for label, test := range tests {
		// Filters don't matter for this test since we just call
		// (defIndex).Covers.
		name, _ := bestCoverageIndex(test.indexes, nil, test.test)
		if name != test.wantBestName {
			t.Errorf("%s: got best index %q, want %q", label, name, test.wantBestName)
		}
	}
}

func TestUnitOffsets_BinaryEncoding(t *testing.T) {
	uofs := unitOffsets{Unit: 123, byteOffsets: []int64{1, 100, 1000, 200, 2}}
	b, err := binary.Marshal(&uofs)
	if err != nil {
		t.Fatal(err)
	}

	var uofs2 unitOffsets
	if err := binary.Unmarshal(b, &uofs2); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(uofs2, uofs) {
		t.Errorf("got %v, want %v", uofs2, uofs)
	}
}
