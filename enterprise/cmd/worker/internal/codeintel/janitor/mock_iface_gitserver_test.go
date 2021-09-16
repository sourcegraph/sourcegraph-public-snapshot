package janitor

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
)

// testUploadExpirerMockGitserverClient returns a mock GitserverClient instance that
// has default behaviors useful for testing the upload expirer.
func testUploadExpirerMockGitserverClient(branchMap map[string]map[string]string, tagMap map[string][]string) *MockGitserverClient {
	gitserverClient := NewMockGitserverClient()

	gitserverClient.RefDescriptionsFunc.SetDefaultHook(func(ctx context.Context, repositoryID int) (map[string][]gitserver.RefDescription, error) {
		refDescriptions := map[string][]gitserver.RefDescription{}
		for commit, branches := range branchMap {
			for branch, tip := range branches {
				if tip != commit {
					continue
				}

				refDescriptions[commit] = append(refDescriptions[commit], gitserver.RefDescription{
					Name: branch,
					Type: gitserver.RefTypeBranch,
				})
			}
		}

		for commit, tags := range tagMap {
			for _, tag := range tags {
				refDescriptions[commit] = append(refDescriptions[commit], gitserver.RefDescription{
					Name: tag,
					Type: gitserver.RefTypeTag,
				})
			}
		}

		return refDescriptions, nil
	})

	gitserverClient.BranchesContainingFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, commit string) ([]string, error) {
		var branches []string
		for branch := range branchMap[commit] {
			branches = append(branches, branch)
		}

		return branches, nil
	})

	return gitserverClient
}
