package commitgraph

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// updater handles updating the commit graph of all repositories marked as dirty.
type updater struct {
	uploadSvc UploadService
	logger    log.Logger
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

func (u *updater) HandleError(err error) {
	var multi errors.MultiError
	switch {
	case errors.HasType(err, store.BackfillIncompleteError{}):
		u.logger.Warn("non-fatal issue encountered", log.Error(err))
	case errors.As(err, &multi):
		var allIncompleteBackfill bool
		for _, err := range multi.Errors() {
			allIncompleteBackfill = allIncompleteBackfill && errors.HasType(err, store.BackfillIncompleteError{})
		}

		if allIncompleteBackfill {
			u.logger.Warn("multiple non-fatal issues encountered", log.Error(err))
			return
		}
		fallthrough
	default:
		u.logger.Error("error updating commit graphs for repositories marked as dirty", log.Error(err))
	}
}
