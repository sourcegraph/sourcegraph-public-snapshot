package dbstore

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRepoNames(t *testing.T) {
	db := dbtest.NewDB(t)
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

func TestRepoIDsByGlobPatterns(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)
	ctx := context.Background()

	insertRepo(t, db, 50, "Darth Vader")
	insertRepo(t, db, 51, "Darth Venamis")
	insertRepo(t, db, 52, "Darth Maul")
	insertRepo(t, db, 53, "Anakin Skywalker")
	insertRepo(t, db, 54, "Luke Skywalker")
	insertRepo(t, db, 55, "7th Sky Corps")

	testCases := []struct {
		patterns              []string
		expectedRepositoryIDs []int
	}{
		{patterns: []string{""}, expectedRepositoryIDs: nil},                                             // No patterns
		{patterns: []string{"*"}, expectedRepositoryIDs: []int{50, 51, 52, 53, 54, 55}},                  // Wildcard
		{patterns: []string{"Darth*"}, expectedRepositoryIDs: []int{50, 51, 52}},                         // Prefix
		{patterns: []string{"Darth V*"}, expectedRepositoryIDs: []int{50, 51}},                           // Prefix
		{patterns: []string{"* Skywalker"}, expectedRepositoryIDs: []int{53, 54}},                        // Suffix
		{patterns: []string{"*er"}, expectedRepositoryIDs: []int{50, 53, 54}},                            // Suffix
		{patterns: []string{"*Sky*"}, expectedRepositoryIDs: []int{53, 54, 55}},                          // Infix
		{patterns: []string{"Darth *", "* Skywalker"}, expectedRepositoryIDs: []int{50, 51, 52, 53, 54}}, // Multiple patterns
		{patterns: []string{"Rey Skywalker"}, expectedRepositoryIDs: nil},                                // No match, never happened
	}

	for _, testCase := range testCases {
		for lo := 0; lo < len(testCase.expectedRepositoryIDs); lo++ {
			hi := lo + 3
			if hi > len(testCase.expectedRepositoryIDs) {
				hi = len(testCase.expectedRepositoryIDs)
			}

			name := fmt.Sprintf(
				"patterns=%v offset=%d",
				testCase.patterns,
				lo,
			)

			t.Run(name, func(t *testing.T) {
				repositoryIDs, _, err := store.RepoIDsByGlobPatterns(ctx, testCase.patterns, 3, lo)
				if err != nil {
					t.Fatalf("unexpected error fetching repository ids by glob pattern: %s", err)
				}

				if diff := cmp.Diff(testCase.expectedRepositoryIDs[lo:hi], repositoryIDs); diff != "" {
					t.Errorf("unexpected repository ids (-want +got):\n%s", diff)
				}
			})
		}
	}

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Enable permissions user mapping forces checking repository permissions
		// against permissions tables in the database, which should effectively block
		// all access because permissions tables are empty.
		before := globals.PermissionsUserMapping()
		globals.SetPermissionsUserMapping(&schema.PermissionsUserMapping{Enabled: true})
		defer globals.SetPermissionsUserMapping(before)

		repoIDs, _, err := store.RepoIDsByGlobPatterns(ctx, []string{"*"}, 10, 0)
		if err != nil {
			t.Fatal(err)
		}
		if len(repoIDs) > 0 {
			t.Fatalf("Want no repositories but got %d repositories", len(repoIDs))
		}
	})
}

func TestUpdateReposMatchingPatterns(t *testing.T) {
	db := dbtest.NewDB(t)
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
	db := dbtest.NewDB(t)
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
