package api

import (
	"context"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

// DefintionMonikersLimit is the maximum number of definition moniker results we'll ask
// for from the bundle manager. Precise definition results should be a very small set so
// we don't have a way to page the results. This limit exists only to stop us from having
// to support a limitless query path in the bundle manager API.
const DefintionMonikersLimit = 100

// Definitions returns the list of source locations that define the symbol at the given position.
// This may include remote definitions if the remote repository is also indexed.
func (api *codeIntelAPI) Definitions(ctx context.Context, file string, line, character, uploadID int) ([]ResolvedLocation, error) {
	dump, exists, err := api.store.GetDumpByID(ctx, uploadID)
	if err != nil {
		return nil, errors.Wrap(err, "store.GetDumpByID")
	}
	if !exists {
		return nil, ErrMissingDump
	}

	pathInBundle := strings.TrimPrefix(file, dump.Root)
	return api.definitionsRaw(ctx, dump, pathInBundle, line, character)
}

func (api *codeIntelAPI) definitionsRaw(ctx context.Context, dump store.Dump, pathInBundle string, line, character int) ([]ResolvedLocation, error) {
	locations, err := api.bundleStore.Definitions(ctx, dump.ID, pathInBundle, line, character)
	if err != nil {
		if err == database.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return nil, nil
		}
		return nil, errors.Wrap(err, "bundleClient.Definitions")
	}
	if len(locations) > 0 {
		return resolveLocationsWithDump(dump, locations), nil
	}

	rangeMonikers, err := api.bundleStore.MonikersByPosition(context.Background(), dump.ID, pathInBundle, line, character)
	if err != nil {
		if err == database.ErrNotFound {
			log15.Warn("Bundle does not exist")
			return nil, nil
		}
		return nil, errors.Wrap(err, "bundleClient.MonikersByPosition")
	}

	for _, monikers := range rangeMonikers {
		for _, moniker := range monikers {
			if moniker.Kind == "import" {
				locations, _, err := lookupMoniker(api.store, api.bundleStore, dump.ID, pathInBundle, "definitions", moniker, 0, DefintionMonikersLimit)
				if err != nil {
					return nil, err
				}
				if len(locations) > 0 {
					return locations, nil
				}
			} else {
				// This symbol was not imported from another bundle. We search the definitions
				// of our own bundle in case there was a definition that wasn't properly attached
				// to a result set but did have the correct monikers attached.

				locations, _, err := api.bundleStore.MonikerResults(context.Background(), dump.ID, "definitions", moniker.Scheme, moniker.Identifier, 0, DefintionMonikersLimit)
				if err != nil {
					if err == database.ErrNotFound {
						log15.Warn("Bundle does not exist")
						return nil, nil
					}
					return nil, errors.Wrap(err, "bundleClient.MonikerResults")
				}
				if len(locations) > 0 {
					return resolveLocationsWithDump(dump, locations), nil
				}
			}
		}
	}

	return nil, nil
}

func (api *codeIntelAPI) definitionRaw(ctx context.Context, dump store.Dump, pathInBundle string, line, character int) (ResolvedLocation, bool, error) {
	resolved, err := api.definitionsRaw(ctx, dump, pathInBundle, line, character)
	if err != nil || len(resolved) == 0 {
		return ResolvedLocation{}, false, errors.Wrap(err, "api.definitionsRaw")
	}

	return resolved[0], true, nil
}
