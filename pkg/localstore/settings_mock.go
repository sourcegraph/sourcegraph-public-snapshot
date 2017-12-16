package localstore

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockSettings struct {
	GetLatest        func(ctx context.Context, subject sourcegraph.ConfigurationSubject) (*sourcegraph.Settings, error)
	CreateIfUpToDate func(ctx context.Context, subject sourcegraph.ConfigurationSubject, lastKnownSettingsID *int32, authorAuth0ID, contents string) (latestSetting *sourcegraph.Settings, err error)
}
