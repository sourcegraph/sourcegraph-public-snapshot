package dbstore

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func TestUpdateReposMatchingPatterns(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	store := testStore(db)
	ctx := context.Background()

	insertRepo(t, db, 50, "r0")
	insertRepo(t, db, 51, "r1")
	insertRepo(t, db, 52, "r2")
	insertRepo(t, db, 53, "r3")
	insertRepo(t, db, 54, "r4")

	// 1. multiple matches
	if err := store.UpdateReposMatchingPatterns(ctx, []string{"r*"}, 100); err != nil {
		t.Fatalf("unexpected error fetching repositories for update: %s", err)
	}

	// 2. exact ids
	if err := store.UpdateReposMatchingPatterns(ctx, []string{"r1"}, 101); err != nil {
		t.Fatalf("unexpected error fetching repositories for update: %s", err)
	}

	if err := store.UpdateReposMatchingPatterns(ctx, []string{"r3"}, 102); err != nil {
		t.Fatalf("unexpected error fetching repositories for update: %s", err)
	}

	// 3. something being overwritten
	if err := store.UpdateReposMatchingPatterns(ctx, []string{"r2"}, 102); err != nil {
		t.Fatalf("unexpected error fetching repositories for update: %s", err)
	}

	// 4. matching one or the other
	if err := store.UpdateReposMatchingPatterns(ctx, []string{"r2", "r1"}, 103); err != nil {
		t.Fatalf("unexpected error fetching repositories for update: %s", err)
	}

	// 5. deletes it when empty
	if err := store.UpdateReposMatchingPatterns(ctx, []string{"r4"}, 104); err != nil {
		t.Fatalf("unexpected error fetching repositories for update: %s", err)
	}

	if err := store.UpdateReposMatchingPatterns(ctx, []string{}, 104); err != nil {
		t.Fatalf("unexpected error fetching repositories for update: %s", err)
	}

	policies := queryRepositoryPatternLookup(t, db)

	expectedPolicies := map[int][]int{
		100: {50, 51, 52, 53, 54},
		101: {51},
		102: {52},
		103: {51, 52},
	}

	if diff := cmp.Diff(expectedPolicies, policies); diff != "" {
		t.Errorf("unexpected job (-want +got):\n%s", diff)
	}
}

func queryRepositoryPatternLookup(t *testing.T, db dbutil.DB) map[int][]int {
	policies, err := scanPolicyRepositories(db.QueryContext(context.Background(), "SELECT policy_id, repo_id FROM lsif_configuration_policies_repository_pattern_lookup"))
	if err != nil {
		t.Fatalf("unexpected error while scanning policies: %s", err)
	}

	return policies
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

func TestRepoIDsByGlobPattern(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	store := testStore(db)
	ctx := context.Background()

	insertRepo(t, db, 50, "Darth/Maul")
	insertRepo(t, db, 51, "DarthVader")
	insertRepo(t, db, 52, "Jedi Anakin")
	insertRepo(t, db, 53, "Jediyoda")
	insertRepo(t, db, 54, "Robot C3PO")

	repoIds, err := store.RepoIDsByGlobPattern(ctx, "Darth*")
	if err != nil {
		t.Fatalf("unexpected error fetching repository IDs by glob pattern: %s", err)
	}

	expectedRepoIds := []int{50, 51}
	if diff := cmp.Diff(expectedRepoIds, repoIds); diff != "" {
		t.Errorf("unexpected job (-want +got):\n%s", diff)
	}

	repoIds, err = store.RepoIDsByGlobPattern(ctx, "Darth/*")
	if err != nil {
		t.Fatalf("unexpected error fetching repository IDs by glob pattern: %s", err)
	}

	expectedRepoIds = []int{50}
	if diff := cmp.Diff(expectedRepoIds, repoIds); diff != "" {
		t.Errorf("unexpected job (-want +got):\n%s", diff)
	}

	repoIds, err = store.RepoIDsByGlobPattern(ctx, "Jedi*")
	if err != nil {
		t.Fatalf("unexpected error fetching repository IDs by glob pattern: %s", err)
	}

	expectedRepoIds = []int{52, 53}
	if diff := cmp.Diff(expectedRepoIds, repoIds); diff != "" {
		t.Errorf("unexpected job (-want +got):\n%s", diff)
	}

	repoIds, err = store.RepoIDsByGlobPattern(ctx, "*C3PO*")
	if err != nil {
		t.Fatalf("unexpected error fetching repository IDs by glob pattern: %s", err)
	}

	expectedRepoIds = []int{54}
	if diff := cmp.Diff(expectedRepoIds, repoIds); diff != "" {
		t.Errorf("unexpected job (-want +got):\n%s", diff)
	}

	repoIds, err = store.RepoIDsByGlobPattern(ctx, "*Human*")
	if err != nil {
		t.Fatalf("unexpected error fetching repository IDs by glob pattern: %s", err)
	}

	expectedRepoIds = nil
	if diff := cmp.Diff(expectedRepoIds, repoIds); diff != "" {
		t.Errorf("unexpected job (-want +got):\n%s", diff)
	}
}
