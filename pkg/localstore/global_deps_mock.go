package localstore

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockGlobalDeps struct {
	Dependencies func(context.Context, DependenciesOptions) ([]*sourcegraph.DependencyReference, error)
}
