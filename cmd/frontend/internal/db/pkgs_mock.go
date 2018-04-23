package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type MockPkgs struct {
	ListPackages func(context.Context, *api.ListPackagesOp) ([]*api.PackageInfo, error)
}
