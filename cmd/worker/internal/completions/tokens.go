package completions

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/completions/tokenusage"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
)

func storeTokenUsageInDb(ctx context.Context, recorder *telemetry.EventRecorder) error {
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
