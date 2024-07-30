package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestGetConfigurationPolicies(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := testStoreWithoutConfigurationPolicies(t, db)
	ctx := context.Background()

	query := `
		INSERT INTO lsif_configuration_policies (
			id,
			repository_id,
			name,
			type,
			pattern,
			repository_patterns,
			retention_enabled,
			retention_duration_hours,
			retain_intermediate_commits,
			indexing_enabled,
			syntactic_indexing_enabled,
			index_commit_max_age_hours,
			index_intermediate_commits,
			embeddings_enabled,
			protected
		) VALUES
			--                        							              ↙ retention_enabled
			--                        							              |    ↙ retention_duration_hours
			--                        							              |    |    ↙ retain_intermediate_commits
			--                        							              |    |    |      ↙ indexing_enabled
			--                        							              |    |    |      |      ↙ syntactic_indexing_enabled
			--                        							              |    |    |      |      |    ↙ index_commit_max_age_hours
			--                        							              |    |    |      |      |    |     ↙ index_intermediate_commits
			--                        							              |    |    |      |      |    |     |     ↙ embeddings_enabled
			--                        							              |    |    |      |      |    |     |     |     ↙ protected
			(101, 42,   'policy  1 abc', 'GIT_TREE', '', null,              false, 0, false, true,  false, 0, false, false, true),
			(102, 42,   'policy  2 def', 'GIT_TREE', '', null,              true , 0, false, false, true,  0, false, false, true),
			(103, 43,   'policy  3 bcd', 'GIT_TREE', '', null,              false, 0, false, true,  false, 0, false, false, true),
			(104, NULL, 'policy  4 abc', 'GIT_TREE', '', null,              true , 0, false, false, true,  0, false, false, false),
			(105, NULL, 'policy  5 bcd', 'GIT_TREE', '', null,              false, 0, false, true,  true,  0, false, false, false),
			(106, NULL, 'policy  6 bcd', 'GIT_TREE', '', '{gitlab.com/*}',  true , 0, false, false, true,  0, false, false, false),
			(107, NULL, 'policy  7 def', 'GIT_TREE', '', '{gitlab.com/*1}', false, 0, false, true,  false, 0, false, false, false),
			(108, NULL, 'policy  8 abc', 'GIT_TREE', '', '{gitlab.com/*2}', true , 0, false, false, true,  0, false, false, false),
			(109, NULL, 'policy  9 def', 'GIT_TREE', '', '{github.com/*}',  false, 0, false, true,  false, 0, false, false, false),
			(110, NULL, 'policy 10 def', 'GIT_TREE', '', '{github.com/*}',  false, 0, false, false, false, 0, false, true,  false),
			(111, 43,   'policy 11 bcd', 'GIT_TREE', '', null,              false, 0, false, false, true,  0, false, false, true)
	`
	if _, err := db.ExecContext(ctx, query); err != nil {
		t.Fatalf("unexpected error while inserting configuration policies: %s", err)
	}

	insertRepo(t, db, 41, "gitlab.com/test1", false)
	insertRepo(t, db, 42, "github.com/test2", false)
	insertRepo(t, db, 43, "bitbucket.org/test3", false)
	insertRepo(t, db, 44, "localhost/secret-repo", false)

	for policyID, patterns := range map[int][]string{
		106: {"gitlab.com/*"},
		107: {"gitlab.com/*1"},
		108: {"gitlab.com/*2"},
		109: {"github.com/*"},
		110: {"github.com/*"},
	} {
		if err := store.UpdateReposMatchingPatterns(ctx, patterns, policyID, nil); err != nil {
			t.Fatalf("unexpected error while updating repositories matching patterns: %s", err)
		}
	}

	var (
		trueValue  = true
		falseValue = false
	)

	type testCase struct {
		repositoryID         int
		term                 string
		forDataRetention     *bool
		forPreciseIndexing   *bool
		forSyntacticIndexing *bool
		forEmbeddings        *bool
		protected            *bool
		expectedIDs          []int
	}
	testCases := []testCase{
		{forEmbeddings: &falseValue, expectedIDs: []int{101, 102, 103, 104, 105, 106, 107, 108, 109, 111}},     // Any flags; all policies
		{forEmbeddings: &falseValue, protected: &trueValue, expectedIDs: []int{101, 102, 103, 111}},            // Only protected
		{forEmbeddings: &falseValue, protected: &falseValue, expectedIDs: []int{104, 105, 106, 107, 108, 109}}, // Only un-protected

		{forEmbeddings: &falseValue, repositoryID: 41, expectedIDs: []int{104, 105, 106, 107}},      // Any flags; matches repo by patterns
		{forEmbeddings: &falseValue, repositoryID: 42, expectedIDs: []int{101, 102, 104, 105, 109}}, // Any flags; matches repo by assignment and pattern
		{forEmbeddings: &falseValue, repositoryID: 43, expectedIDs: []int{103, 104, 105, 111}},      // Any flags; matches repo by assignment
		{forEmbeddings: &falseValue, repositoryID: 44, expectedIDs: []int{104, 105}},                // Any flags; no matches by repo

		{forEmbeddings: &falseValue, forDataRetention: &trueValue, expectedIDs: []int{102, 104, 106, 108}},         // For data retention; all policies
		{forEmbeddings: &falseValue, forDataRetention: &trueValue, repositoryID: 41, expectedIDs: []int{104, 106}}, // For data retention; matches repo by patterns
		{forEmbeddings: &falseValue, forDataRetention: &trueValue, repositoryID: 42, expectedIDs: []int{102, 104}}, // For data retention; matches repo by assignment and pattern
		{forEmbeddings: &falseValue, forDataRetention: &trueValue, repositoryID: 43, expectedIDs: []int{104}},      // For data retention; matches repo by assignment
		{forEmbeddings: &falseValue, forDataRetention: &trueValue, repositoryID: 44, expectedIDs: []int{104}},      // For data retention; no matches by repo

		{forEmbeddings: &falseValue, forPreciseIndexing: &trueValue, expectedIDs: []int{101, 103, 105, 107, 109}},         // For indexing; all policies
		{forEmbeddings: &falseValue, forPreciseIndexing: &trueValue, repositoryID: 41, expectedIDs: []int{105, 107}},      // For indexing; matches repo by patterns
		{forEmbeddings: &falseValue, forPreciseIndexing: &trueValue, repositoryID: 42, expectedIDs: []int{101, 105, 109}}, // For indexing; matches repo by assignment and pattern
		{forEmbeddings: &falseValue, forPreciseIndexing: &trueValue, repositoryID: 43, expectedIDs: []int{103, 105}},      // For indexing; matches repo by assignment
		{forEmbeddings: &falseValue, forPreciseIndexing: &trueValue, repositoryID: 44, expectedIDs: []int{105}},           // For indexing; no matches by repo

		{forSyntacticIndexing: &trueValue, expectedIDs: []int{102, 104, 105, 106, 108, 111}},    // For syntactic indexing; all policies
		{forSyntacticIndexing: &trueValue, repositoryID: 41, expectedIDs: []int{104, 105, 106}}, // For syntactic indexing; matches repo by patterns
		{forSyntacticIndexing: &trueValue, repositoryID: 42, expectedIDs: []int{102, 104, 105}}, // For syntactic indexing; matches repo by assignment and pattern
		{forSyntacticIndexing: &trueValue, repositoryID: 43, expectedIDs: []int{104, 105, 111}}, // For syntactic indexing; matches repo by assignment
		{forSyntacticIndexing: &trueValue, repositoryID: 44, expectedIDs: []int{104, 105}},      // For syntactic indexing; no matches by repo

		{forDataRetention: &falseValue, forPreciseIndexing: &falseValue, forEmbeddings: &trueValue, expectedIDs: []int{110}}, // For embeddings

		{term: "bc", expectedIDs: []int{101, 103, 104, 105, 106, 108, 111}}, // Searches by name (multiple substring matches)
		{term: "abcd", expectedIDs: []int{}},                                // Searches by name (no matches)
	}

	runTest := func(testCase testCase, lo, hi int) (errors int) {
		name := fmt.Sprintf(
			"repositoryID=%d term=%q forDataRetention=%v forIndexing=%v forSyntacticIndexing=%v forEmbeddings=%v protected=%v offset=%d",
			testCase.repositoryID,
			testCase.term,
			// without formatBoolOption the strings for those bool pointers end up looking like `0xc000010b4f`
			// which only helps differentiate between empty value and non-empty
			formatBoolOption(testCase.forDataRetention),
			formatBoolOption(testCase.forPreciseIndexing),
			formatBoolOption(testCase.forSyntacticIndexing),
			formatBoolOption(testCase.forEmbeddings),
			formatBoolOption(testCase.protected),
			lo,
		)

		t.Run(name, func(t *testing.T) {
			policies, totalCount, err := store.GetConfigurationPolicies(ctx, policiesshared.GetConfigurationPoliciesOptions{
				RepositoryID:         testCase.repositoryID,
				Term:                 testCase.term,
				ForDataRetention:     testCase.forDataRetention,
				ForPreciseIndexing:   testCase.forPreciseIndexing,
				ForSyntacticIndexing: testCase.forSyntacticIndexing,
				ForEmbeddings:        testCase.forEmbeddings,
				Protected:            testCase.protected,
				Limit:                3,
				Offset:               lo,
			})
			if err != nil {
				t.Fatalf("unexpected error fetching configuration policies: %s", err)
			}
			if totalCount != 0 {
				var ids []int
				for _, policy := range policies {
					ids = append(ids, policy.ID)
				}
				if diff := cmp.Diff(testCase.expectedIDs[lo:hi], ids); diff != "" {
					t.Errorf("unexpected configuration policy ids at offset %d (-want +got):\n%s", lo, diff)
					errors++
				}
			}
			if totalCount != len(testCase.expectedIDs) {
				t.Errorf("unexpected total count. want=%d have=%d", len(testCase.expectedIDs), totalCount)
				errors++
			}

		})

		return errors
	}

	for _, testCase := range testCases {
		if n := len(testCase.expectedIDs); n == 0 {
			runTest(testCase, 0, 0)
		} else {
			for lo := range n {
				if numErrors := runTest(testCase, lo, min(lo+3, n)); numErrors > 0 {
					break
				}
			}
		}
	}
}

func TestDeleteConfigurationPolicyByID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := testStoreWithoutConfigurationPolicies(t, db)

	repositoryID := 42
	d1 := time.Hour * 5
	d2 := time.Hour * 6

	configurationPolicy := policiesshared.ConfigurationPolicy{
		RepositoryID:              &repositoryID,
		Name:                      "name",
		Type:                      policiesshared.GitObjectTypeCommit,
		Pattern:                   "deadbeef",
		RetentionEnabled:          false,
		RetentionDuration:         &d1,
		RetainIntermediateCommits: true,
		PreciseIndexingEnabled:    false,
		SyntacticIndexingEnabled:  false,
		IndexCommitMaxAge:         &d2,
		IndexIntermediateCommits:  true,
	}

	hydratedConfigurationPolicy, err := store.CreateConfigurationPolicy(context.Background(), configurationPolicy)
	if err != nil {
		t.Fatalf("unexpected error creating configuration policy: %s", err)
	}

	if hydratedConfigurationPolicy.ID == 0 {
		t.Fatalf("hydrated policy does not have an identifier")
	}

	if err := store.DeleteConfigurationPolicyByID(context.Background(), hydratedConfigurationPolicy.ID); err != nil {
		t.Fatalf("unexpected error deleting configuration policy: %s", err)
	}

	_, ok, err := store.GetConfigurationPolicyByID(context.Background(), hydratedConfigurationPolicy.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching configuration policy: %s", err)
	}
	if ok {
		t.Fatalf("unexpected record")
	}
}

func TestDeleteConfigurationProtectedPolicy(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := testStoreWithoutConfigurationPolicies(t, db)

	repositoryID := 42
	d1 := time.Hour * 5
	d2 := time.Hour * 6

	configurationPolicy := policiesshared.ConfigurationPolicy{
		RepositoryID:              &repositoryID,
		Name:                      "name",
		Type:                      policiesshared.GitObjectTypeCommit,
		Pattern:                   "deadbeef",
		RetentionEnabled:          false,
		RetentionDuration:         &d1,
		RetainIntermediateCommits: true,
		PreciseIndexingEnabled:    false,
		SyntacticIndexingEnabled:  false,
		IndexCommitMaxAge:         &d2,
		IndexIntermediateCommits:  true,
	}

	hydratedConfigurationPolicy, err := store.CreateConfigurationPolicy(ctx, configurationPolicy)
	if err != nil {
		t.Fatalf("unexpected error creating configuration policy: %s", err)
	}

	if hydratedConfigurationPolicy.ID == 0 {
		t.Fatalf("hydrated policy does not have an identifier")
	}

	// Mark configuration policy as protected (no other way to do so outside of migrations)
	if _, err := db.ExecContext(ctx, "UPDATE lsif_configuration_policies SET protected = true"); err != nil {
		t.Fatalf("unexpected error marking configuration policy as protected: %s", err)
	}

	if err := store.DeleteConfigurationPolicyByID(ctx, hydratedConfigurationPolicy.ID); err == nil {
		t.Fatalf("expected error deleting configuration policy: %s", err)
	}

	_, ok, err := store.GetConfigurationPolicyByID(ctx, hydratedConfigurationPolicy.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching configuration policy: %s", err)
	}
	if !ok {
		t.Fatalf("expected record")
	}
}

func formatBoolOption(opt *bool) string {
	if opt == nil {
		return "nil"
	} else if *opt {
		return "true"
	} else {
		return "false"
	}
}
