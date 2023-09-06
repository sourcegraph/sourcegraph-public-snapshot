package expirer

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	internaltypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestUploadExpirer(t *testing.T) {
	now := timeutil.Now()
	uploadSvc := setupMockUploadService(now)
	policySvc := setupMockPolicyService()
	policyMatcher := testUploadExpirerMockPolicyMatcher()
	repoStore := defaultMockRepoStore()
	expirationMetrics := NewExpirationMetrics(&observation.TestContext)

	uploadExpirer := &expirer{
		store:         uploadSvc,
		policySvc:     policySvc,
		policyMatcher: policyMatcher,
		repoStore:     repoStore,
	}

	if err := uploadExpirer.HandleExpiredUploadsBatch(context.Background(), expirationMetrics, &Config{
		RepositoryProcessDelay: 24 * time.Hour,
		RepositoryBatchSize:    100,
		UploadProcessDelay:     24 * time.Hour,
		UploadBatchSize:        100,
		CommitBatchSize:        100,
	}); err != nil {
		t.Fatalf("unexpected error from handle: %s", err)
	}

	var protectedIDs []int
	for _, call := range uploadSvc.UpdateUploadRetentionFunc.History() {
		protectedIDs = append(protectedIDs, call.Arg1...)
	}
	sort.Ints(protectedIDs)

	var expiredIDs []int
	for _, call := range uploadSvc.UpdateUploadRetentionFunc.History() {
		expiredIDs = append(expiredIDs, call.Arg2...)
	}
	sort.Ints(expiredIDs)

	expectedProtectedIDs := []int{12, 16, 18, 20, 25, 26, 27, 28}
	if diff := cmp.Diff(expectedProtectedIDs, protectedIDs); diff != "" {
		t.Errorf("unexpected protected upload identifiers (-want +got):\n%s", diff)
	}

	expectedExpiredIDs := []int{11, 13, 14, 15, 17, 19, 21, 22, 23, 24, 29, 30}
	if diff := cmp.Diff(expectedExpiredIDs, expiredIDs); diff != "" {
		t.Errorf("unexpected expired upload identifiers (-want +got):\n%s", diff)
	}

	calls := policyMatcher.CommitsDescribedByPolicyFunc.History()
	if len(calls) != 4 {
		t.Fatalf("unexpected number of calls to CommitsDescribedByPolicy. want=%d have=%d", 4, len(calls))
	}
	for _, call := range calls {
		var policyIDs []int
		for _, policy := range call.Arg3 {
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
}

func setupMockPolicyService() *MockPolicyService {
	policies := []policiesshared.ConfigurationPolicy{
		{ID: 1, RepositoryID: nil},
		{ID: 2, RepositoryID: pointers.Ptr(53)},
		{ID: 3, RepositoryID: nil},
		{ID: 4, RepositoryID: nil},
		{ID: 5, RepositoryID: pointers.Ptr(50)},
	}

	getConfigurationPolicies := func(ctx context.Context, opts policiesshared.GetConfigurationPoliciesOptions) (filtered []policiesshared.ConfigurationPolicy, _ int, _ error) {
		for _, policy := range policies {
			if policy.RepositoryID == nil || *policy.RepositoryID == opts.RepositoryID {
				filtered = append(filtered, policy)
			}
		}

		return filtered, len(filtered), nil
	}

	policySvc := NewMockPolicyService()
	policySvc.GetConfigurationPoliciesFunc.SetDefaultHook(getConfigurationPolicies)

	return policySvc
}

func setupMockUploadService(now time.Time) *MockStore {
	uploads := []shared.Upload{
		{ID: 11, State: "completed", RepositoryID: 50, Commit: "deadbeef01", UploadedAt: daysAgo(now, 1)}, // repo 50
		{ID: 12, State: "completed", RepositoryID: 50, Commit: "deadbeef02", UploadedAt: daysAgo(now, 2)},
		{ID: 13, State: "completed", RepositoryID: 50, Commit: "deadbeef03", UploadedAt: daysAgo(now, 3)},
		{ID: 14, State: "completed", RepositoryID: 50, Commit: "deadbeef04", UploadedAt: daysAgo(now, 4)},
		{ID: 15, State: "completed", RepositoryID: 50, Commit: "deadbeef05", UploadedAt: daysAgo(now, 5)},
		{ID: 16, State: "completed", RepositoryID: 51, Commit: "deadbeef06", UploadedAt: daysAgo(now, 6)}, // repo 51
		{ID: 17, State: "completed", RepositoryID: 51, Commit: "deadbeef07", UploadedAt: daysAgo(now, 7)},
		{ID: 18, State: "completed", RepositoryID: 51, Commit: "deadbeef08", UploadedAt: daysAgo(now, 8)},
		{ID: 19, State: "completed", RepositoryID: 51, Commit: "deadbeef09", UploadedAt: daysAgo(now, 9)},
		{ID: 20, State: "completed", RepositoryID: 51, Commit: "deadbeef10", UploadedAt: daysAgo(now, 1)},
		{ID: 21, State: "completed", RepositoryID: 52, Commit: "deadbeef11", UploadedAt: daysAgo(now, 9)}, // repo 52
		{ID: 22, State: "completed", RepositoryID: 52, Commit: "deadbeef12", UploadedAt: daysAgo(now, 8)},
		{ID: 23, State: "completed", RepositoryID: 52, Commit: "deadbeef13", UploadedAt: daysAgo(now, 7)},
		{ID: 24, State: "completed", RepositoryID: 52, Commit: "deadbeef14", UploadedAt: daysAgo(now, 6)},
		{ID: 25, State: "completed", RepositoryID: 52, Commit: "deadbeef15", UploadedAt: daysAgo(now, 5)},
		{ID: 26, State: "completed", RepositoryID: 53, Commit: "deadbeef16", UploadedAt: daysAgo(now, 4)}, // repo 53
		{ID: 27, State: "completed", RepositoryID: 53, Commit: "deadbeef17", UploadedAt: daysAgo(now, 3)},
		{ID: 28, State: "completed", RepositoryID: 53, Commit: "deadbeef18", UploadedAt: daysAgo(now, 2)},
		{ID: 29, State: "completed", RepositoryID: 53, Commit: "deadbeef19", UploadedAt: daysAgo(now, 1)},
		{ID: 30, State: "completed", RepositoryID: 53, Commit: "deadbeef20", UploadedAt: daysAgo(now, 9)},
	}

	repositoryIDMap := map[int]struct{}{}
	for _, upload := range uploads {
		repositoryIDMap[upload.RepositoryID] = struct{}{}
	}

	repositoryIDs := make([]int, 0, len(repositoryIDMap))
	for repositoryID := range repositoryIDMap {
		repositoryIDs = append(repositoryIDs, repositoryID)
	}

	protected := map[int]time.Time{}
	expired := map[int]struct{}{}

	setRepositoriesForRetentionScanFunc := func(ctx context.Context, processDelay time.Duration, limit int) (scannedIDs []int, _ error) {
		if len(repositoryIDs) <= limit {
			scannedIDs, repositoryIDs = repositoryIDs, nil
		} else {
			scannedIDs, repositoryIDs = repositoryIDs[:limit], repositoryIDs[limit:]
		}

		return scannedIDs, nil
	}

	getUploads := func(ctx context.Context, opts uploadsshared.GetUploadsOptions) ([]shared.Upload, int, error) {
		var filtered []shared.Upload
		for _, upload := range uploads {
			if upload.RepositoryID != opts.RepositoryID {
				continue
			}
			if _, ok := expired[upload.ID]; ok {
				continue
			}
			if lastScanned, ok := protected[upload.ID]; ok && !lastScanned.Before(*opts.LastRetentionScanBefore) {
				continue
			}

			filtered = append(filtered, upload)
		}

		if len(filtered) > opts.Limit {
			filtered = filtered[:opts.Limit]
		}

		return filtered, len(uploads), nil
	}

	updateUploadRetention := func(ctx context.Context, protectedIDs, expiredIDs []int) error {
		for _, id := range protectedIDs {
			protected[id] = time.Now()
		}

		for _, id := range expiredIDs {
			expired[id] = struct{}{}
		}

		return nil
	}

	getCommitsVisibleToUpload := func(ctx context.Context, uploadID, limit int, token *string) ([]string, *string, error) {
		for _, upload := range uploads {
			if upload.ID == uploadID {
				return []string{
					upload.Commit,
					"deadcafe" + upload.Commit[8:],
				}, nil, nil
			}
		}

		return nil, nil, nil
	}

	uploadSvc := NewMockStore()
	uploadSvc.SetRepositoriesForRetentionScanFunc.SetDefaultHook(setRepositoriesForRetentionScanFunc)
	uploadSvc.GetUploadsFunc.SetDefaultHook(getUploads)
	uploadSvc.UpdateUploadRetentionFunc.SetDefaultHook(updateUploadRetention)
	uploadSvc.GetCommitsVisibleToUploadFunc.SetDefaultHook(getCommitsVisibleToUpload)

	return uploadSvc
}

func testUploadExpirerMockPolicyMatcher() *MockPolicyMatcher {
	policyMatches := map[int]map[string][]policies.PolicyMatch{
		50: {
			"deadbeef01": {{PolicyDuration: days(1)}}, // 1 = 1
			"deadbeef02": {{PolicyDuration: days(9)}}, // 9 > 2 (protected)
			"deadbeef03": {{PolicyDuration: days(2)}}, // 2 < 3
			"deadbeef04": {},
			"deadbeef05": {},
		},
		51: {
			// N.B. deadcafe (alt visible commit) used here
			"deadcafe06": {{PolicyDuration: days(7)}}, // 7 > 6 (protected)
			"deadcafe07": {{PolicyDuration: days(6)}}, // 6 < 7
			"deadbeef08": {{PolicyDuration: days(9)}}, // 9 > 8 (protected)
			"deadbeef09": {{PolicyDuration: days(9)}}, // 9 = 9
			"deadbeef10": {{PolicyDuration: days(9)}}, // 9 > 1 (protected)
		},
		52: {
			"deadbeef11": {{PolicyDuration: days(5)}},                        // 5 < 9
			"deadbeef12": {{PolicyDuration: days(5)}},                        // 5 < 8
			"deadbeef13": {{PolicyDuration: days(5)}},                        // 5 < 7
			"deadbeef14": {{PolicyDuration: days(5)}},                        // 5 < 6
			"deadbeef15": {{PolicyDuration: days(5)}, {PolicyDuration: nil}}, // 5 = 5, catch-all (protected)
		},
		53: {
			"deadbeef16": {{PolicyDuration: days(5)}}, // 5 > 4 (protected)
			"deadbeef17": {{PolicyDuration: days(5)}}, // 5 > 3 (protected)
			"deadbeef18": {{PolicyDuration: days(5)}}, // 5 > 2 (protected)
			"deadbeef19": {},
			"deadbeef20": {},
		},
	}

	commitsDescribedByPolicy := func(ctx context.Context, repositoryID int, repoName api.RepoName, policies []policiesshared.ConfigurationPolicy, now time.Time, _ ...string) (map[string][]policies.PolicyMatch, error) {
		return policyMatches[repositoryID], nil
	}

	policyMatcher := NewMockPolicyMatcher()
	policyMatcher.CommitsDescribedByPolicyFunc.SetDefaultHook(commitsDescribedByPolicy)
	return policyMatcher
}

func days(n int) *time.Duration {
	t := time.Hour * 24 * time.Duration(n)
	return &t
}

func daysAgo(now time.Time, n int) time.Time {
	return now.Add(-time.Hour * 24 * time.Duration(n))
}

func defaultMockRepoStore() *dbmocks.MockRepoStore {
	repoStore := dbmocks.NewMockRepoStore()
	repoStore.GetFunc.SetDefaultHook(func(ctx context.Context, id api.RepoID) (*internaltypes.Repo, error) {
		return &internaltypes.Repo{
			ID:   id,
			Name: api.RepoName(fmt.Sprintf("r%d", id)),
		}, nil
	})
	return repoStore
}
