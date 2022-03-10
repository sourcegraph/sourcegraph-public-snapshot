package background

import (
	"testing"

	"github.com/stretchr/testify/require"

	gitprotocol "github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
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
