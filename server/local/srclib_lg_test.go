// +build buildtest,exectest

package local_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/testserver"
	"src.sourcegraph.com/sourcegraph/util/testutil"
)

func TestSrclibPush(t *testing.T) {
	if testserver.Store == "pgsql" {
		t.Skip()
	}

	t.Parallel()

	a, ctx := testserver.NewUnstartedServer()
	a.Config.ServeFlags = append(a.Config.ServeFlags,
		&authutil.Flags{Source: "none", AllowAnonymousReaders: true},
	)
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	_, close, err := testutil.CreateAndPushRepo(t, ctx, "r/rr")
	if err != nil {
		t.Fatal(err)
	}
	defer close()

	repo, err := a.Client.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: "r/rr"})
	if err != nil {
		t.Fatal(err)
	}

	// Clone and build the repo locally.
	if err := cloneAndLocallyBuildRepo(t, a, repo, ""); err != nil {
		t.Fatal(err)
	}

	testutil.CheckImport(t, ctx, "r/rr", "")
}

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
	if err := testutil.CloneRepo(t, authedCloneURL, tmpDir, nil); err != nil {
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
