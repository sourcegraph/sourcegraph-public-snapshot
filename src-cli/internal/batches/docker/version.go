package docker

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// CheckVersion is used to check if docker is running. We use this method instead of
// checkExecutable (https://sourcegraph.com/github.com/sourcegraph/src-cli@main/-/blob/cmd/src/batch_common.go?L547%3A6=&popover=pinned)
// to prevent a case where docker commands take too long and results in `src-cli` freezing for some users.
func CheckVersion(ctx context.Context) error {
	_, err := executeFastCommand(ctx, "version")
	if err != nil {
		return errors.Newf(
			"failed to execute \"docker version\":\n\t%s\n\n'src batch' requires \"docker\" to be available.",
			err,
		)
	}

	return nil
}
