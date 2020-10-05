package gitserver

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

// Text returns the contents of a file in a particular commit of a repository.
func Text(ctx context.Context, store store.Store, repositoryID int, commit, file string) ([]byte, error) {
	out, err := execGitCommand(ctx, store, repositoryID, "show", fmt.Sprintf("%s:%s", commit, file))
	if err != nil {
		return nil, err
	}

	return []byte(out), err
}
