package api

import (
	"context"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
)

// FindClosestDumps returns the set of dumps that can most accurately answer code intelligence
// queries for the given file. These dump IDs should be subsequently passed to invocations of
// Definitions, References, and Hover.
func (api *codeIntelAPI) FindClosestDumps(ctx context.Context, repositoryID int, commit, file string) ([]db.Dump, error) {
	// See if we know about this commit. If not, we need to update our commits table
	// and the visibility of the dumps in this repository.
	if err := api.updateCommitsAndVisibility(ctx, repositoryID, commit); err != nil {
		return nil, err
	}

	candidates, err := api.db.FindClosestDumps(ctx, repositoryID, commit, file)
	if err != nil {
		return nil, errors.Wrap(err, "db.FindClosestDumps")
	}

	var dumps []db.Dump
	for _, dump := range candidates {
		exists, err := api.bundleManagerClient.BundleClient(dump.ID).Exists(ctx, strings.TrimPrefix(file, dump.Root))
		if err != nil {
			if err == client.ErrNotFound {
				log15.Warn("Bundle does not exist")
				return nil, nil
			}
			return nil, errors.Wrap(err, "bundleManagerClient.BundleClient")
		}
		if !exists {
			continue
		}

		dumps = append(dumps, dump)
	}

	return dumps, nil
}

// updateCommits updates the lsif_commits table with the current data known to gitserver, then updates the
// visibility of all dumps for the given repository.
func (api *codeIntelAPI) updateCommitsAndVisibility(ctx context.Context, repositoryID int, commit string) error {
	commitExists, err := api.db.HasCommit(ctx, repositoryID, commit)
	if err != nil {
		return errors.Wrap(err, "db.HasCommit")
	}
	if commitExists {
		return nil
	}

	newCommits, err := api.gitserverClient.CommitsNear(ctx, api.db, repositoryID, commit)
	if err != nil {
		return errors.Wrap(err, "gitserverClient.CommitsNear")
	}
	if err := api.db.UpdateCommits(ctx, repositoryID, newCommits); err != nil {
		return errors.Wrap(err, "db.UpdateCommits")
	}

	tipCommit, err := api.gitserverClient.Head(ctx, api.db, repositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserverClient.Head")
	}
	if err := api.db.UpdateDumpsVisibleFromTip(ctx, repositoryID, tipCommit); err != nil {
		return errors.Wrap(err, "db.UpdateDumpsVisibleFromTip")
	}

	return nil
}
