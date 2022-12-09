package graphqlbackend

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/authz"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestGitTreeEntry_RawZipArchiveURL(t *testing.T) {
	db := database.NewMockDB()
	gitserverClient := gitserver.NewMockClient()
	got := NewGitTreeEntryResolver(db, gitserverClient,
		&GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Name: "my/repo"}),
		},
		CreateFileInfo("a/b", true)).
		RawZipArchiveURL()
	want := "http://example.com/my/repo/-/raw/a/b?format=zip"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGitTreeEntry_Content(t *testing.T) {
	wantPath := "foobar.md"
	wantContent := "foobar"

	db := database.NewMockDB()
	gitserverClient := gitserver.NewMockClient()

	gitserverClient.ReadFileFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, _ api.CommitID, name string, _ authz.SubRepoPermissionChecker) ([]byte, error) {
		if name != wantPath {
			t.Fatalf("wrong name in ReadFile call. want=%q, have=%q", wantPath, name)
		}
		return []byte(wantContent), nil
	})

	gitTree := NewGitTreeEntryResolver(db, gitserverClient,
		&GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Name: "my/repo"}),
		},
		CreateFileInfo(wantPath, true))

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
