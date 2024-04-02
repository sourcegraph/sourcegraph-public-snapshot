package bg

import (
	//nolint:logging // TODO move all logging to sourcegraph/log

	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/completions/tokenusage"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetryrecorder"
)

func ScheduleStoreTokenUsage(ctx context.Context, db database.DB) {
	for {
		// Execute StoreTokenUsage
		err := StoreTokenUsage(ctx, db)
		if err != nil {
			fmt.Printf("Error storing token usage: %v\n", err)
		}

		// Wait for 1 minute before the next execution
		time.Sleep(time.Minute)
	}
}

func StoreTokenUsage(ctx context.Context, db database.DB) error {
	recorder := telemetryrecorder.New(db)
	tokenManager := tokenusage.NewManager()
	tokenUsageData, err := tokenManager.FetchTokenUsageDataForAnalysis()
	if err != nil {
		return err
	}
	convertedTokenUsageData := make(telemetry.EventMetadata)
	for key, value := range tokenUsageData {
		convertedTokenUsageData[telemetry.ConstString(key)] = value
	}

	if err := recorder.Record(ctx, "llmTokenCounter", "modelUsage", &telemetry.EventParameters{
		Metadata: convertedTokenUsageData,
	}); err != nil {
		return err
	}
	return nil
}
