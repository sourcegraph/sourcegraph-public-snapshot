package db

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
)

type MockSiteConfig struct {
	Get func(ctx context.Context) (*types.SiteConfig, error)
}
