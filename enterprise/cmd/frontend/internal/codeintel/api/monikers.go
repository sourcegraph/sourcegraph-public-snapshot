package api

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

func lookupMoniker(
	ctx context.Context,
	dbStore DBStore,
	lsifStore LSIFStore,
	gitserverClient GitserverClient,
	dumpID int,
	path string,
	modelType string,
	moniker lsifstore.MonikerData,
	skip int,
	take int,
) ([]ResolvedLocation, int, error) {
	if moniker.PackageInformationID == "" {
		return nil, 0, nil
	}

	pid, _, err := lsifStore.PackageInformation(ctx, dumpID, path, string(moniker.PackageInformationID))
	if err != nil {
		if err == lsifstore.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return nil, 0, nil
		}
		return nil, 0, errors.Wrap(err, "lsifStore.BundleClient")
	}

	dump, exists, err := dbStore.GetPackage(ctx, moniker.Scheme, pid.Name, pid.Version)
	if err != nil || !exists {
		return nil, 0, errors.Wrap(err, "store.GetPackage")
	}

	commitExists, err := gitserverClient.CommitExists(ctx, dump.RepositoryID, dump.Commit)
	if err != nil {
		return nil, 0, errors.Wrap(err, "gitserverClient.CommitExists")
	}
	if !commitExists {
		return nil, 0, nil
	}

	locations, count, err := lsifStore.MonikerResults(ctx, dump.ID, modelType, moniker.Scheme, moniker.Identifier, skip, take)
	if err != nil {
		if err == lsifstore.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return nil, 0, nil
		}
		return nil, 0, errors.Wrap(err, "lsifStore.BundleClient")
	}

	return resolveLocationsWithDump(dump, locations), count, nil
}
