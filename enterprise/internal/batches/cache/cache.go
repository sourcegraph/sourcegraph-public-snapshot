package cache

import (
	"sort"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution/cache"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
)

func KeyForWorkspace(batchChangeAttributes *template.BatchChangeAttributes, r batcheslib.Repository, path string, onlyFetchWorkspace bool, steps []batcheslib.Step) cache.ExecutionKey {
	sort.Strings(r.FileMatches)

	executionKey := cache.ExecutionKey{
		Repository:            r,
		Path:                  path,
		OnlyFetchWorkspace:    onlyFetchWorkspace,
		Steps:                 steps,
		BatchChangeAttributes: batchChangeAttributes,
	}
	return executionKey
}

func ChangesetSpecsFromCache(spec *batcheslib.BatchSpec, r batcheslib.Repository, result execution.Result) ([]*batcheslib.ChangesetSpec, error) {
	sort.Strings(r.FileMatches)

	input := &batcheslib.ChangesetSpecInput{
		Repository: r,
		BatchChangeAttributes: &template.BatchChangeAttributes{
			Name:        spec.Name,
			Description: spec.Description,
		},
		Template:         spec.ChangesetTemplate,
		TransformChanges: spec.TransformChanges,
		Result:           result,
	}

	return batcheslib.BuildChangesetSpecs(input, batcheslib.ChangesetSpecFeatureFlags{
		IncludeAutoAuthorDetails: true,
		AllowOptionalPublished:   true,
	})
}
