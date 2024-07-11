package gitcli

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGitCLIBackend_CommitLog(t *testing.T) {
	// TODO:
	// Add tests for:
	// - AllRefs
	// - FollowOnlyFirstParent
	// - Order
	// - FollowPathRenames

	ctx := context.Background()

	// Prepare repo state:
	backend := BackendWithRepoCommands(t,
		"echo Hello > f",
		"git add f",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='Foo Author <foo@sourcegraph.com>' --date 2006-01-02T15:04:05Z",
		"git tag testbase",
		"echo World > f2",
		"git add f2",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit -m bar --author='Bar Author <bar@sourcegraph.com>' --date 2006-01-02T15:04:06Z",
	)

	allGitCommits := []*git.GitCommitWithFiles{
		{
			Commit: &gitdomain.Commit{
				ID:        "2b2289762392764ed127587b0d5fd88a2f16b7c1",
				Author:    gitdomain.Signature{Name: "Bar Author", Email: "bar@sourcegraph.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
				Message:   "bar",
				Parents:   []api.CommitID{"5fab3adc1e398e749749271d14ab843759b192cf"},
			},
		},
		{
			Commit: &gitdomain.Commit{
				ID:        "5fab3adc1e398e749749271d14ab843759b192cf",
				Author:    gitdomain.Signature{Name: "Foo Author", Email: "foo@sourcegraph.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Message:   "foo",
				Parents:   nil,
			},
		},
	}

	readCommitIterator := func(t *testing.T, it git.CommitLogIterator) []*git.GitCommitWithFiles {
		cs := []*git.GitCommitWithFiles{}
		for {
			c, err := it.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				require.NoError(t, err)
			}
			cs = append(cs, c)
		}
		return cs
	}

	t.Run("empty repo", func(t *testing.T) {
		backend := BackendWithRepoCommands(t)
		it, err := backend.CommitLog(ctx, git.CommitLogOpts{
			MaxCommits: 2,
			Ranges:     []string{"HEAD"},
		})
		require.NoError(t, err)
		_, err = it.Next()
		require.Error(t, err)
		require.True(t, errors.Is(err, io.EOF))
		require.NoError(t, it.Close())
	})

	t.Run("commit doesn't exist", func(t *testing.T) {
		backend := BackendWithRepoCommands(t)
		it, err := backend.CommitLog(ctx, git.CommitLogOpts{
			MaxCommits: 1,
			Ranges:     []string{"deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"},
		})
		require.NoError(t, err)
		_, err = it.Next()
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
		_ = it.Close()
	})

	t.Run("log commits including root", func(t *testing.T) {
		it, err := backend.CommitLog(ctx, git.CommitLogOpts{
			Ranges: []string{"HEAD"},
		})
		require.NoError(t, err)

		got := readCommitIterator(t, it)
		require.NoError(t, it.Close())

		require.Equal(t, allGitCommits, got)

		t.Run("commit range", func(t *testing.T) {
			it, err := backend.CommitLog(ctx, git.CommitLogOpts{
				Ranges: []string{"HEAD^...HEAD"},
			})
			require.NoError(t, err)

			got := readCommitIterator(t, it)
			require.NoError(t, it.Close())

			require.Equal(t, allGitCommits[:1], got)
		})

		t.Run("MaxCommits", func(t *testing.T) {
			it, err := backend.CommitLog(ctx, git.CommitLogOpts{
				Ranges:     []string{"HEAD"},
				MaxCommits: 1,
			})
			require.NoError(t, err)
			got := readCommitIterator(t, it)
			require.NoError(t, it.Close())
			require.Equal(t, allGitCommits[:1], got)
		})
		t.Run("Skip", func(t *testing.T) {
			it, err := backend.CommitLog(ctx, git.CommitLogOpts{
				Ranges: []string{"HEAD"},
				Skip:   1,
			})
			require.NoError(t, err)
			got := readCommitIterator(t, it)
			require.NoError(t, it.Close())
			require.Equal(t, allGitCommits[1:], got)

			it, err = backend.CommitLog(ctx, git.CommitLogOpts{
				Ranges: []string{"HEAD"},
				Skip:   2,
			})
			require.NoError(t, err)
			got = readCommitIterator(t, it)
			require.NoError(t, it.Close())
			require.Equal(t, []*git.GitCommitWithFiles{}, got)
		})
	})
	t.Run("before", func(t *testing.T) {
		it, err := backend.CommitLog(ctx, git.CommitLogOpts{
			Ranges: []string{"HEAD"},
			Before: mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z"),
		})
		require.NoError(t, err)
		got := readCommitIterator(t, it)
		require.NoError(t, it.Close())
		require.Equal(t, allGitCommits[1:], got)
	})
	t.Run("after", func(t *testing.T) {
		testCases := []struct {
			label       string
			commitDates []string
			after       string
			revspec     string
			want        bool
		}{
			{
				label: "after specific date",
				commitDates: []string{
					"2006-01-02T15:04:05Z",
					"2007-01-02T15:04:05Z",
					"2008-01-02T15:04:05Z",
				},
				after:   "2006-01-02T15:04:05Z",
				revspec: "master",
				want:    true,
			},
			{
				label: "after 1 year ago",
				commitDates: []string{
					"2016-01-02T15:04:05Z",
					"2017-01-02T15:04:05Z",
					"2017-01-02T15:04:06Z",
				},
				after:   "1 year ago",
				revspec: "master",
				want:    false,
			},
			{
				label: "after too recent date",
				commitDates: []string{
					"2006-01-02T15:04:05Z",
					"2007-01-02T15:04:05Z",
					"2008-01-02T15:04:05Z",
				},
				after:   "2010-01-02T15:04:05Z",
				revspec: "HEAD",
				want:    false,
			},
			{
				label: "commit 1 second after",
				commitDates: []string{
					"2006-01-02T15:04:05Z",
					"2007-01-02T15:04:05Z",
					"2007-01-02T15:04:06Z",
				},
				after:   "2007-01-02T15:04:05Z",
				revspec: "HEAD",
				want:    true,
			},
			{
				label: "after 10 years ago",
				commitDates: []string{
					"2016-01-02T15:04:05Z",
					"2017-01-02T15:04:05Z",
					"2017-01-02T15:04:06Z",
				},
				after:   "10 years ago",
				revspec: "HEAD",
				want:    true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.label, func(t *testing.T) {
				gitCommands := make([]string, len(tc.commitDates))
				for i, date := range tc.commitDates {
					gitCommands[i] = fmt.Sprintf("GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=%s git commit --allow-empty -m foo --author='a <a@a.com>'", date)
				}
				backend := BackendWithRepoCommands(t,
					gitCommands...,
				)
				after, err := gitdomain.ParseGitDate(tc.after, time.Now)
				require.NoError(t, err)
				it, err := backend.CommitLog(ctx, git.CommitLogOpts{
					MaxCommits: 2,
					Ranges:     []string{tc.revspec},
					After:      after,
				})
				require.NoError(t, err)

				got := readCommitIterator(t, it)

				require.True(t, len(got) > 0 == tc.want)
			})
		}
	})
	t.Run("include modified files", func(t *testing.T) {
		it, err := backend.CommitLog(ctx, git.CommitLogOpts{
			Ranges:               []string{"HEAD"},
			IncludeModifiedFiles: true,
		})
		require.NoError(t, err)

		got := readCommitIterator(t, it)
		require.NoError(t, it.Close())

		want := []*git.GitCommitWithFiles{
			{
				Commit:        allGitCommits[0].Commit,
				ModifiedFiles: []string{"f2"},
			},
			{
				Commit:        allGitCommits[1].Commit,
				ModifiedFiles: []string{"f"},
			},
		}

		require.Equal(t, want, got)
	})
	t.Run("for path", func(t *testing.T) {
		it, err := backend.CommitLog(ctx, git.CommitLogOpts{
			Ranges: []string{"HEAD"},
			Path:   "f",
		})
		require.NoError(t, err)

		got := readCommitIterator(t, it)
		require.NoError(t, it.Close())

		require.Equal(t, allGitCommits[1:], got)
		t.Run("unknownpath", func(t *testing.T) {
			it, err := backend.CommitLog(ctx, git.CommitLogOpts{
				Ranges: []string{"HEAD"},
				Path:   "notfound",
			})
			require.NoError(t, err)

			got := readCommitIterator(t, it)
			require.NoError(t, it.Close())

			require.Empty(t, got)
		})
		t.Run("non-utf-8", func(t *testing.T) {
			it, err := backend.CommitLog(ctx, git.CommitLogOpts{
				Ranges: []string{"HEAD"},
				Path:   "a\xc0rn",
			})
			require.NoError(t, err)

			got := readCommitIterator(t, it)
			require.NoError(t, it.Close())

			require.Empty(t, got)
		})
	})
	t.Run("message query", func(t *testing.T) {
		it, err := backend.CommitLog(ctx, git.CommitLogOpts{
			Ranges:       []string{"HEAD"},
			MessageQuery: "foo",
		})
		require.NoError(t, err)

		got := readCommitIterator(t, it)
		require.NoError(t, it.Close())

		require.Equal(t, allGitCommits[1:], got)
	})
	t.Run("author query", func(t *testing.T) {
		it, err := backend.CommitLog(ctx, git.CommitLogOpts{
			Ranges:      []string{"HEAD"},
			AuthorQuery: "Foo Author",
		})
		require.NoError(t, err)

		got := readCommitIterator(t, it)
		require.NoError(t, it.Close())

		require.Equal(t, allGitCommits[1:], got)
	})
	t.Run("not found range", func(t *testing.T) {
		// Prepare repo state:
		backend := BackendWithRepoCommands(t,
			"echo line1 > f",
			"git add f",
			"git commit -m foo --author='Foo Author <foo@sourcegraph.com>'",
			"git tag test",
		)

		it, err := backend.CommitLog(ctx, git.CommitLogOpts{
			Ranges: []string{"test", "HEAD", "not found"},
		})
		require.NoError(t, err)
		_, err = it.Next()
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))

		// Verify ordering doesn't matter and we return an error for any missing range:
		it, err = backend.CommitLog(ctx, git.CommitLogOpts{
			Ranges: []string{"not found", "test", "HEAD"},
		})
		require.NoError(t, err)
		_, err = it.Next()
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))

		// Bad commit in range:
		it, err = backend.CommitLog(ctx, git.CommitLogOpts{
			Ranges: []string{"HEAD..deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"},
		})
		require.NoError(t, err)
		_, err = it.Next()
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))

		// Bad commit in range LHS:
		it, err = backend.CommitLog(ctx, git.CommitLogOpts{
			Ranges: []string{"deadbeefdeadbeefdeadbeefdeadbeefdeadbeef..HEAD"},
		})
		require.NoError(t, err)
		_, err = it.Next()
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))

		// Bad ref in range:
		it, err = backend.CommitLog(ctx, git.CommitLogOpts{
			Ranges: []string{"HEAD..unknownbranch"},
		})
		require.NoError(t, err)
		_, err = it.Next()
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))

		// Unknown SHA:
		it, err = backend.CommitLog(ctx, git.CommitLogOpts{
			Ranges: []string{"deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"},
		})
		require.NoError(t, err)
		_, err = it.Next()
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
	})
	// Verify that if the context is canceled, the iterator returns an error.
	t.Run("context cancelation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		t.Cleanup(cancel)

		it, err := backend.CommitLog(ctx, git.CommitLogOpts{AllRefs: true})
		require.NoError(t, err)

		cancel()

		_, err = it.Next()
		require.Error(t, err)
		require.True(t, errors.Is(err, context.Canceled), "unexpected error: %v", err)

		require.True(t, errors.Is(it.Close(), context.Canceled), "unexpected error: %v", err)
	})
}
