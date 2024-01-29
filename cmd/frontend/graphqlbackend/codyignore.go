package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codyignore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *GitTreeEntryResolver) AllowedForCodyContext(ctx context.Context) (bool, error) {
	b, ok := r.ToGitBlob()
	if !ok {
		return false, errors.New("cody ignore is not supported for directories")
	}
	commit := b.Commit()
	repoName := commit.gitRepo
	gitSHA := api.CommitID(commit.oid)
	filePath := b.Path()
	ignored, err := codyignore.NewService(r.gitserverClient).IsIgnored(ctx, repoName, gitSHA, filePath)
	if err != nil {
		// Fail-open: Return error if codyignore file could not be read / repo unavailable.
		return true, err
	}
	return !ignored, nil
}
