package db

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestUpdateCommits(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	if err := db.UpdateCommits(context.Background(), 50, map[string][]string{
		makeCommit(1): {},
		makeCommit(2): {makeCommit(1)},
		makeCommit(3): {makeCommit(1)},
		makeCommit(4): {makeCommit(2), makeCommit(3)},
	}); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

	query := `
		SELECT "commit", "parent_commit"
		FROM lsif_commits
		WHERE repository_id = 50
		ORDER BY "commit", "parent_commit"
	`

	rows, err := dbconn.Global.Query(query)
	if err != nil {
		t.Fatalf("unexpected error querying commits: %s", err)
	}
	defer rows.Close()

	type commitPair struct {
		Commit       string
		ParentCommit *string
	}

	var commitPairs []commitPair
	for rows.Next() {
		var commit string
		var parentCommit *string
		if err := rows.Scan(&commit, &parentCommit); err != nil {
			t.Fatalf("unexpected error scanning row: %s", err)
		}

		commitPairs = append(commitPairs, commitPair{commit, parentCommit})
	}

	expectedCommitPairs := []commitPair{
		{makeCommit(1), nil},
		{makeCommit(2), strPtr(makeCommit(1))},
		{makeCommit(3), strPtr(makeCommit(1))},
		{makeCommit(4), strPtr(makeCommit(2))},
		{makeCommit(4), strPtr(makeCommit(3))},
	}
	if diff := cmp.Diff(expectedCommitPairs, commitPairs); diff != "" {
		t.Errorf("unexpected commits (-want +got):\n%s", diff)
	}
}

func TestUpdateCommitsWithConflicts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	if err := db.UpdateCommits(context.Background(), 50, map[string][]string{
		makeCommit(1): {},
		makeCommit(2): {makeCommit(1)},
		makeCommit(3): {makeCommit(1)},
		makeCommit(4): {makeCommit(2), makeCommit(3)},
	}); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

	if err := db.UpdateCommits(context.Background(), 50, map[string][]string{
		makeCommit(3): {makeCommit(1)},
		makeCommit(4): {makeCommit(3), makeCommit(5)},
		makeCommit(5): {makeCommit(6), makeCommit(7)},
	}); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

	query := `
		SELECT "commit", "parent_commit"
		FROM lsif_commits
		WHERE repository_id = 50
		ORDER BY "commit", "parent_commit"
	`

	rows, err := dbconn.Global.Query(query)
	if err != nil {
		t.Fatalf("unexpected error querying commits: %s", err)
	}
	defer rows.Close()

	type commitPair struct {
		Commit       string
		ParentCommit *string
	}

	var commitPairs []commitPair
	for rows.Next() {
		var commit string
		var parentCommit *string
		if err := rows.Scan(&commit, &parentCommit); err != nil {
			t.Fatalf("unexpected error scanning row: %s", err)
		}

		commitPairs = append(commitPairs, commitPair{commit, parentCommit})
	}

	expectedCommitPairs := []commitPair{
		{makeCommit(1), nil},
		{makeCommit(2), strPtr(makeCommit(1))},
		{makeCommit(3), strPtr(makeCommit(1))},
		{makeCommit(4), strPtr(makeCommit(2))},
		{makeCommit(4), strPtr(makeCommit(3))},
		{makeCommit(4), strPtr(makeCommit(5))},
		{makeCommit(5), strPtr(makeCommit(6))},
		{makeCommit(5), strPtr(makeCommit(7))},
	}
	if diff := cmp.Diff(expectedCommitPairs, commitPairs); diff != "" {
		t.Errorf("unexpected commits (-want +got):\n%s", diff)
	}
}

func strPtr(v string) *string {
	return &v
}
