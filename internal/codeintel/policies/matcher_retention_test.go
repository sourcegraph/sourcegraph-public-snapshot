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

func TestCommitsDescribedByPolicyForRetention(t *testing.T) {
	now := timeutil.Now()
	mainGitserverClient := testUploadExpirerMockGitserverClient("main", now)
	developGitserverClient := testUploadExpirerMockGitserverClient("develop", now)

	runTest := func(t *testing.T, gitserverClient *gitserver.MockClient, policies []policiesshared.ConfigurationPolicy, expectedPolicyMatches map[string][]PolicyMatch) {
		policyMatches, err := NewMatcher(gitserverClient, RetentionExtractor, true, false).CommitsDescribedByPolicy(context.Background(), 50, "r50", policies, now)
		if err != nil {
			t.Fatalf("unexpected error finding matches: %s", err)
		}

		hydrateCommittedAt(expectedPolicyMatches, now)
		sortPolicyMatchesMap(policyMatches)
		sortPolicyMatchesMap(expectedPolicyMatches)

		if diff := cmp.Diff(expectedPolicyMatches, policyMatches); diff != "" {
			t.Errorf("unexpected policy matches (-want +got):\n%s", diff)
		}

		for i, call := range gitserverClient.CommitsFunc.History() {
			if !call.Arg2.After.IsZero() {
				t.Errorf("unexpected restriction of git results by date: call #%d", i)
			}
		}
	}

	policyID := 42
	testDuration := time.Hour * 24

	t.Run("matches tag policies", func(t *testing.T) {
		policies := []policiesshared.ConfigurationPolicy{
			{
				ID:                policyID,
				Type:              "GIT_TAG",
				Pattern:           "v1.*",
				RetentionDuration: &testDuration,
			},
		}

		runTest(t, mainGitserverClient, policies, map[string][]PolicyMatch{
			// N.B. tag v2.2.2 does not match filter
			"deadbeef04": {PolicyMatch{Name: "v1.2.3", PolicyID: &policyID, PolicyDuration: &testDuration}},
			"deadbeef05": {PolicyMatch{Name: "v1.2.2", PolicyID: &policyID, PolicyDuration: &testDuration}},
		})
	})

	t.Run("matches branches tip policies", func(t *testing.T) {
		policies := []policiesshared.ConfigurationPolicy{
			{
				ID:                policyID,
				Type:              "GIT_TREE",
				Pattern:           "xy/*",
				RetentionDuration: &testDuration,
			},
		}

		runTest(t, mainGitserverClient, policies, map[string][]PolicyMatch{
			// N.B. branch zw/* does not match this filter
			"deadbeef07": {PolicyMatch{Name: "xy/feature-x", PolicyID: &policyID, PolicyDuration: &testDuration}},
			"deadbeef09": {PolicyMatch{Name: "xy/feature-y", PolicyID: &policyID, PolicyDuration: &testDuration}},
		})
	})

	t.Run("matches commits on branch policies", func(t *testing.T) {
		policies := []policiesshared.ConfigurationPolicy{
			{
				ID:                        policyID,
				Type:                      "GIT_TREE",
				Pattern:                   "xy/*",
				RetentionDuration:         &testDuration,
				RetainIntermediateCommits: true,
			},
		}

		runTest(t, mainGitserverClient, policies, map[string][]PolicyMatch{
			// N.B. branch zw/* does not match this filter
			"deadbeef07": {PolicyMatch{Name: "xy/feature-x", PolicyID: &policyID, PolicyDuration: &testDuration}},
			"deadbeef08": {PolicyMatch{Name: "xy/feature-x", PolicyID: &policyID, PolicyDuration: &testDuration}},
			"deadbeef09": {PolicyMatch{Name: "xy/feature-y", PolicyID: &policyID, PolicyDuration: &testDuration}},
		})
	})

	t.Run("matches commit policies", func(t *testing.T) {
		policies := []policiesshared.ConfigurationPolicy{
			{
				ID:                policyID,
				Type:              "GIT_COMMIT",
				Pattern:           "deadbeef04",
				RetentionDuration: &testDuration,
			},
		}

		runTest(t, mainGitserverClient, policies, map[string][]PolicyMatch{
			"deadbeef04": {PolicyMatch{Name: "deadbeef04", PolicyID: &policyID, PolicyDuration: &testDuration}},
		})
	})
	t.Run("matches implicit tip of default branch policy", func(t *testing.T) {
		runTest(t, developGitserverClient, nil, map[string][]PolicyMatch{
			"deadbeef01": {{Name: "develop", PolicyID: nil, PolicyDuration: nil}},
		})
	})
}
