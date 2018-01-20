package db

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockSiteConfig struct {
	Get func(ctx context.Context) (*sourcegraph.SiteConfig, error)
}
