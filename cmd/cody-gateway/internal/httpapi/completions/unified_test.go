package completions

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/regexp"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/shared/config"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/tokenizer"
)

func TestIsFlaggedUnifiedRequest(t *testing.T) {
	validPreamble := "You are cody-gateway."

	cfg := config.AnthropicConfig{
		PromptTokenFlaggingLimit:       18000,
		PromptTokenBlockingLimit:       20000,
		MaxTokensToSampleFlaggingLimit: 1000,
		ResponseTokenBlockingLimit:     1000,
	}
	tk, err := tokenizer.NewAnthropicClaudeTokenizer()
	require.NoError(t, err)

	t.Run("works for known preamble", func(t *testing.T) {
		r := unifiedRequest{Model: "anthropic/claude-3-sonnet-20240229", Messages: []unifiedMessage{unifiedMessage{Role: "system", Content: []unifiedContent{unifiedContent{Type: "text", Text: validPreamble}}}}}
		result, err := isFlaggedUnifiedRequest(tk, r, []*regexp.Regexp{regexp.MustCompile(validPreamble)}, cfg)
		require.NoError(t, err)
		require.Nil(t, result)
	})

	t.Run("missing known preamble", func(t *testing.T) {
		r := unifiedRequest{Model: "anthropic/claude-3-sonnet-20240229", Messages: []unifiedMessage{unifiedMessage{Role: "system", Content: []unifiedContent{unifiedContent{Type: "text", Text: "some prompt without known preamble"}}}}}
		result, err := isFlaggedUnifiedRequest(tk, r, []*regexp.Regexp{regexp.MustCompile(validPreamble)}, cfg)
		require.NoError(t, err)
		require.True(t, result.IsFlagged())
		require.False(t, result.shouldBlock)
		require.Contains(t, result.reasons, "unknown_prompt")
	})

	t.Run("preamble not configured ", func(t *testing.T) {
		r := unifiedRequest{Model: "anthropic/claude-3-sonnet-20240229", Messages: []unifiedMessage{unifiedMessage{Role: "system", Content: []unifiedContent{unifiedContent{Type: "text", Text: "some prompt without known preamble"}}}}}
		result, err := isFlaggedUnifiedRequest(tk, r, []*regexp.Regexp{}, cfg)
		require.NoError(t, err)
		require.False(t, result.IsFlagged())
	})

	t.Run("high max tokens to sample", func(t *testing.T) {
		r := unifiedRequest{Model: "anthropic/claude-3-sonnet-20240229", MaxTokens: 10000, Messages: []unifiedMessage{unifiedMessage{Role: "system", Content: []unifiedContent{unifiedContent{Type: "text", Text: validPreamble}}}}}

		result, err := isFlaggedUnifiedRequest(tk, r, []*regexp.Regexp{}, cfg)
		require.NoError(t, err)
		require.True(t, result.IsFlagged())
		require.True(t, result.shouldBlock)
		require.Contains(t, result.reasons, "high_max_tokens_to_sample")
		require.Equal(t, int32(result.maxTokensToSample), r.MaxTokens)
	})

	t.Run("high prompt token count (below block limit)", func(t *testing.T) {
		tokenLengths, err := tk.Tokenize(validPreamble)
		require.NoError(t, err)

		validPreambleTokens := len(tokenLengths)
		longPrompt := strings.Repeat("word ", cfg.PromptTokenFlaggingLimit+1)
		r := unifiedRequest{Model: "anthropic/claude-3-sonnet-20240229", Messages: []unifiedMessage{
			{Role: "system", Content: []unifiedContent{{Type: "text", Text: validPreamble}}},
			{Role: "user", Content: []unifiedContent{{Type: "text", Text: longPrompt}}},
		}}

		result, err := isFlaggedUnifiedRequest(tk, r, []*regexp.Regexp{regexp.MustCompile(validPreamble)}, cfg)
		require.NoError(t, err)
		require.True(t, result.IsFlagged())
		require.False(t, result.shouldBlock)
		require.Contains(t, result.reasons, "high_prompt_token_count")
		require.Equal(t, result.promptTokenCount, validPreambleTokens+4+cfg.PromptTokenFlaggingLimit+4, cfg)
	})

	t.Run("high prompt token count (below block limit)", func(t *testing.T) {
		tokenLengths, err := tk.Tokenize(validPreamble)
		require.NoError(t, err)

		validPreambleTokens := len(tokenLengths)
		longPrompt := strings.Repeat("word ", cfg.PromptTokenBlockingLimit+1)
		r := unifiedRequest{Model: "anthropic/claude-3-sonnet-20240229", Messages: []unifiedMessage{
			{Role: "system", Content: []unifiedContent{{Type: "text", Text: validPreamble}}},
			{Role: "user", Content: []unifiedContent{{Type: "text", Text: longPrompt}}},
		}}

		result, err := isFlaggedUnifiedRequest(tk, r, []*regexp.Regexp{regexp.MustCompile(validPreamble)}, cfg)
		require.NoError(t, err)
		require.True(t, result.IsFlagged())
		require.True(t, result.shouldBlock)
		require.Contains(t, result.reasons, "high_prompt_token_count")
		require.Equal(t, result.promptTokenCount, validPreambleTokens+4+cfg.PromptTokenBlockingLimit+4)
	})
}

func TestUnifiedRequestJSON(t *testing.T) {
	tk, err := tokenizer.NewAnthropicClaudeTokenizer()
	require.NoError(t, err)

	r := unifiedRequest{Model: "anthropic/claude-3-sonnet-20240229", Messages: []unifiedMessage{{Role: "user", Content: []unifiedContent{{Type: "text", Text: "Hello world"}}}}}
	_, _ = r.GetPromptTokenCount(tk) // hydrate values that should not be marshalled

	b, err := json.MarshalIndent(r, "", "\t")
	require.NoError(t, err)

	autogold.Expect(`{
"messages": [
	{
		"role": "user",
		"content": [
			{
				"type": "text",
				"text": "Hello world"
			}
		]
	}
],
"model": "anthropic/claude-3-sonnet-20240229"
}`).Equal(t, string(b))
}
