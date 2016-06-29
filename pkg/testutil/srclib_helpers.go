package testutil

import (
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
func CheckImport(t *testing.T, ctx context.Context, repo int32, repoPath string, commitID string) {
	cl, _ := sourcegraph.NewClientFromContext(ctx)

	if len(commitID) != 40 {
		res, err := cl.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
			Repo: repo,
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
}
