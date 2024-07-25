package completions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/completions/tokenusage"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetrytest"
	v1 "github.com/sourcegraph/sourcegraph/lib/telemetrygateway/v1"
)

func TestStoreTokenUsageInDB(t *testing.T) {
	kv := rcache.SetupForTest(t)
	cache := rcache.NewWithTTL(kv, "LLMUsage", 1800)
	cache.SetInt("LLMUsage:model1:feature1:stream:input", 10)
	cache.SetInt("LLMUsage:model1:feature1:stream:output", 20)
	manager := tokenusage.NewManagerWithCache(cache)

	mockEventStore := telemetrytest.NewMockEventsStore()
	var sentEvent []*v1.Event
	mockEventStore.StoreEventsFunc.SetDefaultHook(func(ctx context.Context, event []*v1.Event) error {
		sentEvent = event
		return nil
	})
	recorder := telemetry.NewEventRecorder(mockEventStore)

	err := recordTokenUsage(context.Background(), manager, recorder)
	require.NoError(t, err)
	require.Equal(t, len(sentEvent), 1)
	require.Equal(t, sentEvent[0].Feature, "cody.llmTokenCounter")
	require.Equal(t, map[string]float64{
		"LLMUsage:model1:feature1:stream:input":  10,
		"LLMUsage:model1:feature1:stream:output": 20,
		"FinalFetchAndSync":                      0.0,
	}, sentEvent[0].Parameters.Metadata)
}
