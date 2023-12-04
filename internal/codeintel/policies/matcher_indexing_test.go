package policies

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestCommitsDescribedByPolicyForIndexing(t *testing.T) {
	now := timeutil.Now()
	mainGitserverClient := testUploadExpirerMockGitserverClient("main", now)
	developGitserverClient := testUploadExpirerMockGitserverClient("develop", now)

	runTest := func(t *testing.T, gitserverClient gitserver.Client, policies []policiesshared.ConfigurationPolicy, expectedPolicyMatches map[string][]PolicyMatch) {
		policyMatches, err := NewMatcher(gitserverClient, IndexingExtractor, false, true).CommitsDescribedByPolicy(context.Background(), 50, "r50", policies, now)
		if err != nil {
			t.Fatalf("unexpected error finding matches: %s", err)
		}

		hydrateCommittedAt(expectedPolicyMatches, now)
		sortPolicyMatchesMap(policyMatches)
		sortPolicyMatchesMap(expectedPolicyMatches)

		if diff := cmp.Diff(expectedPolicyMatches, policyMatches); diff != "" {
			t.Errorf("unexpected policy matches (-want +got):\n%s", diff)
		}
	}

	policyID := 42
	testDuration := time.Hour * 10

	t.Run("matches tag policies", func(t *testing.T) {
		policies := []policiesshared.ConfigurationPolicy{
			{
				ID:                policyID,
				Type:              "GIT_TAG",
				Pattern:           "v1.*",
				IndexCommitMaxAge: &testDuration,
			},
		}

		runTest(t, mainGitserverClient, policies, map[string][]PolicyMatch{
			// N.B. tag v2.2.2 does not match filter
			// N.B. tag v1.2.2 does not fall within policy duration
			"deadbeef04": {PolicyMatch{Name: "v1.2.3", PolicyID: &policyID, PolicyDuration: &testDuration}},
		})
	})

	t.Run("matches branches tip policies", func(t *testing.T) {
		policies := []policiesshared.ConfigurationPolicy{
			{
				ID:                policyID,
				Type:              "GIT_TREE",
				Pattern:           "xy/*",
				IndexCommitMaxAge: &testDuration,
			},
		}

		runTest(t, mainGitserverClient, policies, map[string][]PolicyMatch{
			// N.B. branch zw/* does not match this filter
			// N.B. xy/feature-y does not fall within policy duration
			"deadbeef07": {PolicyMatch{Name: "xy/feature-x", PolicyID: &policyID, PolicyDuration: &testDuration}},
		})
	})

	t.Run("matches commits on branch policies", func(t *testing.T) {
		policies := []policiesshared.ConfigurationPolicy{
			{
				ID:                       policyID,
				Type:                     "GIT_TREE",
				Pattern:                  "xy/*",
				IndexCommitMaxAge:        &testDuration,
				IndexIntermediateCommits: true,
			},
		}

		runTest(t, mainGitserverClient, policies, map[string][]PolicyMatch{
			// N.B. branch zw/* does not match this filter
			// N.B. xy/feature-y does not fall within policy duration
			"deadbeef07": {PolicyMatch{Name: "xy/feature-x", PolicyID: &policyID, PolicyDuration: &testDuration}},
			"deadbeef08": {PolicyMatch{Name: "xy/feature-x", PolicyID: &policyID, PolicyDuration: &testDuration}},
		})
	})

	t.Run("return all matching policies for each commit", func(t *testing.T) {
		policyID1 := policyID
		policyID2 := policyID1 + 1
		policyID3 := policyID1 + 2

		testDuration1 := testDuration
		testDuration2 := time.Hour * 13
		testDuration3 := time.Hour * 20

		policies := []policiesshared.ConfigurationPolicy{
			{
				ID:                       policyID1,
				Type:                     "GIT_TREE",
				Pattern:                  "develop",
				IndexCommitMaxAge:        &testDuration1,
				IndexIntermediateCommits: true,
			},
			{
				ID:                       policyID2,
				Type:                     "GIT_TREE",
				Pattern:                  "*",
				IndexCommitMaxAge:        &testDuration2,
				IndexIntermediateCommits: true,
			},
			{
				ID:                       policyID3,
				Type:                     "GIT_TREE",
				Pattern:                  "feat/*",
				IndexCommitMaxAge:        &testDuration3,
				IndexIntermediateCommits: true,
			},
		}

		runTest(t, mainGitserverClient, policies, map[string][]PolicyMatch{
			"deadbeef01": {
				PolicyMatch{Name: "develop", PolicyID: &policyID1, PolicyDuration: &testDuration1},
				PolicyMatch{Name: "develop", PolicyID: &policyID2, PolicyDuration: &testDuration2},
			},
			"deadbeef02": {
				PolicyMatch{Name: "feat/blank", PolicyID: &policyID2, PolicyDuration: &testDuration2},
				PolicyMatch{Name: "feat/blank", PolicyID: &policyID3, PolicyDuration: &testDuration3},
			},
			"deadbeef03": {
				PolicyMatch{Name: "develop", PolicyID: &policyID1, PolicyDuration: &testDuration1},
				PolicyMatch{Name: "develop", PolicyID: &policyID2, PolicyDuration: &testDuration2},
			},
			"deadbeef04": {
				PolicyMatch{Name: "develop", PolicyID: &policyID1, PolicyDuration: &testDuration1},
				PolicyMatch{Name: "develop", PolicyID: &policyID2, PolicyDuration: &testDuration2},
			},

			// N.B. deadbeef05 too old to match policy 1
			// N.B. deadbeef06 and deadbeef09 are too old for any matching policy
			"deadbeef05": {PolicyMatch{Name: "develop", PolicyID: &policyID2, PolicyDuration: &testDuration2}},
			"deadbeef07": {PolicyMatch{Name: "xy/feature-x", PolicyID: &policyID2, PolicyDuration: &testDuration2}},
			"deadbeef08": {PolicyMatch{Name: "xy/feature-x", PolicyID: &policyID2, PolicyDuration: &testDuration2}},
		})
	})

	t.Run("matches commit policies", func(t *testing.T) {
		policies := []policiesshared.ConfigurationPolicy{
			{
				ID:                policyID,
				Type:              "GIT_COMMIT",
				Pattern:           "deadbeef04",
				IndexCommitMaxAge: &testDuration,
			},
		}

		runTest(t, mainGitserverClient, policies, map[string][]PolicyMatch{
			"deadbeef04": {PolicyMatch{Name: "deadbeef04", PolicyID: &policyID, PolicyDuration: &testDuration}},
		})
	})

	t.Run("does not match commit policies outside of policy duration", func(t *testing.T) {
		policies := []policiesshared.ConfigurationPolicy{
			{
				ID:                policyID,
				Type:              "GIT_COMMIT",
				Pattern:           "deadbeef05",
				IndexCommitMaxAge: &testDuration,
			},
		}

		runTest(t, mainGitserverClient, policies, map[string][]PolicyMatch{
			// N.B. deadbeef05 does not fall within policy duration
		})
	})

	t.Run("does not match a default policy", func(t *testing.T) {
		runTest(t, developGitserverClient, nil, nil)
	})
}
