package background

import (
	"testing"

	"github.com/stretchr/testify/require"

	gitprotocol "github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/search/commit"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/textsearch"
)

func TestHashArgs(t *testing.T) {
	t.Parallel()

	t.Run("same everything", func(t *testing.T) {
		args1 := &gitprotocol.SearchRequest{
			Repo:        "a",
			Revisions:   []gitprotocol.RevisionSpecifier{{RevSpec: "b"}},
			Query:       &gitprotocol.AuthorMatches{Expr: "camden"},
			IncludeDiff: true,
		}
		args2 := &gitprotocol.SearchRequest{
			Repo:        "a",
			Revisions:   []gitprotocol.RevisionSpecifier{{RevSpec: "b"}},
			Query:       &gitprotocol.AuthorMatches{Expr: "camden"},
			IncludeDiff: true,
		}
		require.Equal(t, hashArgs(args1), hashArgs(args2))
	})

	// Requests that search different things should
	// not have the same hash.

	t.Run("different repos", func(t *testing.T) {
		args1 := &gitprotocol.SearchRequest{Repo: "a", Query: &gitprotocol.AuthorMatches{Expr: "camden"}}
		args2 := &gitprotocol.SearchRequest{Repo: "b", Query: &gitprotocol.AuthorMatches{Expr: "camden"}}
		require.NotEqual(t, hashArgs(args1), hashArgs(args2))
	})

	t.Run("different revisions", func(t *testing.T) {
		args1 := &gitprotocol.SearchRequest{
			Revisions: []gitprotocol.RevisionSpecifier{{RevSpec: "a"}, {RefGlob: "b"}},
			Query:     &gitprotocol.AuthorMatches{Expr: "a"},
		}
		args2 := &gitprotocol.SearchRequest{
			Revisions: []gitprotocol.RevisionSpecifier{{RevSpec: "a"}, {ExcludeRefGlob: "b"}},
			Query:     &gitprotocol.AuthorMatches{Expr: "a"},
		}
		require.NotEqual(t, hashArgs(args1), hashArgs(args2))
	})

	t.Run("different queries", func(t *testing.T) {
		args1 := &gitprotocol.SearchRequest{Query: &gitprotocol.AuthorMatches{Expr: "a"}}
		args2 := &gitprotocol.SearchRequest{Query: &gitprotocol.AuthorMatches{Expr: "b"}}
		require.NotEqual(t, hashArgs(args1), hashArgs(args2))
	})
}

func TestAddCodeMonitorHook(t *testing.T) {
	t.Parallel()

	t.Run("errors on non-commit search", func(t *testing.T) {
		erroringJobs := []job.Job{
			job.NewParallelJob(&run.RepoSearch{}, &commit.CommitSearch{}),
			&run.RepoSearch{},
			job.NewAndJob(&textsearch.RepoUniverseTextSearch{}, &commit.CommitSearch{}),
			job.NewTimeoutJob(0, &run.RepoSearch{}),
		}

		for _, j := range erroringJobs {
			t.Run("", func(t *testing.T) {
				_, err := addCodeMonitorHook(j, 0)
				require.Error(t, err)
			})
		}
	})

	t.Run("no errors on only commit search", func(t *testing.T) {
		nonErroringJobs := []job.Job{
			job.NewParallelJob(&commit.CommitSearch{}, &commit.CommitSearch{}),
			job.NewAndJob(&commit.CommitSearch{}, &commit.CommitSearch{}),
			&commit.CommitSearch{},
			job.NewTimeoutJob(0, &commit.CommitSearch{}),
		}

		for _, j := range nonErroringJobs {
			t.Run("", func(t *testing.T) {
				_, err := addCodeMonitorHook(j, 0)
				require.NoError(t, err)
			})
		}
	})
}

func TestCodeMonitorHook(t *testing.T) {

}
