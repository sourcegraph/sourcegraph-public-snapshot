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

func TestRepoIDsByGlobPatternToo(t *testing.T) {
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
		{50, "Darth/Maul"},
		{51, "DarthVader"},
		{52, "Jedi Anakin"},
		{53, "Jediyoda"},
		{54, "Robot C3PO"},
	}

	// insert repo test data
	for _, mock := range mockInsertedRepos {
		insertRepo(t, db, mock.ID, mock.name)
	}

	// set test data and expected
	testData := []struct {
		pattern  string
		expected []int
	}{
		{pattern: "Darth*", expected: []int{50, 51}},
		{pattern: "Darth/*", expected: []int{50}},
		{pattern: "Jedi*", expected: []int{52, 53}},
		{pattern: "*C3PO*", expected: []int{54}},
		{pattern: "*Human*", expected: nil},
	}

	for _, data := range testData {
		// find pattern
		repoIds, err := store.RepoIDsByGlobPattern(ctx, data.pattern)
		if err != nil {
			t.Fatalf("unexpected error fetching repository IDs by glob pattern: %s", err)
		}

		// Actual test what you get with what is expected
		if diff := cmp.Diff(data.expected, repoIds); diff != "" {
			t.Errorf("unexpected job (-want +got):\n%s", diff)
		}
	}
}
