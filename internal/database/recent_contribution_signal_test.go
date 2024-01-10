package database

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestRecentContributionSignalStore(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	store := RecentContributionSignalStoreWith(db)

	ctx := context.Background()
	repo := mustCreate(ctx, t, db, &types.Repo{Name: "a/b"})

	for i, commit := range []Commit{
		{
			RepoID:       repo.ID,
			AuthorName:   "alice",
			AuthorEmail:  "alice@example.com",
			FilesChanged: []string{"file1.txt", "dir/file2.txt"},
		},
		{
			RepoID:       repo.ID,
			AuthorName:   "alice",
			AuthorEmail:  "alice@example.com",
			FilesChanged: []string{"file1.txt", "dir/file3.txt"},
		},
		{
			RepoID:       repo.ID,
			AuthorName:   "alice",
			AuthorEmail:  "alice@example.com",
			FilesChanged: []string{"file1.txt", "dir/file2.txt", "dir/subdir/file.txt"},
		},
		{
			RepoID:       repo.ID,
			AuthorName:   "bob",
			AuthorEmail:  "bob@example.com",
			FilesChanged: []string{"file1.txt", "dir2/file2.txt", "dir2/subdir/file.txt"},
		},
	} {
		commit.Timestamp = time.Now()
		commit.CommitSHA = gitSha(fmt.Sprintf("%d", i))
		if err := store.AddCommit(ctx, commit); err != nil {
			t.Fatal(err)
		}
	}

	for p, w := range map[string][]RecentContributorSummary{
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
			got, err := store.FindRecentAuthors(ctx, repo.ID, path)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, want, got)
		})
	}
}

func gitSha(val string) string {
	writer := sha1.New()
	writer.Write([]byte(val))
	return hex.EncodeToString(writer.Sum(nil))
}
