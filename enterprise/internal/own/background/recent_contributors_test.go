package background

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func Test_RecentContributorIndexFromGitserver(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Background()

	err := db.Repos().Create(ctx, &types.Repo{
		ID:   1,
		Name: "own/repo1",
	})
	require.NoError(t, err)

	commits := []fakeCommit{
		{
			name:         "alice",
			email:        "alice@example.com",
			changedFiles: []string{"file1.txt", "dir/file2.txt"},
		},
		{
			name:         "alice",
			email:        "alice@example.com",
			changedFiles: []string{"file1.txt", "dir/file3.txt"},
		},
		{
			name:         "alice",
			email:        "alice@example.com",
			changedFiles: []string{"file1.txt", "dir/file2.txt", "dir/subdir/file.txt"},
		},
		{
			name:         "bob",
			email:        "bob@example.com",
			changedFiles: []string{"file1.txt", "dir2/file2.txt", "dir2/subdir/file.txt"},
		},
	}

	client := gitserver.NewMockClient()
	client.CommitLogFunc.SetDefaultReturn(fakeCommitsToLog(commits), nil)

	indexer := newRecentContributorsIndexer(client, db, logger)
	err = indexer.indexRepo(ctx, api.RepoID(1))
	require.NoError(t, err)

	for p, w := range map[string][]database.RecentContributorSummary{
		"dir": {
			{
				AuthorName:        "alice",
				AuthorEmail:       "alice@example.com",
				ContributionCount: 4,
			},
		},
		"file1.txt": {
			{
				AuthorName:        "alice",
				AuthorEmail:       "alice@example.com",
				ContributionCount: 3,
			},
			{
				AuthorName:        "bob",
				AuthorEmail:       "bob@example.com",
				ContributionCount: 1,
			},
		},
		"": {
			{
				AuthorName:        "alice",
				AuthorEmail:       "alice@example.com",
				ContributionCount: 7,
			},
			{
				AuthorName:        "bob",
				AuthorEmail:       "bob@example.com",
				ContributionCount: 3,
			},
		},
	} {
		path := p
		want := w
		t.Run(path, func(t *testing.T) {
			got, err := db.RecentContributionSignals().FindRecentAuthors(ctx, 1, path)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, want, got)
		})
	}
}

func fakeCommitsToLog(commits []fakeCommit) (results []gitserver.CommitLog) {
	for i, commit := range commits {
		results = append(results, gitserver.CommitLog{
			AuthorEmail:  commit.email,
			AuthorName:   commit.name,
			Timestamp:    time.Now(),
			SHA:          gitSha(fmt.Sprintf("%d", i)),
			ChangedFiles: commit.changedFiles,
		})
	}
	return results
}

type fakeCommit struct {
	email        string
	name         string
	changedFiles []string
}

func gitSha(val string) string {
	writer := sha1.New()
	writer.Write([]byte(val))
	return hex.EncodeToString(writer.Sum(nil))
}
