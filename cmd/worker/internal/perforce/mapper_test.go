package perforce

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestPerforceChangelistMapper_Handle(t *testing.T) {
	ctx := context.Background()

	gs := gitserver.NewMockClient()
	gs.GetDefaultBranchFunc.SetDefaultReturn("refs/heads/main", "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", nil)
	commits := []*gitdomain.Commit{}
	for range 100 {
		commits = append(commits, &gitdomain.Commit{
			ID:      api.CommitID(fmt.Sprintf("deadbeefdeadbeefdeadbeefdeadbe%02d", len(commits))),
			Message: gitdomain.Message(p4FusionCommitMessage("Commitmessage", len(commits))),
		})
	}
	gs.CommitsFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, opts gitserver.CommitsOptions) ([]*gitdomain.Commit, error) {
		return commits[opts.Skip:min(opts.N+opts.Skip, uint(len(commits)))], nil
	})

	logger := logtest.NoOp(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	r := &types.Repo{
		Name: api.RepoName("sourcegraph/sourcegraph"),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "sourcegraph/sourcegraph",
			ServiceType: extsvc.TypePerforce,
		},
	}
	require.NoError(t, db.Repos().Create(ctx, r))

	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				PerforceChangelistMapping: "enabled",
			},
		},
	})
	t.Cleanup(func() {
		conf.Mock(nil)
	})
	m := &perforceChangelistMapper{cfg: &Config{RepositoryBatchSize: 1}, db: db, logger: logger, gs: gs}

	err := m.Handle(ctx)
	require.NoError(t, err)

	for i := 1; i < 100; i++ {
		c, err := db.RepoCommitsChangelists().GetRepoCommitChangelist(ctx, r.ID, int64(i))
		require.NoError(t, err)
		require.Equal(t, dbutil.CommitBytea(fmt.Sprintf("deadbeefdeadbeefdeadbeefdeadbe%02d", i)), c.CommitSHA)
		require.Equal(t, int64(i), c.PerforceChangelistID)
	}
}

func TestParseChangelistID(t *testing.T) {
	t.Run("passes valid perforce commit", func(t *testing.T) {
		testCases := []string{
			p4FusionCommitMessage("test abc\n\nsomething something", 123),
			// Empty
			p4FusionCommitMessage("", 123),
		}

		for _, tc := range testCases {
			got, err := parseChangelistID(tc)
			require.NoError(t, err)
			require.Equal(t, int64(123), got)
		}
	})

	t.Run("fails invalid message", func(t *testing.T) {
		testCases := []string{
			p4FusionCommitMessage("", -123),
		}

		for _, tc := range testCases {
			_, err := parseChangelistID(tc)
			require.Error(t, err)
		}
	})
}

func p4FusionCommitMessage(message string, changelist int) string {
	return fmt.Sprintf(`%s

[p4-fusion: depot-paths = "//test-perms/": change = %d]`, message, changelist)
}

func TestExtractChangelistsFromCommits(t *testing.T) {
	ctx := context.Background()
	repo := api.RepoName("test/repo")
	headSHA := api.CommitID("abcdef")

	t.Run("no commits", func(t *testing.T) {
		gs := gitserver.NewMockClient()
		gs.CommitsFunc.SetDefaultReturn(nil, nil)

		changelists, err := extractChangelistsFromCommits(ctx, gs, repo, "", headSHA)
		require.NoError(t, err)
		require.Empty(t, changelists)
	})

	t.Run("valid changelists", func(t *testing.T) {
		gs := gitserver.NewMockClient()
		gs.CommitsFunc.SetDefaultReturn([]*gitdomain.Commit{
			{ID: "abc123", Message: gitdomain.Message(p4FusionCommitMessage("commit 1", 1234))},
			{ID: "def456", Message: gitdomain.Message(p4FusionCommitMessage("commit 2", 5678))},
		}, nil)

		changelists, err := extractChangelistsFromCommits(ctx, gs, repo, "", headSHA)
		require.NoError(t, err)
		require.Equal(t, []types.PerforceChangelist{
			{CommitSHA: "abc123", ChangelistID: 1234},
			{CommitSHA: "def456", ChangelistID: 5678},
		}, changelists)
	})

	t.Run("with last mapped commit", func(t *testing.T) {
		gs := gitserver.NewMockClient()
		gs.CommitsFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, opt gitserver.CommitsOptions) ([]*gitdomain.Commit, error) {
			require.Equal(t, []string{"abc123..abcdef"}, opt.Ranges)
			return []*gitdomain.Commit{
				{ID: "def456", Message: gitdomain.Message(p4FusionCommitMessage("commit 2", 5678))},
			}, nil
		})

		changelists, err := extractChangelistsFromCommits(ctx, gs, repo, "abc123", headSHA)
		require.NoError(t, err)
		require.Equal(t, []types.PerforceChangelist{
			{CommitSHA: "def456", ChangelistID: 5678},
		}, changelists)
	})
}
