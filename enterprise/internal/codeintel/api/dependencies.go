package api

import (
	"context"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type ResolvedDependency struct {
	Dump       store.Dump
	Dependency bundles.PackageInformationData
}

// Dependencies returns the dependencies for documents with the given path prefix.
func (api *codeIntelAPI) Dependencies(ctx context.Context, prefix string, uploadID, limit, offset int) ([]ResolvedDependency, int, error) {
	dump, exists, err := api.store.GetDumpByID(ctx, uploadID)
	if err != nil {
		return nil, 0, errors.Wrap(err, "store.GetDumpByID")
	}
	if !exists {
		return nil, 0, ErrMissingDump
	}

	pathInBundle := strings.TrimPrefix(prefix, dump.Root)
	bundleClient := api.bundleManagerClient.BundleClient(dump.ID)

	packageInformations, totalCount, err := bundleClient.PackageInformations(ctx, pathInBundle, offset, limit)
	if err != nil {
		if err == bundles.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return nil, 0, nil
		}
		return nil, 0, errors.Wrap(err, "bundleClient.PackageInformations")
	}

	return resolveDependenciesWithDump(dump, packageInformations), totalCount, nil
}

func resolveDependenciesWithDump(dump store.Dump, packageInformations []bundles.PackageInformationData) []ResolvedDependency {
	var resolvedDependencies []ResolvedDependency
	for _, packageInformation := range packageInformations {
		resolvedDependencies = append(resolvedDependencies, ResolvedDependency{
			Dump:       dump,
			Dependency: packageInformation,
		})
	}

	return resolvedDependencies
}
