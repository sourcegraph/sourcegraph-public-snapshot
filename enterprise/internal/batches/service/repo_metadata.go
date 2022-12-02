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

// getRepoMetadata returns the repo metadata for the given repo, recalculating
// it if necessary.
//
// 🚨 SECURITY: calling code is responsible for validating that the given repo
// can be seen by the current user; although GetRepoMetadata performs an authz
// check as part of its query, a failed authz check will still result in the
// gitserver being hit for the repo, which could expose a side channel of
// information about the existence or not of the given repo.
func getRepoMetadata(ctx context.Context, tx *store.Store, client gitserver.Client, repo *types.Repo) (*btypes.RepoMetadata, error) {
	meta, err := tx.GetRepoMetadata(ctx, repo.ID)
	if err != nil && err != store.ErrNoResults {
		return nil, errors.Wrap(err, "getting repo metadata")
	}

	// Check if we need to refresh the metadata.
	if (err == store.ErrNoResults) ||
		(!meta.UpdatedAt.IsZero() && meta.UpdatedAt.Before(repo.UpdatedAt)) ||
		meta.UpdatedAt.Before(repo.CreatedAt) {
		meta, err = CalculateRepoMetadata(ctx, client, CalculateRepoMetadataOpts{
			ID:   repo.ID,
			Name: repo.Name,
		})
		if err != nil {
			return nil, errors.Wrap(err, "refreshing repo metadata")
		}

		if err := tx.UpsertRepoMetadata(ctx, meta); err != nil {
			return nil, errors.Wrap(err, "upserting repo metadata")
		}
	}

	return meta, nil
}

const batchIgnoreFilePath = ".batchignore"

type CalculateRepoMetadataOpts struct {
	ID   api.RepoID
	Name api.RepoName
}

// CalculateRepoMetadata calculates and persists the repo metadata for the given
// repo.
//
// 🚨 SECURITY: calling code is responsible for validating that the given repo
// can be updated by the given user.
func CalculateRepoMetadata(ctx context.Context, client gitserver.Client, opts CalculateRepoMetadataOpts) (meta *btypes.RepoMetadata, err error) {
	traceTitle := fmt.Sprintf("RepoID: %q", opts.ID)
	tr, ctx := trace.New(ctx, "calculateRepoMetadata", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// Figure out the head commit, since we need it to stat the file.
	commit, ok, err := client.Head(ctx, opts.Name, authz.DefaultSubRepoPermsChecker)
	if err != nil {
		return nil, errors.Wrapf(err, "resolving head commit in repo %q", string(opts.Name))
	}
	if !ok {
		return nil, errors.Newf("no head commit for repo %q", string(opts.Name))
	}

	meta = &btypes.RepoMetadata{RepoID: opts.ID, Ignored: false}
	meta.Ignored, err = hasBatchIgnoreFile(ctx, client, opts.Name, api.CommitID(commit))
	if err != nil {
		return nil, errors.Wrapf(err, "looking for %s file in repo %q", batchIgnoreFilePath, string(opts.Name))
	}

	return meta, nil
}

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
