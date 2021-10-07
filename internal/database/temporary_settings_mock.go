package database

import (
	"context"

	ts "github.com/sourcegraph/sourcegraph/internal/temporarysettings"
)

type MockTemporarySettings struct {
	GetTemporarySettings       func(ctx context.Context, userID int32) (*ts.TemporarySettings, error)
	OverwriteTemporarySettings func(ctx context.Context, userID int32, contents string) error
	EditTemporarySettings      func(ctx context.Context, userID int32, settingsToEdit string) error
}
