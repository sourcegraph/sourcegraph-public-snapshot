package store

import (
	"fmt"
	"reflect"
	"testing"

	"sort"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

func testRepoStore(t *testing.T, newFn func() RepoStoreImporter) {
	testRepoStore_uninitialized(t, newFn())
	testRepoStore_Import_empty(t, newFn())
	testRepoStore_Import(t, newFn())
	testRepoStore_Versions(t, newFn())
	testRepoStore_Units(t, newFn())
	testRepoStore_Defs(t, newFn())
	testRepoStore_Defs_ByCommitIDs(t, newFn())
	testRepoStore_Defs_ByCommitIDs_ByFile(t, newFn())
	testRepoStore_Refs(t, newFn())
}

func testRepoStore_uninitialized(t *testing.T, rs RepoStore) {
	versions, _ := rs.Versions()
	if len(versions) != 0 {
		t.Errorf("%s: Versions(): got versions %v, want empty", rs, versions)
	}

	testTreeStore_uninitialized(t, rs)
}

func testRepoStore_Import_empty(t *testing.T, rs RepoStoreImporter) {
	if err := rs.Import("c", nil, graph.Output{}); err != nil {
		t.Errorf("%s: Import(c, nil, empty): %s", rs, err)
	}
	if rs, ok := rs.(RepoIndexer); ok {
		if err := rs.Index("c"); err != nil {
			t.Fatalf("%s: Index: %s", rs, err)
		}
	}
	testTreeStore_empty(t, rs)
}

func testRepoStore_Import(t *testing.T, rs RepoStoreImporter) {
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
	if err := rs.Import("c", unit, data); err != nil {
		t.Errorf("%s: Import(c, %v, data): %s", rs, unit, err)
	}
	if rs, ok := rs.(RepoIndexer); ok {
		if err := rs.Index("c"); err != nil {
			t.Fatalf("%s: Index: %s", rs, err)
		}
	}
}

func testRepoStore_Versions(t *testing.T, rs RepoStoreImporter) {
	for _, version := range []string{"c1", "c2"} {
		unit := &unit.SourceUnit{Type: "t1", Name: "u1"}
		if err := rs.Import(version, unit, graph.Output{}); err != nil {
			t.Errorf("%s: Import(%s, %v, empty data): %s", rs, version, unit, err)
		}
		if rs, ok := rs.(RepoIndexer); ok {
			if err := rs.Index(version); err != nil {
				t.Fatalf("%s: Index: %s", rs, err)
			}
		}
	}

	want := []*Version{{CommitID: "c1"}, {CommitID: "c2"}}

	versions, err := rs.Versions()
	if err != nil {
		t.Errorf("%s: Versions(): %s", rs, err)
	}
	if !reflect.DeepEqual(versions, want) {
		t.Errorf("%s: Versions(): got %v, want %v", rs, versions, want)
	}

	versions, err = rs.Versions(ByCommitIDs("c2"))
	if err != nil {
		t.Errorf("%s: Versions(c2): %s", rs, err)
	}
	if want := []*Version{{CommitID: "c2"}}; !reflect.DeepEqual(versions, want) {
		t.Errorf("%s: Versions(c2): got %v, want %v", rs, versions, want)
	}
}

func testRepoStore_Units(t *testing.T, rs RepoStoreImporter) {
	units := []*unit.SourceUnit{
		{Type: "t1", Name: "u1"},
		{Type: "t2", Name: "u2"},
	}
	for _, unit := range units {
		if err := rs.Import("c", unit, graph.Output{}); err != nil {
			t.Errorf("%s: Import(c, %v, empty data): %s", rs, unit, err)
		}
	}
	if rs, ok := rs.(RepoIndexer); ok {
		if err := rs.Index("c"); err != nil {
			t.Fatalf("%s: Index: %s", rs, err)
		}
	}

	want := []*unit.SourceUnit{
		{CommitID: "c", Type: "t1", Name: "u1"},
		{CommitID: "c", Type: "t2", Name: "u2"},
	}

	units, err := rs.Units()
	if err != nil {
		t.Errorf("%s: Units(): %s", rs, err)
	}
	if !reflect.DeepEqual(units, want) {
		t.Errorf("%s: Units(): got %v, want %v", rs, units, want)
	}

	units, err = rs.Units(ByCommitIDs("c"), ByUnits(unit.ID2{Type: "t2", Name: "u2"}))
	if err != nil {
		t.Errorf("%s: Units: %s", rs, err)
	}
	if want := []*unit.SourceUnit{{CommitID: "c", Type: "t2", Name: "u2"}}; !reflect.DeepEqual(units, want) {
		t.Errorf("%s: Units: got %v, want %v", rs, units, want)
	}

	if units, err = rs.Units(ByCommitIDs("c"), ByUnits(unit.ID2{Type: "t3", Name: "u3"})); err != nil {
		t.Errorf("%s: Units: %s", rs, err)
	} else if len(units) != 0 {
		t.Errorf("%s: Units: got %v, want none", rs, units)
	}
}

func testRepoStore_Defs(t *testing.T, rs RepoStoreImporter) {
	u := &unit.SourceUnit{Type: "t", Name: "u"}
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
	if err := rs.Import("c", u, data); err != nil {
		t.Errorf("%s: Import(c, %v, data): %s", rs, u, err)
	}
	if rs, ok := rs.(RepoIndexer); ok {
		if err := rs.Index("c"); err != nil {
			t.Fatalf("%s: Index: %s", rs, err)
		}
	}

	want := []*graph.Def{
		{
			DefKey: graph.DefKey{CommitID: "c", UnitType: "t", Unit: "u", Path: "p1"},
			Name:   "n1",
		},
		{
			DefKey: graph.DefKey{CommitID: "c", UnitType: "t", Unit: "u", Path: "p2"},
			Name:   "n2",
		},
	}

	defs, err := rs.Defs()
	if err != nil {
		t.Errorf("%s: Defs(): %s", rs, err)
	}
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs(): got defs %v, want %v", rs, defs, want)
	}

	want = []*graph.Def{
		{
			DefKey: graph.DefKey{CommitID: "c", UnitType: "t", Unit: "u", Path: "p1"},
			Name:   "n1",
		},
	}
	defs, err = rs.Defs(ByCommitIDs("c"), ByUnits(unit.ID2{Type: "t", Name: "u"}), ByDefPath("p1"))
	if err != nil {
		t.Errorf("%s: Defs: %s", rs, err)
	}
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs: got defs %v, want %v", rs, defs, want)
	}
}

func testRepoStore_Defs_ByCommitIDs(t *testing.T, rs RepoStoreImporter) {
	const numCommits = 3
	for c := 1; c <= numCommits; c++ {
		unit := &unit.SourceUnit{Type: "t", Name: "u"}
		data := graph.Output{Defs: []*graph.Def{{DefKey: graph.DefKey{Path: "p"}}}}
		commitID := fmt.Sprintf("c%d", c)
		if err := rs.Import(commitID, unit, data); err != nil {
			t.Errorf("%s: Import(%s, %v, data): %s", rs, commitID, unit, err)
		}
		if rs, ok := rs.(RepoIndexer); ok {
			if err := rs.Index(commitID); err != nil {
				t.Fatalf("%s: Index: %s", rs, err)
			}
		}
	}

	want := []*graph.Def{
		{DefKey: graph.DefKey{CommitID: "c1", UnitType: "t", Unit: "u", Path: "p"}},
		{DefKey: graph.DefKey{CommitID: "c3", UnitType: "t", Unit: "u", Path: "p"}},
	}

	defs, err := rs.Defs(ByCommitIDs("c1", "c3"))
	if err != nil {
		t.Fatalf("%s: Defs: %s", rs, err)
	}
	sort.Sort(graph.Defs(defs))
	sort.Sort(graph.Defs(want))
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs: got defs %v, want %v", rs, defs, want)
	}
}

func testRepoStore_Defs_ByCommitIDs_ByFile(t *testing.T, rs RepoStoreImporter) {
	const numCommits = 2
	for c := 1; c <= numCommits; c++ {
		unit := &unit.SourceUnit{Type: "t", Name: "u", Files: []string{"f1", "f2"}}
		data := graph.Output{
			Defs: []*graph.Def{
				{DefKey: graph.DefKey{Path: "p1"}, File: "f1"},
				{DefKey: graph.DefKey{Path: "p2"}, File: "f2"},
			},
		}
		commitID := fmt.Sprintf("c%d", c)
		if err := rs.Import(commitID, unit, data); err != nil {
			t.Errorf("%s: Import(%s, %v, data): %s", rs, commitID, unit, err)
		}
		if rs, ok := rs.(RepoIndexer); ok {
			if err := rs.Index(commitID); err != nil {
				t.Fatalf("%s: Index: %s", rs, err)
			}
		}
	}

	want := []*graph.Def{
		{DefKey: graph.DefKey{CommitID: "c2", UnitType: "t", Unit: "u", Path: "p1"}, File: "f1"},
	}

	c_unitFilesIndex_getByPath.set(0)
	defs, err := rs.Defs(ByCommitIDs("c2"), ByFiles(false, "f1"))
	if err != nil {
		t.Fatalf("%s: Defs: %s", rs, err)
	}
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs: got defs %v, want %v", rs, defs, want)
	}
	if isIndexedStore(rs) {
		if want := 1; c_unitFilesIndex_getByPath.get() != want {
			t.Errorf("%s: Defs: got %d unitFilesIndex hits, want %d", rs, c_unitFilesIndex_getByPath.get(), want)
		}
	}
}

func testRepoStore_Refs(t *testing.T, rs RepoStoreImporter) {
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
	if err := rs.Import("c", unit, data); err != nil {
		t.Errorf("%s: Import(c, %v, data): %s", rs, unit, err)
	}
	if rs, ok := rs.(RepoIndexer); ok {
		if err := rs.Index("c"); err != nil {
			t.Fatalf("%s: Index: %s", rs, err)
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
			CommitID:    "c",
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
			CommitID:    "c",
		},
	}

	refs, err := rs.Refs()
	if err != nil {
		t.Errorf("%s: Refs(): %s", rs, err)
	}
	if !reflect.DeepEqual(refs, want) {
		t.Errorf("%s: Refs(): got refs %v, want %v", rs, refs, want)
	}
}
