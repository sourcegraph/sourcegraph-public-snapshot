package graphqlbackend

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmock"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestGitTreeEntry_RawZipArchiveURL(t *testing.T) {
	db := dbmock.NewMockDB()
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

	git.Mocks.ReadFile = func(commit api.CommitID, name string) ([]byte, error) {
		if name != wantPath {
			t.Fatalf("wrong name in ReadFile call. want=%q, have=%q", wantPath, name)
		}
		return []byte(wantContent), nil
	}
	t.Cleanup(func() { git.Mocks.ReadFile = nil })

	db := dbmock.NewMockDB()
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

func TestGitTreeEntry_Content_Sub_Repo_Filtering(t *testing.T) {
	wantPath := "foobar.md"
	wantContent := ""

	ctx := setupSubRepoDeny(context.Background(), t, []string{"foobar.md"})

	git.Mocks.ReadFile = func(commit api.CommitID, name string) ([]byte, error) {
		if name != wantPath {
			t.Fatalf("wrong name in ReadFile call. want=%q, have=%q", wantPath, name)
		}
		return []byte("something"), nil
	}
	t.Cleanup(func() { git.Mocks.ReadFile = nil })

	db := dbmock.NewMockDB()
	gitTree := &GitTreeEntryResolver{
		db: db,
		commit: &GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, &types.Repo{Name: "my/repo"}),
		},
		stat: CreateFileInfo(wantPath, false),
	}

	newFileContent, err := gitTree.Content(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(newFileContent, wantContent); diff != "" {
		t.Fatalf("wrong newFileContent: %s", diff)
	}
}

func setupSubRepoDeny(ctx context.Context, t *testing.T, denyPaths []string) context.Context {
	oldFn := subRepoPermsClient
	t.Cleanup(func() { subRepoPermsClient = oldFn })

	subRepoPermsClient = func(db dbutil.DB) authz.SubRepoPermissionChecker {
		m := authz.NewMockSubRepoPermissionChecker()
		m.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
			for _, p := range denyPaths {
				if p == content.Path {
					return authz.None, nil
				}
			}
			return authz.Read, nil
		})
		m.EnabledFunc.SetDefaultHook(func() bool {
			return true
		})
		return m
	}

	return actor.WithActor(ctx, &actor.Actor{
		UID: 1,
	})
}
