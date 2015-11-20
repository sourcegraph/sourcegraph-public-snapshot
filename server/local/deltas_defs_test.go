package local

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/go-diff/diff"
	"sourcegraph.com/sourcegraph/srclib/graph"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestDeltasService_ListDefs(t *testing.T) {
	tests := []struct {
		description string
		baseRevSpec sourcegraph.RepoRevSpec
		headRevSpec sourcegraph.RepoRevSpec
	}{{
		description: "same repo",
		baseRevSpec: sourcegraph.RepoRevSpec{
			RepoSpec: sourcegraph.RepoSpec{URI: "github.com/owner/foo"},
			Rev:      "baserev", CommitID: "basecommit",
		},
		headRevSpec: sourcegraph.RepoRevSpec{
			RepoSpec: sourcegraph.RepoSpec{URI: "github.com/owner/foo"},
			Rev:      "headrev", CommitID: "headcommit",
		},
	}, {
		description: "different repo",
		baseRevSpec: sourcegraph.RepoRevSpec{
			RepoSpec: sourcegraph.RepoSpec{URI: "github.com/owner/foo"},
			Rev:      "baserev", CommitID: "basecommit",
		},
		headRevSpec: sourcegraph.RepoRevSpec{
			RepoSpec: sourcegraph.RepoSpec{URI: "github.com/owner2/foob"},
			Rev:      "headrev", CommitID: "headcommit",
		},
	}}

	for _, test := range tests {
		func() {
			var s deltas
			ctx, mock := testContext()

			ds := sourcegraph.DeltaSpec{
				Base: test.baseRevSpec,
				Head: test.headRevSpec,
			}
			data, diffStat := makeDeltaDefsTestDefns(ds)
			mock.stores.Graph = newMockMultiRepoStoreWithDefs(data)

			dd, err := s.ListDefs(ctx, &sourcegraph.DeltasListDefsOp{Ds: ds, Opt: nil})
			if err != nil {
				t.Fatal(err)
			}

			checkDeltaDefs(t, dd, diffStat)
		}()
	}
}

func TestDeltasService_ListDefs_golangCrossRepo(t *testing.T) {
	var s deltas
	ctx, mock := testContext()

	baseRevSpec := sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: "github.com/baseowner/foo"},
		Rev:      "baserev", CommitID: "basecommit",
	}
	headRevSpec := sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: "github.com/headowner/foo"},
		Rev:      "headrev", CommitID: "headcommit",
	}

	ds := sourcegraph.DeltaSpec{
		Base: baseRevSpec,
		Head: headRevSpec,
	}
	data, diffStat := makeDeltaDefsTestDefns_golangCrossRepo(ds)
	mock.stores.Graph = newMockMultiRepoStoreWithDefs(data)

	dd, err := s.ListDefs(ctx, &sourcegraph.DeltasListDefsOp{Ds: ds, Opt: nil})
	if err != nil {
		t.Fatal(err)
	}

	checkDeltaDefs(t, dd, diffStat)
}

func newMockMultiRepoStoreWithDefs(defs []*graph.Def) srcstore.MockMultiRepoStore {
	mockGraph := srcstore.MockMultiRepoStore{}
	mockGraph.Defs_ = func(filters ...srcstore.DefFilter) ([]*graph.Def, error) {
		var selectedDefs []*graph.Def
		for _, def := range defs {
			selected := true
			for _, filter := range filters {
				selected = selected && filter.SelectDef(def)
			}
			if selected {
				selectedDefs = append(selectedDefs, def)
			}
		}
		return selectedDefs, nil
	}
	return mockGraph
}

func makeDeltaDefsTestDefns_golangCrossRepo(ds sourcegraph.DeltaSpec) ([]*graph.Def, diff.Stat) {
	// Cross-repo comparisons for Go require special handling because
	// the non-vendored pkg unit names (Go import paths) of Go defns
	// in 2 repos are always different, all defns from head are shown
	// as added and all defns from base are shown as deleted.
	//
	// But we also need to make sure it handles vendored pkg names
	// correctly. For example, packages in a repo's Godeps/_workspace
	// GOPATH do NOT have their import paths (unit names) prefixed
	// with the current repo.
	//
	// This test tests that package names are only considered different if
	// they differ after the repository URI.

	// The following test data has some defns that are the "same"
	// logically but just have different Go import paths.
	defns := []*graph.Def{
		&graph.Def{
			DefKey: graph.DefKey{
				Repo:     ds.Base.RepoSpec.URI,
				CommitID: ds.Base.CommitID,
				UnitType: "GoPackage",
				Unit:     ds.Base.RepoSpec.URI,
				Path:     "unchanged-defn",
			},
			Exported: true,
			Data:     []byte("{}"),
		},
		&graph.Def{
			DefKey: graph.DefKey{
				Repo:     ds.Head.RepoSpec.URI,
				CommitID: ds.Head.CommitID,
				UnitType: "GoPackage",
				Unit:     ds.Head.RepoSpec.URI,
				Path:     "unchanged-defn",
			},
			Exported: true,
			Data:     []byte("{}"),
		},
		&graph.Def{
			DefKey: graph.DefKey{
				Repo:     ds.Base.RepoSpec.URI,
				CommitID: ds.Base.CommitID,
				UnitType: "GoPackage",
				Unit:     ds.Base.RepoSpec.URI,
				Path:     "deleted-defn",
			},
			Exported: true,
			Data:     []byte("{}"),
		},
		&graph.Def{
			DefKey: graph.DefKey{
				Repo:     ds.Head.RepoSpec.URI,
				CommitID: ds.Head.CommitID,
				UnitType: "GoPackage",
				Unit:     ds.Head.RepoSpec.URI,
				Path:     "added-defn",
			},
			Exported: true,
			Data:     []byte("{}"),
		},

		// defs from vendored packages
		&graph.Def{
			DefKey: graph.DefKey{
				Repo:     ds.Base.RepoSpec.URI,
				CommitID: ds.Base.CommitID,
				UnitType: "GoPackage",
				Unit:     "github.com/myvendored/pkg",
				Path:     "vendored-defn",
			},
			Exported: true,
			Data:     []byte("{}"),
		},
		&graph.Def{
			DefKey: graph.DefKey{
				Repo:     ds.Head.RepoSpec.URI,
				CommitID: ds.Head.CommitID,
				UnitType: "GoPackage",
				Unit:     "github.com/myvendored/pkg",
				Path:     "vendored-defn",
			},
			Exported: true,
			Data:     []byte("{}"),
		},
	}
	return defns, diff.Stat{Added: 1, Deleted: 1}
}

// makeDeltaDefsTestDefns returns test data that includes added, changed, unchanged, and deleted defs for the given
// DeltaSpec. Should be used in conjunction with checkDeltaDefs.
func makeDeltaDefsTestDefns(ds sourcegraph.DeltaSpec) ([]*graph.Def, diff.Stat) {
	return []*graph.Def{
		&graph.Def{
			DefKey: graph.DefKey{
				Repo:     ds.Base.RepoSpec.URI,
				CommitID: ds.Base.CommitID,
				UnitType: "t",
				Unit:     "u",
				Path:     "unchanged-defn",
			},
			Exported: true,
			Data:     []byte("{}"),
		},
		&graph.Def{
			DefKey: graph.DefKey{
				Repo:     ds.Head.RepoSpec.URI,
				CommitID: ds.Head.CommitID,
				UnitType: "t",
				Unit:     "u",
				Path:     "unchanged-defn",
			},
			Exported: true,
			Data:     []byte("{}"),
		},
		&graph.Def{
			DefKey: graph.DefKey{
				Repo:     ds.Base.RepoSpec.URI,
				CommitID: ds.Base.CommitID,
				UnitType: "t",
				Unit:     "u",
				Path:     "changed-defn",
			},
			DefStart: 10,
			DefEnd:   20,
			Exported: true,
			Data:     []byte("{}"),
		},
		&graph.Def{
			DefKey: graph.DefKey{
				Repo:     ds.Head.RepoSpec.URI,
				CommitID: ds.Head.CommitID,
				UnitType: "t",
				Unit:     "u",
				Path:     "changed-defn",
			},
			DefStart: 10,
			DefEnd:   30, // added 10 chars from base repo
			Exported: true,
			Data:     []byte("{}"),
		},
		&graph.Def{
			DefKey: graph.DefKey{
				Repo:     ds.Base.RepoSpec.URI,
				CommitID: ds.Base.CommitID,
				UnitType: "t",
				Unit:     "u",
				Path:     "deleted-defn",
			},
			Exported: true,
			Data:     []byte("{}"),
		},
		&graph.Def{
			DefKey: graph.DefKey{
				Repo:     ds.Head.RepoSpec.URI,
				CommitID: ds.Head.CommitID,
				UnitType: "t",
				Unit:     "u",
				Path:     "added-defn",
			},
			Exported: true,
			Data:     []byte("{}"),
		},
	}, diff.Stat{Added: 1, Deleted: 1, Changed: 1}
}

// checkDeltaDefs checks that dd's added defns all have the path
// "added-defn", dd's changed defns all have the path "changed-defn",
// and dd's deleted defns all have the path "deleted-defn", and also
// that its diff stat matches wantDiffStat.
func checkDeltaDefs(t *testing.T, dd *sourcegraph.DeltaDefs, wantDiffStat diff.Stat) {
	// Ensure that the only added and deleted defns are added-defn and
	// deleted-defn, respectively. The unchanged-defn should not
	// appear in either the added or deleted list.
	for _, dd := range dd.Defs {
		if dd.Added() {
			if want := "added-defn"; string(dd.Head.Path) != want {
				t.Errorf("got added defn %v, want only defn with Path=%q", dd.Head.DefKey, want)
			}
		}
		if dd.Changed() {
			if want := "changed-defn"; string(dd.Head.Path) != want {
				t.Errorf("got changed defn %v, want only defn with Path=%q", dd.Head.DefKey, want)
			}
		}
		if dd.Deleted() {
			if want := "deleted-defn"; string(dd.Base.Path) != want {
				t.Errorf("got deleted defn %v, want only defn with Path=%q", dd.Base.DefKey, want)
			}
		}
	}
	if dd.DiffStat != wantDiffStat {
		t.Errorf("got diffstat %+v, want %+v", dd.DiffStat, wantDiffStat)
	}
}

func TestChunkDiffOps(t *testing.T) {
	baseEntrySpec := sourcegraph.TreeEntrySpec{
		Path: "basefile",
		RepoRev: sourcegraph.RepoRevSpec{
			RepoSpec: sourcegraph.RepoSpec{URI: "github.com/baseowner/foo"},
			Rev:      "baserev", CommitID: "basecommit",
		},
	}
	headEntrySpec := sourcegraph.TreeEntrySpec{
		Path: "headfile",
		RepoRev: sourcegraph.RepoRevSpec{
			RepoSpec: sourcegraph.RepoSpec{URI: "github.com/headowner/foo"},
			Rev:      "headrev", CommitID: "headcommit",
		},
	}

	tests := map[string]struct {
		hunk *sourcegraph.Hunk
		want []*repoTreeGetOp
	}{
		"basic": {
			hunk: &sourcegraph.Hunk{Hunk: diff.Hunk{
				Body: []byte(`
 a
-b
+c
+d
 e
+f
 g
-h
`),
			}},
			want: []*repoTreeGetOp{
				{baseEntrySpec, sourcegraph.RepoTreeGetOptions{Formatted: true, GetFileOptions: vcsclient.GetFileOptions{FileRange: vcsclient.FileRange{StartLine: 0, EndLine: 1}}}},
				{headEntrySpec, sourcegraph.RepoTreeGetOptions{Formatted: true, GetFileOptions: vcsclient.GetFileOptions{FileRange: vcsclient.FileRange{StartLine: 1, EndLine: 2}}}},
				{baseEntrySpec, sourcegraph.RepoTreeGetOptions{Formatted: true, GetFileOptions: vcsclient.GetFileOptions{FileRange: vcsclient.FileRange{StartLine: 2, EndLine: 2}}}},
				{headEntrySpec, sourcegraph.RepoTreeGetOptions{Formatted: true, GetFileOptions: vcsclient.GetFileOptions{FileRange: vcsclient.FileRange{StartLine: 4, EndLine: 4}}}},
				{baseEntrySpec, sourcegraph.RepoTreeGetOptions{Formatted: true, GetFileOptions: vcsclient.GetFileOptions{FileRange: vcsclient.FileRange{StartLine: 3, EndLine: 4}}}},
			},
		},
	}
	for label, test := range tests {
		test.hunk.Body = bytes.TrimPrefix(test.hunk.Body, []byte{'\n'})
		ops := chunkDiffOps(baseEntrySpec, headEntrySpec, test.hunk)
		if !reflect.DeepEqual(ops, test.want) {
			t.Errorf("%s: got ops != want\n\ngot ops =======\n%+v\n\nwant ops =======\n%+v", label, diffOpString(ops), diffOpString(test.want))
			continue
		}
	}
}

func diffOpString(ops []*repoTreeGetOp) string {
	var s []string
	for _, op := range ops {
		s = append(s, fmt.Sprintf("entrySpec = %+v\n  opt = %+v\n\n", op.file, op.opt.GetFileOptions.FileRange))
	}
	return strings.Join(s, "\n")
}

func TestSetHunkLines(t *testing.T) {
	tests := map[string]struct {
		origBody string
		fmtBody  string
		want     string
	}{
		"basic": {
			origBody: `+added
-deleted
 same
+added`,
			fmtBody: `ADDED
DELETED
SAME
ADDED`,
			want: `+ADDED
-DELETED
 SAME
+ADDED`,
		},
	}
	for label, test := range tests {
		got, err := setHunkLines([]byte(test.origBody), []byte(test.fmtBody))
		if err != nil {
			t.Errorf("%s: setHunkLines: %s", label, err)
			continue
		}
		if string(got) != test.want {
			t.Errorf("%s: got != want\n\ngot =======\n%s\n\nwant =======\n%s", label, got, test.want)
			continue
		}
	}
}
