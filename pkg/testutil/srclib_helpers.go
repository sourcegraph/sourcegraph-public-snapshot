package testutil

import (
	"reflect"
	"strconv"
	"testing"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

// CheckImport checks whether the defs produced by the srclib-sample
// toolchain have been imported. It is hardcoded dependent on the
// vendored srclib-sample toolchain in
// pkg/testutil/testdata/srclibpath.
//
// It's important to call CheckImport because it ensures that both the
// per-source-unit import step and the index step run correctly. If
// you just checked that specific defs exist, then the index creation
// could fail (or have a bug) without you realizing it.
func CheckImport(t *testing.T, ctx context.Context, repo, commitID string) {
	cl, _ := sourcegraph.NewClientFromContext(ctx)

	if len(commitID) != 40 {
		res, err := cl.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
			Repo: sourcegraph.RepoSpec{URI: repo},
			Rev:  commitID,
		})
		if err != nil {
			t.Fatal(err)
		}
		commitID = res.CommitID
	}

	const n = 6 // hardcoded (must be same as the number of units srclib-sample produces)
	var sampleDefs []sourcegraph.DefSpec
	for i := 0; i < n; i++ {
		sampleDefs = append(sampleDefs, sourcegraph.DefSpec{Repo: repo, CommitID: commitID, UnitType: "sample", Unit: "myunit" + strconv.Itoa(i), Path: "mydef" + strconv.Itoa(i)})
	}

	// Specific def lookup.
	for _, defSpec := range sampleDefs {
		def, err := cl.Defs.Get(ctx, &sourcegraph.DefsGetOp{Def: defSpec, Opt: nil})
		if err != nil {
			t.Errorf("failed to get def %#v: %s", defSpec, err)
			continue
		}

		// Not important to test; omitting this lets us not hardcode
		// it in sampleDefs, which looks nicer.
		def.CommitID = ""

		if def.Path != defSpec.Path {
			t.Errorf("got def path %q, want %q", def.Path, defSpec.Path)
		}
		if want := defSpec.Path + "x"; def.Name != want {
			t.Errorf("got def name %q, want %q", def.Name, want)
		}
	}

	// Search by def name (hits the index).
	for _, defSpec := range sampleDefs {
		query := defSpec.Path + "x" // "x" suffix prevents undesired prefix matches
		defs, err := cl.Defs.List(ctx, &sourcegraph.DefListOptions{
			RepoRevs: []string{repo + "@" + commitID},
			Query:    query,
		})
		if err != nil {
			t.Errorf("failed to query defs for %q: %s", query, err)
			continue
		}
		if len(defs.Defs) != 1 {
			t.Errorf("query %q: got len(defs) == %d, want 1", query, len(defs.Defs))
			continue
		}

		def := defs.Defs[0]

		// Not important to test; omitting this lets us not hardcode
		// it in sampleDefs, which looks nicer.
		def.CommitID = ""
		defSpec.CommitID = ""
		for _, def := range defs.Defs {
			def.CommitID = ""
		}

		if name, want := def.Name, defSpec.Path+"x"; name != want {
			t.Errorf("got def name %q, want %q", name, want)
		}
		if !reflect.DeepEqual(def.DefSpec(), defSpec) {
			t.Errorf("got def %#v, want %#v", def.DefSpec(), defSpec)
		}
	}
}
