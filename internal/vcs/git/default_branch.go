package git

import (
	"bytes"
	"context"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

// GetDefaultBranch returns the name of the default branch and the commit it's
// currently at from the given repository.
//
// If the repository is empty or currently being cloned, empty values and no
// error are returned.
func GetDefaultBranch(ctx context.Context, repo api.RepoName) (refName string, commit api.CommitID, err error) {
	if Mocks.GetDefaultBranch != nil {
		return Mocks.GetDefaultBranch(repo)
	}

	refBytes, _, exitCode, err := ExecSafe(ctx, repo, []string{"symbolic-ref", "HEAD"})
	refName = string(bytes.TrimSpace(refBytes))

	if err == nil && exitCode == 0 {
		// Check that our repo is not empty
		commit, err = ResolveRevision(ctx, repo, "HEAD", ResolveRevisionOptions{NoEnsureRevision: true})
	}

	// If we fail to get the default branch due to cloning or being empty, we return nothing.
	if err != nil {
		if gitdomain.IsCloneInProgress(err) || errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
			return "", "", nil
		}
		return "", "", err
	}

	return refName, commit, nil
}
