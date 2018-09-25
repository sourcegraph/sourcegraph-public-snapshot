package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type MockSettings struct {
	GetLatest        func(ctx context.Context, subject api.ConfigurationSubject) (*api.Settings, error)
	CreateIfUpToDate func(ctx context.Context, subject api.ConfigurationSubject, lastID *int32, authorUserID int32, contents string) (latestSetting *api.Settings, err error)
}
