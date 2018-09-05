package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/pkg/types"
)

type MockSiteConfig struct {
	Get func(ctx context.Context) (*types.SiteConfig, error)
}
