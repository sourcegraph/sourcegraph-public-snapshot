package git

import (
	"bytes"
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GetDefaultBranch returns the name of the default branch and the commit it's
// currently at from the given repository.
//
// If the repository is empty or currently being cloned, empty values and no
// error are returned.
func GetDefaultBranch(ctx context.Context, db database.DB, repo api.RepoName) (refName string, commit api.CommitID, err error) {
	if Mocks.GetDefaultBranch != nil {
		return Mocks.GetDefaultBranch(repo)
	}
	return getDefaultBranch(ctx, db, repo, false)
}

// GetDefaultBranchShort returns the short name of the default branch for the
// given repository and the commit it's currently at. A short name would return
// something like `main` instead of `refs/heads/main`.
//
// If the repository is empty or currently being cloned, empty values and no
// error are returned.
func GetDefaultBranchShort(ctx context.Context, db database.DB, repo api.RepoName) (refName string, commit api.CommitID, err error) {
	if Mocks.GetDefaultBranchShort != nil {
		return Mocks.GetDefaultBranchShort(repo)
	}
	return getDefaultBranch(ctx, db, repo, true)
}

// GetDefaultBranch returns the name of the default branch and the commit it's
// currently at from the given repository.
//
// If the repository is empty or currently being cloned, empty values and no
// error are returned.
func getDefaultBranch(ctx context.Context, db database.DB, repo api.RepoName, short bool) (refName string, commit api.CommitID, err error) {
	args := []string{"symbolic-ref", "HEAD"}
	if short {
		args = append(args, "--short")
	}
	refBytes, _, exitCode, err := execSafe(ctx, db, repo, args)
	refName = string(bytes.TrimSpace(refBytes))

	if err == nil && exitCode == 0 {
		// Check that our repo is not empty
		commit, err = gitserver.NewClient(db).ResolveRevision(ctx, repo, "HEAD", gitserver.ResolveRevisionOptions{NoEnsureRevision: true})
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
