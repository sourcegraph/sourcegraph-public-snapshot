package localstore_test

import (
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/testutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/testserver"
)

// TestRepos_CreateDeleteWorks_lg tests that creating and deleting a mirrored
// repository does remove the filesystem-stored git repository (which acts as
// a working directory for git ops).
func TestRepos_CreateDeleteWorks_lg(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

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
	repoObj, err := pgsqlServer.Client.Repos.Create(pgsqlCtx, &sourcegraph.ReposCreateOp{
		Op: &sourcegraph.ReposCreateOp_New{
			New: &sourcegraph.ReposCreateOp_NewRepo{
				URI:      repo,
				CloneURL: fsServer.Config.Serve.AppURL + "/myrepo",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Wait for the repo to be initialized.
	time.Sleep(2 * time.Second)

	// Delete the repo.
	_, err = pgsqlServer.Client.Repos.Delete(pgsqlCtx, &sourcegraph.RepoSpec{ID: repoObj.ID})
	if err != nil {
		t.Fatal(err)
	}
}
