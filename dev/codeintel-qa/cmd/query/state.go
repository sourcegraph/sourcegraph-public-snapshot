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
	uploadedCommitsByRepo, err := queryUploads(ctx)
	if err != nil {
		return "", err
	}
	for _, commits := range uploadedCommitsByRepo {
		sort.Strings(commits)
	}

	commitsByRepo, err := internal.CommitsByRepo(indexDir)
	if err != nil {
		return "", err
	}

	expectedCommitsByRepo := map[string][]string{}
	for repoName, commits := range commitsByRepo {
		sort.Strings(commits)
		expectedCommitsByRepo[internal.MakeTestRepoName(repoName)] = commits
	}

	return cmp.Diff(expectedCommitsByRepo, uploadedCommitsByRepo), nil
}
