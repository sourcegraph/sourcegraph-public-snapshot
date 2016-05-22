// +build exectest,pgsqltest

package localstore_test

import (
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/server/testserver"
	"sourcegraph.com/sourcegraph/sourcegraph/util/testutil"
)

// TestRepos_CreateStartsBuild_lg tests that creating a mirror repository
// properly enqueues a new build for that repo.
func TestRepos_CreateStartsBuild_lg(t *testing.T) {
	t.Skip("flaky")
	t.Parallel()

	// Start a server to act as our repository host for mirroring.
	fsServer, fsCtx := testserver.NewUnstartedServer()
	if err := fsServer.Start(); err != nil {
		t.Fatal(err)
	}
	defer fsServer.Close()

	// Create and push a repo to the host.
	_, commitID, done, err := testutil.CreateAndPushRepo(t, fsCtx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	// Start our primary pgsql-backed server.
	pgsqlServer, pgsqlCtx := testserver.NewUnstartedServer()
	if err := pgsqlServer.Start(); err != nil {
		t.Fatal(err)
	}
	defer pgsqlServer.Close()

	// Create a mirror repo against the fs-backed instance.
	repo := "myrepo/name"
	_, err = pgsqlServer.Client.Repos.Create(pgsqlCtx, &sourcegraph.ReposCreateOp{
		Op: &sourcegraph.ReposCreateOp_New{
			New: &sourcegraph.ReposCreateOp_NewRepo{
				URI:      repo,
				CloneURL: fsServer.Config.Serve.AppURL + "/myrepo",
				Mirror:   true,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Wait for a build to succeeded for up to 10s.
	for i := 0; i < 10; i++ {
		builds, err := pgsqlServer.Client.Builds.List(pgsqlCtx, &sourcegraph.BuildListOptions{
			Succeeded:   true,
			Repo:        repo,
			CommitID:    commitID,
			ListOptions: sourcegraph.ListOptions{PerPage: 10},
		})
		if err != nil {
			t.Log(err)
		} else if len(builds.Builds) > 0 {
			return // Success!
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatal("timed out waiting for build to enqueue")
}

// TestRepos_CreateDeleteWorks_lg tests that creating and deleting a mirrored
// repository does remove the filesystem-stored git repository (which acts as
// a working directory for git ops).
func TestRepos_CreateDeleteWorks_lg(t *testing.T) {
	t.Parallel()

	// Start a server to act as our repository host for mirroring.
	fsServer, fsCtx := testserver.NewUnstartedServer()
	if err := fsServer.Start(); err != nil {
		t.Fatal(err)
	}
	defer fsServer.Close()

	// Create and push a repo to the host.
	_, _, done, err := testutil.CreateAndPushRepo(t, fsCtx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	// Start our primary pgsql-backed server.
	pgsqlServer, pgsqlCtx := testserver.NewUnstartedServer()
	if err := pgsqlServer.Start(); err != nil {
		t.Fatal(err)
	}
	defer pgsqlServer.Close()

	// Create a mirror repo against the fs-backed instance.
	repo := "myrepo/name"
	_, err = pgsqlServer.Client.Repos.Create(pgsqlCtx, &sourcegraph.ReposCreateOp{
		Op: &sourcegraph.ReposCreateOp_New{
			New: &sourcegraph.ReposCreateOp_NewRepo{
				URI:      repo,
				CloneURL: fsServer.Config.Serve.AppURL + "/myrepo",
				Mirror:   true,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Wait for the repo to be initialized.
	time.Sleep(2 * time.Second)

	// Delete the repo.
	_, err = pgsqlServer.Client.Repos.Delete(pgsqlCtx, &sourcegraph.RepoSpec{
		URI: repo,
	})
	if err != nil {
		t.Fatal(err)
	}
}
