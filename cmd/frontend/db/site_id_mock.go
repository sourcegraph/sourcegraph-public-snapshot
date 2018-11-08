package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type MockSiteIDInfo struct {
	Get func(ctx context.Context) (*types.SiteIDInfo, error)
}
