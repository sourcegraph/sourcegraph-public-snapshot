package tokenusage_test

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/completions/tokenusage"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

func TestGetAllTokenUsageData(t *testing.T) {
	rcache.SetupForTest(t)
	manager := tokenusage.NewManager()
	cache := rcache.NewWithTTL(redispool.Store, "LLMUsage", 1800)
	cache.SetInt("LLMUsage:model1:feature1:stream:input", 10)
	cache.SetInt("LLMUsage:model1:feature1:stream:output", 20)

	usageSummary, err := manager.RetrieveAndResetTokenUsageData()

	if err != nil {
		t.Error(err)
	}

	llmUsage, ok := usageSummary["llm_usage"].([]map[string]interface{})
	if !ok {
		t.Fatalf("Expected llm_usage key to be present and be a map")
	}

	models, ok := llmUsage[0]["models"].([]tokenusage.ModelData)
	if !ok || len(models) != 2 {
		t.Fatalf("Expected models to be a slice of map with 2 items, got %d", len(models))
	}

	// Prepare expected results for easier comparison
	expected := map[string]float64{
		"LLMUsage:model1:feature1:stream:input":  10,
		"LLMUsage:model1:feature1:stream:output": 20,
	}

	for _, model := range models {
		if expectedTokens, exists := expected[model.Description]; !exists || expectedTokens != model.Tokens {
			t.Errorf("Expected %f tokens for %s, got %f", expectedTokens, model.Description, model.Tokens)
		}
	}
}
