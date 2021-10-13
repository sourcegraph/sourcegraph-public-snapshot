package policies

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
)

func testUploadExpirerMockGitserverClient(defaultBranchName string, now time.Time) *MockGitserverClient {
	// Test repository:
	//
	//                                              v2.2.2                              02 -- feat/blank
	//                                             /                                   /
	//  09               08 ---- 07              06              05 ------ 04 ------ 03 ------ 01
	//   \                        \               \               \         \                   \
	//    xy/feature-y            xy/feature-x    zw/feature-z     v1.2.2    v1.2.3              develop

	var branchHeads = map[string]string{
		"develop":      "deadbeef01",
		"feat/blank":   "deadbeef02",
		"xy/feature-x": "deadbeef07",
		"zw/feature-z": "deadbeef06",
		"xy/feature-y": "deadbeef09",
	}

	var tagHeads = map[string]string{
		"v1.2.3": "deadbeef04",
		"v1.2.2": "deadbeef05",
		"v2.2.2": "deadbeef06",
	}

	var branchMembers = map[string][]string{
		"develop":      {"deadbeef01", "deadbeef03", "deadbeef04", "deadbeef05"},
		"feat/blank":   {"deadbeef02"},
		"xy/feature-x": {"deadbeef07", "deadbeef08"},
		"xy/feature-y": {"deadbeef09"},
		"zw/feature-z": {"deadbeef06"},
	}

	var createdAt = map[string]time.Time{
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

	commitDate := func(ctx context.Context, repositoryID int, commit string) (time.Time, error) {
		return createdAt[commit], nil
	}

	refDescriptions := func(ctx context.Context, repositoryID int) (map[string][]gitserver.RefDescription, error) {
		refDescriptions := map[string][]gitserver.RefDescription{}
		for branch, commit := range branchHeads {
			refDescriptions[commit] = append(refDescriptions[commit], gitserver.RefDescription{
				Name:            branch,
				Type:            gitserver.RefTypeBranch,
				IsDefaultBranch: branch == defaultBranchName,
				CreatedDate:     createdAt[commit],
			})
		}

		for tag, commit := range tagHeads {
			refDescriptions[commit] = append(refDescriptions[commit], gitserver.RefDescription{
				Name:        tag,
				Type:        gitserver.RefTypeTag,
				CreatedDate: createdAt[commit],
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
