package api

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/db"
)

// FindClosestDumps returns the set of dumps that can most accurately answer code intelligence
// queries for the given file. These dump IDs should be subsequently passed to invocations of
// Definitions, References, and Hover.
func (api *codeIntelAPI) FindClosestDumps(ctx context.Context, repositoryID int, commit, file string) ([]db.Dump, error) {
	candidates, err := api.db.FindClosestDumps(ctx, repositoryID, commit, file)
	if err != nil {
		return nil, err
	}

	var dumps []db.Dump
	for _, dump := range candidates {
		exists, err := api.bundleManagerClient.BundleClient(dump.ID).Exists(ctx, strings.TrimPrefix(file, dump.Root))
		if err != nil {
			return nil, err
		}
		if !exists {
			continue
		}

		dumps = append(dumps, dump)
	}

	return dumps, nil
}
