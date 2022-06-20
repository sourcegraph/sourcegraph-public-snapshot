package policies

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

func testUploadExpirerMockGitserverClient(defaultBranchName string, now time.Time) *MockGitserverClient {
	// Test repository:
	//
	//                                              v2.2.2                              02 -- feat/blank
	//                                             /                                   /
	//  09               08 ---- 07              06              05 ------ 04 ------ 03 ------ 01
	//   \                        \               \               \         \                   \
	//    xy/feature-y            xy/feature-x    zw/feature-z     v1.2.2    v1.2.3              develop

	branchHeads := map[string]string{
		"develop":      "deadbeef01",
		"feat/blank":   "deadbeef02",
		"xy/feature-x": "deadbeef07",
		"zw/feature-z": "deadbeef06",
		"xy/feature-y": "deadbeef09",
	}

	tagHeads := map[string]string{
		"v1.2.3": "deadbeef04",
		"v1.2.2": "deadbeef05",
		"v2.2.2": "deadbeef06",
	}

	branchMembers := map[string][]string{
		"develop":      {"deadbeef01", "deadbeef03", "deadbeef04", "deadbeef05"},
		"feat/blank":   {"deadbeef02"},
		"xy/feature-x": {"deadbeef07", "deadbeef08"},
		"xy/feature-y": {"deadbeef09"},
		"zw/feature-z": {"deadbeef06"},
		"deadbeef01":   {"deadbeef01", "deadbeef03", "deadbeef04", "deadbeef05"},
		"deadbeef02":   {"deadbeef02"},
		"deadbeef06":   {"deadbeef06"},
		"deadbeef07":   {"deadbeef07", "deadbeef08"},
		"deadbeef09":   {"deadbeef09"},
	}

	createdAt := map[string]time.Time{
		"deadbeef01": now.Add(-time.Hour * 5),
		"deadbeef02": now.Add(-time.Hour * 5),
		"deadbeef03": now.Add(-time.Hour * 5),
		"deadbeef04": now.Add(-time.Hour * 5),
		"deadbeef07": now.Add(-time.Hour * 5),
		"deadbeef08": now.Add(-time.Hour * 5),
		"deadbeef05": now.Add(-time.Hour * 12),
		"deadbeef06": now.Add(-time.Hour * 15),
		"deadbeef09": now.Add(-time.Hour * 15),
	}

	commitDate := func(ctx context.Context, repositoryID int, commit string) (string, time.Time, bool, error) {
		commitDate, ok := createdAt[commit]
		return commit, commitDate, ok, nil
	}

	refDescriptions := func(ctx context.Context, repositoryID int, _ ...string) (map[string][]gitdomain.RefDescription, error) {
		refDescriptions := map[string][]gitdomain.RefDescription{}
		for branch, commit := range branchHeads {
			branchHeadCreateDate := createdAt[commit]
			refDescriptions[commit] = append(refDescriptions[commit], gitdomain.RefDescription{
				Name:            branch,
				Type:            gitdomain.RefTypeBranch,
				IsDefaultBranch: branch == defaultBranchName,
				CreatedDate:     &branchHeadCreateDate,
			})
		}

		for tag, commit := range tagHeads {
			tagCreateDate := createdAt[commit]
			refDescriptions[commit] = append(refDescriptions[commit], gitdomain.RefDescription{
				Name:        tag,
				Type:        gitdomain.RefTypeTag,
				CreatedDate: &tagCreateDate,
			})
		}

		return refDescriptions, nil
	}

	commitsUniqueToBranch := func(ctx context.Context, repositoryID int, branchName string, isDefaultBranch bool, maxAge *time.Time) (map[string]time.Time, error) {
		branches := map[string]time.Time{}
		for _, commit := range branchMembers[branchName] {
			if maxAge == nil || !createdAt[commit].Before(*maxAge) {
				branches[commit] = createdAt[commit]
			}
		}

		return branches, nil
	}

	gitserverClient := NewMockGitserverClient()
	gitserverClient.CommitDateFunc.SetDefaultHook(commitDate)
	gitserverClient.RefDescriptionsFunc.SetDefaultHook(refDescriptions)
	gitserverClient.CommitsUniqueToBranchFunc.SetDefaultHook(commitsUniqueToBranch)
	return gitserverClient
}
