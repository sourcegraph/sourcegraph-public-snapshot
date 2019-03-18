package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type MockGlobalDeps struct {
	Dependencies func(context.Context, DependenciesOptions) ([]*api.DependencyReference, error)
}
