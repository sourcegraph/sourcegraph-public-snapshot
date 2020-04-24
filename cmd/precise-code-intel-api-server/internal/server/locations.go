package server

import (
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/api"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/bundles"
)

type APILocation struct {
	RepositoryID int           `json:"repositoryId"`
	Commit       string        `json:"commit"`
	Path         string        `json:"path"`
	Range        bundles.Range `json:"range"`
}

func serializeLocations(resolvedLocations []api.ResolvedLocation) ([]APILocation, error) {
	var apiLocations []APILocation
	for _, res := range resolvedLocations {
		apiLocations = append(apiLocations, APILocation{
			RepositoryID: res.Dump.RepositoryID,
			Commit:       res.Dump.Commit,
			Path:         res.Path,
			Range:        res.Range,
		})
	}

	return apiLocations, nil
}
