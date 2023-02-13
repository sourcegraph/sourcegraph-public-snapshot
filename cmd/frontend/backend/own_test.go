package backend_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
)

type repoPath struct {
	Repo     api.RepoName
	CommitID api.CommitID
	Path     string
}

// repoFiles is a fake git client mapping a file
type repoFiles map[repoPath]string

func (fs repoFiles) ReadFile(_ context.Context, _ authz.SubRepoPermissionChecker, repoName api.RepoName, commitID api.CommitID, file string) ([]byte, error) {
	content, ok := fs[repoPath{Repo: repoName, CommitID: commitID, Path: file}]
	if !ok {
		return nil, os.ErrNotExist
	}
	return []byte(content), nil
}

func TestOwnersServesFilesAtVariousLocations(t *testing.T) {
	codeownersText := (&codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{
				Pattern: "README.md",
				Owner:   []*codeownerspb.Owner{{Email: "owner@example.com"}},
			},
		},
	}).Repr()
	for name, repo := range map[string]repoFiles{
		"top-level": {{"repo", "SHA", "CODEOWNERS"}: codeownersText},
		".github":   {{"repo", "SHA", ".github/CODEOWNERS"}: codeownersText},
		".gitlab":   {{"repo", "SHA", ".gitlab/CODEOWNERS"}: codeownersText},
	} {
		t.Run(name, func(t *testing.T) {
			git := gitserver.NewMockClient()
			git.ReadFileFunc.SetDefaultHook(repo.ReadFile)
			got, err := backend.NewOwnService(git, database.NewMockDB()).OwnersFile(context.Background(), "repo", "SHA")
			require.NoError(t, err)
			assert.Equal(t, codeownersText, got.Repr())
		})
	}
}

func TestOwnersCannotFindFile(t *testing.T) {
	codeownersFile := &codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{
				Pattern: "README.md",
				Owner:   []*codeownerspb.Owner{{Email: "owner@example.com"}},
			},
		},
	}
	repo := repoFiles{
		{"repo", "SHA", "notCODEOWNERS"}: codeownersFile.Repr(),
	}
	git := gitserver.NewMockClient()
	git.ReadFileFunc.SetDefaultHook(repo.ReadFile)
	got, err := backend.NewOwnService(git, database.NewMockDB()).OwnersFile(context.Background(), "repo", "SHA")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestResolveOwnersWithType(t *testing.T) {
	t.Run("no owners returns empty", func(t *testing.T) {
		git := gitserver.NewMockClient()
		got, err := backend.NewOwnService(git, database.NewMockDB()).ResolveOwnersWithType(context.Background(), nil)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
	t.Run("no user or team match returns unknown owner", func(t *testing.T) {
		git := gitserver.NewMockClient()
		mockUserStore := database.NewMockUserStore()
		mockTeamStore := database.NewMockTeamStore()
		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(mockUserStore)
		db.TeamsFunc.SetDefaultReturn(mockTeamStore)
		ownService := backend.NewOwnService(git, db)

		mockUserStore.GetByUsernameFunc.SetDefaultReturn(nil, database.MockUserNotFoundErr)
		mockTeamStore.GetTeamByNameFunc.SetDefaultReturn(nil, database.TeamNotFoundError{})
		owners := []*codeownerspb.Owner{
			{
				Handle: "unknown",
			},
		}

		got, err := ownService.ResolveOwnersWithType(context.Background(), owners)
		require.NoError(t, err)
		assert.Equal(t, []codeowners.ResolvedOwner{
			NewUnknownOwner("unknown", ""),
		}, got)
	})
	t.Run("user match from handle returns person owner", func(t *testing.T) {

	})
	t.Run("user match from email returns person owner", func(t *testing.T) {

	})
	t.Run("team match from handle returns team owner", func(t *testing.T) {

	})
	t.Run("team match from email returns team owner", func(t *testing.T) {

	})
	t.Run("mix of person and team match", func(t *testing.T) {

	})
	t.Run("makes use of cache", func(t *testing.T) {

	})
	t.Run("errors", func(t *testing.T) {

	})
}

func NewUnknownOwner(handle, email string) codeowners.ResolvedOwner {
	return codeowners.UnknownOwner{
		Handle: handle,
		Email:  email,
	}
}
