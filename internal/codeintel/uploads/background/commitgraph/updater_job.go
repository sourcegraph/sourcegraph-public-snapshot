package commitgraph

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// HandleUpdateDirtyRepositories periodically re-calculates the commit and upload visibility graph for repositories
// that are marked as dirty by the worker process. This is done out-of-band from the rest of
// the upload processing as it is likely that we are processing multiple uploads concurrently
// for the same repository and should not repeat the work since the last calculation performed
// will always be the one we want.

// HandleUpdater checks for dirty repositories and invokes the underlying updater on each one.
func (u *updater) HandleUpdateDirtyRepositories(ctx context.Context) error {
	err := u.uploadSvc.UpdateDirtyRepositories(ctx, ConfigInst.MaxAgeForNonStaleBranches, ConfigInst.MaxAgeForNonStaleTags)
	if err != nil {
		return errors.Wrap(err, "uploadSvc.UpdateUploadsVisibleToCommits")
	}

	return nil
}
