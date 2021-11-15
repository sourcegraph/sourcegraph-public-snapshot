package indexing

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func init() {
	autoIndexingEnabled = func() bool { return true }
}

func TestIndexScheduler(t *testing.T) {
	now := timeutil.Now()
	dbStore := testIndexSchedulerMockDBStore()
	policyMatcher := testIndexSchedulerMockPolicyMatcher(now)
	indexEnqueuer := NewMockIndexEnqueuer()

	scheduler := &IndexScheduler{
		dbStore:                dbStore,
		policyMatcher:          policyMatcher,
		indexEnqueuer:          indexEnqueuer,
		repositoryProcessDelay: 24 * time.Hour,
		repositoryBatchSize:    100,
		operations:             newOperations(&observation.TestContext),
	}

	if err := scheduler.Handle(context.Background()); err != nil {
		t.Fatalf("unexpected error performing update: %s", err)
	}

	var repoCommits []string
	for _, call := range indexEnqueuer.QueueIndexesFunc.History() {
		repoCommits = append(repoCommits, fmt.Sprintf("%d@%s", call.Arg1, call.Arg2))
	}
	sort.Strings(repoCommits)

	expectedRepoCommits := []string{
		"50@deadbeef02",
		"51@deadbeef06",
		"51@deadbeef08",
		"51@deadbeef10",
		"52@deadbeef15",
		"53@deadbeef16",
		"53@deadbeef17",
		"53@deadbeef18",
	}
	if diff := cmp.Diff(expectedRepoCommits, repoCommits); diff != "" {
		t.Errorf("unexpected repository IDs (-want +got):\n%s", diff)
	}

	commitCalls := policyMatcher.CommitsDescribedByPolicyFunc.History()
	if len(commitCalls) != 4 {
		t.Fatalf("unexpected number of calls to CommitsDescribedByPolicy. want=%d have=%d", 4, len(commitCalls))
	}
	for _, call := range commitCalls {
		var policyIDs []int
		for _, policy := range call.Arg2 {
			policyIDs = append(policyIDs, policy.ID)
		}
		sort.Ints(policyIDs)

		expectedPolicyIDs := map[int][]int{
			50: {1, 3, 4, 5},
			51: {1, 3, 4},
			52: {1, 3, 4},
			53: {1, 2, 3, 4},
		}
		if diff := cmp.Diff(expectedPolicyIDs[call.Arg1], policyIDs); diff != "" {
			t.Errorf("unexpected policies supplied to CommitsDescribedByPolicy(%d) (-want +got):\n%s", call.Arg1, diff)
		}
	}

	enqueueCalls := indexEnqueuer.QueueIndexesFunc.History()
	if len(enqueueCalls) != 8 {
		t.Fatalf("unexpected number of calls to QueueIndexes. want=%d have=%d", 8, len(indexEnqueuer.QueueIndexesFunc.History()))
	}

}

func testIndexSchedulerMockDBStore() *MockDBStore {
	repositoryIDs := []int{
		50,
		51,
		52,
		53,
	}

	policies := []dbstore.ConfigurationPolicy{
		{ID: 1, RepositoryID: nil},
		{ID: 2, RepositoryID: intPtr(53)},
		{ID: 3, RepositoryID: nil},
		{ID: 4, RepositoryID: nil},
		{ID: 5, RepositoryID: intPtr(50)},
	}

	selectRepositoriesForIndexScan := func(ctx context.Context, processDelay time.Duration, limit int) (scannedIDs []int, _ error) {
		if len(repositoryIDs) <= limit {
			scannedIDs, repositoryIDs = repositoryIDs, nil
		} else {
			scannedIDs, repositoryIDs = repositoryIDs[:limit], repositoryIDs[limit:]
		}

		return scannedIDs, nil
	}

	getConfigurationPolicies := func(ctx context.Context, opts dbstore.GetConfigurationPoliciesOptions) (filtered []dbstore.ConfigurationPolicy, _ error) {
		for _, policy := range policies {
			if opts.RepositoryID == 0 {
				if policy.RepositoryID != nil {
					continue
				}
			} else if policy.RepositoryID == nil || *policy.RepositoryID != opts.RepositoryID {
				continue
			}

			filtered = append(filtered, policy)
		}

		return filtered, nil
	}

	dbStore := NewMockDBStore()
	dbStore.SelectRepositoriesForIndexScanFunc.SetDefaultHook(selectRepositoriesForIndexScan)
	dbStore.GetConfigurationPoliciesFunc.SetDefaultHook(getConfigurationPolicies)
	return dbStore
}

func testIndexSchedulerMockPolicyMatcher(now time.Time) *MockPolicyMatcher {
	policyMatches := map[int]map[string][]policies.PolicyMatch{
		50: {
			"deadbeef01": {},
			"deadbeef02": {{PolicyDuration: days(9)}},
			"deadbeef03": {},
			"deadbeef04": {},
			"deadbeef05": {},
		},
		51: {
			"deadbeef06": {{PolicyDuration: days(7)}},
			"deadbeef07": {},
			"deadbeef08": {{PolicyDuration: days(9)}},
			"deadbeef09": {},
			"deadbeef10": {{PolicyDuration: days(9)}},
		},
		52: {
			"deadbeef11": {},
			"deadbeef12": {},
			"deadbeef13": {},
			"deadbeef14": {},
			"deadbeef15": {{PolicyDuration: nil}},
		},
		53: {
			"deadbeef16": {{PolicyDuration: days(5)}},
			"deadbeef17": {{PolicyDuration: days(5)}},
			"deadbeef18": {{PolicyDuration: days(5)}},
			"deadbeef19": {},
			"deadbeef20": {},
		},
	}

	commitsDescribedByPolicy := func(ctx context.Context, repositoryID int, policies []dbstore.ConfigurationPolicy, now time.Time) (map[string][]policies.PolicyMatch, error) {
		return policyMatches[repositoryID], nil
	}

	policyMatcher := NewMockPolicyMatcher()
	policyMatcher.CommitsDescribedByPolicyFunc.SetDefaultHook(commitsDescribedByPolicy)
	return policyMatcher
}

func intPtr(v int) *int {
	return &v
}

func days(n int) *time.Duration {
	t := time.Hour * 24 * time.Duration(n)
	return &t
}
