package api

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

func lookupMoniker(
	dbStore DBStore,
	lsifStore LSIFStore,
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

	pid, _, err := lsifStore.PackageInformation(context.Background(), dumpID, path, string(moniker.PackageInformationID))
	if err != nil {
		if err == lsifstore.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return nil, 0, nil
		}
		return nil, 0, errors.Wrap(err, "lsifStore.BundleClient")
	}

	dump, exists, err := dbStore.GetPackage(context.Background(), moniker.Scheme, pid.Name, pid.Version)
	if err != nil || !exists {
		return nil, 0, errors.Wrap(err, "store.GetPackage")
	}

	locations, count, err := lsifStore.MonikerResults(context.Background(), dump.ID, modelType, moniker.Scheme, moniker.Identifier, skip, take)
	if err != nil {
		if err == lsifstore.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return nil, 0, nil
		}
		return nil, 0, errors.Wrap(err, "lsifStore.BundleClient")
	}

	return resolveLocationsWithDump(dump, locations), count, nil
}
