// +build exectest,pgsqltest

package pgsql_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/testserver"
	"src.sourcegraph.com/sourcegraph/util/testutil"
)

// TestRepos_CreateStartsBuild_lg tests that creating a mirror repository
// properly enqueues a new build for that repo.
func TestRepos_CreateStartsBuild_lg(t *testing.T) {
	if testserver.Store != "pgsql" {
		t.Skip()
	}
	t.Parallel()

	// Start a fs-backed server to act as our repository host for mirroring.
	fsServer, fsCtx := testserver.NewUnstartedServerWithStore("fs")
	fsServer.Config.ServeFlags = append(fsServer.Config.ServeFlags,
		&authutil.Flags{Source: "none", AllowAnonymousReaders: true},
	)
	if err := fsServer.Start(); err != nil {
		t.Fatal(err)
	}
	defer fsServer.Close()

	// Create and push a repo to the host.
	commitID, done, err := testutil.CreateAndPushRepo(t, fsCtx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	// Start our primary pgsql-backed server.
	pgsqlServer, pgsqlCtx := testserver.NewUnstartedServer()
	pgsqlServer.Config.ServeFlags = append(pgsqlServer.Config.ServeFlags,
		&authutil.Flags{Source: "none", AllowAnonymousReaders: true},
	)
	if err := pgsqlServer.Start(); err != nil {
		t.Fatal(err)
	}
	defer pgsqlServer.Close()

	// Create a mirror repo against the fs-backed instance.
	repo := "myrepo/name"
	_, err = pgsqlServer.Client.Repos.Create(pgsqlCtx, &sourcegraph.ReposCreateOp{
		URI:      repo,
		VCS:      "git",
		CloneURL: fsServer.Config.Serve.AppURL + "/myrepo",
		Mirror:   true,
		Private:  true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Artificial delay for caching. This speeds up the test as Builds.List
	// performs a veryShortCache on the first lack-of-builds query which lasts for
	// 7s. By introducing a delay here just long enough for the build to complete
	// on average, we shave off 5s from the test run time.
	time.Sleep(2 * time.Second)

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
	if testserver.Store != "pgsql" {
		t.Skip()
	}
	t.Parallel()

	// Start a fs-backed server to act as our repository host for mirroring.
	fsServer, fsCtx := testserver.NewUnstartedServerWithStore("fs")
	fsServer.Config.ServeFlags = append(fsServer.Config.ServeFlags,
		&authutil.Flags{Source: "none", AllowAnonymousReaders: true},
	)
	if err := fsServer.Start(); err != nil {
		t.Fatal(err)
	}
	defer fsServer.Close()

	// Create and push a repo to the host.
	_, done, err := testutil.CreateAndPushRepo(t, fsCtx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	// Start our primary pgsql-backed server.
	pgsqlServer, pgsqlCtx := testserver.NewUnstartedServer()
	pgsqlServer.Config.ServeFlags = append(pgsqlServer.Config.ServeFlags,
		&authutil.Flags{Source: "none", AllowAnonymousReaders: true},
	)
	if err := pgsqlServer.Start(); err != nil {
		t.Fatal(err)
	}
	defer pgsqlServer.Close()

	// Create a mirror repo against the fs-backed instance.
	repo := "myrepo/name"
	_, err = pgsqlServer.Client.Repos.Create(pgsqlCtx, &sourcegraph.ReposCreateOp{
		URI:      repo,
		VCS:      "git",
		CloneURL: fsServer.Config.Serve.AppURL + "/myrepo",
		Mirror:   true,
		Private:  true,
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

	// Manually check $SGPATH/repos/myrepo/name for the directory and confirm it
	// was deleted.
	_, err = os.Stat(filepath.Join(pgsqlServer.Config.ServeFSFlags.ReposDir, repo))
	if !os.IsNotExist(err) {
		t.Fatal("Repos.Delete did not properly remove the repository directory")
	}
}
