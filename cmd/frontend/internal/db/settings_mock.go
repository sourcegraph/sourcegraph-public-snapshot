package db

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockSettings struct {
	GetLatest        func(ctx context.Context, subject api.ConfigurationSubject) (*types.Settings, error)
	CreateIfUpToDate func(ctx context.Context, subject api.ConfigurationSubject, lastKnownSettingsID *int32, authorUserID int32, contents string) (latestSetting *types.Settings, err error)
}
