package localstore

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockPkgs struct {
	ListPackages func(context.Context, *sourcegraph.ListPackagesOp) ([]sourcegraph.PackageInfo, error)
}
