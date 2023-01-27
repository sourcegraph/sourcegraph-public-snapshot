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
	"github.com/sourcegraph/sourcegraph/internal/gitserver"

	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/proto"
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
			got, err := backend.NewOwnService(git).OwnersFile(context.Background(), "repo", "SHA")
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
	got, err := backend.NewOwnService(git).OwnersFile(context.Background(), "repo", "SHA")
	require.NoError(t, err)
	assert.Nil(t, got)
}
