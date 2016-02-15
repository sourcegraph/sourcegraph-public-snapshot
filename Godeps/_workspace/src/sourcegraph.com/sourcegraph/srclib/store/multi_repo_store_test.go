package store

import (
	"reflect"
	"testing"

	"sort"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

func testMultiRepoStore(t *testing.T, newFn func() MultiRepoStoreImporter) {
	testMultiRepoStore_uninitialized(t, newFn())
	testMultiRepoStore_Import_empty(t, newFn())
	testMultiRepoStore_Import(t, newFn())
	testMultiRepoStore_Repos(t, newFn())
	testMultiRepoStore_Repos_ByRepos(t, newFn())
	testMultiRepoStore_Versions(t, newFn())
	testMultiRepoStore_Units(t, newFn())
	testMultiRepoStore_Def(t, newFn())
	testMultiRepoStore_Defs(t, newFn())
	testMultiRepoStore_Defs_filter(t, newFn())
	testMultiRepoStore_Defs_ByRepos(t, newFn())
	testMultiRepoStore_Defs_ByRepos_ByDefQuery(t, newFn())
	testMultiRepoStore_Defs_ByRepoCommitIDs(t, newFn())
	testMultiRepoStore_Defs_ByRepoCommitIDs_ByDefQuery(t, newFn())
	testMultiRepoStore_Refs(t, newFn())
	testMultiRepoStore_Refs_filterByRepoCommitAndFile(t, newFn())
	testMultiRepoStore_Refs_filterByDef(t, newFn())
}

func testMultiRepoStore_uninitialized(t *testing.T, mrs MultiRepoStore) {
	repos, _ := mrs.Repos()
	if len(repos) != 0 {
		t.Errorf("%s: Repos(): got repos %v, want empty", mrs, repos)
	}

	testRepoStore_uninitialized(t, mrs)
}

func testMultiRepoStore_Import_empty(t *testing.T, mrs MultiRepoStoreImporter) {
	if err := mrs.Import("r", "c", nil, graph.Output{}); err != nil {
		t.Errorf("%s: Import(c, nil, empty): %s", mrs, err)
	}
	testTreeStore_empty(t, mrs)
}

func testMultiRepoStore_Import(t *testing.T, mrs MultiRepoStoreImporter) {
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
	if err := mrs.Import("r", "c", unit, data); err != nil {
		t.Errorf("%s: Import(c, %v, data): %s", mrs, unit, err)
	}
}

func testMultiRepoStore_Repos(t *testing.T, mrs MultiRepoStoreImporter) {
	for _, repo := range []string{"r1", "r2"} {
		unit := &unit.SourceUnit{Type: "t1", Name: "u1"}
		if err := mrs.Import(repo, "c", unit, graph.Output{}); err != nil {
			t.Errorf("%s: Import(%s, c, %v, empty data): %s", mrs, repo, unit, err)
		}
	}

	want := []string{"r1", "r2"}

	repos, err := mrs.Repos()
	if err != nil {
		t.Errorf("%s: Repos(): %s", mrs, err)
	}
	sort.Strings(repos)
	sort.Strings(want)
	if !reflect.DeepEqual(repos, want) {
		t.Errorf("%s: Repos(): got %v, want %v", mrs, repos, want)
	}

	repos2, err := mrs.Repos(ByRepos("r1"))
	if err != nil {
		t.Fatalf("%s: Repos(ByRepos r1): %s", mrs, err)
	}
	if want := []string{"r1"}; !reflect.DeepEqual(repos2, want) {
		t.Errorf("%s: Repos(ByRepos r1): got %v, want %v", mrs, repos2, want)
	}
}

func testMultiRepoStore_Repos_ByRepos(t *testing.T, mrs MultiRepoStoreImporter) {
	for _, repo := range []string{"r1", "r2", "r3"} {
		unit := &unit.SourceUnit{Type: "t1", Name: "u1"}
		if err := mrs.Import(repo, "c", unit, graph.Output{}); err != nil {
			t.Errorf("%s: Import(%s, c, %v, empty data): %s", mrs, repo, unit, err)
		}
	}

	{
		repos, err := mrs.Repos(ByRepos("r1", "r3"))
		if err != nil {
			t.Errorf("%s: Repos: %s", mrs, err)
		}
		want := []string{"r1", "r3"}
		sort.Strings(repos)
		sort.Strings(want)
		if !reflect.DeepEqual(repos, want) {
			t.Errorf("%s: Repos: got %v, want %v", mrs, repos, want)
		}
	}

	{
		repos, err := mrs.Repos(ByRepos("r1"))
		if err != nil {
			t.Fatalf("%s: Repos: %s", mrs, err)
		}
		if want := []string{"r1"}; !reflect.DeepEqual(repos, want) {
			t.Errorf("%s: Repos: got %v, want %v", mrs, repos, want)
		}
	}

	{
		repos, err := mrs.Repos(ByRepos())
		if err != nil {
			t.Fatalf("%s: Repos: %s", mrs, err)
		}
		if want := []string{}; !reflect.DeepEqual(repos, want) {
			t.Errorf("%s: Repos: got %v, want %v", mrs, repos, want)
		}
	}
}

func TestFSMultiRepoStore_Repos_customPathFuncs(t *testing.T) {
	tests := map[string]struct{ conf *FSMultiRepoStoreConf }{
		"nil struct":         {conf: nil},
		"evenly distributed": {conf: &FSMultiRepoStoreConf{RepoPaths: &customRepoPaths{}}},
	}
	for label, test := range tests {
		mrs := NewFSMultiRepoStore(newTestFS(), test.conf)

		allRepos := []string{"r1", "r2", "r3"}
		for _, repo := range allRepos {
			unit := &unit.SourceUnit{Type: "t1", Name: "u1"}
			if err := mrs.Import(repo, "c", unit, graph.Output{}); err != nil {
				t.Errorf("%s: Import(%s, c, %v, empty data): %s", label, repo, unit, err)
				continue
			}
		}

		repos, err := mrs.Repos()
		if err != nil {
			t.Errorf("%s: Repos(): %s", label, err)
		}
		sort.Strings(repos)
		if want := allRepos; !reflect.DeepEqual(repos, want) {
			t.Errorf("%s: Repos(): got %v, want %v", label, repos, want)
		}

		for _, repo := range allRepos {
			repos, err := mrs.Repos(ByRepos(repo))
			if err != nil {
				t.Errorf("%s: Repos(ByRepos %s): %s", label, repo, err)
				continue
			}
			if want := []string{repo}; !reflect.DeepEqual(repos, want) {
				t.Errorf("%s: Repos(ByRepos %s): got %v, want %v", label, repo, repos, want)
			}
		}
	}
}

func testMultiRepoStore_Versions(t *testing.T, mrs MultiRepoStoreImporter) {
	for _, version := range []string{"c1", "c2"} {
		unit := &unit.SourceUnit{Type: "t1", Name: "u1"}
		if err := mrs.Import("r", version, unit, graph.Output{}); err != nil {
			t.Errorf("%s: Import(%s, %v, empty data): %s", mrs, version, unit, err)
		}
	}

	want := []*Version{{Repo: "r", CommitID: "c1"}, {Repo: "r", CommitID: "c2"}}

	versions, err := mrs.Versions()
	if err != nil {
		t.Errorf("%s: Versions(): %s", mrs, err)
	}
	if !reflect.DeepEqual(versions, want) {
		t.Errorf("%s: Versions(): got %v, want %v", mrs, versions, want)
	}

	versions2, err := mrs.Versions(ByCommitIDs("c2"))
	if err != nil {
		t.Errorf("%s: Versionss(ByCommitIDs c2): %s", mrs, err)
	}
	if want := []*Version{{Repo: "r", CommitID: "c2"}}; !reflect.DeepEqual(versions2, want) {
		t.Errorf("%s: Versions(ByCommitIDs c2): got %v, want %v", mrs, versions2, want)
	}
}

func testMultiRepoStore_Units(t *testing.T, mrs MultiRepoStoreImporter) {
	units := []*unit.SourceUnit{
		{Type: "t1", Name: "u1"},
		{Type: "t2", Name: "u2"},
	}
	for _, unit := range units {
		if err := mrs.Import("r", "c", unit, graph.Output{}); err != nil {
			t.Errorf("%s: Import(c, %v, empty data): %s", mrs, unit, err)
		}
	}
	if mrs, ok := mrs.(MultiRepoIndexer); ok {
		if err := mrs.Index("r", "c"); err != nil {
			t.Fatalf("%s: Index: %s", mrs, err)
		}
	}

	want := []*unit.SourceUnit{
		{Repo: "r", CommitID: "c", Type: "t1", Name: "u1"},
		{Repo: "r", CommitID: "c", Type: "t2", Name: "u2"},
	}

	units, err := mrs.Units()
	if err != nil {
		t.Errorf("%s: Units(): %s", mrs, err)
	}
	sort.Sort(unit.SourceUnits(units))
	sort.Sort(unit.SourceUnits(want))
	if !reflect.DeepEqual(units, want) {
		t.Errorf("%s: Units(): got %v, want %v", mrs, units, want)
	}

	units2, err := mrs.Units(ByUnits(unit.ID2{Type: "t2", Name: "u2"}))
	if err != nil {
		t.Errorf("%s: Units(t2 u2): %s", mrs, err)
	}
	if want := []*unit.SourceUnit{{Repo: "r", CommitID: "c", Type: "t2", Name: "u2"}}; !reflect.DeepEqual(units2, want) {
		t.Errorf("%s: Units(t2 u2): got %v, want %v", mrs, units2, want)
	}
}

func testMultiRepoStore_Def(t *testing.T, mrs MultiRepoStoreImporter) {
	unit := &unit.SourceUnit{Type: "t", Name: "u"}
	data := graph.Output{
		Defs: []*graph.Def{
			{
				DefKey: graph.DefKey{Path: "p"},
				Name:   "n",
			},
		},
	}
	if err := mrs.Import("r", "c", unit, data); err != nil {
		t.Errorf("%s: Import(c, %v, data): %s", mrs, unit, err)
	}

	want := []*graph.Def{
		{
			DefKey: graph.DefKey{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"},
			Name:   "n",
		},
	}
	defs, err := mrs.Defs(ByDefKey(graph.DefKey{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"}))
	if err != nil {
		t.Errorf("%s: Defs: %s", mrs, err)
	}
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs: got defs %v, want %v", mrs, defs, want)
	}

	defs2, err := mrs.Defs(ByDefKey(graph.DefKey{Repo: "r2", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"}))
	if err != nil {
		t.Errorf("%s: Defs: %s", mrs, err)
	}
	if len(defs2) != 0 {
		t.Errorf("%s: Defs: got defs %v, want none", mrs, defs2)
	}
}

func testMultiRepoStore_Defs(t *testing.T, mrs MultiRepoStoreImporter) {
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
	if err := mrs.Import("r", "c", unit, data); err != nil {
		t.Errorf("%s: Import(c, %v, data): %s", mrs, unit, err)
	}

	want := []*graph.Def{
		{
			DefKey: graph.DefKey{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u", Path: "p1"},
			Name:   "n1",
		},
		{
			DefKey: graph.DefKey{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u", Path: "p2"},
			Name:   "n2",
		},
	}

	defs, err := mrs.Defs()
	if err != nil {
		t.Errorf("%s: Defs(): %s", mrs, err)
	}
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs(): got defs %v, want %v", mrs, defs, want)
	}
}

func testMultiRepoStore_Defs_filter(t *testing.T, mrs MultiRepoStoreImporter) {
	if err := mrs.Import("r", "c", &unit.SourceUnit{Type: "t", Name: "u"}, graph.Output{Defs: []*graph.Def{
		{DefKey: graph.DefKey{Path: "p"}},
		{DefKey: graph.DefKey{Path: "p2"}},
	}}); err != nil {
		t.Errorf("%s: Import: %s", mrs, err)
	}
	if err := mrs.Import("r", "c2", &unit.SourceUnit{Type: "t", Name: "u"}, graph.Output{Defs: []*graph.Def{{DefKey: graph.DefKey{Path: "p"}}}}); err != nil {
		t.Errorf("%s: Import: %s", mrs, err)
	}
	if err := mrs.Import("r2", "c2", &unit.SourceUnit{Type: "t", Name: "u"}, graph.Output{Defs: []*graph.Def{{DefKey: graph.DefKey{Path: "p"}}}}); err != nil {
		t.Errorf("%s: Import: %s", mrs, err)
	}
	if mrs, ok := mrs.(MultiRepoIndexer); ok {
		if err := mrs.Index("r", "c"); err != nil {
			t.Fatalf("%s: Index: %s", mrs, err)
		}
		if err := mrs.Index("r", "c2"); err != nil {
			t.Fatalf("%s: Index: %s", mrs, err)
		}
		if err := mrs.Index("r2", "c2"); err != nil {
			t.Fatalf("%s: Index: %s", mrs, err)
		}
	}

	want := []*graph.Def{
		{
			DefKey: graph.DefKey{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"},
		},
	}

	defs, err := mrs.Defs(ByRepos("r"), ByCommitIDs("c"), ByDefPath("p"))
	if err != nil {
		t.Errorf("%s: Defs(): %s", mrs, err)
	}
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs(): got defs %v, want %v", mrs, defs, want)
	}
}

func testMultiRepoStore_Defs_ByRepos(t *testing.T, mrs MultiRepoStoreImporter) {
	repos := []string{"r1", "r2", "r3"}
	for _, repo := range repos {
		if err := mrs.Import(repo, "c", &unit.SourceUnit{Type: "t", Name: "u"}, graph.Output{Defs: []*graph.Def{
			{DefKey: graph.DefKey{Path: "p"}},
		}}); err != nil {
			t.Errorf("%s: Import: %s", mrs, err)
		}
		if mrs, ok := mrs.(MultiRepoIndexer); ok {
			if err := mrs.Index(repo, "c"); err != nil {
				t.Fatalf("%s: Index: %s", mrs, err)
			}
		}
	}

	want := []*graph.Def{
		{DefKey: graph.DefKey{Repo: "r1", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"}},
		{DefKey: graph.DefKey{Repo: "r3", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"}},
	}

	defs, err := mrs.Defs(ByRepos("r1", "r3"))
	if err != nil {
		t.Errorf("%s: Defs: %s", mrs, err)
	}
	sort.Sort(graph.Defs(defs))
	sort.Sort(graph.Defs(want))
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs: got defs %v, want %v", mrs, defs, want)
	}
}

func testMultiRepoStore_Defs_ByRepos_ByDefQuery(t *testing.T, mrs MultiRepoStoreImporter) {
	repos := []string{"r1", "r2", "r3"}
	for _, repo := range repos {
		data := graph.Output{
			Defs: []*graph.Def{
				{DefKey: graph.DefKey{Path: "p1"}, Name: "abc-" + repo},
				{DefKey: graph.DefKey{Path: "p2"}, Name: "xyz-" + repo},
			},
		}
		if err := mrs.Import(repo, "c", &unit.SourceUnit{Type: "t", Name: "u"}, data); err != nil {
			t.Errorf("%s: Import: %s", mrs, err)
		}
		if mrs, ok := mrs.(MultiRepoIndexer); ok {
			if err := mrs.Index(repo, "c"); err != nil {
				t.Fatalf("%s: Index: %s", mrs, err)
			}
		}
	}

	c_defQueryTreeIndex_getByQuery.set(0)

	want := []*graph.Def{
		{DefKey: graph.DefKey{Repo: "r1", CommitID: "c", UnitType: "t", Unit: "u", Path: "p1"}, Name: "abc-r1"},
		{DefKey: graph.DefKey{Repo: "r3", CommitID: "c", UnitType: "t", Unit: "u", Path: "p1"}, Name: "abc-r3"},
	}

	defs, err := mrs.Defs(ByRepos("r1", "r3"), ByDefQuery("abc"))
	if err != nil {
		t.Errorf("%s: Defs: %s", mrs, err)
	}
	sort.Sort(graph.Defs(defs))
	sort.Sort(graph.Defs(want))
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs: got defs %v, want %v", mrs, defs, want)
	}
	if isIndexedStore(mrs) {
		if want := 2; c_defQueryTreeIndex_getByQuery.get() != want {
			t.Errorf("%s: Defs: got %d index hits on tree def query index, want %d", mrs, c_defQueryTreeIndex_getByQuery.get(), want)
		}
	}
}

func testMultiRepoStore_Defs_ByRepoCommitIDs(t *testing.T, mrs MultiRepoStoreImporter) {
	repos := []string{"r1", "r2", "r3"}
	commitIDs := []string{"c1", "c2"}
	for _, repo := range repos {
		for _, commitID := range commitIDs {
			if err := mrs.Import(repo, commitID, &unit.SourceUnit{Type: "t", Name: "u"}, graph.Output{Defs: []*graph.Def{
				{DefKey: graph.DefKey{Path: "p"}},
			}}); err != nil {
				t.Errorf("%s: Import: %s", mrs, err)
			}
			if mrs, ok := mrs.(MultiRepoIndexer); ok {
				if err := mrs.Index(repo, commitID); err != nil {
					t.Fatalf("%s: Index: %s", mrs, err)
				}
			}
		}
	}

	want := []*graph.Def{
		{DefKey: graph.DefKey{Repo: "r1", CommitID: "c2", UnitType: "t", Unit: "u", Path: "p"}},
		{DefKey: graph.DefKey{Repo: "r3", CommitID: "c1", UnitType: "t", Unit: "u", Path: "p"}},
	}

	defs, err := mrs.Defs(ByRepoCommitIDs(Version{Repo: "r1", CommitID: "c2"}, Version{Repo: "r3", CommitID: "c1"}))
	if err != nil {
		t.Errorf("%s: Defs: %s", mrs, err)
	}
	sort.Sort(graph.Defs(defs))
	sort.Sort(graph.Defs(want))
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs: got defs %v, want %v", mrs, defs, want)
	}
}

func testMultiRepoStore_Defs_ByRepoCommitIDs_ByDefQuery(t *testing.T, mrs MultiRepoStoreImporter) {
	repos := []string{"r1", "r2", "r3"}
	commitIDs := []string{"c1", "c2"}
	for _, repo := range repos {
		for _, commitID := range commitIDs {
			data := graph.Output{
				Defs: []*graph.Def{
					{DefKey: graph.DefKey{Path: "p1"}, Name: "abc-" + repo},
					{DefKey: graph.DefKey{Path: "p2"}, Name: "xyz-" + repo},
				},
			}
			if err := mrs.Import(repo, commitID, &unit.SourceUnit{Type: "t", Name: "u"}, data); err != nil {
				t.Errorf("%s: Import: %s", mrs, err)
			}
			if mrs, ok := mrs.(MultiRepoIndexer); ok {
				if err := mrs.Index(repo, commitID); err != nil {
					t.Fatalf("%s: Index: %s", mrs, err)
				}
			}
		}
	}

	c_defQueryTreeIndex_getByQuery.set(0)

	want := []*graph.Def{
		{DefKey: graph.DefKey{Repo: "r1", CommitID: "c2", UnitType: "t", Unit: "u", Path: "p1"}, Name: "abc-r1"},
		{DefKey: graph.DefKey{Repo: "r3", CommitID: "c1", UnitType: "t", Unit: "u", Path: "p1"}, Name: "abc-r3"},
	}

	defs, err := mrs.Defs(ByRepoCommitIDs(Version{Repo: "r1", CommitID: "c2"}, Version{Repo: "r3", CommitID: "c1"}), ByDefQuery("abc"))
	if err != nil {
		t.Errorf("%s: Defs: %s", mrs, err)
	}
	sort.Sort(graph.Defs(defs))
	sort.Sort(graph.Defs(want))
	if !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs: got defs %v, want %v", mrs, defs, want)
	}
	if isIndexedStore(mrs) {
		if want := 2; c_defQueryTreeIndex_getByQuery.get() != want {
			t.Errorf("%s: Defs: got %d index hits on tree def query index, want %d", mrs, c_defQueryTreeIndex_getByQuery.get(), want)
		}
	}
}

func testMultiRepoStore_Refs(t *testing.T, mrs MultiRepoStoreImporter) {
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
	if err := mrs.Import("r", "c", unit, data); err != nil {
		t.Errorf("%s: Import(c, %v, data): %s", mrs, unit, err)
	}

	want := []*graph.Ref{
		{
			DefRepo:     "r",
			DefUnitType: "t",
			DefUnit:     "u",
			DefPath:     "p1",
			File:        "f1",
			Start:       1,
			End:         2,
			Repo:        "r",
			UnitType:    "t",
			Unit:        "u",
			CommitID:    "c",
		},
		{
			DefRepo:     "r",
			DefUnitType: "t",
			DefUnit:     "u",
			DefPath:     "p2",
			File:        "f2",
			Start:       2,
			End:         3,
			Repo:        "r",
			UnitType:    "t",
			Unit:        "u",
			CommitID:    "c",
		},
	}

	refs, err := mrs.Refs()
	if err != nil {
		t.Errorf("%s: Refs(): %s", mrs, err)
	}
	if !reflect.DeepEqual(refs, want) {
		t.Errorf("%s: Refs(): got refs %v, want %v", mrs, refs, want)
	}
}

func testMultiRepoStore_Refs_filterByRepoCommitAndFile(t *testing.T, mrs MultiRepoStoreImporter) {
	data1 := graph.Output{
		Refs: []*graph.Ref{
			{File: "f1"},
			{File: "f2"},
			{File: "f3"},
		},
	}
	if err := mrs.Import("r", "c", &unit.SourceUnit{Type: "t", Name: "u", Files: []string{"f1", "f2", "f3"}}, data1); err != nil {
		t.Errorf("%s: Import: %s", mrs, err)
	}
	data2 := graph.Output{
		Refs: []*graph.Ref{
			{File: "f4"},
			{File: "f5"},
			{File: "f6"},
		},
	}
	if err := mrs.Import("r", "c2", &unit.SourceUnit{Type: "t", Name: "u", Files: []string{"f4", "f5", "f6"}}, data2); err != nil {
		t.Errorf("%s: Import: %s", mrs, err)
	}
	data3 := graph.Output{
		Refs: []*graph.Ref{
			{File: "f7"},
			{File: "f8"},
			{File: "f9"},
		},
	}
	if err := mrs.Import("r2", "c", &unit.SourceUnit{Type: "t", Name: "u", Files: []string{"f7", "f8", "f9"}}, data3); err != nil {
		t.Errorf("%s: Import: %s", mrs, err)
	}
	if mrs, ok := mrs.(MultiRepoIndexer); ok {
		if err := mrs.Index("r", "c"); err != nil {
			t.Fatalf("%s: Index: %s", mrs, err)
		}
		if err := mrs.Index("r", "c2"); err != nil {
			t.Fatalf("%s: Index: %s", mrs, err)
		}
		if err := mrs.Index("r2", "c"); err != nil {
			t.Fatalf("%s: Index: %s", mrs, err)
		}
	}

	want := []*graph.Ref{
		{
			DefRepo:     "r",
			DefUnitType: "t",
			DefUnit:     "u",
			File:        "f1",
			CommitID:    "c",
			Repo:        "r",
			Unit:        "u",
			UnitType:    "t",
		},
		{
			DefRepo:     "r",
			DefUnitType: "t",
			DefUnit:     "u",
			File:        "f3",
			CommitID:    "c",
			Repo:        "r",
			Unit:        "u",
			UnitType:    "t",
		},
	}

	byFiles := RefFilterFunc(func(ref *graph.Ref) bool { return ref.File == "f1" || ref.File == "f3" })
	refs, err := mrs.Refs(ByRepos("r"), ByCommitIDs("c"), byFiles)
	if err != nil {
		t.Errorf("%s: Refs(): %s", mrs, err)
	}
	if !reflect.DeepEqual(refs, want) {
		t.Errorf("%s: Refs(): got refs %v, want %v", mrs, refs, want)
	}
}

func testMultiRepoStore_Refs_filterByDef(t *testing.T, mrs MultiRepoStoreImporter) {
	data := graph.Output{
		Refs: []*graph.Ref{
			{
				DefRepo:     "",
				DefUnitType: "",
				DefUnit:     "",
				DefPath:     "p",
				File:        "f",
			},
		},
	}
	if err := mrs.Import("r", "c", &unit.SourceUnit{Type: "t", Name: "u", Files: []string{"f"}}, data); err != nil {
		t.Errorf("%s: Import: %s", mrs, err)
	}
	if mrs, ok := mrs.(MultiRepoIndexer); ok {
		if err := mrs.Index("r", "c"); err != nil {
			t.Fatalf("%s: Index: %s", mrs, err)
		}
	}

	want := []*graph.Ref{
		{
			DefRepo:     "r",
			DefUnitType: "t",
			DefUnit:     "u",
			DefPath:     "p",
			File:        "f",
			CommitID:    "c",
			Repo:        "r",
			Unit:        "u",
			UnitType:    "t",
		},
	}

	// Note: this filter does not work because DefRepo is populated
	// sparsely. See the docs on byRefDefFilter for more info.
	//
	//   RefFilterFunc(func(ref *graph.Ref) bool { return ref.DefRepo == "r" })
	//

	refs, err := mrs.Refs(ByRefDef(graph.RefDefKey{DefPath: "p", DefRepo: "r", DefUnitType: "t", DefUnit: "u"}))
	if err != nil {
		t.Errorf("%s: Refs(): %s", mrs, err)
	}
	if !reflect.DeepEqual(refs, want) {
		t.Errorf("%s: Refs(): got refs %v, want %v", mrs, refs, want)
	}
}
