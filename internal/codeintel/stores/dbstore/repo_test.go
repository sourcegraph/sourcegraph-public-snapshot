package dbstore

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestRepoNames(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)
	ctx := context.Background()

	insertRepo(t, db, 50, "A")
	insertRepo(t, db, 51, "B")
	insertRepo(t, db, 52, "C")
	insertRepo(t, db, 53, "D")
	insertRepo(t, db, 54, "E")
	insertRepo(t, db, 55, "F")

	names, err := store.RepoNames(ctx, 50, 52, 53, 54, 57)
	if err != nil {
		t.Fatalf("unexpected error querying repository names: %s", err)
	}

	expected := map[int]string{
		50: "A",
		52: "C",
		53: "D",
		54: "E",
	}
	if diff := cmp.Diff(expected, names); diff != "" {
		t.Errorf("unexpected repository names (-want +got):\n%s", diff)
	}
}

func TestUpdateReposMatchingPatterns(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)
	ctx := context.Background()

	insertRepo(t, db, 50, "r1")
	insertRepo(t, db, 51, "r2")
	insertRepo(t, db, 52, "r3")
	insertRepo(t, db, 53, "r4")
	insertRepo(t, db, 54, "r5")

	updates := []struct {
		policyID int
		pattern  []string
	}{
		// multiple matches
		{100, []string{"r*"}},

		// exact identifiers
		{101, []string{"r1"}},

		// multiple exact identifiers
		{102, []string{"r2", "r3"}},

		// updated patterns (disjoint)
		{103, []string{"r4"}},
		{103, []string{"r5"}},

		// updated patterns (intersecting)
		{104, []string{"r1", "r2", "r3"}},
		{104, []string{"r2", "r3", "r4"}},

		// deleted matches
		{105, []string{"r5"}},
		{105, []string{}},
	}
	for _, update := range updates {
		if err := store.UpdateReposMatchingPatterns(ctx, update.pattern, update.policyID, nil); err != nil {
			t.Fatalf("unexpected error updating repositories matching patterns: %s", err)
		}
	}

	policies, err := scanPolicyRepositories(db.QueryContext(context.Background(), `
		SELECT policy_id, repo_id
		FROM lsif_configuration_policies_repository_pattern_lookup
	`))
	if err != nil {
		t.Fatalf("unexpected error while scanning policies: %s", err)
	}

	for _, repositoryIDs := range policies {
		sort.Ints(repositoryIDs)
	}

	expectedPolicies := map[int][]int{
		100: {50, 51, 52, 53, 54}, // multiple matches
		101: {50},                 // exact identifiers
		102: {51, 52},             // multiple exact identifiers
		103: {54},                 // updated patterns (disjoint)
		104: {51, 52, 53},         // updated patterns (intersecting)
	}
	if diff := cmp.Diff(expectedPolicies, policies); diff != "" {
		t.Errorf("unexpected repository identifiers for policies (-want +got):\n%s", diff)
	}
}

func TestUpdateReposMatchingPatternsOverLimit(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)
	ctx := context.Background()

	limit := 50
	ids := make([]int, 0, limit*3)
	for i := 0; i < cap(ids); i++ {
		ids = append(ids, 50+i)
	}

	for _, id := range ids {
		insertRepo(t, db, id, fmt.Sprintf("r%03d", id))
	}

	if err := store.UpdateReposMatchingPatterns(ctx, []string{"r*"}, 100, &limit); err != nil {
		t.Fatalf("unexpected error updating repositories matching patterns: %s", err)
	}

	policies, err := scanPolicyRepositories(db.QueryContext(context.Background(), `
		SELECT policy_id, repo_id
		FROM lsif_configuration_policies_repository_pattern_lookup
	`))
	if err != nil {
		t.Fatalf("unexpected error while scanning policies: %s", err)
	}

	for _, repositoryIDs := range policies {
		sort.Ints(repositoryIDs)
	}

	expectedPolicies := map[int][]int{
		100: ids[:limit],
	}
	if diff := cmp.Diff(expectedPolicies, policies); diff != "" {
		t.Errorf("unexpected repository identifiers for policies (-want +got):\n%s", diff)
	}
}

// scanPolicyRepositories returns a map of policyIDs that have a slice of their correspondent repoIDs (repoIDs associated with that policyIDs).
func scanPolicyRepositories(rows *sql.Rows, queryErr error) (_ map[int][]int, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	policies := map[int][]int{}
	for rows.Next() {
		var policyID int
		var repoID int
		if err := rows.Scan(&policyID, &repoID); err != nil {
			return nil, err
		}

		policies[policyID] = append(policies[policyID], repoID)
	}

	return policies, nil
}
