package commitgraph

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type updater struct {
	uploadSvc UploadService
}

var (
	_ goroutine.Handler      = &updater{}
	_ goroutine.ErrorHandler = &updater{}
)

// Handle periodically re-calculates the commit and upload visibility graph for repositories
// that are marked as dirty by the worker process. This is done out-of-band from the rest of
// the upload processing as it is likely that we are processing multiple uploads concurrently
// for the same repository and should not repeat the work since the last calculation performed
// will always be the one we want.

// Handle checks for dirty repositories and invokes the underlying updater on each one.
func (u *updater) Handle(ctx context.Context) error {
	err := u.uploadSvc.UpdateDirtyRepositories(ctx, ConfigInst.MaxAgeForNonStaleBranches, ConfigInst.MaxAgeForNonStaleTags)
	if err != nil {
		return errors.Wrap(err, "uploadSvc.UpdateDirtyRepositories")
	}

	return nil
}

func (u *updater) HandleError(err error) {}
