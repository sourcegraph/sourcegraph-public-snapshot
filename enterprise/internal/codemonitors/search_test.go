package codemonitors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	gitprotocol "github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/search/commit"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/types"
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
		args1 := &gitprotocol.SearchRequest{Repo: "a"}
		args2 := &gitprotocol.SearchRequest{Repo: "b"}
		require.NotEqual(t, hashArgs(args1), hashArgs(args2))
	})

	t.Run("different revisions", func(t *testing.T) {
		args1 := &gitprotocol.SearchRequest{
			Revisions: []gitprotocol.RevisionSpecifier{{RevSpec: "a"}, {RefGlob: "b"}},
		}
		args2 := &gitprotocol.SearchRequest{
			Revisions: []gitprotocol.RevisionSpecifier{{RevSpec: "a"}, {ExcludeRefGlob: "b"}},
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
			job.NewAndJob(&searcher.SymbolSearcher{}, &commit.CommitSearch{}),
			job.NewTimeoutJob(0, &run.RepoSearch{}),
		}

		for _, j := range erroringJobs {
			t.Run("", func(t *testing.T) {
				_, err := addCodeMonitorHook(j, nil)
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
				_, err := addCodeMonitorHook(j, nil)
				require.NoError(t, err)
			})
		}
	})
}

func TestCodeMonitorHook(t *testing.T) {
	t.Parallel()

	type testFixtures struct {
		User    *types.User
		Monitor *edb.Monitor
	}
	populateFixtures := func(db edb.EnterpriseDB) testFixtures {
		ctx := context.Background()
		u, err := db.Users().Create(ctx, database.NewUser{Email: "test", Username: "test", EmailVerificationCode: "test"})
		require.NoError(t, err)
		ctx = actor.WithActor(ctx, actor.FromUser(u.ID))
		m, err := db.CodeMonitors().CreateMonitor(ctx, edb.MonitorArgs{NamespaceUserID: &u.ID})
		require.NoError(t, err)
		return testFixtures{User: u, Monitor: m}
	}

	db := database.NewDB(dbtest.NewDB(t))
	fixtures := populateFixtures(edb.NewEnterpriseDB(db))
	ctx := context.Background()

	gs := gitserver.NewMockClient()
	gs.ResolveRevisionsFunc.PushReturn([]string{"hash1", "hash2"}, nil)
	gs.ResolveRevisionsFunc.PushReturn([]string{"hash3", "hash4"}, nil)

	// The first time, doSearch should receive only the resolved hashes
	doSearch := func(args *gitprotocol.SearchRequest) error {
		require.Equal(t, args.Revisions, []gitprotocol.RevisionSpecifier{{
			RevSpec: "hash1",
		}, {
			RevSpec: "hash2",
		}})
		return nil
	}
	err := hookWithID(ctx, db, gs, &gitprotocol.SearchRequest{}, doSearch, fixtures.Monitor.ID)
	require.NoError(t, err)

	// The next time, doSearch should receive the new resolved hashes plus the
	// hashes from last time, but excluded
	doSearch = func(args *gitprotocol.SearchRequest) error {
		require.Equal(t, args.Revisions, []gitprotocol.RevisionSpecifier{{
			RevSpec: "hash3",
		}, {
			RevSpec: "hash4",
		}, {
			RevSpec: "^hash1",
		}, {
			RevSpec: "^hash2",
		}})
		return nil
	}
	err = hookWithID(ctx, db, gs, &gitprotocol.SearchRequest{}, doSearch, fixtures.Monitor.ID)
	require.NoError(t, err)
}
