package graphqlbackend

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestGitTreeEntry_RawZipArchiveURL(t *testing.T) {
	db := new(dbtesting.MockDB)
	got := (&GitTreeEntryResolver{
		db: db,
		commit: &GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, &types.Repo{Name: "my/repo"}),
		},
		stat: CreateFileInfo("a/b", true),
	}).RawZipArchiveURL()
	want := "http://example.com/my/repo/-/raw/a/b?format=zip"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGitTreeEntry_Content(t *testing.T) {
	wantPath := "foobar.md"
	wantContent := "foobar"
	db := new(dbtesting.MockDB)

	git.Mocks.ReadFile = func(commit api.CommitID, name string) ([]byte, error) {
		if name != wantPath {
			t.Fatalf("wrong name in ReadFile call. want=%q, have=%q", wantPath, name)
		}
		return []byte(wantContent), nil
	}
	t.Cleanup(func() { git.Mocks.ReadFile = nil })

	database.Mocks.Repos.Get = func(ctx context.Context, repo api.RepoID) (*types.Repo, error) {
		return &types.Repo{
			ID:        1,
			Name:      "github.com/foo/bar",
			CreatedAt: time.Now(),
		}, nil
	}
	defer func() { database.Mocks.Repos = database.MockRepos{} }()

	gitTree := &GitTreeEntryResolver{
		db: db,
		commit: &GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, &types.Repo{Name: "my/repo"}),
		},
		stat: CreateFileInfo(wantPath, true),
	}

	newFileContent, err := gitTree.Content(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(newFileContent, wantContent); diff != "" {
		t.Fatalf("wrong newFileContent: %s", diff)
	}

	newByteSize, err := gitTree.ByteSize(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if have, want := newByteSize, int32(len([]byte(wantContent))); have != want {
		t.Fatalf("wrong file size, want=%d have=%d", want, have)
	}
}
