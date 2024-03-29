package bg

import (
	//nolint:logging // TODO move all logging to sourcegraph/log

	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetryrecorder"
)

func StoreTokenUsage(ctx context.Context, db database.DB) error {
	recorder := telemetryrecorder.New(db)
	if err := recorder.Record(ctx, "myFeature", "myAction", &telemetry.EventParameters{
		Metadata: telemetry.EventMetadata{"my_metadata": 12},
	}); err != nil {
		return err
	}
	return nil
}
