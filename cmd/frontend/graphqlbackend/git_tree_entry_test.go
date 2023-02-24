package graphqlbackend

import (
	"context"
	"strings"
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
	opts := GitTreeEntryResolverOpts{
		commit: &GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Name: "my/repo"}),
		},
		stat: CreateFileInfo("a/b", true),
	}
	got := NewGitTreeEntryResolver(db, gitserverClient, opts).RawZipArchiveURL()
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

	gitserverClient.ReadFileFunc.SetDefaultHook(func(_ context.Context, _ authz.SubRepoPermissionChecker, _ api.RepoName, _ api.CommitID, name string) ([]byte, error) {
		if name != wantPath {
			t.Fatalf("wrong name in ReadFile call. want=%q, have=%q", wantPath, name)
		}
		return []byte(wantContent), nil
	})
	opts := GitTreeEntryResolverOpts{
		commit: &GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Name: "my/repo"}),
		},
		stat: CreateFileInfo(wantPath, true),
	}
	gitTree := NewGitTreeEntryResolver(db, gitserverClient, opts)

	newFileContent, err := gitTree.Content(context.Background(), &GitTreeContentPageArgs{})
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

func TestGitTreeEntry_ContentPagination(t *testing.T) {
	wantPath := "foobar.md"
	fullContent := `1
2
3
4
5
6`

	db := database.NewMockDB()
	gitserverClient := gitserver.NewMockClient()

	gitserverClient.ReadFileFunc.SetDefaultHook(func(_ context.Context, _ authz.SubRepoPermissionChecker, _ api.RepoName, _ api.CommitID, name string) ([]byte, error) {
		if name != wantPath {
			t.Fatalf("wrong name in ReadFile call. want=%q, have=%q", wantPath, name)
		}
		return []byte(fullContent), nil
	})

	tests := []struct {
		startLine   int32
		endLine     int32
		wantContent string
	}{
		{
			startLine:   2,
			endLine:     6,
			wantContent: "2\n3\n4\n5\n6",
		},
		{
			startLine:   0,
			endLine:     2,
			wantContent: "1\n2",
		},
		{
			startLine:   0,
			endLine:     0,
			wantContent: "",
		},
		{
			startLine:   6,
			endLine:     6,
			wantContent: "6",
		},
		{
			startLine:   -1,
			endLine:     -1,
			wantContent: fullContent,
		},
		{
			startLine:   7,
			endLine:     7,
			wantContent: "",
		},
		{
			startLine:   5,
			endLine:     2,
			wantContent: fullContent,
		},
	}

	for _, tc := range tests {
		opts := GitTreeEntryResolverOpts{
			commit: &GitCommitResolver{
				repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Name: "my/repo"}),
			},
			stat: CreateFileInfo(wantPath, true),
		}
		gitTree := NewGitTreeEntryResolver(db, gitserverClient, opts)

		newFileContent, err := gitTree.Content(context.Background(), &GitTreeContentPageArgs{
			StartLine: &tc.startLine,
			EndLine:   &tc.endLine,
		})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(newFileContent, tc.wantContent); diff != "" {
			t.Fatalf("wrong newFileContent: %s", diff)
		}

		newByteSize, err := gitTree.ByteSize(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		if have, want := newByteSize, int32(len([]byte(fullContent))); have != want {
			t.Fatalf("wrong file size, want=%d have=%d", want, have)
		}

		newTotalLines, err := gitTree.TotalLines(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		if have, want := newTotalLines, int32(len(strings.Split(fullContent, "\n"))); have != want {
			t.Fatalf("wrong file size, want=%d have=%d", want, have)
		}
	}

	// Testing default (nils) for pagination.
	opts := GitTreeEntryResolverOpts{
		commit: &GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Name: "my/repo"}),
		},
		stat: CreateFileInfo(wantPath, true),
	}
	gitTree := NewGitTreeEntryResolver(db, gitserverClient, opts)

	newFileContent, err := gitTree.Content(context.Background(), &GitTreeContentPageArgs{})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(newFileContent, fullContent); diff != "" {
		t.Fatalf("wrong newFileContent: %s", diff)
	}

	newByteSize, err := gitTree.ByteSize(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if have, want := newByteSize, int32(len([]byte(fullContent))); have != want {
		t.Fatalf("wrong file size, want=%d have=%d", want, have)
	}
}
