package store

import (
	"fmt"
	"reflect"
	"testing"

	"sort"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// mockNeverCalledUnitStore calls t.Error if any of its methods are
// called.
func mockNeverCalledUnitStore(t *testing.T) MockUnitStore {
	return MockUnitStore{
		Defs_: func(f ...DefFilter) ([]*graph.Def, error) {
			t.Fatalf("(UnitStore).Defs called, but wanted it not to be called (arg f was %v)", f)
			return nil, nil
		},
		Refs_: func(f ...RefFilter) ([]*graph.Ref, error) {
			t.Fatalf("(UnitStore).Refs called, but wanted it not to be called (arg f was %v)", f)
			return nil, nil
		},
	}
}

type emptyUnitStore struct{}

func (m emptyUnitStore) Defs(f ...DefFilter) ([]*graph.Def, error) {
	return []*graph.Def{}, nil
}

func (m emptyUnitStore) Refs(f ...RefFilter) ([]*graph.Ref, error) {
	return []*graph.Ref{}, nil
}

type mapUnitStoreOpener map[unit.ID2]UnitStore

func (m mapUnitStoreOpener) openUnitStore(u unit.ID2) UnitStore {
	return m[u]
}
func (m mapUnitStoreOpener) openAllUnitStores() (map[unit.ID2]UnitStore, error) { return m, nil }

type recordingUnitStoreOpener struct {
	opened    map[unit.ID2]int // how many times openUnitStore was called for each unit
	openedAll int              // how many times openAllUnitStores was called
	unitStoreOpener
}

func (m *recordingUnitStoreOpener) openUnitStore(u unit.ID2) UnitStore {
	if m.opened == nil {
		m.opened = map[unit.ID2]int{}
	}
	m.opened[u]++
	return m.unitStoreOpener.openUnitStore(u)
}
func (m *recordingUnitStoreOpener) openAllUnitStores() (map[unit.ID2]UnitStore, error) {
	m.openedAll++
	return m.unitStoreOpener.openAllUnitStores()
}
func (m *recordingUnitStoreOpener) reset() { m.opened = map[unit.ID2]int{}; m.openedAll = 0 }

func TestUnitStores_filterByUnits(t *testing.T) {
	// Test that filters by source unit cause unit stores for other
	// source units to not be called.

	o := &recordingUnitStoreOpener{unitStoreOpener: mapUnitStoreOpener{
		unit.ID2{Type: "t", Name: "u"}:  emptyUnitStore{},
		unit.ID2{Type: "t", Name: "u2"}: mockNeverCalledUnitStore(t),
		unit.ID2{Type: "t2", Name: "u"}: mockNeverCalledUnitStore(t),
	}}
	uss := unitStores{opener: o}

	if defs, err := uss.Defs(ByUnits(unit.ID2{Type: "t", Name: "u"}), ByDefPath("p")); err != nil {
		t.Error(err)
	} else if len(defs) > 0 {
		t.Errorf("got defs %v, want none", defs)
	}
	if want := map[unit.ID2]int{unit.ID2{Type: "t", Name: "u"}: 1}; !reflect.DeepEqual(o.opened, want) {
		t.Errorf("got opened %v, want %v", o.opened, want)
	}
	o.reset()

	if defs, err := uss.Defs(ByUnits(unit.ID2{Type: "t", Name: "u"})); err != nil {
		t.Error(err)
	} else if len(defs) > 0 {
		t.Errorf("got defs %v, want none", defs)
	}
	if want := map[unit.ID2]int{unit.ID2{Type: "t", Name: "u"}: 1}; !reflect.DeepEqual(o.opened, want) {
		t.Errorf("got opened %v, want %v", o.opened, want)
	}
	o.reset()

	if refs, err := uss.Refs(ByUnits(unit.ID2{Type: "t", Name: "u"})); err != nil {
		t.Error(err)
	} else if len(refs) > 0 {
		t.Errorf("got refs %v, want none", refs)
	}
	if want := map[unit.ID2]int{unit.ID2{Type: "t", Name: "u"}: 1}; !reflect.DeepEqual(o.opened, want) {
		t.Errorf("got opened %v, want %v", o.opened, want)
	}
	o.reset()
}

func TestScopeUnits(t *testing.T) {
	tests := []struct {
		filters []interface{}
		want    []unit.ID2
	}{
		{
			filters: nil,
			want:    nil,
		},
		{
			filters: []interface{}{ByUnits(unit.ID2{Type: "t", Name: "u"})},
			want:    []unit.ID2{{Type: "t", Name: "u"}},
		},
		{
			filters: []interface{}{nil, ByUnits(unit.ID2{Type: "t", Name: "u"})},
			want:    []unit.ID2{{Type: "t", Name: "u"}},
		},
		{
			filters: []interface{}{ByUnits(unit.ID2{Type: "t", Name: "u"}), nil},
			want:    []unit.ID2{{Type: "t", Name: "u"}},
		},
		{
			filters: []interface{}{ByUnits(), nil},
			want:    []unit.ID2{},
		},
		{
			filters: []interface{}{ByUnits(unit.ID2{Type: "t", Name: "u"}), ByUnits(), nil},
			want:    []unit.ID2{},
		},
		{
			filters: []interface{}{ByUnits(unit.ID2{Type: "t", Name: "u"}, unit.ID2{Type: "t2", Name: "u2"}), nil},
			want:    []unit.ID2{{Type: "t", Name: "u"}, {"t2", "u2"}},
		},
		{
			filters: []interface{}{nil, ByUnits(unit.ID2{Type: "t", Name: "u"}), nil},
			want:    []unit.ID2{{Type: "t", Name: "u"}},
		},
		{
			filters: []interface{}{ByUnits(unit.ID2{Type: "t", Name: "u"}), ByUnits(unit.ID2{Type: "t", Name: "u"})},
			want:    []unit.ID2{{Type: "t", Name: "u"}},
		},
		{
			filters: []interface{}{ByUnits(unit.ID2{Type: "t1", Name: "u1"}), ByUnits(unit.ID2{Type: "t2", Name: "u2"})},
			want:    []unit.ID2{},
		},
		{
			filters: []interface{}{ByUnits(unit.ID2{Type: "t1", Name: "u1"}), ByUnits(unit.ID2{Type: "t2", Name: "u2"}), ByUnits(unit.ID2{Type: "t1", Name: "u1"})},
			want:    []unit.ID2{},
		},
		{
			filters: []interface{}{ByUnits(unit.ID2{Type: "t", Name: "u1"}), ByUnits(unit.ID2{Type: "t", Name: "u2"})},
			want:    []unit.ID2{},
		},
		{
			filters: []interface{}{ByUnits(unit.ID2{Type: "t", Name: "u1"}), ByUnits(unit.ID2{Type: "t2", Name: "u"})},
			want:    []unit.ID2{},
		},
		{
			filters: []interface{}{ByUnitKey(unit.Key{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u"})},
			want:    []unit.ID2{{Type: "t", Name: "u"}},
		},
		{
			filters: []interface{}{
				ByUnitKey(unit.Key{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u"}),
				ByUnitKey(unit.Key{Repo: "r2", CommitID: "c2", UnitType: "t", Unit: "u"}),
			},
			want: []unit.ID2{{Type: "t", Name: "u"}},
		},
		{
			filters: []interface{}{
				ByUnitKey(unit.Key{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u"}),
				ByUnitKey(unit.Key{Repo: "r", CommitID: "c", UnitType: "t2", Unit: "u2"}),
			},
			want: []unit.ID2{},
		},
		{
			filters: []interface{}{ByDefKey(graph.DefKey{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"})},
			want:    []unit.ID2{{Type: "t", Name: "u"}},
		},
		{
			filters: []interface{}{
				ByDefKey(graph.DefKey{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"}),
				ByDefKey(graph.DefKey{Repo: "r2", CommitID: "c2", UnitType: "t", Unit: "u", Path: "p2"}),
			},
			want: []unit.ID2{{Type: "t", Name: "u"}},
		},
		{
			filters: []interface{}{
				ByDefKey(graph.DefKey{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"}),
				ByDefKey(graph.DefKey{Repo: "r", CommitID: "c", UnitType: "t2", Unit: "u2", Path: "p"}),
			},
			want: []unit.ID2{},
		},
		{
			filters: []interface{}{UnitFilterFunc(func(*unit.SourceUnit) bool { return false })},
			want:    nil,
		},
		{
			filters: []interface{}{ByRepos("r")},
			want:    nil,
		},
	}
	for _, test := range tests {
		units, err := scopeUnits(test.filters)
		if err != nil {
			t.Errorf("%+v: %v", test.filters, err)
			continue
		}
		sort.Sort(unitID2s(units))
		sort.Sort(unitID2s(test.want))
		if !reflect.DeepEqual(units, test.want) {
			t.Errorf("%+v: got units %v, want %v", test.filters, units, test.want)
		}
	}
}

func TestFiltersForUnit(t *testing.T) {
	tests := []struct {
		filters    interface{}
		wantByUnit map[unit.ID2]interface{}
	}{
		{
			filters:    nil,
			wantByUnit: nil,
		},
		{
			filters: []DefFilter{ByUnits(unit.ID2{Type: "t", Name: "u"})},
			wantByUnit: map[unit.ID2]interface{}{
				unit.ID2{Type: "t", Name: "u"}: []DefFilter{},
			},
		},
		{
			filters: []DefFilter{ByRepos("r"), ByUnits(unit.ID2{Type: "t", Name: "u"})},
			wantByUnit: map[unit.ID2]interface{}{
				unit.ID2{Type: "t", Name: "u"}: []DefFilter{ByRepos("r")},
			},
		},
	}
	for _, test := range tests {
		for unit, want := range test.wantByUnit {
			pre := fmt.Sprintf("%+v", test.filters)
			unitFilters := filtersForUnit(unit, test.filters)
			if !reflect.DeepEqual(unitFilters, want) {
				t.Errorf("%+v: unit %q: got unit filters %v, want %v", test.filters, unit, unitFilters, want)
			}
			if post := fmt.Sprintf("%+v", test.filters); pre != post {
				t.Errorf("%+v: filters modified: post filtersToUnit, filters == %v", test.filters, post)
			}
		}
	}
}
