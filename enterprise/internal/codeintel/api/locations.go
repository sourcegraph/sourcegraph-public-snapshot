package api

import (
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type ResolvedLocation struct {
	Dump  store.Dump    `json:"dump"`
	Path  string        `json:"path"`
	Range bundles.Range `json:"range"`
}

func sliceLocations(locations []bundles.Location, lo, hi int) []bundles.Location {
	if lo >= len(locations) {
		return nil
	}
	if hi >= len(locations) {
		hi = len(locations)
	}
	return locations[lo:hi]
}

func filterLocationsWithPath(path string, locations []bundles.Location) []bundles.Location {
	filtered := make([]bundles.Location, 0, len(locations))
	for _, l := range locations {
		if l.Path == path {
			filtered = append(filtered, l)
		}
	}

	return filtered
}

func resolveLocationsWithDump(dump store.Dump, locations []bundles.Location) []ResolvedLocation {
	resolved := make([]ResolvedLocation, 0, len(locations))
	for _, location := range locations {
		resolved = append(resolved, ResolvedLocation{
			Dump:  dump,
			Path:  dump.Root + location.Path,
			Range: location.Range,
		})
	}

	return resolved
}
