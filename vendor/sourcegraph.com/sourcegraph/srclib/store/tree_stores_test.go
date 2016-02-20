package store

import (
	"reflect"
	"testing"

	"sort"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// mockNeverCalledTreeStore calls t.Error if any of its methods are
// called.
func mockNeverCalledTreeStore(t *testing.T) MockTreeStore {
	return MockTreeStore{
		Units_: func(f ...UnitFilter) ([]*unit.SourceUnit, error) {
			t.Fatalf("(TreeStore).Units called, but wanted it not to be called (arg f was %v)", f)
			return nil, nil
		},
		MockUnitStore: mockNeverCalledUnitStore(t),
	}
}

type emptyTreeStore struct{ emptyUnitStore }

func (m emptyTreeStore) Units(f ...UnitFilter) ([]*unit.SourceUnit, error) {
	return []*unit.SourceUnit{}, nil
}

type mapTreeStoreOpener map[string]TreeStore

func (m mapTreeStoreOpener) openTreeStore(commitID string) TreeStore {
	return m[commitID]
}
func (m mapTreeStoreOpener) openAllTreeStores() (map[string]TreeStore, error) { return m, nil }

type recordingTreeStoreOpener struct {
	opened    map[string]int // how many times openTreeStore was called for each tree
	openedAll int            // how many times openAllTreeStores was called
	treeStoreOpener
}

func (m *recordingTreeStoreOpener) openTreeStore(commitID string) TreeStore {
	if m.opened == nil {
		m.opened = map[string]int{}
	}
	m.opened[commitID]++
	return m.treeStoreOpener.openTreeStore(commitID)
}
func (m *recordingTreeStoreOpener) openAllTreeStores() (map[string]TreeStore, error) {
	m.openedAll++
	return m.treeStoreOpener.openAllTreeStores()
}
func (m *recordingTreeStoreOpener) reset() { m.opened = map[string]int{}; m.openedAll = 0 }

func TestTreeStores_filterByCommit(t *testing.T) {
	// Test that filters by a specific commit cause tree stores for
	// other commits to not be called.

	o := &recordingTreeStoreOpener{treeStoreOpener: mapTreeStoreOpener{
		"c":  emptyTreeStore{},
		"c2": mockNeverCalledTreeStore(t),
	}}
	tss := treeStores{opener: o}

	if defs, err := tss.Defs(ByCommitIDs("c"), ByUnits(unit.ID2{Type: "t", Name: "u"}), ByDefPath("p")); err != nil {
		t.Error(err)
	} else if len(defs) > 0 {
		t.Errorf("got defs %v, want none", defs)
	}
	if want := map[string]int{"c": 1}; !reflect.DeepEqual(o.opened, want) {
		t.Errorf("got opened %v, want %v", o.opened, want)
	}
	o.reset()

	if defs, err := tss.Defs(ByCommitIDs("c")); err != nil {
		t.Error(err)
	} else if len(defs) > 0 {
		t.Errorf("got defs %v, want none", defs)
	}

	if refs, err := tss.Refs(ByCommitIDs("c")); err != nil {
		t.Error(err)
	} else if len(refs) > 0 {
		t.Errorf("got refs %v, want none", refs)
	}

	if units, err := tss.Units(ByCommitIDs("c"), ByUnits(unit.ID2{Type: "t", Name: "u"})); err != nil {
		t.Error(err)
	} else if len(units) > 0 {
		t.Errorf("got units %v, want none", units)
	}

	if units, err := tss.Units(ByCommitIDs("c")); err != nil {
		t.Error(err)
	} else if len(units) > 0 {
		t.Errorf("got units %v, want none", units)
	}
}

func TestScopeTrees(t *testing.T) {
	tests := []struct {
		filters []interface{}
		want    []string
	}{
		{
			filters: nil,
			want:    nil,
		},
		{
			filters: []interface{}{ByCommitIDs("c")},
			want:    []string{"c"},
		},
		{
			filters: []interface{}{nil, ByCommitIDs("c")},
			want:    []string{"c"},
		},
		{
			filters: []interface{}{ByCommitIDs("c"), nil},
			want:    []string{"c"},
		},
		{
			filters: []interface{}{nil, ByCommitIDs("c"), nil},
			want:    []string{"c"},
		},
		{
			filters: []interface{}{ByCommitIDs("c"), ByCommitIDs("c")},
			want:    []string{"c"},
		},
		{
			filters: []interface{}{ByCommitIDs("c1"), ByCommitIDs("c2")},
			want:    []string{},
		},
		{
			filters: []interface{}{ByCommitIDs("c2", "c1"), ByCommitIDs("c1", "c2")},
			want:    []string{"c1", "c2"},
		},
		{
			filters: []interface{}{ByCommitIDs("c1", "c2"), ByCommitIDs("c2")},
			want:    []string{"c2"},
		},
		{
			filters: []interface{}{ByCommitIDs("c1"), ByCommitIDs("c2"), ByCommitIDs("c1")},
			want:    []string{},
		},
		{
			filters: []interface{}{ByUnitKey(unit.Key{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u"})},
			want:    []string{"c"},
		},
		{
			filters: []interface{}{
				ByUnitKey(unit.Key{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u"}),
				ByUnitKey(unit.Key{Repo: "r2", CommitID: "c", UnitType: "t2", Unit: "u2"}),
			},
			want: []string{"c"},
		},
		{
			filters: []interface{}{
				ByUnitKey(unit.Key{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u"}),
				ByUnitKey(unit.Key{Repo: "r", CommitID: "c2", UnitType: "t", Unit: "u"}),
			},
			want: []string{},
		},
		{
			filters: []interface{}{ByDefKey(graph.DefKey{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"})},
			want:    []string{"c"},
		},
		{
			filters: []interface{}{
				ByDefKey(graph.DefKey{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"}),
				ByDefKey(graph.DefKey{Repo: "r2", CommitID: "c", UnitType: "t2", Unit: "u2", Path: "p2"}),
			},
			want: []string{"c"},
		},
		{
			filters: []interface{}{
				ByDefKey(graph.DefKey{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"}),
				ByDefKey(graph.DefKey{Repo: "r", CommitID: "c2", UnitType: "t", Unit: "u", Path: "p"}),
			},
			want: []string{},
		},
		{
			filters: []interface{}{VersionFilterFunc(func(*Version) bool { return false })},
			want:    nil,
		},
		{
			filters: []interface{}{ByUnits(unit.ID2{Type: "t", Name: "u"})},
			want:    nil,
		},
	}
	for _, test := range tests {
		trees, err := scopeTrees(test.filters)
		if err != nil {
			t.Errorf("%+v: %v", test.filters, err)
			continue
		}
		sort.Strings(trees)
		sort.Strings(test.want)
		if !reflect.DeepEqual(trees, test.want) {
			t.Errorf("%+v: got trees %v, want %v", test.filters, trees, test.want)
		}
	}
}
