package main

import (
	"context"
	"sort"

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
	commitsByRepo, err := internal.CommitsByRepo(indexDir)
	if err != nil {
		return "", err
	}
	expectedCommitsByRepo := map[string][]string{}
	for repoName, commits := range commitsByRepo {
		sort.Strings(commits)
		expectedCommitsByRepo[internal.MakeTestRepoName(repoName)] = commits
	}

	uploadedCommitsByRepo, err := queryUploads(ctx)
	if err != nil {
		return "", err
	}
	for _, commits := range uploadedCommitsByRepo {
		sort.Strings(commits)
	}
	if allowDirtyInstance {
		// We allow other upload records to exist on the instance, but we still
		// need to ensure that the set of uploads we require for the tests remain
		// accessible on the instance. Here, we remove references to uploads and
		// commits that don't exist in our expected list, and check only that we
		// have a superset of our expected state.

		for repoName, commits := range uploadedCommitsByRepo {
			if expectedCommits, ok := expectedCommitsByRepo[repoName]; !ok {
				delete(uploadedCommitsByRepo, repoName)
			} else {
				filtered := commits[:0]
				for _, commit := range commits {
					if i := sort.SearchStrings(expectedCommits, commit); i < len(expectedCommits) && expectedCommits[i] == commit {
						filtered = append(filtered, commit)
					}

					uploadedCommitsByRepo[repoName] = filtered
				}
			}
		}
	}

	return cmp.Diff(expectedCommitsByRepo, uploadedCommitsByRepo), nil
}
