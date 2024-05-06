package gitcli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGitCLIBackend_ReadFile(t *testing.T) {
	ctx := context.Background()

	// Prepare repo state:
	backend := BackendWithRepoCommands(t,
		// simple file
		"echo abcd > file1",
		"git add file1",
		"git commit -m commit --author='Foo Author <foo@sourcegraph.com>'",

		// test we handle file names with .. (git show by default interprets
		// this). Ensure past the .. exists as a branch. Then if we use git
		// show it would return a diff instead of file contents.
		"mkdir subdir",
		"echo old > subdir/name",
		"echo old > subdir/name..dev",
		"git add subdir",
		"git commit -m commit --author='Foo Author <foo@sourcegraph.com>'",
		"echo dotdot > subdir/name..dev",
		"git add subdir",
		"git commit -m commit --author='Foo Author <foo@sourcegraph.com>'",
		"git branch dev",
	)

	commitID, err := backend.RevParseHead(ctx)
	require.NoError(t, err)

	t.Run("read simple file", func(t *testing.T) {
		r, err := backend.ReadFile(ctx, commitID, "file1")
		require.NoError(t, err)
		t.Cleanup(func() { r.Close() })
		contents, err := io.ReadAll(r)
		require.NoError(t, err)
		require.Equal(t, "abcd\n", string(contents))
	})

	t.Run("non existent file", func(t *testing.T) {
		_, err := backend.ReadFile(ctx, commitID, "filexyz")
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("non existent commit", func(t *testing.T) {
		_, err := backend.ReadFile(ctx, "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", "file1")
		require.Error(t, err)
		require.True(t, errors.HasType(err, &gitdomain.RevisionNotFoundError{}))
	})

	t.Run("special file paths", func(t *testing.T) {
		// File with .. in path name:
		{
			r, err := backend.ReadFile(ctx, commitID, "subdir/name..dev")
			require.NoError(t, err)
			t.Cleanup(func() { r.Close() })
			contents, err := io.ReadAll(r)
			require.NoError(t, err)
			require.Equal(t, "dotdot\n", string(contents))
		}
		// File with .. in path name that doesn't exist:
		{
			_, err := backend.ReadFile(ctx, commitID, "subdir/404..dev")
			require.Error(t, err)
			require.True(t, os.IsNotExist(err))
		}
		// This test case ensures we do not return a log with diff for the
		// specially crafted "git show HASH:..branch". IE a way to bypass
		// sub-repo permissions.
		{
			_, err := backend.ReadFile(ctx, commitID, "..dev")
			require.Error(t, err)
			require.True(t, os.IsNotExist(err))
		}

		// 3 dots ... as a prefix when using git show will return an error like
		// error: object b5462a7c880ce339ba3f93ac343706c0fa35babc is a tree, not a commit
		// fatal: Invalid symmetric difference expression 269e2b9bda9a95ad4181a7a6eb2058645d9bad82:...dev
		{
			_, err := backend.ReadFile(ctx, commitID, "...dev")
			require.Error(t, err)
			require.True(t, os.IsNotExist(err))
		}
	})

	t.Run("submodule", func(t *testing.T) {
		submodDir := RepoWithCommands(t,
			// simple file
			"echo abcd > file1",
			"git add file1",
			"git commit -m commit --author='Foo Author <foo@sourcegraph.com>'",
		)

		// Prepare repo state:
		backend := BackendWithRepoCommands(t,
			// simple file
			"echo abcd > file1",
			"git add file1",
			"git commit -m commit --author='Foo Author <foo@sourcegraph.com>'",

			// Add submodule
			"git -c protocol.file.allow=always submodule add "+filepath.ToSlash(string(submodDir))+" submod",
			"git commit -m 'add submodule' --author='Foo Author <foo@sourcegraph.com>'",
		)

		commitID, err := backend.RevParseHead(ctx)
		require.NoError(t, err)

		r, err := backend.ReadFile(ctx, commitID, "submod")
		require.NoError(t, err)
		t.Cleanup(func() { r.Close() })
		contents, err := io.ReadAll(r)
		require.NoError(t, err)
		// A submodule should read like an empty file for now.
		require.Equal(t, "", string(contents))
	})
}

func TestGitCLIBackend_ReadFile_GoroutineLeak(t *testing.T) {
	ctx := context.Background()

	// Prepare repo state:
	backend := BackendWithRepoCommands(t,
		// simple file
		"echo abcd > file1",
		"git add file1",
		"git commit -m commit --author='Foo Author <foo@sourcegraph.com>'",
	)

	commitID, err := backend.RevParseHead(ctx)
	require.NoError(t, err)

	routinesBefore := runtime.NumGoroutine()

	r, err := backend.ReadFile(ctx, commitID, "file1")
	require.NoError(t, err)

	// Read just a few bytes, but not enough to complete.
	buf := make([]byte, 2)
	n, err := r.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 2, n)

	// Don't complete reading all the output, instead, bail and close the reader.
	require.NoError(t, r.Close())

	time.Sleep(time.Millisecond)

	// Expect no leaked routines.
	routinesAfter := runtime.NumGoroutine()
	require.Equal(t, routinesBefore, routinesAfter)
}

func TestRepository_GetCommit(t *testing.T) {
	ctx := context.Background()

	// Prepare repo state:
	backend := BackendWithRepoCommands(t,
		// simple file
		"echo abcd > file1",
		"git add file1",
		"git commit -m commit --author='Foo Author <foo@sourcegraph.com>'",
		"echo efgh > file2",
		"git add file2",
		"git commit -m commit2 --author='Foo Author <foo@sourcegraph.com>'",
	)

	commitID, err := backend.RevParseHead(ctx)
	require.NoError(t, err)

	t.Run("non existent commit", func(t *testing.T) {
		_, err := backend.GetCommit(ctx, "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", false)
		require.Error(t, err)
		require.True(t, errors.HasType(err, &gitdomain.RevisionNotFoundError{}))
	})

	t.Run("read commit", func(t *testing.T) {
		c, err := backend.GetCommit(ctx, commitID, false)
		require.NoError(t, err)
		require.Equal(t, &git.GitCommitWithFiles{
			Commit: &gitdomain.Commit{
				ID:      commitID,
				Message: "commit2",
				Author: gitdomain.Signature{
					Name:  "Foo Author",
					Email: "foo@sourcegraph.com",
					Date:  c.Author.Date, // Hard to test
				},
				Committer: &gitdomain.Signature{
					Name:  "a",
					Email: "a@a.com",
					Date:  c.Committer.Date, // Hard to test
				},
				Parents: []api.CommitID{"405b565ed446e271bc1998a91dbf4fb50dbfabfe"},
			},
		}, c)

		c2, err := backend.GetCommit(ctx, c.Parents[0], false)
		require.NoError(t, err)
		require.Equal(t, &git.GitCommitWithFiles{
			Commit: &gitdomain.Commit{
				ID:      c.Parents[0],
				Message: "commit",
				Author: gitdomain.Signature{
					Name:  "Foo Author",
					Email: "foo@sourcegraph.com",
					Date:  c2.Author.Date, // Hard to test
				},
				Committer: &gitdomain.Signature{
					Name:  "a",
					Email: "a@a.com",
					Date:  c2.Committer.Date, // Hard to test
				},
				Parents: nil,
			},
		}, c2)
	})

	t.Run("include modified files", func(t *testing.T) {
		c, err := backend.GetCommit(ctx, commitID, true)
		require.NoError(t, err)
		require.Equal(t, &git.GitCommitWithFiles{
			Commit: &gitdomain.Commit{
				ID:      commitID,
				Message: "commit2",
				Author: gitdomain.Signature{
					Name:  "Foo Author",
					Email: "foo@sourcegraph.com",
					Date:  c.Author.Date, // Hard to test
				},
				Committer: &gitdomain.Signature{
					Name:  "a",
					Email: "a@a.com",
					Date:  c.Committer.Date, // Hard to test
				},
				Parents: []api.CommitID{"405b565ed446e271bc1998a91dbf4fb50dbfabfe"},
			},
			ModifiedFiles: []string{"file2"},
		}, c)
	})
}

func TestRepository_FirstEverCommit(t *testing.T) {
	testCases := []struct {
		commitDates []string
		want        string
	}{
		{
			commitDates: []string{
				"2006-01-02T15:04:05Z",
				"2007-01-02T15:04:05Z",
				"2008-01-02T15:04:05Z",
			},
			want: "2006-01-02T15:04:05Z",
		},
		{
			commitDates: []string{
				"2007-01-02T15:04:05Z", // Don't think this is possible, but if it is we still want the first commit (not strictly "oldest")
				"2006-01-02T15:04:05Z",
				"2007-01-02T15:04:06Z",
			},
			want: "2007-01-02T15:04:05Z",
		},
	}

	t.Run("basic", func(t *testing.T) {
		for _, tc := range testCases {
			ctx := context.Background()

			gitCommands := make([]string, len(tc.commitDates))
			for i, date := range tc.commitDates {
				gitCommands[i] = fmt.Sprintf("GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=%s git commit --allow-empty -m foo --author='a <a@a.com>'", date)
			}

			backend := BackendWithRepoCommands(
				t,
				gitCommands...,
			)

			id, err := backend.FirstEverCommit(ctx)
			if err != nil {
				t.Fatal(err)
			}

			commit, err := backend.GetCommit(ctx, id, false)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.want, commit.Committer.Date.Format(time.RFC3339)); diff != "" {
				t.Fatalf("unexpected commit date (-want +got):\n%s", diff)
			}
		}
	})

	// Added for awareness if this error message changes.
	// Insights skip over empty repos and check against this error type
	t.Run("empty repo", func(t *testing.T) {
		backend := BackendWithRepoCommands(
			t,
		)

		_, err := backend.FirstEverCommit(context.Background())

		var repoErr *gitdomain.RevisionNotFoundError
		if !errors.As(err, &repoErr) {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGitCLIBackend_GetBehindAhead(t *testing.T) {
	ctx := context.Background()

	// Prepare repo state:
	backend := BackendWithRepoCommands(t,
		// This is the commit graph we are creating
		//
		//         +-----> 3  -----> 4           (branch1)
		//         |
		//         |
		// 0 ----> 1  ----> 2 -----> 5  -----> 6  (master)
		//
		"echo abcd > file0",
		"git add file0",
		"git commit -m commit0 --author='Foo Author <foo@sourcegraph.com>'",

		"echo abcd > file1",
		"git add file1",
		"git commit -m commit1 --author='Foo Author <foo@sourcegraph.com>'",

		"git branch branch1",

		"echo efgh > file2",
		"git add file2",
		"git commit -m commit2 --author='Foo Author <foo@sourcegraph.com>'",

		"git checkout branch1",

		"echo ijkl > file3",
		"git add file3",
		"git commit -m commit3 --author='Foo Author <foo@sourcegraph.com>'",

		"echo ijkl > file4",
		"git add file4",
		"git commit -m commit4 --author='Foo Author <foo@sourcegraph.com>'",

		"git checkout master",

		"echo ijkl > file5",
		"git add file5",
		"git commit -m commit5 --author='Foo Author <foo@sourcegraph.com>'",

		"echo ijkl > file6",
		"git add file6",
		"git commit -m commit6 --author='Foo Author <foo@sourcegraph.com>'",
	)

	left := "branch1"
	right := "master"

	t.Run("valid branches", func(t *testing.T) {
		behindAhead, err := backend.BehindAhead(ctx, left, right)
		require.NoError(t, err)
		require.Equal(t, &gitdomain.BehindAhead{Behind: 2, Ahead: 3}, behindAhead)
	})

	t.Run("missing left branch", func(t *testing.T) {
		_, err := backend.BehindAhead(ctx, left, "")
		require.NoError(t, err) // Should compare to HEAD
	})

	t.Run("missing right branch", func(t *testing.T) {
		_, err := backend.BehindAhead(ctx, "", right)
		require.NoError(t, err) // Should compare to HEAD
	})

	t.Run("invalid left branch", func(t *testing.T) {
		_, err := backend.BehindAhead(ctx, "invalid-branch", right)
		require.Error(t, err)
		var e *gitdomain.RevisionNotFoundError
		require.True(t, errors.As(err, &e))
	})

	t.Run("invalid right branch", func(t *testing.T) {
		_, err := backend.BehindAhead(ctx, left, "invalid-branch")
		require.Error(t, err)
		var e *gitdomain.RevisionNotFoundError
		require.True(t, errors.As(err, &e))
	})

	t.Run("same branch", func(t *testing.T) {
		behindAhead, err := backend.BehindAhead(ctx, left, left)
		require.NoError(t, err)
		require.Equal(t, &gitdomain.BehindAhead{Behind: 0, Ahead: 0}, behindAhead)
	})

	t.Run("invalid object id", func(t *testing.T) {
		_, err := backend.BehindAhead(ctx, "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", right)
		require.Error(t, err)
		var e *gitdomain.RevisionNotFoundError
		require.True(t, errors.As(err, &e))
	})
}
