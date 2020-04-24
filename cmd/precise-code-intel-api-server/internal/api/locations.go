package api

import (
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/bundles"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/db"
)

type ResolvedLocation struct {
	Dump  db.Dump       `json:"dump"`
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

func resolveLocationsWithDump(dump db.Dump, locations []bundles.Location) []ResolvedLocation {
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
