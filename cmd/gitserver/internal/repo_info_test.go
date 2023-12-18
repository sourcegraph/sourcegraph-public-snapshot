package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestHandleRepoDelete(t *testing.T) {
	testHandleRepoDelete(t, false)
}

func TestHandleRepoDeleteWhenDeleteInDB(t *testing.T) {
	// We also want to ensure that we can delete repo data on disk for a repo that
	// has already been deleted in the DB.
	testHandleRepoDelete(t, true)
}

func testHandleRepoDelete(t *testing.T, deletedInDB bool) {
	logger := logtest.Scoped(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	remote := t.TempDir()
	repoName := api.RepoName("example.com/foo/bar")
	db := database.NewDB(logger, dbtest.NewDB(t))

	dbRepo := &types.Repo{
		Name:        repoName,
		Description: "Test",
	}

	// Insert the repo into our database
	if err := db.Repos().Create(ctx, dbRepo); err != nil {
		t.Fatal(err)
	}

	repo := remote
	cmd := func(name string, arg ...string) string {
		t.Helper()
		return runCmd(t, repo, name, arg...)
	}
	_ = makeSingleCommitRepo(cmd)
	// Add a bad tag
	cmd("git", "tag", "HEAD")

	rr := httptest.NewRecorder()

	updateReq := protocol.RepoUpdateRequest{
		Repo: repoName,
	}
	body, err := json.Marshal(updateReq)
	if err != nil {
		t.Fatal(err)
	}

	// This will perform an initial clone
	req := newRequest("GET", "/repo-update", bytes.NewReader(body))
	s.handleRepoUpdate(rr, req)

	size := gitserverfs.DirSize(gitserverfs.RepoDirFromName(s.ReposDir, repoName).Path("."))
	want := &types.GitserverRepo{
		RepoID:        dbRepo.ID,
		ShardID:       "",
		CloneStatus:   types.CloneStatusCloned,
		RepoSizeBytes: size,
	}
	fromDB, err := db.GitserverRepos().GetByID(ctx, dbRepo.ID)
	if err != nil {
		t.Fatal(err)
	}

	// We don't expect an error
	if diff := cmp.Diff(want, fromDB, ignoreVolatileGitserverRepoFields); diff != "" {
		t.Fatal(diff)
	}

	if deletedInDB {
		if err := db.Repos().Delete(ctx, dbRepo.ID); err != nil {
			t.Fatal(err)
		}
		repos, err := db.Repos().List(ctx, database.ReposListOptions{IncludeDeleted: true, IDs: []api.RepoID{dbRepo.ID}})
		if err != nil {
			t.Fatal(err)
		}
		if len(repos) != 1 {
			t.Fatalf("Expected 1 repo, got %d", len(repos))
		}
		dbRepo = repos[0]
	}

	reposDir := t.TempDir()

	// Now we can delete it
	require.NoError(t, deleteRepo(ctx, logger, db, "test-gitserver", reposDir, dbRepo.Name))

	size = gitserverfs.DirSize(gitserverfs.RepoDirFromName(reposDir, repoName).Path("."))
	if size != 0 {
		t.Fatalf("Size should be 0, got %d", size)
	}

	// Check status in gitserver_repos
	want = &types.GitserverRepo{
		RepoID:        dbRepo.ID,
		ShardID:       "",
		CloneStatus:   types.CloneStatusNotCloned,
		RepoSizeBytes: size,
	}
	fromDB, err = db.GitserverRepos().GetByID(ctx, dbRepo.ID)
	if err != nil {
		t.Fatal(err)
	}

	// We don't expect an error
	if diff := cmp.Diff(want, fromDB, ignoreVolatileGitserverRepoFields); diff != "" {
		t.Fatal(diff)
	}
}
