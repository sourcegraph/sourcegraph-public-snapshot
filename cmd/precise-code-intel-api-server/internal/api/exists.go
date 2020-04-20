package api

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/db"
)

func (api *codeIntelAPI) FindClosestDumps(repositoryID int, commit, file string) ([]db.Dump, error) {
	candidates, err := api.db.FindClosestDumps(context.Background(), repositoryID, commit, file)
	if err != nil {
		return nil, err
	}

	var dumps []db.Dump
	for _, dump := range candidates {
		exists, err := api.bundleManagerClient.BundleClient(dump.ID).Exists(context.Background(), strings.TrimPrefix(file, dump.Root))
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
