package api

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client_types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

func lookupMoniker(
	store store.Store,
	bundleStore database.Database,
	dumpID int,
	path string,
	modelType string,
	moniker bundles.MonikerData,
	skip int,
	take int,
) ([]ResolvedLocation, int, error) {
	if moniker.PackageInformationID == "" {
		return nil, 0, nil
	}

	pid, _, err := bundleStore.PackageInformation(context.Background(), dumpID, path, moniker.PackageInformationID)
	if err != nil {
		if err == database.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return nil, 0, nil
		}
		return nil, 0, errors.Wrap(err, "bundleStore.BundleClient")
	}

	dump, exists, err := store.GetPackage(context.Background(), moniker.Scheme, pid.Name, pid.Version)
	if err != nil || !exists {
		return nil, 0, errors.Wrap(err, "store.GetPackage")
	}

	locations, count, err := bundleStore.MonikerResults(context.Background(), dump.ID, modelType, moniker.Scheme, moniker.Identifier, skip, take)
	if err != nil {
		if err == database.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return nil, 0, nil
		}
		return nil, 0, errors.Wrap(err, "bundleStore.BundleClient")
	}

	return resolveLocationsWithDump(dump, locations), count, nil
}
