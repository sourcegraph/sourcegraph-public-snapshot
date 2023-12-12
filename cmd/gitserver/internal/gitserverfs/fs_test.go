package gitserverfs

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestIgnorePath(t *testing.T) {
	reposDir := "/data/repos"

	for _, tc := range []struct {
		path         string
		shouldIgnore bool
	}{
		{path: filepath.Join(reposDir, TempDirName), shouldIgnore: true},
		{path: filepath.Join(reposDir, P4HomeName), shouldIgnore: true},
		// Double check handling of trailing space
		{path: filepath.Join(reposDir, P4HomeName+"   "), shouldIgnore: true},
		{path: filepath.Join(reposDir, "sourcegraph/sourcegraph"), shouldIgnore: false},
	} {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tc.shouldIgnore, IgnorePath(reposDir, tc.path))
		})
	}
}

func TestRemoveRepoDirectory(t *testing.T) {
	logger := logtest.Scoped(t)
	root := t.TempDir()

	mkFiles(t, root,
		"github.com/foo/baz/.git/HEAD",
		"github.com/foo/survivor/.git/HEAD",
		"github.com/bam/bam/.git/HEAD",
		"example.com/repo/.git/HEAD",
	)

	// Set them up in the DB
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db := database.NewDB(logger, dbtest.NewDB(t))

	idMapping := make(map[api.RepoName]api.RepoID)

	// Set them all as cloned in the DB
	for _, r := range []string{
		"github.com/foo/baz",
		"github.com/foo/survivor",
		"github.com/bam/bam",
		"example.com/repo",
	} {
		repo := &types.Repo{
			Name: api.RepoName(r),
		}
		if err := db.Repos().Create(ctx, repo); err != nil {
			t.Fatal(err)
		}
		if err := db.GitserverRepos().Update(ctx, &types.GitserverRepo{
			RepoID:      repo.ID,
			ShardID:     "test",
			CloneStatus: types.CloneStatusCloned,
		}); err != nil {
			t.Fatal(err)
		}
		idMapping[repo.Name] = repo.ID
	}

	// Remove everything but github.com/foo/survivor
	for _, d := range []string{
		"github.com/foo/baz/.git",
		"github.com/bam/bam/.git",
		"example.com/repo/.git",
	} {
		if err := RemoveRepoDirectory(ctx, logger, db, "test-gitserver", root, common.GitDir(filepath.Join(root, d)), true); err != nil {
			t.Fatalf("failed to remove %s: %s", d, err)
		}
	}

	// Removing them a second time is safe
	for _, d := range []string{
		"github.com/foo/baz/.git",
		"github.com/bam/bam/.git",
		"example.com/repo/.git",
	} {
		if err := RemoveRepoDirectory(ctx, logger, db, "test-gitserver", root, common.GitDir(filepath.Join(root, d)), true); err != nil {
			t.Fatalf("failed to remove %s: %s", d, err)
		}
	}

	assertPaths(t, root,
		"github.com/foo/survivor/.git/HEAD",
		".tmp",
	)

	for _, tc := range []struct {
		name   api.RepoName
		status types.CloneStatus
	}{
		{"github.com/foo/baz", types.CloneStatusNotCloned},
		{"github.com/bam/bam", types.CloneStatusNotCloned},
		{"example.com/repo", types.CloneStatusNotCloned},
		{"github.com/foo/survivor", types.CloneStatusCloned},
	} {
		id, ok := idMapping[tc.name]
		if !ok {
			t.Fatal("id mapping not found")
		}
		r, err := db.GitserverRepos().GetByID(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if r.CloneStatus != tc.status {
			t.Errorf("Want %q, got %q for %q", tc.status, r.CloneStatus, tc.name)
		}
	}
}

func TestRemoveRepoDirectory_Empty(t *testing.T) {
	root := t.TempDir()

	mkFiles(t, root,
		"github.com/foo/baz/.git/HEAD",
	)
	db := dbmocks.NewMockDB()
	gr := dbmocks.NewMockGitserverRepoStore()
	db.GitserverReposFunc.SetDefaultReturn(gr)
	logger := logtest.Scoped(t)

	if err := RemoveRepoDirectory(context.Background(), logger, db, "test-gitserver", root, common.GitDir(filepath.Join(root, "github.com/foo/baz/.git")), true); err != nil {
		t.Fatal(err)
	}

	assertPaths(t, root,
		".tmp",
	)

	if len(gr.SetCloneStatusFunc.History()) == 0 {
		t.Fatal("expected gitserverRepos.SetLastError to be called, but wasn't")
	}
	require.Equal(t, gr.SetCloneStatusFunc.History()[0].Arg2, types.CloneStatusNotCloned)
}

func TestRemoveRepoDirectory_UpdateCloneStatus(t *testing.T) {
	logger := logtest.Scoped(t)

	db := database.NewDB(logger, dbtest.NewDB(t))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := &types.Repo{
		Name: api.RepoName("github.com/foo/baz/"),
	}
	if err := db.Repos().Create(ctx, repo); err != nil {
		t.Fatal(err)
	}

	if err := db.GitserverRepos().Update(ctx, &types.GitserverRepo{
		RepoID:      repo.ID,
		ShardID:     "test",
		CloneStatus: types.CloneStatusCloned,
	}); err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	mkFiles(t, root, "github.com/foo/baz/.git/HEAD")

	if err := RemoveRepoDirectory(ctx, logger, db, "test-gitserver", root, common.GitDir(filepath.Join(root, "github.com/foo/baz/.git")), false); err != nil {
		t.Fatal(err)
	}

	assertPaths(t, root, ".tmp")

	r, err := db.Repos().GetByName(ctx, repo.Name)
	if err != nil {
		t.Fatal(err)
	}

	gsRepo, err := db.GitserverRepos().GetByID(ctx, r.ID)
	if err != nil {
		t.Fatal(err)
	}

	if gsRepo.CloneStatus != types.CloneStatusCloned {
		t.Fatalf("Expected clone_status to be %s, but got %s", types.CloneStatusCloned, gsRepo.CloneStatus)
	}
}
