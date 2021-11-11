package graphqlbackend

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmock"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestGitTreeEntry_RawZipArchiveURL(t *testing.T) {
	db := dbmock.NewMockDB()
	got := NewGitTreeEntryResolver(db,
		&GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, &types.Repo{Name: "my/repo"}),
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

	git.Mocks.ReadFile = func(commit api.CommitID, name string) ([]byte, error) {
		if name != wantPath {
			t.Fatalf("wrong name in ReadFile call. want=%q, have=%q", wantPath, name)
		}
		return []byte(wantContent), nil
	}
	t.Cleanup(func() { git.Mocks.ReadFile = nil })

	db := dbmock.NewMockDB()
	gitTree := NewGitTreeEntryResolver(db,
		&GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, &types.Repo{Name: "my/repo"}),
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

func TestGitTreeEntry_Content_Sub_Repo_Filtering(t *testing.T) {
	wantPath := "foobar.md"
	wantContent := ""

	git.Mocks.ReadFile = func(commit api.CommitID, name string) ([]byte, error) {
		if name != wantPath {
			t.Fatalf("wrong name in ReadFile call. want=%q, have=%q", wantPath, name)
		}
		return []byte("something"), nil
	}
	t.Cleanup(func() { git.Mocks.ReadFile = nil })

	db := dbmock.NewMockDB()
	gitTree := NewGitTreeEntryResolver(db,
		NewGitCommitResolver(db, NewRepositoryResolver(db, &types.Repo{Name: "my/repo"}), "", nil),
		CreateFileInfo(wantPath, false))

	srp := authz.NewMockSubRepoPermissionChecker()
	srp.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
		for _, p := range []string{"foobar.md"} {
			if p == content.Path {
				return authz.None, nil
			}
		}
		return authz.Read, nil
	})
	srp.EnabledFunc.SetDefaultReturn(true)

	// Apply mocks to all resolvers
	gitTree.subRepoPerms = srp
	gitTree.commit.subRepoPerms = srp

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	newFileContent, err := gitTree.Content(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(newFileContent, wantContent); diff != "" {
		t.Fatalf("wrong newFileContent: %s", diff)
	}
}
