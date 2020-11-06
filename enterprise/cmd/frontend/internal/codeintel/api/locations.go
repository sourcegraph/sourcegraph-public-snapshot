package api

import (
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

type ResolvedLocation struct {
	Dump  store.Dump
	Path  string
	Range lsifstore.Range
}

func sliceLocations(locations []lsifstore.Location, lo, hi int) []lsifstore.Location {
	if lo >= len(locations) {
		return nil
	}
	if hi >= len(locations) {
		hi = len(locations)
	}
	return locations[lo:hi]
}

func resolveLocationsWithDump(dump store.Dump, locations []lsifstore.Location) []ResolvedLocation {
	var resolvedLocations []ResolvedLocation
	for _, location := range locations {
		resolvedLocations = append(resolvedLocations, ResolvedLocation{
			Dump:  dump,
			Path:  dump.Root + location.Path,
			Range: location.Range,
		})
	}

	return resolvedLocations
}
