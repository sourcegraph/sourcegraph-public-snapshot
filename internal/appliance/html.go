package appliance

import (
	"context"

	"github.com/life4/genesis/slices"

	"github.com/sourcegraph/sourcegraph/internal/releaseregistry"
)

func (a *Appliance) getVersions(ctx context.Context) ([]string, error) {
	versions, err := a.releaseRegistryClient.ListVersions(ctx, "sourcegraph")
	if err != nil {
		return nil, err
	}
	return slices.MapFilter(versions, func(version releaseregistry.ReleaseInfo) (string, bool) {
		return version.Version, version.Public
	}), nil
}
