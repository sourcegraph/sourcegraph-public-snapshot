package db

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockGlobalDeps struct {
	Dependencies func(context.Context, DependenciesOptions) ([]*api.DependencyReference, error)
}
