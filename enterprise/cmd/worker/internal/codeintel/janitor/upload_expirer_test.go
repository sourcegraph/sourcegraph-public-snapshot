package janitor

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestUploadExpirer(t *testing.T) {
	now := timeutil.Now()
	t1 := now.Add(-time.Hour)                 // 1 hour old
	t2 := now.Add(-time.Hour * 24 * 7)        // 1 week ago
	t3 := now.Add(-time.Hour * 24 * 30 * 5)   // 5 months ago
	t4 := now.Add(-time.Hour * 24 * 30 * 9)   // 9 months ago
	t5 := now.Add(-time.Hour * 24 * 30 * 18)  // 18 months ago
	t6 := now.Add(-time.Hour * 24 * 365 * 2)  // 3 years ago
	t8 := now.Add(-time.Hour * 24 * 365 * 15) // 15 years ago

	uploads := []dbstore.Upload{
		//
		// Repository 50

		// 1 week old
		// tip of develop (PROTECTED, younger than 3 months)
		{ID: 1, RepositoryID: 50, Commit: "deadbeef01", State: "completed", FinishedAt: &t2},

		// 1 week old
		// on develop (UNPROTECTED, not tip)
		// tip of feat/blank (PROTECTED, younger than 3 months)
		{ID: 2, RepositoryID: 50, Commit: "deadbeef02", State: "completed", FinishedAt: &t2},

		// 5 months old
		// on develop (UNPROTECTED, not tip)
		{ID: 3, RepositoryID: 50, Commit: "deadbeef03", State: "completed", FinishedAt: &t3},

		// 5 months old
		// on develop (UNPROTECTED, not tip)
		// tag v1.2.3 (PROTECTED, younger than 6 months)
		{ID: 4, RepositoryID: 50, Commit: "deadbeef04", State: "completed", FinishedAt: &t3},

		// 9 months old
		// on develop (UNPROTECTED, not tip)
		// tag v1.2.2 (UNPROTECTED, older than 6 months)
		{ID: 5, RepositoryID: 50, Commit: "deadbeef05", State: "completed", FinishedAt: &t4},

		// 5 months old
		// tip of es/feature-z (UNPROTECTED, older than 3 months)
		{ID: 6, RepositoryID: 50, Commit: "deadbeef06", State: "completed", FinishedAt: &t3},

		// 9 months old
		// tip of ef/feature-x (PROTECTED, younger than 2 years)
		{ID: 7, RepositoryID: 50, Commit: "deadbeef07", State: "completed", FinishedAt: &t4},

		// 18 months old
		// on ef/feature-x (PROTECTED, younger than 2 years)
		{ID: 8, RepositoryID: 50, Commit: "deadbeef08", State: "completed", FinishedAt: &t5},

		// 3 years old
		// tip of ef/feature-y (UNPROTECTED, older than 2 years)
		{ID: 9, RepositoryID: 50, Commit: "deadbeef09", State: "completed", FinishedAt: &t6},

		//
		// Repository 51

		// 9 months old
		// tip of ef/feature-w (UNPROTECTED, policy does not apply to this repo)
		{ID: 10, RepositoryID: 51, Commit: "deadbeef10", State: "completed", FinishedAt: &t4},

		//
		// Repository 52

		// 15 years old
		// tip of main (PROTECTED, no duration)
		{ID: 11, RepositoryID: 52, Commit: "deadbeef11", State: "completed", FinishedAt: &t8},

		// 15 years old
		// on main (UNPROTECTED, not tip)
		{ID: 12, RepositoryID: 52, Commit: "deadbeef12", State: "completed", FinishedAt: &t8},

		//
		// Repository 53

		// 1 hour old
		// covered by catch-all (PROTECTED, younger than 1 day)
		{ID: 13, RepositoryID: 53, Commit: "deadbeef13", State: "completed", FinishedAt: &t1},
	}

	// Repository 50:
	//
	//    05 ------ 04 ------ 03 ------ 02 ------ 01
	//     \         \                   \         \
	//      v1.2.2    v1.2.3              \         develop
	//                                     feat/blank
	//
	//              08 ---- 07
	//  09                   \                     06
	//   \                   ef/feature-x           \
	//    ef/feature-y                              es/feature-z

	branchMap := map[string]map[string]string{
		"deadbeef01": {"develop": "deadbeef01"},
		"deadbeef02": {"develop": "deadbeef01", "feat/blank": "deadbeef02"},
		"deadbeef03": {"develop": "deadbeef01"},
		"deadbeef04": {"develop": "deadbeef01"},
		"deadbeef05": {"develop": "deadbeef01"},
		"deadbeef06": {"es/feature-z": "deadbeef06"},
		"deadbeef07": {"ef/feature-x": "deadbeef07"},
		"deadbeef08": {"ef/feature-x": "deadbeef07"},
		"deadbeef09": {"ef/feature-y": "deadbeef09"},
		"deadbeef10": {"ef/feature-w": "deadbeef10"},
		"deadbeef11": {"main": "deadbeef11"},
		"deadbeef12": {"main": "deadbeef11"},
	}

	tagMap := map[string][]string{
		"deadbeef01": nil,
		"deadbeef02": nil,
		"deadbeef03": nil,
		"deadbeef04": {"v1.2.3"},
		"deadbeef05": {"v1.2.2"},
		"deadbeef06": nil,
		"deadbeef07": nil,
		"deadbeef08": nil,
		"deadbeef09": nil,
		"deadbeef10": nil,
		"deadbeef11": nil,
		"deadbeef12": nil,
	}

	d1 := time.Hour * 24           // 1 day
	d2 := time.Hour * 24 * 90      // 3 months
	d3 := time.Hour * 24 * 180     // 6 months
	d4 := time.Hour * 24 * 365 * 2 // 2 years

	globalPolicies := []dbstore.ConfigurationPolicy{
		{
			Type:              "GIT_TREE",
			Pattern:           "*",
			RetentionEnabled:  true,
			RetentionDuration: &d2,
		},
		{
			Type:              "GIT_TAG",
			Pattern:           "*",
			RetentionEnabled:  true,
			RetentionDuration: &d3,
		},
		{
			Type:              "GIT_TREE",
			Pattern:           "main",
			RetentionEnabled:  true,
			RetentionDuration: nil, // indefinite
		},
	}

	repositoryPolicies := map[int][]dbstore.ConfigurationPolicy{
		50: {
			dbstore.ConfigurationPolicy{
				Type:                      "GIT_TREE",
				Pattern:                   "ef/*",
				RetentionEnabled:          true,
				RetainIntermediateCommits: true,
				RetentionDuration:         &d4,
			},
		},
		51: {},
		52: {},
		53: {
			dbstore.ConfigurationPolicy{
				Type:              "GIT_COMMIT",
				Pattern:           "*",
				RetentionEnabled:  true,
				RetentionDuration: &d1,
			},
		},
	}

	dbStore := testUploadExpirerMockDBStore(globalPolicies, repositoryPolicies, uploads)
	gitserverClient := testUploadExpirerMockGitserverClient(branchMap, tagMap)

	uploadExpirer := &uploadExpirer{
		dbStore:                dbStore,
		gitserverClient:        gitserverClient,
		metrics:                newMetrics(&observation.TestContext),
		repositoryProcessDelay: 24 * time.Hour,
		repositoryBatchSize:    100,
		uploadProcessDelay:     24 * time.Hour,
		uploadBatchSize:        100,
		commitBatchSize:        100,
		branchesCacheMaxKeys:   10000,
	}

	if err := uploadExpirer.Handle(context.Background()); err != nil {
		t.Fatalf("unexpected error from handle: %s", err)
	}

	assertProtectedAndExpiredIDs(
		t,
		dbStore,
		[]int{1, 2, 4, 7, 8, 11, 13},
		[]int{3, 5, 6, 9, 10, 12},
	)
}

func TestUploadExpirerDefaultBranch(t *testing.T) {
	now := timeutil.Now()
	t1 := now.Add(-time.Hour * 24 * 365 * 15) // 15 years ago

	uploads := []dbstore.Upload{
		{ID: 1, RepositoryID: 50, Commit: "deadbeef01", State: "completed", FinishedAt: &t1},
		{ID: 2, RepositoryID: 50, Commit: "deadbeef02", State: "completed", FinishedAt: &t1},
		{ID: 3, RepositoryID: 50, Commit: "deadbeef03", State: "completed", FinishedAt: &t1},
		{ID: 4, RepositoryID: 50, Commit: "deadbeef04", State: "completed", FinishedAt: &t1},
		{ID: 5, RepositoryID: 50, Commit: "deadbeef05", State: "completed", FinishedAt: &t1},
	}

	branchMap := map[string]map[string]string{
		"deadbeef01": {"main": "deadbeef01"},
		"deadbeef02": {"main": "deadbeef01"},
		"deadbeef03": {"main": "deadbeef01"},
		"deadbeef04": {"main": "deadbeef01"},
		"deadbeef05": {"main": "deadbeef01"},
	}

	tagMap := map[string][]string{
		"deadbeef01": nil,
		"deadbeef02": nil,
		"deadbeef03": nil,
		"deadbeef04": nil,
		"deadbeef05": nil,
	}

	globalPolicies := []dbstore.ConfigurationPolicy{}
	repositoryPolicies := map[int][]dbstore.ConfigurationPolicy{
		50: {},
	}

	dbStore := testUploadExpirerMockDBStore(globalPolicies, repositoryPolicies, uploads)
	gitserverClient := testUploadExpirerMockGitserverClient(branchMap, tagMap)

	uploadExpirer := &uploadExpirer{
		dbStore:                dbStore,
		gitserverClient:        gitserverClient,
		metrics:                newMetrics(&observation.TestContext),
		repositoryProcessDelay: 24 * time.Hour,
		repositoryBatchSize:    100,
		uploadProcessDelay:     24 * time.Hour,
		uploadBatchSize:        100,
		commitBatchSize:        100,
		branchesCacheMaxKeys:   10000,
	}

	if err := uploadExpirer.Handle(context.Background()); err != nil {
		t.Fatalf("unexpected error from handle: %s", err)
	}

	assertProtectedAndExpiredIDs(
		t,
		dbStore,
		[]int{1},
		[]int{2, 3, 4, 5},
	)
}

func TestUploadExpirerCachingStrategy(t *testing.T) {
	now := timeutil.Now()

	var uploads []dbstore.Upload
	for i := 0; i < 10; i++ {
		commit := fmt.Sprintf("deadbeef%02d", i+1)
		finishedAt := now.Add(-time.Hour * time.Duration(i))

		uploads = append(uploads, dbstore.Upload{
			ID:           i + 1,
			RepositoryID: 50,
			State:        "completed",
			Commit:       commit,
			Root:         fmt.Sprintf("r%d/", i+1),
			FinishedAt:   &finishedAt,
		})
	}

	branchMap := map[string]map[string]string{
		"deadbeef01": {"main": "deadbeef01"},
		"deadbeef02": {"main": "deadbeef01"},
		"deadbeef03": {"main": "deadbeef01"},
		"deadbeef04": {"main": "deadbeef01"},
		"deadbeef05": {"main": "deadbeef01"},
		"deadbeef06": {"main": "deadbeef01"},
		"deadbeef07": {"main": "deadbeef01"},
		"deadbeef08": {"main": "deadbeef01"},
		"deadbeef09": {"main": "deadbeef01"},
		"deadbeef10": {"main": "deadbeef01"},
	}

	tagMap := map[string][]string{
		"deadbeef01": nil,
		"deadbeef02": nil,
		"deadbeef03": nil,
		"deadbeef04": nil,
		"deadbeef05": nil,
		"deadbeef06": nil,
		"deadbeef07": nil,
		"deadbeef08": nil,
		"deadbeef09": nil,
		"deadbeef10": nil,
	}

	t.Run("CachedUnprotectedCommits", func(t *testing.T) {
		d1 := time.Hour * 1
		d2 := time.Hour * 5

		globalPolicies := []dbstore.ConfigurationPolicy{
			{
				Type:                      "GIT_TREE",
				Pattern:                   "main",
				RetentionEnabled:          true,
				RetainIntermediateCommits: false,
				RetentionDuration:         &d1,
			},
			{
				Type:                      "GIT_TREE",
				Pattern:                   "main",
				RetentionEnabled:          true,
				RetainIntermediateCommits: true,
				RetentionDuration:         &d2,
			},
		}

		repositoryPolicies := map[int][]dbstore.ConfigurationPolicy{
			50: {},
		}

		dbStore := testUploadExpirerMockDBStore(globalPolicies, repositoryPolicies, uploads)
		dbStore.CommitsVisibleToUploadFunc.SetDefaultHook(testMultipleCommitsVisibleToUpload)
		gitserverClient := testUploadExpirerMockGitserverClient(branchMap, tagMap)

		uploadExpirer := &uploadExpirer{
			dbStore:                dbStore,
			gitserverClient:        gitserverClient,
			metrics:                newMetrics(&observation.TestContext),
			repositoryProcessDelay: 24 * time.Hour,
			repositoryBatchSize:    100,
			uploadProcessDelay:     24 * time.Hour,
			uploadBatchSize:        100,
			commitBatchSize:        100,
			branchesCacheMaxKeys:   10000,
		}

		if err := uploadExpirer.Handle(context.Background()); err != nil {
			t.Fatalf("unexpected error from handle: %s", err)
		}

		assertProtectedAndExpiredIDs(
			t,
			dbStore,
			[]int{1, 2, 3, 4, 5},
			[]int{6, 7, 8, 9, 10},
		)
	})

	t.Run("CachedBranches", func(t *testing.T) {
		d1 := time.Hour * 1
		d2 := time.Hour * 2
		d3 := time.Hour * 5

		globalPolicies := []dbstore.ConfigurationPolicy{
			{
				Type:                      "GIT_TREE",
				Pattern:                   "main",
				RetentionEnabled:          true,
				RetainIntermediateCommits: false,
				RetentionDuration:         &d1,
			},
			{
				Type:                      "GIT_TREE",
				Pattern:                   "main",
				RetentionEnabled:          true,
				RetainIntermediateCommits: false,
				RetentionDuration:         &d2,
			},
			{
				Type:                      "GIT_TREE",
				Pattern:                   "main",
				RetentionEnabled:          true,
				RetainIntermediateCommits: false,
				RetentionDuration:         &d3,
			},
		}

		repositoryPolicies := map[int][]dbstore.ConfigurationPolicy{
			50: {},
		}

		dbStore := testUploadExpirerMockDBStore(globalPolicies, repositoryPolicies, uploads)
		dbStore.CommitsVisibleToUploadFunc.SetDefaultHook(testMultipleCommitsVisibleToUpload)
		gitserverClient := testUploadExpirerMockGitserverClient(branchMap, tagMap)

		uploadExpirer := &uploadExpirer{
			dbStore:                dbStore,
			gitserverClient:        gitserverClient,
			metrics:                newMetrics(&observation.TestContext),
			repositoryProcessDelay: 24 * time.Hour,
			repositoryBatchSize:    100,
			uploadProcessDelay:     24 * time.Hour,
			uploadBatchSize:        100,
			commitBatchSize:        100,
			branchesCacheMaxKeys:   10000,
		}

		if err := uploadExpirer.Handle(context.Background()); err != nil {
			t.Fatalf("unexpected error from handle: %s", err)
		}

		assertProtectedAndExpiredIDs(
			t,
			dbStore,
			[]int{1, 2},
			[]int{3, 4, 5, 6, 7, 8, 9, 10},
		)
	})
}

func assertProtectedAndExpiredIDs(t *testing.T, dbStore *MockDBStore, expectedProtectedIDs, expectedExpiredIDs []int) {
	var protectedIDs []int
	for _, call := range dbStore.UpdateUploadRetentionFunc.History() {
		protectedIDs = append(protectedIDs, call.Arg1...)
	}
	sort.Ints(protectedIDs)

	var expiredIDs []int
	for _, call := range dbStore.UpdateUploadRetentionFunc.History() {
		expiredIDs = append(expiredIDs, call.Arg2...)
	}
	sort.Ints(expiredIDs)

	if diff := cmp.Diff(expectedProtectedIDs, protectedIDs); diff != "" {
		t.Errorf("unexpected protected upload identifiers (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(expectedExpiredIDs, expiredIDs); diff != "" {
		t.Errorf("unexpected expired upload identifiers (-want +got):\n%s", diff)
	}
}

// testMultipleCommitsVisibleToUpload is an alternate implementation of the mock DBStore
// CommitsVisibleToUpload method. The default behavior mocked in this package returns _only_
// the commit on which the upload is defined. We instead want to have each upload be visible
// to one ancestor and one descendant commit as well so that we can test the slow path.
//
// This function assumes that the direct parent of deadbeef{c} is deadbeef{c+1}.
func testMultipleCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) ([]string, *string, error) {
	lo := uploadID - 1
	if lo < 1 {
		lo = 1
	}

	hi := uploadID + 1
	if hi > 10 {
		hi = 10
	}

	var commits []string
	for i := lo; i <= hi; i++ {
		commits = append(commits, fmt.Sprintf("deadbeef%02d", i))
	}

	return commits, nil, nil
}
