package db

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
)

type MockSettings struct {
	GetLatest        func(ctx context.Context, subject types.ConfigurationSubject) (*types.Settings, error)
	CreateIfUpToDate func(ctx context.Context, subject types.ConfigurationSubject, lastKnownSettingsID *int32, authorUserID int32, contents string) (latestSetting *types.Settings, err error)
}
