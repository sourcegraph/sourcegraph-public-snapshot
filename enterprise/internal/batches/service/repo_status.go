package service

import (
	"context"
	"fmt"
	"os"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// getRepoStatus returns the repo status for the given repo HEAD, recalculating
// it if necessary.
//
// 🚨 SECURITY: calling code is responsible for validating that the given repo
// can be seen by the current user.
func getRepoStatus(ctx context.Context, tx *store.Store, client gitserver.Client, repo *types.Repo) (_ *btypes.RepoStatus, err error) {
	traceTitle := fmt.Sprintf("RepoID: %q", repo.ID)
	tr, ctx := trace.New(ctx, "getRepoStatus", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// First we need to know the actual commit of the repo HEAD.
	head, ok, err := client.Head(ctx, repo.Name, authz.DefaultSubRepoPermsChecker)
	if err != nil {
		return nil, errors.Wrap(err, "getting HEAD commit")
	}
	if !ok {
		return nil, errors.New("no HEAD commit")
	}

	// Now we can see if it's in the cache.
	rs, err := tx.GetRepoStatus(ctx, repo.ID, head)
	if err != nil && err != store.ErrNoResults {
		return nil, errors.Wrap(err, "getting repo status from cache")
	} else if err == nil {
		return rs, nil
	}

	// The unlucky case lands here, where we have to ask gitserver.
	rs = &btypes.RepoStatus{RepoID: repo.ID, Commit: head}
	rs.Ignored, err = hasBatchIgnoreFile(ctx, client, repo.Name, api.CommitID(head))
	if err != nil {
		return nil, errors.Wrap(err, "looking for batch ignore file")
	}

	// Let's update the cache.
	if err := tx.UpsertRepoStatus(ctx, rs); err != nil {
		return nil, errors.Wrap(err, "upserting repo status to cache")
	}

	return rs, nil
}

const batchIgnoreFilePath = ".batchignore"

func hasBatchIgnoreFile(ctx context.Context, client gitserver.Client, repoName api.RepoName, commit api.CommitID) (bool, error) {
	stat, err := client.Stat(ctx, authz.DefaultSubRepoPermsChecker, repoName, commit, batchIgnoreFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !stat.Mode().IsRegular() {
		return false, errors.Errorf("not a blob: %q", batchIgnoreFilePath)
	}
	return true, nil
}
