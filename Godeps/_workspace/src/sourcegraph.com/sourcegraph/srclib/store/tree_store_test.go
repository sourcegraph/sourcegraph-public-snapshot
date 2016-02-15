package store

import (
	"fmt"
	"reflect"
	"testing"

	"sort"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

func isIndexedStore(s interface{}) bool {
	switch s.(type) {
	case *indexedTreeStore:
		return true
	case *indexedUnitStore:
		return true
	case *fsRepoStore:
		return useIndexedStore
	case *fsMultiRepoStore:
		return useIndexedStore
	default:
		return false
	}
}

func testTreeStore(t *testing.T, newFn func() TreeStoreImporter) {
	testTreeStore_uninitialized(t, newFn())
	testTreeStore_Import_empty(t, newFn())
	testTreeStore_Import(t, newFn())
	testTreeStore_Unit(t, newFn())
	testTreeStore_Units(t, newFn())
	testTreeStore_Units_ByFile(t, newFn())
	testTreeStore_Def(t, newFn())
	testTreeStore_Defs(t, newFn())
	testTreeStore_Defs_Query(t, newFn())
	testTreeStore_Defs_Query_ByUnit(t, newFn())
	testTreeStore_Defs_ByUnits(t, newFn())
	testTreeStore_Defs_ByFiles(t, newFn())
	testTreeStore_Refs(t, newFn())
	testTreeStore_Refs_ByFiles(t, newFn())
	testTreeStore_Refs_ByDef(t, newFn())
}

func testTreeStore_uninitialized(t *testing.T, ts TreeStore) {
	units, _ := ts.Units()
	if len(units) != 0 {
		t.Errorf("%s: Units(): got units %v, want empty", ts, units)
	}

	testUnitStore_uninitialized(t, ts)
}

func testTreeStore_empty(t *testing.T, ts TreeStore) {
	units, err := ts.Units()
	if err != nil {
		t.Errorf("%s: Units(): %s", ts, err)
	}
	if len(units) != 0 {
		t.Errorf("%s: Units(): got units %v, want empty", ts, units)
	}

	testUnitStore_empty(t, ts)
}

func testTreeStore_Import_empty(t *testing.T, ts TreeStoreImporter) {
	if err := ts.Import(nil, graph.Output{}); err != nil {
		t.Errorf("%s: Import(nil, empty): %s", ts, err)
	}
	if ts, ok := ts.(TreeIndexer); ok {
		if err := ts.Index(); err != nil {
			t.Fatalf("%s: Index: %s", ts, err)
		}
	}
	testTreeStore_empty(t, ts)
}

func testTreeStore_Import(t *testing.T, ts TreeStoreImporter) {
	unit := &unit.SourceUnit{Type: "t", Name: "u", Files: []string{"f"}}
	data := graph.Output{
		Defs: []*graph.Def{
			{
				DefKey: graph.DefKey{Path: "p"},
				Name:   "n",
			},
		},
		Refs: []*graph.Ref{
			{
				DefPath: "p",
				File:    "f",
				Start:   1,
				End:     2,
			},
		},
	}
	if err := ts.Import(unit, data); err != nil {
		t.Errorf("%s: Import(%v, data): %s", ts, unit, err)
	}
	if ts, ok := ts.(TreeIndexer); ok {
		if err := ts.Index(); err != nil {
			t.Fatalf("%s: Index: %s", ts, err)
		}
	}
}

func testTreeStore_Unit(t *testing.T, ts TreeStoreImporter) {
	if err := ts.Import(&unit.SourceUnit{Type: "t", Name: "u"}, graph.Output{}); err != nil {
		t.Errorf("%s: Import(empty data): %s", ts, err)
	}
	if ts, ok := ts.(TreeIndexer); ok {
		if err := ts.Index(); err != nil {
			t.Fatalf("%s: Index: %s", ts, err)
		}
	}

	units, err := ts.Units(ByCommitIDs("c"), ByUnits(unit.ID2{Type: "t", Name: "u"}))
	if err != nil {
		t.Errorf("%s: Units: %s", ts, err)
	}
	if want := []*unit.SourceUnit{{Type: "t", Name: "u"}}; !reflect.DeepEqual(units, want) {
		t.Errorf("%s: Units: got %v, want %v", ts, units, want)
	}
}

func testTreeStore_Units(t *testing.T, ts TreeStoreImporter) {
	want := []*unit.SourceUnit{
		{Type: "t1", Name: "u1"},
		{Type: "t2", Name: "u2"},
		{Type: "t3", Name: "u3"},
	}
	for _, unit := range want {
		if err := ts.Import(unit, graph.Output{}); err != nil {
			t.Errorf("%s: Import(%v, empty data): %s", ts, unit, err)
		}
	}
	if ts, ok := ts.(TreeIndexer); ok {
		if err := ts.Index(); err != nil {
			t.Fatalf("%s: Index: %s", ts, err)
		}
	}

	{
		c_fsTreeStore_unitsOpened.set(0)
		c_unitsIndex_listUnits.set(0)
		units, err := ts.Units()
		if err != nil {
			t.Errorf("%s: Units(): %s", ts, err)
		}
		sort.Sort(unit.SourceUnits(units))
		sort.Sort(unit.SourceUnits(want))
		if !reflect.DeepEqual(units, want) {
			t.Errorf("%s: Units(): got %v, want %v", ts, units, want)
		}
		if isIndexedStore(ts) {
			if want := 1; c_unitsIndex_listUnits.get() != want {
				t.Errorf("%s: Units: listed unitsIndex %dx, want %dx", ts, c_unitsIndex_listUnits.get(), want)
			}
			if want := 0; c_fsTreeStore_unitsOpened.get() != want {
				t.Errorf("%s: Units: got %d units opened, want %d (should use unitsIndex)", ts, c_fsTreeStore_unitsOpened.get(), want)
			}
		}
	}

	{
		c_fsTreeStore_unitsOpened.set(0)
		c_unitsIndex_listUnits.set(0)

		origMaxIndividualFetches := maxIndividualFetches
		maxIndividualFetches = 1
		defer func() {
			maxIndividualFetches = origMaxIndividualFetches
		}()

		units, err := ts.Units(ByUnits(unit.ID2{Type: "t3", Name: "u3"}, unit.ID2{Type: "t1", Name: "u1"}))
		if err != nil {
			t.Errorf("%s: Units(3 and 1): %s", ts, err)
		}
		want := []*unit.SourceUnit{
			{Type: "t1", Name: "u1"},
			{Type: "t3", Name: "u3"},
		}
		sort.Sort(unit.SourceUnits(units))
		sort.Sort(unit.SourceUnits(want))
		if !reflect.DeepEqual(units, want) {
			t.Errorf("%s: Units(3 and 1): got %v, want %v", ts, units, want)
		}
		if isIndexedStore(ts) {
			if want := 1; c_unitsIndex_listUnits.get() != want {
				t.Errorf("%s: Units: listed unitsIndex %dx, want %dx", ts, c_unitsIndex_listUnits.get(), want)
			}
			if want := 0; c_fsTreeStore_unitsOpened.get() != want {
				t.Errorf("%s: Units: got %d units opened, want %d (should use unitsIndex)", ts, c_fsTreeStore_unitsOpened.get(), want)
			}
		}
	}
}

func testTreeStore_Units_ByFile(t *testing.T, ts TreeStoreImporter) {
	want := []*unit.SourceUnit{
		{Type: "t1", Name: "u1", Files: []string{"f1"}},
		{Type: "t2", Name: "u2", Files: []string{"f1", "f2"}},
		{Type: "t3", Name: "u3", Files: []string{"f1", "f3"}},
	}
	for _, unit := range want {
		if err := ts.Import(unit, graph.Output{}); err != nil {
			t.Errorf("%s: Import(%v, empty data): %s", ts, unit, err)
		}
	}
	if ts, ok := ts.(TreeIndexer); ok {
		if err := ts.Index(); err != nil {
			t.Fatalf("%s: Index: %s", ts, err)
		}
	}

	c_unitFilesIndex_getByPath.set(0)
	units, err := ts.Units(ByFiles(false, "f1"))
	if err != nil {
		t.Errorf("%s: Units(ByFiles f1): %s", ts, err)
	}
	sort.Sort(unit.SourceUnits(units))
	sort.Sort(unit.SourceUnits(want))
	if !reflect.DeepEqual(units, want) {
		t.Errorf("%s: Units(ByFiles f1): got %v, want %v", ts, units, want)
	}
	if isIndexedStore(ts) {
		if want := 1; c_unitFilesIndex_getByPath.get() != want {
			t.Errorf("%s: Units(ByFiles f1): got %d index hits, want %d", ts, c_unitFilesIndex_getByPath.get(), want)
		}
	}

	c_unitFilesIndex_getByPath.set(0)
	units2, err := ts.Units(ByFiles(false, "f2"))
	if err != nil {
		t.Errorf("%s: Units(ByFiles f2): %s", ts, err)
	}
	want2 := []*unit.SourceUnit{
		{Type: "t2", Name: "u2", Files: []string{"f1", "f2"}},
	}
	if !reflect.DeepEqual(units2, want2) {
		t.Errorf("%s: Units(ByFiles f2): got %v, want %v", ts, units2, want2)
	}
	if isIndexedStore(ts) {
		if want := 1; c_unitFilesIndex_getByPath.get() != want {
			t.Errorf("%s: Units(ByFiles f1): got %d index hits, want %d", ts, c_unitFilesIndex_getByPath.get(), want)
		}
	}
}

func testTreeStore_Def(t *testing.T, ts TreeStoreImporter) {
	u := &unit.SourceUnit{Type: "t", Name: "u"}
	data := graph.Output{
		Defs: []*graph.Def{
			{
				DefKey: graph.DefKey{Path: "p"},
				Name:   "n",
			},
		},
	}
	if err := ts.Import(u, data); err != nil {
		t.Errorf("%s: Import(%v, data): %s", ts, u, err)
	}
	if ts, ok := ts.(TreeIndexer); ok {
		if err := ts.Index(); err != nil {
			t.Fatalf("%s: Index: %s", ts, err)
		}
	}

	want := []*graph.Def{
		{
			DefKey: graph.DefKey{UnitType: "t", Unit: "u", Path: "p"},
			Name:   "n",
		},
	}

	defs, err := ts.Defs(ByDefPath("p"))
	if err != nil {
		t.Fatalf("%s: Defs: %s", ts, err)
	}
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs: got defs %v, want %v", ts, defs, want)
	}

	defs, err = ts.Defs(ByUnits(unit.ID2{Type: "t", Name: "u"}), ByDefPath("p"))
	if err != nil {
		t.Errorf("%s: Defs: %s", ts, err)
	}
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs: got defs %v, want %v", ts, defs, want)
	}

	defs, err = ts.Defs(ByUnits(unit.ID2{Type: "t2", Name: "u2"}), ByDefPath("p"))
	if err != nil {
		t.Fatalf("%s: Defs: %s", ts, err)
	}
	if len(defs) != 0 {
		t.Errorf("%s: Defs: got defs %v, want none", ts, defs)
	}
}

func testTreeStore_Defs(t *testing.T, ts TreeStoreImporter) {
	unit := &unit.SourceUnit{Type: "t", Name: "u"}
	data := graph.Output{
		Defs: []*graph.Def{
			{
				DefKey: graph.DefKey{Path: "p1"},
				Name:   "n1",
			},
			{
				DefKey: graph.DefKey{Path: "p2"},
				Name:   "n2",
			},
		},
	}
	if err := ts.Import(unit, data); err != nil {
		t.Errorf("%s: Import(%v, data): %s", ts, unit, err)
	}
	if ts, ok := ts.(TreeIndexer); ok {
		if err := ts.Index(); err != nil {
			t.Fatalf("%s: Index: %s", ts, err)
		}
	}

	want := []*graph.Def{
		{
			DefKey: graph.DefKey{UnitType: "t", Unit: "u", Path: "p1"},
			Name:   "n1",
		},
		{
			DefKey: graph.DefKey{UnitType: "t", Unit: "u", Path: "p2"},
			Name:   "n2",
		},
	}

	defs, err := ts.Defs()
	if err != nil {
		t.Errorf("%s: Defs(): %s", ts, err)
	}
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs(): got defs %v, want %v", ts, defs, want)
	}
}

func testTreeStore_Defs_Query(t *testing.T, ts TreeStoreImporter) {
	defsByUnit := map[string][]*graph.Def{
		"u1": []*graph.Def{
			{
				DefKey: graph.DefKey{Path: "p1"},
				Name:   "a",
			},
			{
				DefKey: graph.DefKey{Path: "p2"},
				Name:   "ab",
			},
		},
		"u2": []*graph.Def{
			{
				DefKey: graph.DefKey{Path: "p3"},
				Name:   "abcdef",
			},
			{
				DefKey: graph.DefKey{Path: "p4"},
				Name:   "abcxxx",
			},
			{
				DefKey: graph.DefKey{Path: "p5"},
				Name:   "x",
			},
		},
	}
	for unitName, defs := range defsByUnit {
		u := &unit.SourceUnit{Type: "t", Name: unitName}
		data := graph.Output{Defs: defs}
		if err := ts.Import(u, data); err != nil {
			t.Errorf("%s: Import(%v, data): %s", ts, u, err)
		}
		if ts, ok := ts.(TreeIndexer); ok {
			if err := ts.Index(); err != nil {
				t.Fatalf("%s: Index: %s", ts, err)
			}
		}
	}

	tests := []struct {
		q             string
		wantDefPaths  []string
		wantIndexHits int
	}{
		{
			q:             "a",
			wantDefPaths:  []string{"p1", "p2", "p3", "p4"},
			wantIndexHits: 1,
		},
		{
			q:             "ab",
			wantDefPaths:  []string{"p2", "p3", "p4"},
			wantIndexHits: 1,
		},
		{
			q:             "Abc",
			wantDefPaths:  []string{"p3", "p4"},
			wantIndexHits: 1,
		},
		{
			q:             "abc000",
			wantDefPaths:  []string{},
			wantIndexHits: 1,
		},
		{
			q:             "abcde",
			wantDefPaths:  []string{"p3"},
			wantIndexHits: 1,
		},
		{
			q:             "abcdef",
			wantDefPaths:  []string{"p3"},
			wantIndexHits: 1,
		},
		{
			q:             "abcdefg",
			wantDefPaths:  []string{},
			wantIndexHits: 1,
		},
		{
			q:             "x",
			wantDefPaths:  []string{"p5"},
			wantIndexHits: 1,
		},
		{
			q:             "z",
			wantDefPaths:  []string{},
			wantIndexHits: 1,
		},
	}
	for _, test := range tests {
		c_defQueryTreeIndex_getByQuery.set(0)
		c_defQueryIndex_getByQuery.set(0)
		defs, err := ts.Defs(ByDefQuery(test.q))
		if err != nil {
			t.Errorf("%s: Defs(ByDefQuery %q): %s", ts, test.q, err)
		}
		if got, want := defPaths(defs), test.wantDefPaths; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: Defs(ByDefQuery %q): got defs %v, want %v", ts, test.q, got, want)
		}
		if isIndexedStore(ts) {
			if want := test.wantIndexHits; c_defQueryTreeIndex_getByQuery.get() != want {
				t.Errorf("%s: Defs(ByDefQuery %q): got %d index hits, want %d", ts, test.q, c_defQueryTreeIndex_getByQuery.get(), want)
			}
			if c_defQueryIndex_getByQuery.get() != 0 {
				// This query should only hit the tree-level def query
				// index, not the def query indexes for each unit.
				t.Errorf("%s: Defs(ByDefQuery %q): got %d index hits on non-tree index, want none", ts, test.q, c_defQueryIndex_getByQuery.get())
			}
		}
	}
}

func testTreeStore_Defs_Query_ByUnit(t *testing.T, ts TreeStoreImporter) {
	defsByUnit := map[string][]*graph.Def{
		"u1": []*graph.Def{
			{
				DefKey: graph.DefKey{Path: "p1"},
				Name:   "a",
			},
			{
				DefKey: graph.DefKey{Path: "p2"},
				Name:   "ab",
			},
		},
		"u2": []*graph.Def{
			{
				DefKey: graph.DefKey{Path: "p3"},
				Name:   "abcdef",
			},
			{
				DefKey: graph.DefKey{Path: "p4"},
				Name:   "abcxxx",
			},
			{
				DefKey: graph.DefKey{Path: "p5"},
				Name:   "x",
			},
		},
	}
	for unitName, defs := range defsByUnit {
		u := &unit.SourceUnit{Type: "t", Name: unitName}
		data := graph.Output{Defs: defs}
		if err := ts.Import(u, data); err != nil {
			t.Errorf("%s: Import(%v, data): %s", ts, u, err)
		}
		if ts, ok := ts.(TreeIndexer); ok {
			if err := ts.Index(); err != nil {
				t.Fatalf("%s: Index: %s", ts, err)
			}
		}
	}

	c_defQueryTreeIndex_getByQuery.set(0)
	c_defQueryIndex_getByQuery.set(0)
	defs, err := ts.Defs(ByDefQuery("a"), ByUnits(unit.ID2{Type: "t", Name: "u1"}))
	if err != nil {
		t.Errorf("%s: Defs(ByDefQuery, ByUnit): %s", ts, err)
	}
	wantDefPaths := []string{"p1", "p2"}
	if got, want := defPaths(defs), wantDefPaths; !reflect.DeepEqual(got, want) {
		t.Errorf("%s: Defs(ByDefQuery, ByUnit): got defs %v, want %v", ts, got, want)
	}
	if isIndexedStore(ts) {
		if want := 1; c_defQueryIndex_getByQuery.get() != want {
			t.Errorf("%s: Defs(ByDefQuery, ByUnit): got %d index hits, want %d", ts, c_defQueryIndex_getByQuery.get(), want)
		}
		if c_defQueryTreeIndex_getByQuery.get() != 0 {
			// This query should only hit the unit-level def query
			// index, not the tree-wide def query indexes.
			t.Errorf("%s: Defs(ByDefQuery, ByUnit): got %d index hits on tree index, want none", ts, c_defQueryTreeIndex_getByQuery.get())
		}
	}
}

func testTreeStore_Defs_ByUnits(t *testing.T, ts TreeStoreImporter) {
	units := []*unit.SourceUnit{
		{Type: "t1", Name: "u1"},
		{Type: "t2", Name: "u2"},
		{Type: "t3", Name: "u3"},
	}
	for i, unit := range units {
		data := graph.Output{
			Defs: []*graph.Def{{DefKey: graph.DefKey{Path: fmt.Sprintf("p%d", i+1)}}},
		}
		if err := ts.Import(unit, data); err != nil {
			t.Errorf("%s: Import(%v, data): %s", ts, unit, err)
		}
	}
	if ts, ok := ts.(TreeIndexer); ok {
		if err := ts.Index(); err != nil {
			t.Fatalf("%s: Index: %s", ts, err)
		}
	}

	want := []*graph.Def{
		{DefKey: graph.DefKey{UnitType: "t1", Unit: "u1", Path: "p1"}},
		{DefKey: graph.DefKey{UnitType: "t3", Unit: "u3", Path: "p3"}},
	}

	c_fsTreeStore_unitsOpened.set(0)
	defs, err := ts.Defs(ByUnits(unit.ID2{Type: "t3", Name: "u3"}, unit.ID2{Type: "t1", Name: "u1"}))
	if err != nil {
		t.Errorf("%s: Defs(ByUnits): %s", ts, err)
	}
	sort.Sort(graph.Defs(defs))
	sort.Sort(graph.Defs(want))
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs(ByUnits): got defs %v, want %v", ts, defs, want)
	}
	if isIndexedStore(ts) {
		if c_fsTreeStore_unitsOpened.get() != 0 {
			t.Errorf("%s: Defs(ByUnits): got %d units opened, want none (should be able to use ByUnits filter to avoid needing to open any units)", ts, c_fsTreeStore_unitsOpened.get())
		}
	}
}

func testTreeStore_Defs_ByFiles(t *testing.T, ts TreeStoreImporter) {
	units := []*unit.SourceUnit{
		{Type: "t1", Name: "u1", Files: []string{"f1"}},
		{Type: "t2", Name: "u2", Files: []string{"f2"}},
	}
	for i, unit := range units {
		data := graph.Output{
			Defs: []*graph.Def{{DefKey: graph.DefKey{Path: fmt.Sprintf("p%d", i+1)}, File: fmt.Sprintf("f%d", i+1)}},
		}
		if err := ts.Import(unit, data); err != nil {
			t.Errorf("%s: Import(%v, data): %s", ts, unit, err)
		}
	}
	if ts, ok := ts.(TreeIndexer); ok {
		if err := ts.Index(); err != nil {
			t.Fatalf("%s: Index: %s", ts, err)
		}
	}

	want := []*graph.Def{
		{DefKey: graph.DefKey{UnitType: "t2", Unit: "u2", Path: "p2"}, File: "f2"},
	}

	c_unitFilesIndex_getByPath.set(0)
	defs, err := ts.Defs(ByFiles(false, "f2"))
	if err != nil {
		t.Errorf("%s: Defs(ByFiles f2): %s", ts, err)
	}
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs(ByFiles f2): got defs %v, want %v", ts, defs, want)
	}
	if isIndexedStore(ts) {
		if want := 1; c_unitFilesIndex_getByPath.get() != want {
			t.Errorf("%s: Defs(ByFiles f2): got %d index hits, want %d", ts, c_unitFilesIndex_getByPath.get(), want)
		}
	}
}

func testTreeStore_Refs(t *testing.T, ts TreeStoreImporter) {
	unit := &unit.SourceUnit{Type: "t", Name: "u", Files: []string{"f1", "f2"}}
	data := graph.Output{
		Refs: []*graph.Ref{
			{
				DefPath: "p1",
				File:    "f1",
				Start:   1,
				End:     2,
			},
			{
				DefPath: "p2",
				File:    "f2",
				Start:   2,
				End:     3,
			},
		},
	}
	if err := ts.Import(unit, data); err != nil {
		t.Errorf("%s: Import(%v, data): %s", ts, unit, err)
	}
	if ts, ok := ts.(TreeIndexer); ok {
		if err := ts.Index(); err != nil {
			t.Fatalf("%s: Index: %s", ts, err)
		}
	}

	want := []*graph.Ref{
		{
			DefUnitType: "t",
			DefUnit:     "u",
			DefPath:     "p1",
			File:        "f1",
			Start:       1,
			End:         2,
			UnitType:    "t",
			Unit:        "u",
		},
		{
			DefUnitType: "t",
			DefUnit:     "u",
			DefPath:     "p2",
			File:        "f2",
			Start:       2,
			End:         3,
			UnitType:    "t",
			Unit:        "u",
		},
	}

	refs, err := ts.Refs()
	if err != nil {
		t.Errorf("%s: Refs(): %s", ts, err)
	}
	if !reflect.DeepEqual(refs, want) {
		t.Errorf("%s: Refs(): got refs %v, want %v", ts, refs, want)
	}
}

func testTreeStore_Refs_ByFiles(t *testing.T, ts TreeStoreImporter) {
	refsByUnitByFile := map[string]map[string][]*graph.Ref{
		"u1": {
			"f1": {
				{DefPath: "p1", Start: 0, End: 5},
			},
			"f2": {
				{DefPath: "p1", Start: 0, End: 5},
				{DefPath: "p2", Start: 5, End: 10},
			},
		},
		"u2": {
			"f1": {
				{DefPath: "p1", Start: 5, End: 10},
			},
		},
	}
	refsByFile := map[string][]*graph.Ref{}
	for unitName, refsByFile0 := range refsByUnitByFile {
		u := &unit.SourceUnit{Type: "t", Name: unitName}
		var data graph.Output
		for file, refs := range refsByFile0 {
			u.Files = append(u.Files, file)
			for _, ref := range refs {
				ref.File = file
			}
			data.Refs = append(data.Refs, refs...)
			refsByFile[file] = append(refsByFile[file], refs...)
		}
		if err := ts.Import(u, data); err != nil {
			t.Errorf("%s: Import(%v, data): %s", ts, u, err)
		}
	}
	if ts, ok := ts.(TreeIndexer); ok {
		if err := ts.Index(); err != nil {
			t.Fatalf("%s: Index: %s", ts, err)
		}
	}

	for file, wantRefs := range refsByFile {
		c_unitStores_Refs_last_numUnitsQueried.set(0)
		c_refFileIndex_getByFile.set(0)
		refs, err := ts.Refs(ByFiles(false, file))
		if err != nil {
			t.Fatalf("%s: Refs(ByFiles %s): %s", ts, file, err)
		}

		distinctRefUnits := map[string]struct{}{}
		for _, ref := range refs {
			distinctRefUnits[ref.Unit] = struct{}{}
		}

		// for test equality
		sort.Sort(refsByFileStartEnd(refs))
		sort.Sort(refsByFileStartEnd(wantRefs))
		cleanForImport(&graph.Output{Refs: refs}, "", "t", "u1")
		cleanForImport(&graph.Output{Refs: refs}, "", "t", "u2")

		if want := wantRefs; !reflect.DeepEqual(refs, want) {
			t.Errorf("%s: Refs(ByFiles %s): got refs %v, want %v", ts, file, refs, want)
		}
		if isIndexedStore(ts) {
			if want := len(distinctRefUnits); c_refFileIndex_getByFile.get() != want {
				t.Errorf("%s: Refs(ByFiles %s): got %d index hits, want %d", ts, file, c_refFileIndex_getByFile.get(), want)
			}
			if want := len(distinctRefUnits); c_unitStores_Refs_last_numUnitsQueried.get() != want {
				t.Errorf("%s: Refs(ByFiles %s): got %d units queried, want %d", ts, file, c_unitStores_Refs_last_numUnitsQueried.get(), want)
			}
		}
	}
}

func testTreeStore_Refs_ByDef(t *testing.T, ts TreeStoreImporter) {
	refsByUnit := map[string][]*graph.Ref{
		"u1": {
			{DefPath: "p1", Start: 0, End: 1},
			{DefPath: "p1", Unit: "u2", Start: 1, End: 2},
			{DefPath: "p2", Start: 0, End: 2},
			{DefPath: "p2", Unit: "u2", Start: 2, End: 4},
			{DefPath: "p2", Unit: "u2", Start: 4, End: 6},
		},
		"u2": {
			{DefPath: "p1", Start: 0, End: 1},
			{DefPath: "p1", Unit: "u1", Start: 1, End: 2},
			{DefPath: "p1", Unit: "u1", Start: 2, End: 3},
		},
	}
	refsByDefUnitByDefPath := map[string]map[string][]*graph.Ref{}
	for unitName, refs := range refsByUnit {
		u := &unit.SourceUnit{Type: "t", Name: unitName}
		data := graph.Output{Refs: refs}
		for _, ref := range data.Refs {
			defUnit := ref.DefUnit
			if defUnit == "" {
				defUnit = unitName
			}
			if _, present := refsByDefUnitByDefPath[defUnit]; !present {
				refsByDefUnitByDefPath[defUnit] = map[string][]*graph.Ref{}
			}
			refsByDefUnitByDefPath[defUnit][ref.DefPath] = append(refsByDefUnitByDefPath[defUnit][ref.DefPath], ref)
		}
		if err := ts.Import(u, data); err != nil {
			t.Errorf("%s: Import(%v, data): %s", ts, u, err)
		}
	}
	if ts, ok := ts.(TreeIndexer); ok {
		if err := ts.Index(); err != nil {
			t.Fatalf("%s: Index: %s", ts, err)
		}
	}

	for defUnit, refsByDefPath := range refsByDefUnitByDefPath {
		for defPath, wantRefs := range refsByDefPath {
			defLabel := defUnit + ":" + defPath

			c_unitStores_Refs_last_numUnitsQueried.set(0)
			c_defRefsIndex_getByDef.set(0)
			c_defRefUnitsIndex_getByDef.set(0)
			refs, err := ts.Refs(ByRefDef(graph.RefDefKey{DefUnitType: "t", DefUnit: defUnit, DefPath: defPath}))
			if err != nil {
				t.Fatalf("%s: Refs(ByDef %s): %s", ts, defLabel, err)
			}

			distinctRefUnits := map[string]struct{}{}
			for _, ref := range refs {
				distinctRefUnits[ref.Unit] = struct{}{}
			}

			// for test equality
			sort.Sort(refsByFileStartEnd(refs))
			sort.Sort(refsByFileStartEnd(wantRefs))
			cleanForImport(&graph.Output{Refs: refs}, "", "t", "u1")
			cleanForImport(&graph.Output{Refs: refs}, "", "t", "u2")

			if want := wantRefs; !reflect.DeepEqual(refs, want) {
				t.Errorf("%s: Refs(ByDef %s): got refs %v, want %v", ts, defLabel, refs, want)
			}
			if isIndexedStore(ts) {
				if want := len(distinctRefUnits); c_defRefsIndex_getByDef.get() != want {
					t.Errorf("%s: Refs(ByDef %s): got %d c_defRefsIndex_getByDef index hits, want %d", ts, defLabel, c_defRefsIndex_getByDef.get(), want)
				}
				if want := 1; c_defRefUnitsIndex_getByDef.get() != want {
					t.Errorf("%s: Refs(ByDef %s): got %d c_defRefUnitsIndex_getByDef index hits, want %d", ts, defLabel, c_defRefUnitsIndex_getByDef.get(), want)
				}
				if want := len(distinctRefUnits); c_unitStores_Refs_last_numUnitsQueried.get() != want {
					t.Errorf("%s: Refs(ByDef %s): got %d units queried, want %d", ts, defLabel, c_unitStores_Refs_last_numUnitsQueried.get(), want)
				}
			}
		}
	}
}
