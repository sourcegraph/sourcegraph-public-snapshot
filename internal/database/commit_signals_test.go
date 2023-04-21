package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestOwnSignalStore_AddCommit(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
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
		commit.CommitSHA = fmt.Sprintf("sha%d", i)
		if err := store.AddCommit(ctx, commit); err != nil {
			t.Fatal(err)
		}
	}

	for p, w := range map[string][]RecentAuthor{
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
