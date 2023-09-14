package main

import (
	"context"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/dev/codeintel-qa/internal"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func checkInstanceState(ctx context.Context) error {
	if diff, err := instanceStateDiff(ctx); err != nil {
		return err
	} else if diff != "" {
		return errors.Newf("unexpected instance state: %s", diff)
	}

	return nil
}

func instanceStateDiff(ctx context.Context) (string, error) {
	extensionAndCommitsByRepo, err := internal.ExtensionAndCommitsByRepo(indexDir)
	if err != nil {
		return "", err
	}
	expectedCommitAndRootsByRepo := map[string][]CommitAndRoot{}
	for repoName, extensionAndCommits := range extensionAndCommitsByRepo {
		commitAndRoots := make([]CommitAndRoot, 0, len(extensionAndCommits))
		for _, e := range extensionAndCommits {
			root := strings.ReplaceAll(e.Root, "_", "/")
			if root == "/" {
				root = ""
			}

			commitAndRoots = append(commitAndRoots, CommitAndRoot{e.Commit, root})
		}

		expectedCommitAndRootsByRepo[internal.MakeTestRepoName(repoName)] = commitAndRoots
	}

	uploadedCommitAndRootsByRepo, err := queryPreciseIndexes(ctx)
	if err != nil {
		return "", err
	}

	for _, commitAndRoots := range uploadedCommitAndRootsByRepo {
		sortCommitAndRoots(commitAndRoots)
	}
	for _, commitAndRoots := range expectedCommitAndRootsByRepo {
		sortCommitAndRoots(commitAndRoots)
	}

	if allowDirtyInstance {
		// We allow other upload records to exist on the instance, but we still
		// need to ensure that the set of uploads we require for the tests remain
		// accessible on the instance. Here, we remove references to uploads and
		// commits that don't exist in our expected list, and check only that we
		// have a superset of our expected state.

		for repoName, commitAndRoots := range uploadedCommitAndRootsByRepo {
			if expectedCommits, ok := expectedCommitAndRootsByRepo[repoName]; !ok {
				delete(uploadedCommitAndRootsByRepo, repoName)
			} else {
				filtered := commitAndRoots[:0]
				for _, commitAndRoot := range commitAndRoots {
					found := false
					for _, ex := range expectedCommits {
						if ex.Commit == commitAndRoot.Commit && ex.Root == commitAndRoot.Root {
							found = true
							break
						}
					}
					if !found {
						filtered = append(filtered, commitAndRoot)
					}
				}

				uploadedCommitAndRootsByRepo[repoName] = filtered
			}
		}
	}

	return cmp.Diff(expectedCommitAndRootsByRepo, uploadedCommitAndRootsByRepo), nil
}

func sortCommitAndRoots(commitAndRoots []CommitAndRoot) {
	sort.Slice(commitAndRoots, func(i, j int) bool {
		if commitAndRoots[i].Commit != commitAndRoots[j].Commit {
			return commitAndRoots[i].Commit < commitAndRoots[j].Commit
		}

		return commitAndRoots[i].Root < commitAndRoots[j].Root
	})
}
