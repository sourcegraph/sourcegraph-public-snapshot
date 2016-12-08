package localstore

import (
	"context"

	"github.com/pkg/errors"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

type globalRefs struct{}

// RefreshIndex refreshes the global refs index for the specified repository.
func (g *globalRefs) RefreshIndex(ctx context.Context, source, version string) error {
	return errors.New("GlobalRefs.RefreshIndex not implemented")
}

func (g *globalRefs) TotalRefs(ctx context.Context, source string) (int, error) {
	return 0, errors.New("GlobalRefs.TotalRefs not implemented")
}

func (g *globalRefs) RefLocations(ctx context.Context, op sourcegraph.RefLocationsOptions) (*sourcegraph.RefLocations, error) {
	return nil, errors.New("GlobalRefs.RefLocations not implemented")
}
