package api

import (
	"context"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

// FindClosestDumps returns the set of dumps that can most accurately answer code intelligence
// queries for the given path. If exactPath is true, then only dumps that definitely contain the
// exact document path are returned. Otherwise, dumps containing any document for which the given
// path is a prefix are returned. These dump IDs should be subsequently passed to invocations of
// Definitions, References, and Hover.
func (api *codeIntelAPI) FindClosestDumps(ctx context.Context, repositoryID int, commit, path string, exactPath bool, indexer string) ([]store.Dump, error) {
	// See if we know about this commit. If not, we need to update our commits table
	// and the visibility of the dumps in this repository.
	if err := api.updateCommitsAndVisibility(ctx, repositoryID, commit); err != nil {
		return nil, err
	}

	// The parameters exactPath and rootMustEnclosePath align here: if we're looking for dumps
	// that can answer queries for a directory (e.g. diagnostics), we want any dump that happens
	// to intersect the target directory. If we're looking for dumps that can answer queries for
	// a single file, then we need a dump with a root that properly encloses that file.
	candidates, err := api.store.FindClosestDumps(ctx, repositoryID, commit, path, exactPath, indexer)
	if err != nil {
		return nil, errors.Wrap(err, "store.FindClosestDumps")
	}

	var dumps []store.Dump
	for _, dump := range candidates {
		// TODO(efritz) - ensure there's a valid document path
		// for the other condition. This should probably look like
		// an additional parameter on the following exists query.
		if exactPath {
			exists, err := api.bundleManagerClient.BundleClient(dump.ID).Exists(ctx, strings.TrimPrefix(path, dump.Root))
			if err != nil {
				if err == bundles.ErrNotFound {
					log15.Warn("Bundle does not exist")
					return nil, nil
				}
				return nil, errors.Wrap(err, "bundleManagerClient.BundleClient")
			}
			if !exists {
				continue
			}
		}

		dumps = append(dumps, dump)
	}

	return dumps, nil
}

// updateCommits updates the lsif_commits table with the current data known to gitserver, then updates the
// visibility of all dumps for the given repository.
func (api *codeIntelAPI) updateCommitsAndVisibility(ctx context.Context, repositoryID int, commit string) error {
	commitExists, err := api.store.HasCommit(ctx, repositoryID, commit)
	if err != nil {
		return errors.Wrap(err, "store.HasCommit")
	}
	if commitExists {
		return nil
	}

	newCommits, err := api.gitserverClient.CommitsNear(ctx, api.store, repositoryID, commit)
	if err != nil {
		return errors.Wrap(err, "gitserverClient.CommitsNear")
	}
	if err := api.store.UpdateCommits(ctx, repositoryID, newCommits); err != nil {
		return errors.Wrap(err, "store.UpdateCommits")
	}

	tipCommit, err := api.gitserverClient.Head(ctx, api.store, repositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserverClient.Head")
	}
	if err := api.store.UpdateDumpsVisibleFromTip(ctx, repositoryID, tipCommit); err != nil {
		return errors.Wrap(err, "store.UpdateDumpsVisibleFromTip")
	}

	return nil
}
