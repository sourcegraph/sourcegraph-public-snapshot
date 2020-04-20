package api

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/bundles"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/db"
)

func (api *codeIntelAPI) Definitions(file string, line, character, uploadID int) ([]ResolvedLocation, error) {
	dump, exists, err := api.db.GetDumpByID(context.Background(), uploadID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrMissingDump
	}

	pathInBundle := strings.TrimPrefix(file, dump.Root)
	bundleClient := api.bundleManagerClient.BundleClient(dump.ID)
	return api.definitionsRaw(dump, bundleClient, pathInBundle, line, character)
}

func (api *codeIntelAPI) definitionsRaw(dump db.Dump, bundleClient bundles.BundleClient, pathInBundle string, line, character int) ([]ResolvedLocation, error) {
	locations, err := bundleClient.Definitions(context.Background(), pathInBundle, line, character)
	if err != nil {
		return nil, err
	}
	if len(locations) > 0 {
		return resolveLocationsWithDump(dump, locations), nil
	}

	rangeMonikers, err := bundleClient.MonikersByPosition(context.Background(), pathInBundle, line, character)
	if err != nil {
		return nil, err
	}

	for _, monikers := range rangeMonikers {
		for _, moniker := range monikers {
			if moniker.Kind == "import" {
				locations, _, err := lookupMoniker(api.db, api.bundleManagerClient, dump.ID, pathInBundle, "definitions", moniker, 0, 0)
				if err != nil {
					return nil, err
				}
				if len(locations) > 0 {
					return locations, nil
				}
			} else {
				// This symbol was not imported from another database. We search the definitions
				// table of our own database in case there was a definition that wasn't properly
				// attached to a result set but did have the correct monikers attached.

				locations, _, err := bundleClient.MonikerResults(context.Background(), "definitions", moniker.Scheme, moniker.Identifier, 0, 0)
				if err != nil {
					return nil, err
				}
				if len(locations) > 0 {
					return resolveLocationsWithDump(dump, locations), nil
				}
			}
		}
	}

	return nil, nil
}

func (api *codeIntelAPI) definitionRaw(dump db.Dump, bundleClient bundles.BundleClient, pathInBundle string, line, character int) (ResolvedLocation, bool, error) {
	resolved, err := api.definitionsRaw(dump, bundleClient, pathInBundle, line, character)
	if err != nil || len(resolved) == 0 {
		return ResolvedLocation{}, false, err
	}

	return resolved[0], true, nil
}
