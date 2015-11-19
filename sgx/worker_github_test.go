// +build exectest,buildtest

package sgx_test

import (
	"testing"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/testserver"
	"src.sourcegraph.com/sourcegraph/util/testutil"
)

func TestBuildRepo_serverside_github_lg(t *testing.T) {
	t.Skip("flaky") // https://magnum.travis-ci.com/sourcegraph/sourcegraph/jobs/21544094

	t.Parallel()

	a, ctx := testserver.NewServer()
	defer a.Close()

	testutil.EnsureRepoExists(t, ctx, "github.com/sgtest/tiny-go-repo")

	// Build the repo.
	build, _, err := testutil.BuildRepoAndWait(t, ctx, "github.com/sgtest/tiny-go-repo", "3446340574e3601b322d30f29e0d8477193e23af")
	if err != nil {
		t.Fatal(err)
	}
	if !build.Success {
		t.Errorf("build #%d for %s@%s failed", build.Attempt, build.Repo, build.CommitID)
	}

	checkImport(t, ctx, a.Client, "github.com/sgtest/tiny-go-repo")
}

// func TestBuildRepo_serverside_private_github_lg(t *testing.T) {
// t.Parallel()

// 	t.Log("NOTE: if this test fails in the ssh cloning step, try killing all ssh processes that are connected to github and try again. see https://github.com/sourcegraph/go-vcs/issues/27.")

// 	const repoURI = "github.com/sgtest-alice99-admin-public/buildtest-private"

// 	a := testserver.NewServer()
// 	defer a.Close()

// 	// Add alice99 user.
// 	u, _, err := a.Client.Users.Get(cliCtx, &sourcegraph.UserSpec{Login: "alice99"})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Logf("created user %+v", u)

// 	createAndStoreGitHubAuthzToken(t, "alice99", aliceGitHubPasswd, a, []string{"repo"})
// 	alice99Client, err := a.NewSourcegraphClientAs("alice99")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	repoSpec := sourcegraph.RepoSpec{URI: repoURI}
// 	repo, _, err := alice99Client.Repos.Get(cliCtx, &repoSpec)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Logf("got repo %s", repo.URI)
// 	if _, err := alice99Client.Repos.UpdateSettings(repoSpec, sourcegraph.RepoSettings{UseSSHPrivateKey: github.Bool(true)}); err != nil {
// 		t.Fatal(err)
// 	}

// 	// Wait until VCS data is loaded.
// 	const commitID = "aa348f1b997603311442e247c4087a4d6d8d3c7e"
// 	if _, err := alice99Client.MirrorRepos.RefreshVCS(cliCtx, &repoSpec); err != nil {
// 		t.Fatal(err)
// 	}
// 	for start, max := time.Now(), 10*time.Second; ; {
// 		if time.Since(start) > max {
// 			t.Fatalf("VCS wait exceeded %s", max)
// 		}
// 		c, _, err := alice99Client.Repos.GetCommit(cliCtx, &sourcegraph.RepoRevSpec{RepoSpec: repoSpec, Rev: commitID, CommitID: commitID})
// 		if err == nil {
// 			t.Logf("commit is ready: %v %q", c.ID, c.Message)
// 			break
// 		}
// 	}

// 	// Build the repo.
// 	build, _, err := BuildRepoAndWait(t, alice99Client, repoURI, commitID)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !build.Success {
// 		t.Errorf("build #%d failed", build.BID)
// 	}

// 	def, _, err := alice99Client.Defs.Get(cliCtx, &sourcegraph.DefsGetOp{Def: sourcegraph.DefSpec{
// 		Repo:     repoURI,
// 		UnitType: "sample",
// 		Unit:     "myunit",
// 		Path:     "mydef",
// 	}, Opt: nil})

// 	if err != nil {
// 		t.Fatalf("failed to get def: %s", err)
// 	}
// 	t.Logf("got def: %s", def.Name)
// }

func TestBuildRepo_push_github_lg(t *testing.T) {
	t.Skip("flaky") // https://magnum.travis-ci.com/sourcegraph/sourcegraph/jobs/21564920

	t.Parallel()

	a, ctx := testserver.NewServer()
	defer a.Close()

	testutil.EnsureRepoExists(t, ctx, "github.com/alice99/buildtest-go")

	repo, err := a.Client.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: "github.com/alice99/buildtest-go"})
	if err != nil {
		t.Fatal(err)
	}

	// Clone and build the repo locally.
	if err := cloneAndLocallyBuildRepo(t, a, repo, ""); err != nil {
		t.Fatal(err)
	}

	checkImport(t, ctx, a.Client, "github.com/alice99/buildtest-go")
}
