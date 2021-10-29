package dbstore

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRepoIDsByGlobPattern(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	store := testStore(db)
	ctx := context.Background()

	insertRepo(t, db, 50, "Darth Vader")
	insertRepo(t, db, 51, "Darth Venamis")
	insertRepo(t, db, 52, "Darth Maul")
	insertRepo(t, db, 53, "Anakin Skywalker")
	insertRepo(t, db, 54, "Luke Skywalker")
	insertRepo(t, db, 55, "7th Sky Corps")

	testCases := []struct {
		pattern               string
		expectedRepositoryIDs []int
	}{
		{pattern: "Darth*", expectedRepositoryIDs: []int{50, 51, 52}},  // Prefix
		{pattern: "Darth V*", expectedRepositoryIDs: []int{50, 51}},    // Prefix
		{pattern: "* Skywalker", expectedRepositoryIDs: []int{53, 54}}, // Suffix
		{pattern: "*er", expectedRepositoryIDs: []int{50, 53, 54}},     // Suffix
		{pattern: "*Sky*", expectedRepositoryIDs: []int{53, 54, 55}},   // Infix
		{pattern: "Rey Skywalker", expectedRepositoryIDs: nil},         // No match, never happened
	}

	for _, testCase := range testCases {
		repositoryIDs, err := store.RepoIDsByGlobPattern(ctx, testCase.pattern)
		if err != nil {
			t.Fatalf("unexpected error fetching repository ids by glob pattern: %s", err)
		}

		if diff := cmp.Diff(testCase.expectedRepositoryIDs, repositoryIDs); diff != "" {
			t.Errorf("unexpected repository ids (-want +got):\n%s", diff)
		}
	}

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Enable permissions user mapping forces checking repository permissions
		// against permissions tables in the database, which should effectively block
		// all access because permissions tables are empty.
		before := globals.PermissionsUserMapping()
		globals.SetPermissionsUserMapping(&schema.PermissionsUserMapping{Enabled: true})
		defer globals.SetPermissionsUserMapping(before)

		repoIDs, err := store.RepoIDsByGlobPattern(ctx, "*")
		if err != nil {
			t.Fatal(err)
		}
		if len(repoIDs) > 0 {
			t.Fatalf("Want no repositories but got %d repositories", len(repoIDs))
		}
	})
}

func TestUpdateReposMatchingPatterns(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	store := testStore(db)
	ctx := context.Background()

	// set repo test data
	mockInsertedRepos := []struct {
		ID   int
		name string
	}{
		{50, "r0"},
		{51, "r1"},
		{52, "r2"},
		{53, "r3"},
		{54, "r4"},
	}

	// insert repo test data
	for _, mock := range mockInsertedRepos {
		insertRepo(t, db, mock.ID, mock.name)
	}

	// set pattern data
	testDataList := []struct {
		pattern     []string
		policyID    int
		description string
	}{
		{pattern: []string{"r*"}, policyID: 100, description: "multiple matches"},
		{pattern: []string{"r1"}, policyID: 101, description: "exact ids"},
		{pattern: []string{"r3"}, policyID: 102, description: "exact ids"},
		{pattern: []string{"r2"}, policyID: 102, description: "something being overwritten"},
		{pattern: []string{"r2", "r1"}, policyID: 103, description: "matching one or the other"},
		{pattern: []string{"r4"}, policyID: 104, description: "deletes it when empty"},
		{pattern: []string{}, policyID: 104, description: "deletes it when empty"},
	}

	// execute update repos matching patterns with pattern data above.
	for _, data := range testDataList {
		if err := store.UpdateReposMatchingPatterns(ctx, data.pattern, data.policyID); err != nil {
			t.Fatalf("unexpected error fetching repositories for update: %s for %s", err, data.description)
		}
	}

	// get everything from the lookup table
	policies := queryRepositoryPatternLookup(t, db)

	// if all goes well with the insertion above this is what we should see
	expectedPolicies := map[int][]int{
		100: {50, 51, 52, 53, 54}, // "multiple matches"
		101: {51},                 // "exact ids"
		102: {52},                 // "something being overwritten"
		103: {51, 52},             // "matching one or the other"
		// deletes it when empty
	}

	// actual test
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
