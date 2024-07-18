package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGetOldestCommitDate(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(observation.TestContextTB(t), db)

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(time.Minute)
	t3 := t1.Add(time.Minute * 4)
	t4 := t1.Add(time.Minute * 6)

	insertUploads(t, db,
		shared.Upload{ID: 1, State: "completed"},
		shared.Upload{ID: 2, State: "completed"},
		shared.Upload{ID: 3, State: "completed"},
		shared.Upload{ID: 4, State: "errored"},
		shared.Upload{ID: 5, State: "completed"},
		shared.Upload{ID: 6, State: "completed", RepositoryID: 51},
		shared.Upload{ID: 7, State: "completed", RepositoryID: 51},
		shared.Upload{ID: 8, State: "completed", RepositoryID: 51},
	)

	if err := store.UpdateCommittedAt(context.Background(), 50, makeCommit(3), "-infinity"); err != nil {
		t.Fatalf("unexpected error updating commit date %s", err)
	}

	// Repo 50
	for commit, committedAtStr := range map[string]string{
		makeCommit(1): t3.Format(time.RFC3339),
		makeCommit(2): t4.Format(time.RFC3339),
		makeCommit(3): "-infinity",
		makeCommit(4): t1.Format(time.RFC3339),
		// commit for upload 5 is initially missing
	} {
		if err := store.UpdateCommittedAt(context.Background(), 50, commit, committedAtStr); err != nil {
			t.Fatalf("unexpected error updating commit date %s", err)
		}
	}

	if _, err := store.GetCommitAndDateForOldestUpload(context.Background(), 50); err == nil {
		t.Fatalf("expected error getting oldest commit date")
	} else if !errors.Is(err, &backfillIncompleteError{50}) {
		t.Fatalf("unexpected backfill error, got %q", err)
	}

	// Finish backfill
	if err := store.UpdateCommittedAt(context.Background(), 50, makeCommit(5), "-infinity"); err != nil {
		t.Fatalf("unexpected error updating commit date %s", err)
	}

	if commitWithDate, err := store.GetCommitAndDateForOldestUpload(context.Background(), 50); err != nil {
		t.Fatalf("unexpected error getting oldest commit date: %s", err)
	} else if commitWithDate.IsNone() {
		t.Fatalf("expected commit date for repository")
	} else {
		require.Equal(t, CommitWithDate{Commit: api.CommitID(makeCommit(1)), CommitterDate: t3}, commitWithDate.Unwrap())
	}

	// Repo 51
	for commit, committedAtStr := range map[string]string{
		makeCommit(6): t2.Format(time.RFC3339),
		makeCommit(7): "-infinity",
		makeCommit(8): "-infinity",
	} {
		if err := store.UpdateCommittedAt(context.Background(), 51, commit, committedAtStr); err != nil {
			t.Fatalf("unexpected error updating commit date %s", err)
		}
	}

	if commitAndDate, err := store.GetCommitAndDateForOldestUpload(context.Background(), 51); err != nil {
		t.Fatalf("unexpected error getting oldest commit date: %s", err)
	} else if commitAndDate.IsNone() {
		t.Fatalf("expected commit date for repository")
	} else {
		require.Equal(t, CommitWithDate{Commit: api.CommitID(makeCommit(6)), CommitterDate: t2}, commitAndDate.Unwrap())
	}

	// Missing repository
	if commitDate, err := store.GetCommitAndDateForOldestUpload(context.Background(), 52); err != nil {
		t.Fatalf("unexpected error getting oldest commit date: %s", err)
	} else if commitDate.IsSome() {
		t.Fatalf("unexpected commit date for repository")
	}
}

func TestSourcedCommitsWithoutCommittedAt(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(observation.TestContextTB(t), db)

	now := time.Unix(1587396557, 0).UTC()

	insertUploads(t, db,
		shared.Upload{ID: 1, RepositoryID: 50, Commit: makeCommit(1), State: "completed"},
		shared.Upload{ID: 2, RepositoryID: 50, Commit: makeCommit(1), State: "completed", Root: "sub/"},
		shared.Upload{ID: 3, RepositoryID: 51, Commit: makeCommit(4), State: "completed"},
		shared.Upload{ID: 4, RepositoryID: 51, Commit: makeCommit(5), State: "completed"},
		shared.Upload{ID: 5, RepositoryID: 52, Commit: makeCommit(7), State: "completed"},
		shared.Upload{ID: 6, RepositoryID: 52, Commit: makeCommit(8), State: "completed"},
	)

	sourcedCommits, err := store.SourcedCommitsWithoutCommittedAt(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error getting stale sourced commits: %s", err)
	}
	expectedCommits := []SourcedCommits{
		{RepositoryID: 50, RepositoryName: "n-50", Commits: []string{makeCommit(1)}},
		{RepositoryID: 51, RepositoryName: "n-51", Commits: []string{makeCommit(4), makeCommit(5)}},
		{RepositoryID: 52, RepositoryName: "n-52", Commits: []string{makeCommit(7), makeCommit(8)}},
	}
	if diff := cmp.Diff(expectedCommits, sourcedCommits); diff != "" {
		t.Errorf("unexpected sourced commits (-want +got):\n%s", diff)
	}

	// Update commits 1 and 4
	if err := store.UpdateCommittedAt(context.Background(), 50, makeCommit(1), now.Format(time.RFC3339)); err != nil {
		t.Fatalf("unexpected error refreshing commit resolvability: %s", err)
	}
	if err := store.UpdateCommittedAt(context.Background(), 51, makeCommit(4), now.Format(time.RFC3339)); err != nil {
		t.Fatalf("unexpected error refreshing commit resolvability: %s", err)
	}

	sourcedCommits, err = store.SourcedCommitsWithoutCommittedAt(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error getting stale sourced commits: %s", err)
	}
	expectedCommits = []SourcedCommits{
		{RepositoryID: 51, RepositoryName: "n-51", Commits: []string{makeCommit(5)}},
		{RepositoryID: 52, RepositoryName: "n-52", Commits: []string{makeCommit(7), makeCommit(8)}},
	}
	if diff := cmp.Diff(expectedCommits, sourcedCommits); diff != "" {
		t.Errorf("unexpected sourced commits (-want +got):\n%s", diff)
	}
}
