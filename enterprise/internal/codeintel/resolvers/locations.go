package resolvers

import (
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/sourcegraph/internal/api"
	codeintelapi "github.com/sourcegraph/sourcegraph/internal/codeintel/api"
)

type APILocation struct {
	RepositoryID api.RepoID
	Commit       string
	Path         string
	Range        lsp.Range
}

// TODO(efritz) - cleanup
func serializeLocations(resolvedLocations []codeintelapi.ResolvedLocation) ([]*APILocation, error) {
	var apiLocations []*APILocation
	for _, res := range resolvedLocations {
		apiLocations = append(apiLocations, &APILocation{
			RepositoryID: api.RepoID(res.Dump.RepositoryID),
			Commit:       res.Dump.Commit,
			Path:         res.Path,
			Range: lsp.Range{
				Start: lsp.Position{Line: res.Range.Start.Line, Character: res.Range.Start.Character},
				End:   lsp.Position{Line: res.Range.End.Line, Character: res.Range.End.Character},
			},
		})
	}

	return apiLocations, nil
}
