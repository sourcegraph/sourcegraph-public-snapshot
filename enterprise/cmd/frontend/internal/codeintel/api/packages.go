package api

import (
	"context"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type ResolvedPackage struct {
	lsifstore.Package
	Dump store.Dump
}

// Packages returns the packages defined in the given path prefix.
func (api *CodeIntelAPI) Packages(ctx context.Context, prefix string, uploadID, limit, offset int) (_ []ResolvedPackage, _ int, err error) {
	ctx, endObservation := api.operations.packages.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("prefix", prefix),
		log.Int("uploadID", uploadID),
		log.Int("limit", limit),
		log.Int("offset", offset),
	}})
	defer endObservation(1, observation.Args{})

	dump, exists, err := api.dbStore.GetDumpByID(ctx, uploadID)
	if err != nil {
		return nil, 0, errors.Wrap(err, "store.GetDumpByID")
	}
	if !exists {
		return nil, 0, ErrMissingDump
	}

	pathInBundle := strings.TrimPrefix(prefix, dump.Root)
	packages, totalCount, err := api.lsifStore.Packages(ctx, dump.ID, pathInBundle, offset, limit)
	if err != nil {
		if err == lsifstore.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return nil, 0, nil
		}
		return nil, 0, errors.Wrap(err, "bundleClient.Packages")
	}

	return resolvePackagesWithDump(dump, packages), totalCount, nil
}

func resolvePackagesWithDump(dump store.Dump, packages []lsifstore.Package) []ResolvedPackage {
	var resolvedPackages []ResolvedPackage
	for _, pkg := range packages {
		resolvedPackages = append(resolvedPackages, ResolvedPackage{
			Dump:    dump,
			Package: pkg,
		})
	}
	return resolvedPackages
}
