package executor

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/batches"
)

func CheckCache(ctx context.Context, cache ExecutionCache, clearCache bool, features batches.FeatureFlags, task *Task) (specs []*batches.ChangesetSpec, found bool, err error) {
	// Check if the task is cached.
	cacheKey := task.cacheKey()
	if clearCache {
		if err = cache.Clear(ctx, cacheKey); err != nil {
			return specs, false, errors.Wrapf(err, "clearing cache for %q", task.Repository.Name)
		}

		return specs, false, nil
	}

	var result executionResult
	result, found, err = cache.Get(ctx, cacheKey)
	if err != nil {
		return specs, false, errors.Wrapf(err, "checking cache for %q", task.Repository.Name)
	}

	if !found {
		return specs, false, nil
	}

	// If the cached result resulted in an empty diff, we don't need to
	// add it to the list of specs that are displayed to the user and
	// send to the server. Instead, we can just report that the task is
	// complete and move on.
	if result.Diff == "" {
		return specs, true, nil
	}

	specs, err = createChangesetSpecs(task, result, features)
	if err != nil {
		return specs, false, err
	}

	return specs, true, nil
}
