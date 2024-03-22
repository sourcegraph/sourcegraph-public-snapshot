package bg

import (
	//nolint:logging // TODO move all logging to sourcegraph/log

	"context"

	"github.com/sourcegraph/sourcegraph/internal/completions/tokenusage"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetryrecorder"
)

func StoreTokenUsage(ctx context.Context, db database.DB) error {
	recorder := telemetryrecorder.New(db)
	tokenManager := tokenusage.NewManager()
	data, err := tokenManager.FetchTokenUsageDataForAnalysis()
	if err != nil {
		return err
	}
	convertedData := make(telemetry.EventMetadata)
	for key, value := range data {
		convertedData[telemetry.ConstString(key)] = value
	}

	if err := recorder.Record(ctx, "myFeature", "myAction", &telemetry.EventParameters{
		Metadata: convertedData,
	}); err != nil {
		return err
	}
	return nil
}
