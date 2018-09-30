package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/inventory"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
)

type GlobalDepsProvider interface {
	TotalRefs(ctx context.Context, repo *types.Repo, langs []*inventory.Lang) (int, error)
	ListTotalRefs(ctx context.Context, repo *types.Repo, langs []*inventory.Lang) ([]api.RepoID, error)
	UpdateIndexForLanguage(ctx context.Context, language string, repo api.RepoID, deps []lspext.DependencyReference) (err error)
	Dependencies(ctx context.Context, op DependenciesOptions) (refs []*api.DependencyReference, err error)
	Delete(ctx context.Context, repo api.RepoID) error
}

type globalDeps struct{}

// DependenciesOptions specifies options for querying locations that reference
// a definition.
type DependenciesOptions struct {
	// Language is the type of language whose references are being queried.
	// e.g. "go" or "java".
	Language string

	// DepData is data that matches the output of xdependencies with a psql
	// jsonb containment operator. It may be a subset of data.
	DepData map[string]interface{}

	// Repo filters the returned list of DependencyReference instances
	// by repo. It should be used mutually exclusively with DepData.
	Repo api.RepoID

	// Limit limits the number of returned dependency references to the
	// specified number.
	Limit int
}

func (g *globalDeps) TotalRefs(ctx context.Context, repo *types.Repo, langs []*inventory.Lang) (int, error) {
	return 0, nil
}

func (g *globalDeps) ListTotalRefs(ctx context.Context, repo *types.Repo, langs []*inventory.Lang) ([]api.RepoID, error) {
	return nil, nil
}

func (g *globalDeps) UpdateIndexForLanguage(ctx context.Context, language string, repo api.RepoID, deps []lspext.DependencyReference) (err error) {
	return nil
}

func (g *globalDeps) Dependencies(ctx context.Context, op DependenciesOptions) (refs []*api.DependencyReference, err error) {
	if Mocks.GlobalDeps.Dependencies != nil {
		return Mocks.GlobalDeps.Dependencies(ctx, op)
	}
	return nil, nil
}

func (g *globalDeps) Delete(ctx context.Context, repo api.RepoID) error {
	return nil
}
