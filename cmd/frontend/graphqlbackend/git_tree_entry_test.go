package graphqlbackend

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestGitTreeEntry_RawZipArchiveURL(t *testing.T) {
	db := dbmocks.NewMockDB()
	gitserverClient := gitserver.NewMockClient()
	opts := GitTreeEntryResolverOpts{
		Commit: &GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Name: "my/repo"}),
		},
		Stat: CreateFileInfo("a/b", true),
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

	db := dbmocks.NewMockDB()
	gitserverClient := gitserver.NewMockClient()

	gitserverClient.NewFileReaderFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, ci api.CommitID, name string) (io.ReadCloser, error) {
		if name != wantPath {
			t.Fatalf("wrong name in ReadFile call. want=%q, have=%q", wantPath, name)
		}
		return io.NopCloser(bytes.NewReader([]byte(wantContent))), nil
	})
	opts := GitTreeEntryResolverOpts{
		Commit: &GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Name: "my/repo"}),
		},
		Stat: CreateFileInfo(wantPath, true),
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

	db := dbmocks.NewMockDB()
	gitserverClient := gitserver.NewMockClient()

	gitserverClient.NewFileReaderFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, ci api.CommitID, name string) (io.ReadCloser, error) {
		if name != wantPath {
			t.Fatalf("wrong name in ReadFile call. want=%q, have=%q", wantPath, name)
		}
		return io.NopCloser(bytes.NewReader(testContent)), nil
	})

	tests := []struct {
		startLine   int32
		endLine     int32
		wantContent string
	}{
		{
			startLine:   2,
			endLine:     6,
			wantContent: "2\n3\n4\n5\n6\n",
		},
		{
			startLine:   0,
			endLine:     2,
			wantContent: "1\n2\n",
		},
		{
			startLine:   0,
			endLine:     0,
			wantContent: string(testContent),
		},
		{
			startLine:   6,
			endLine:     6,
			wantContent: "6\n",
		},
		{
			startLine:   -1,
			endLine:     -1,
			wantContent: string(testContent),
		},
		{
			startLine:   7,
			endLine:     7,
			wantContent: "",
		},
		{
			startLine:   5,
			endLine:     2,
			wantContent: string(testContent),
		},
	}

	for i, tc := range tests {
		opts := GitTreeEntryResolverOpts{
			Commit: &GitCommitResolver{
				repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Name: "my/repo"}),
			},
			Stat: CreateFileInfo(wantPath, true),
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
			t.Fatalf("wrong newFileContent %d: %s", i, diff)
		}

		newByteSize, err := gitTree.ByteSize(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		if have, want := newByteSize, int32(len(testContent)); have != want {
			t.Fatalf("wrong file size, want=%d have=%d", want, have)
		}

		newTotalLines, err := gitTree.TotalLines(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		if have, want := newTotalLines, int32(lineCount(testContent)); have != want {
			t.Fatalf("wrong file size, want=%d have=%d", want, have)
		}
	}

	// Testing default (nils) for pagination.
	opts := GitTreeEntryResolverOpts{
		Commit: &GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Name: "my/repo"}),
		},
		Stat: CreateFileInfo(wantPath, true),
	}
	gitTree := NewGitTreeEntryResolver(db, gitserverClient, opts)

	newFileContent, err := gitTree.Content(context.Background(), &GitTreeContentPageArgs{})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(newFileContent, string(testContent)); diff != "" {
		t.Fatalf("wrong newFileContent: %s", diff)
	}

	newByteSize, err := gitTree.ByteSize(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if have, want := newByteSize, int32(len(testContent)); have != want {
		t.Fatalf("wrong file size, want=%d have=%d", want, have)
	}
}

var testContent = []byte(`1
2
3
4
5
6
`)

var testContentNoNewline = []byte(`1
2
3`)

func TestPageContent(t *testing.T) {
	t.Run("No pagination", func(t *testing.T) {
		have := pageContent(testContent, nil, nil)
		assert.Equal(t, string(testContent), string(have))
		have = pageContent(testContentNoNewline, nil, nil)
		assert.Equal(t, string(testContentNoNewline), string(have))
	})
	t.Run("Single line", func(t *testing.T) {
		// At beginning.
		have := pageContent(testContent, pointers.Ptr(1), pointers.Ptr(1))
		assert.Equal(t, "1\n", string(have))
		// In the middle.
		have = pageContent(testContent, pointers.Ptr(3), pointers.Ptr(3))
		assert.Equal(t, "3\n", string(have))
		// At the end.
		have = pageContent(testContent, pointers.Ptr(6), pointers.Ptr(6))
		assert.Equal(t, "6\n", string(have))
		// At the end without newline.
		have = pageContent(testContentNoNewline, pointers.Ptr(3), pointers.Ptr(3))
		assert.Equal(t, "3", string(have))
	})
	t.Run("Multi line", func(t *testing.T) {
		// At beginning.
		have := pageContent(testContent, pointers.Ptr(1), pointers.Ptr(2))
		assert.Equal(t, "1\n2\n", string(have))
		// In the middle.
		have = pageContent(testContent, pointers.Ptr(3), pointers.Ptr(4))
		assert.Equal(t, "3\n4\n", string(have))
		// At the end.
		have = pageContent(testContent, pointers.Ptr(5), pointers.Ptr(6))
		assert.Equal(t, "5\n6\n", string(have))
		// At the end without newline.
		have = pageContent(testContentNoNewline, pointers.Ptr(2), pointers.Ptr(3))
		assert.Equal(t, "2\n3", string(have))
	})
	t.Run("No startLine", func(t *testing.T) {
		// Whole file.
		have := pageContent(testContent, nil, pointers.Ptr(6))
		assert.Equal(t, string(testContent), string(have))
		// Subsection of the file.
		have = pageContent(testContent, nil, pointers.Ptr(3))
		assert.Equal(t, "1\n2\n3\n", string(have))
		// First line only.
		have = pageContent(testContent, nil, pointers.Ptr(1))
		assert.Equal(t, "1\n", string(have))
		// No final newline.
		have = pageContent(testContentNoNewline, nil, pointers.Ptr(3))
		assert.Equal(t, "1\n2\n3", string(have))
		// Invalid input.
		have = pageContent(testContent, nil, pointers.Ptr(0))
		assert.Equal(t, string(testContent), string(have))
		have = pageContent(testContent, nil, pointers.Ptr(10000))
		assert.Equal(t, string(testContent), string(have))
		have = pageContent(testContent, nil, pointers.Ptr(-1))
		assert.Equal(t, string(testContent), string(have))
	})
	t.Run("No endLine", func(t *testing.T) {
		// Whole file.
		have := pageContent(testContent, pointers.Ptr(1), nil)
		assert.Equal(t, string(testContent), string(have))
		// Subsection of the file.
		have = pageContent(testContent, pointers.Ptr(4), nil)
		assert.Equal(t, "4\n5\n6\n", string(have))
		// Last line only.
		have = pageContent(testContent, pointers.Ptr(6), nil)
		assert.Equal(t, "6\n", string(have))
		// No final newline.
		have = pageContent(testContentNoNewline, pointers.Ptr(1), nil)
		assert.Equal(t, "1\n2\n3", string(have))
		// Invalid input.
		have = pageContent(testContent, pointers.Ptr(0), nil)
		assert.Equal(t, string(testContent), string(have))
		have = pageContent(testContent, pointers.Ptr(10000), nil)
		assert.Equal(t, "", string(have))
		have = pageContent(testContent, pointers.Ptr(-1), nil)
		assert.Equal(t, string(testContent), string(have))
	})
	t.Run("Invalid range", func(t *testing.T) {
		// Out of bounds.
		have := pageContent(testContent, pointers.Ptr(1), pointers.Ptr(1000))
		assert.Equal(t, string(testContent), string(have))
		// Negative start.
		have = pageContent(testContent, pointers.Ptr(-1), pointers.Ptr(4))
		assert.Equal(t, "1\n2\n3\n4\n", string(have))
		// Negative end.
		have = pageContent(testContent, pointers.Ptr(1), pointers.Ptr(-4))
		assert.Equal(t, string(testContent), string(have))
		// End < start.
		have = pageContent(testContent, pointers.Ptr(4), pointers.Ptr(1))
		assert.Equal(t, string(testContent), string(have))
	})
}
