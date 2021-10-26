package database

import (
	"context"
)

type MockFeatureFlags struct {
	GetOrgFeatureFlag func(ctx context.Context, orgID int32, flagName string) (bool, error)
}
