package bg

import (
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
		err := storeTokenUsageinDb(ctx, db)
		if err != nil {
			fmt.Printf("Error storing token usage: %v\n", err)
		}

		// Wait for 5 minutes before the next execution
		time.Sleep(5 * time.Minute)
	}
}

func storeTokenUsageinDb(ctx context.Context, db database.DB) error {
	recorder := telemetryrecorder.New(db)
	tokenManager := tokenusage.NewManager()
	tokenUsageData, err := tokenManager.FetchTokenUsageDataForAnalysis()
	if err != nil {
		return err
	}
	convertedTokenUsageData := make(telemetry.EventMetadata)
	for key, value := range tokenUsageData {
		convertedTokenUsageData[telemetry.SafeMetadataKey(key)] = value
	}

	// This extra variable helps demarcate that this was NOT the final fetch and sync before the redis was reset
	convertedTokenUsageData["FinalFetchAndSync"] = 0.0

	if err := recorder.Record(ctx, "cody.llmTokenCounter", "record", &telemetry.EventParameters{
		Metadata: convertedTokenUsageData,
	}); err != nil {
		return err
	}
	return nil
}
