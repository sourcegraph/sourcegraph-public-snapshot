// +build exectest,buildtest

package sgx_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"strconv"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/testserver"
	"src.sourcegraph.com/sourcegraph/util/testutil"
)

func cloneAndLocallyBuildRepo(t *testing.T, a *testserver.Server, repo *sourcegraph.Repo, asUser string) (err error) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// Add auth to HTTP clone URL so that `git clone`, `git push`,
	// etc., commands are authenticated.
	authedCloneURL, err := testutil.AddSystemAuthToURL(a.Ctx, repo.HTTPCloneURL)
	if err != nil {
		return err
	}

	// Clone the repo locally.
	if err := testutil.CloneRepo(t, authedCloneURL, tmpDir); err != nil {
		t.Fatal(err)
	}
	repoDir := filepath.Join(tmpDir, repo.Name)

	// Fix up the remote so the URI is "r/r". (HACK)
	cmd := exec.Command("git", "remote", "add", "srclib", "http://"+repo.URI)
	if err := testutil.RunCmd(cmd, repoDir); err != nil {
		return err
	}

	// Build the repo.
	cmd = a.SrclibCmd(nil, []string{"config"})
	if err := testutil.RunCmd(cmd, repoDir); err != nil {
		return err
	}
	cmd = a.SrclibCmd(nil, []string{"make"})
	if err := testutil.RunCmd(cmd, repoDir); err != nil {
		return err
	}

	// Push the repo.
	if asUser != "" {
		cmd, err = a.CmdAs(asUser, []string{"push"})
		if err != nil {
			return err
		}
	} else {
		cmd, err = a.CmdAsSystem([]string{"push"})
		if err != nil {
			return err
		}
	}
	if err := testutil.RunCmd(cmd, repoDir); err != nil {
		return err
	}

	return nil
}

// checkImport checks whether the defs produced by the srclib-sample
// toolchain have been imported. It is hardcoded dependent on the
// vendored srclib-sample toolchain in
// util/testutil/testdata/srclibpath.
//
// It's important to call checkImport because it ensures that both the
// per-source-unit import step and the index step run correctly. If
// you just checked that specific defs exist, then the index creation
// could fail (or have a bug) without you realizing it.
func checkImport(t *testing.T, ctx context.Context, cl *sourcegraph.Client, repo string) {
	const n = 6 // hardcoded (must be same as the number of units srclib-sample produces)
	var sampleDefs []sourcegraph.DefSpec
	for i := 0; i < n; i++ {
		sampleDefs = append(sampleDefs, sourcegraph.DefSpec{Repo: repo, UnitType: "sample", Unit: "myunit" + strconv.Itoa(i), Path: "mydef" + strconv.Itoa(i)})
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
			RepoRevs: []string{repo},
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

		if name, want := def.Name, defSpec.Path+"x"; name != want {
			t.Errorf("got def name %q, want %q", name, want)
		}
		if !reflect.DeepEqual(def.DefSpec(), defSpec) {
			t.Errorf("got def %#v, want %#v", def.DefSpec(), defSpec)
		}
	}
}
