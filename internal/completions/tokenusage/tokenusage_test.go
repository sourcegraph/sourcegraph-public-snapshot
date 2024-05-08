package tokenusage_test

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/completions/tokenizer"
	"github.com/sourcegraph/sourcegraph/internal/completions/tokenusage"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

func TestTokenizeAndCalculateUsage(t *testing.T) {
	rcache.SetupForTest(t)
	mockCache := rcache.NewWithTTL("LLMUsage", 1800)
	manager := tokenusage.NewManager()
	messages := []types.Message{
		{Speaker: "human", Text: "Hello"},
		{Speaker: "user", Text: "Hi"},
	}
	err := manager.TokenizeAndCalculateUsage(messages, "output text", tokenizer.OpenAIModel+"/gpt-4", "feature1", tokenusage.OpenAI)
	if err != nil {
		t.Fatalf("TokenizeAndCalculateUsage returned an error: %v", err)
	}

	// Verify that token counts are updated in the cache
	inputKey := "openai:openai/gpt-4:feature1:input"
	outputKey := "openai:openai/gpt-4:feature1:output"

	if val, exists, _ := mockCache.GetInt64(inputKey); !exists || val <= 0 {
		t.Errorf("Expected input token count to be updated in cache, but key %s was not found or value is not positive", inputKey)
	}

	if val, exists, _ := mockCache.GetInt64(outputKey); !exists || val <= 0 {
		t.Errorf("Expected output token count to be updated in cache, but key %s was not found or value is not positive", outputKey)
	}
}

func TestGetAllTokenUsageData(t *testing.T) {
	rcache.SetupForTest(t)
	manager := tokenusage.NewManager()
	cache := rcache.NewWithTTL("LLMUsage", 1800)
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
