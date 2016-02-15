package store

import (
	"reflect"
	"sort"
	"testing"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

func testUnitStore(t *testing.T, newFn func() UnitStoreImporter) {
	testUnitStore_uninitialized(t, newFn())
	testUnitStore_Import_empty(t, newFn())
	testUnitStore_Import(t, newFn())
	testUnitStore_Def(t, newFn())
	testUnitStore_Defs(t, newFn())
	testUnitStore_Defs_SortByName(t, newFn())
	testUnitStore_Defs_Query(t, newFn())
	testUnitStore_Refs(t, newFn())
	testUnitStore_Refs_ByFiles(t, newFn())
	testUnitStore_Refs_ByDef(t, newFn())
}

func testUnitStore_uninitialized(t *testing.T, us UnitStore) {
	defs, err := us.Defs()
	if err == nil {
		t.Errorf("%s: Defs(): got nil err", us)
	}
	if len(defs) != 0 {
		t.Errorf("%s: Defs(): got defs %v, want empty", us, defs)
	}

	refs, err := us.Refs()
	if err == nil {
		t.Errorf("%s: Refs(): got nil err", us)
	}
	if len(refs) != 0 {
		t.Errorf("%s: Refs(): got refs %v, want empty", us, refs)
	}
}

func testUnitStore_empty(t *testing.T, us UnitStore) {
	defs, err := us.Defs()
	if err != nil {
		t.Errorf("%s: Defs(): %s", us, err)
	}
	if len(defs) != 0 {
		t.Errorf("%s: Defs(): got defs %v, want empty", us, defs)
	}

	refs, err := us.Refs()
	if err != nil {
		t.Errorf("%s: Refs(): %s", us, err)
	}
	if len(refs) != 0 {
		t.Errorf("%s: Refs(): got refs %v, want empty", us, refs)
	}
}

func testUnitStore_Import_empty(t *testing.T, us UnitStoreImporter) {
	data := graph.Output{
		Defs: []*graph.Def{},
		Refs: []*graph.Ref{},
	}
	if err := us.Import(data); err != nil {
		t.Errorf("%s: Import(empty data): %s", us, err)
	}
	testUnitStore_empty(t, us)
}

func testUnitStore_Import(t *testing.T, us UnitStoreImporter) {
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
	if err := us.Import(data); err != nil {
		t.Errorf("%s: Import(data): %s", us, err)
	}
}

func testUnitStore_Def(t *testing.T, us UnitStoreImporter) {
	data := graph.Output{
		Defs: []*graph.Def{
			{
				DefKey: graph.DefKey{Path: "p"},
				Name:   "n",
			},
		},
	}
	if err := us.Import(data); err != nil {
		t.Errorf("%s: Import(data): %s", us, err)
	}

	defs, err := us.Defs(ByDefPath("p"))
	if err != nil {
		t.Errorf("%s: Defs: %s", us, err)
	}
	if want := data.Defs; !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs: got def %v, want %v", us, defs, want)
	}

	defs, err = us.Defs(ByDefPath("p2"))
	if err != nil {
		t.Errorf("%s: Defs: %s", us, err)
	}
	if len(defs) != 0 {
		t.Errorf("%s: Defs: got defs %v, want none", us, defs)
	}
}

func testUnitStore_Defs(t *testing.T, us UnitStoreImporter) {
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
	if err := us.Import(data); err != nil {
		t.Errorf("%s: Import(data): %s", us, err)
	}

	defs, err := us.Defs()
	if err != nil {
		t.Errorf("%s: Defs(): %s", us, err)
	}
	if want := data.Defs; !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs(): got defs %v, want %v", us, defs, want)
	}
}

func testUnitStore_Defs_SortByName(t *testing.T, us UnitStoreImporter) {
	data := graph.Output{
		Defs: []*graph.Def{
			{
				DefKey: graph.DefKey{Path: "p1"},
				Name:   "b",
			},
			{
				DefKey: graph.DefKey{Path: "p2"},
				Name:   "c",
			},
			{
				DefKey: graph.DefKey{Path: "p3"},
				Name:   "a",
			},
		},
	}
	if err := us.Import(data); err != nil {
		t.Errorf("%s: Import(data): %s", us, err)
	}

	defs, err := us.Defs(DefsSortByName{})
	if err != nil {
		t.Errorf("%s: Defs(): %s", us, err)
	}
	DefsSortByName{}.DefsSort(data.Defs)
	if want := data.Defs; !reflect.DeepEqual(defs, want) {
		t.Errorf("%s: Defs(): got defs %v, want %v", us, defs, want)
	}
}

func testUnitStore_Defs_Query(t *testing.T, us UnitStoreImporter) {
	data := graph.Output{
		Defs: []*graph.Def{
			{
				DefKey: graph.DefKey{Path: "p1"},
				Name:   "a",
			},
			{
				DefKey: graph.DefKey{Path: "p2"},
				Name:   "ab",
			},
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
	if err := us.Import(data); err != nil {
		t.Errorf("%s: Import(data): %s", us, err)
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
		c_defQueryIndex_getByQuery.set(0)
		defs, err := us.Defs(ByDefQuery(test.q))
		if err != nil {
			t.Errorf("%s: Defs(ByDefQuery %q): %s", us, test.q, err)
		}
		if got, want := defPaths(defs), test.wantDefPaths; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: Defs(ByDefQuery %q): got defs %v, want %v", us, test.q, got, want)
		}
		if isIndexedStore(us) {
			if want := test.wantIndexHits; c_defQueryIndex_getByQuery.get() != want {
				t.Errorf("%s: Defs(ByDefQuery %q): got %d index hits, want %d", us, test.q, c_defQueryIndex_getByQuery.get(), want)
			}
		}
	}
}

func testUnitStore_Refs(t *testing.T, us UnitStoreImporter) {
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
	if err := us.Import(data); err != nil {
		t.Errorf("%s: Import(data): %s", us, err)
	}

	refs, err := us.Refs()
	if err != nil {
		t.Errorf("%s: Refs(): %s", us, err)
	}
	if want := data.Refs; !reflect.DeepEqual(refs, want) {
		t.Errorf("%s: Refs(): got refs %v, want %v", us, refs, want)
	}
}

func testUnitStore_Refs_ByFiles(t *testing.T, us UnitStoreImporter) {
	refsByFile := map[string][]*graph.Ref{
		"f1": {
			{DefPath: "p1", Start: 0, End: 5},
		},
		"f2": {
			{DefPath: "p1", Start: 0, End: 5},
			{DefPath: "p2", Start: 5, End: 10},
		},
		"f3": {
			{DefPath: "p1", Start: 0, End: 5},
			{DefPath: "p2", Start: 5, End: 10},
			{DefPath: "p3", Start: 10, End: 15},
		},
	}
	var data graph.Output
	for file, refs := range refsByFile {
		for _, ref := range refs {
			ref.File = file
		}
		data.Refs = append(data.Refs, refs...)
	}

	if err := us.Import(data); err != nil {
		t.Errorf("%s: Import(data): %s", us, err)
	}

	for file, wantRefs := range refsByFile {
		c_refFileIndex_getByFile.set(0)
		refs, err := us.Refs(ByFiles(false, file))
		if err != nil {
			t.Fatalf("%s: Refs(ByFiles %s): %s", us, file, err)
		}
		sort.Sort(refsByFileStartEnd(refs))
		sort.Sort(refsByFileStartEnd(wantRefs))
		if want := wantRefs; !reflect.DeepEqual(refs, want) {
			t.Errorf("%s: Refs(ByFiles %s): got refs %v, want %v", us, file, refs, want)
		}
		if isIndexedStore(us) {
			if want := 1; c_refFileIndex_getByFile.get() != want {
				t.Errorf("%s: Refs(ByFiles %s): got %d index hits, want %d", us, file, c_refFileIndex_getByFile.get(), want)
			}
		}
	}
}

func testUnitStore_Refs_ByDef(t *testing.T, us UnitStoreImporter) {
	refsByDef := map[string][]*graph.Ref{
		"p1": {
			{File: "f1", Start: 0, End: 5},
		},
		"p2": {
			{File: "f1", Start: 0, End: 5},
			{File: "f2", Start: 5, End: 10},
		},
		"p3": {
			{File: "f3", Start: 0, End: 5},
			{File: "f1", Start: 5, End: 10},
			{File: "f1", Start: 10, End: 15},
		},
	}
	var data graph.Output
	for defPath, refs := range refsByDef {
		for _, ref := range refs {
			ref.DefPath = defPath
		}
		data.Refs = append(data.Refs, refs...)
	}

	if err := us.Import(data); err != nil {
		t.Errorf("%s: Import(data): %s", us, err)
	}

	for defPath, wantRefs := range refsByDef {
		c_defRefsIndex_getByDef.set(0)
		refs, err := us.Refs(ByRefDef(graph.RefDefKey{DefPath: defPath}))
		if err != nil {
			t.Fatalf("%s: Refs(ByDefs %s): %s", us, defPath, err)
		}
		sort.Sort(refsByFileStartEnd(refs))
		sort.Sort(refsByFileStartEnd(wantRefs))
		if want := wantRefs; !reflect.DeepEqual(refs, want) {
			t.Errorf("%s: Refs(ByDefs %s): got refs %v, want %v", us, defPath, refs, want)
		}
		if isIndexedStore(us) {
			if want := 1; c_defRefsIndex_getByDef.get() != want {
				t.Errorf("%s: Refs(ByDefs %s): got %d index hits, want %d", us, defPath, c_defRefsIndex_getByDef.get(), want)
			}
		}
	}
}

func defPaths(defs []*graph.Def) []string {
	dps := make([]string, len(defs))
	for i, def := range defs {
		dps[i] = def.Path
	}
	sort.Strings(dps)
	return dps
}
