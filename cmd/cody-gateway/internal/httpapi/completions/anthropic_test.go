package completions

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/grafana/regexp"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/tokenizer"
)

func TestIsFlaggedAnthropicRequest(t *testing.T) {
	validPreamble := "You are cody-gateway."

	tk, err := tokenizer.NewAnthropicClaudeTokenizer()
	require.NoError(t, err)

	t.Run("missing known preamble", func(t *testing.T) {
		ar := anthropicRequest{Model: "claude-2", Prompt: "some prompt without known preamble"}
		flagged, reason, err := isFlaggedAnthropicRequest(tk, ar, []*regexp.Regexp{regexp.MustCompile(validPreamble)})
		require.NoError(t, err)
		require.True(t, flagged)
		require.Equal(t, "unknown_prompt", reason)
	})

	t.Run("preamble not configured ", func(t *testing.T) {
		ar := anthropicRequest{Model: "claude-2", Prompt: "some prompt without known preamble"}
		flagged, _, err := isFlaggedAnthropicRequest(tk, ar, []*regexp.Regexp{})
		require.NoError(t, err)
		require.False(t, flagged)
	})

	t.Run("high max tokens to sample", func(t *testing.T) {
		ar := anthropicRequest{Model: "claude-2", MaxTokensToSample: 10000, Prompt: validPreamble}
		flagged, reason, err := isFlaggedAnthropicRequest(tk, ar, []*regexp.Regexp{})
		require.NoError(t, err)
		require.True(t, flagged)
		require.Equal(t, "high_max_tokens_to_sample_10000", reason)
	})

	t.Run("high prompt token count", func(t *testing.T) {
		tokenLengths, err := tk.Tokenize(validPreamble)
		require.NoError(t, err)

		validPreambleTokens := len(tokenLengths)
		longPrompt := strings.Repeat("word ", promptTokenLimit+1)
		ar := anthropicRequest{Model: "claude-2", Prompt: validPreamble + " " + longPrompt}
		flagged, reason, err := isFlaggedAnthropicRequest(tk, ar, []*regexp.Regexp{regexp.MustCompile(validPreamble)})
		require.NoError(t, err)
		require.True(t, flagged)
		require.Equal(t, fmt.Sprintf("high_prompt_token_count_%d", promptTokenLimit+1+validPreambleTokens+1), reason)
	})
}

func TestAnthropicRequestJSON(t *testing.T) {
	tk, err := tokenizer.NewAnthropicClaudeTokenizer()
	require.NoError(t, err)

	ar := anthropicRequest{Prompt: "Hello world"}
	_, _ = ar.GetPromptTokenCount(tk) // hydrate values that should not be marshalled

	b, err := json.MarshalIndent(ar, "", "\t")
	require.NoError(t, err)

	autogold.Expect(`{
"prompt": "Hello world",
"model": "",
"max_tokens_to_sample": 0
}`).Equal(t, string(b))
}

func TestAnthropicRequestGetPromptTokenCount(t *testing.T) {
	tk, err := tokenizer.NewAnthropicClaudeTokenizer()
	require.NoError(t, err)

	originalRequest := anthropicRequest{Prompt: "Hello world"}

	t.Run("values are hydrated", func(t *testing.T) {
		count, err := originalRequest.GetPromptTokenCount(tk)
		require.NoError(t, err)
		assert.Equal(t, originalRequest.promptTokens.count, count)
		assert.Nil(t, originalRequest.promptTokens.err)
	})

	t.Run("values are not recomputed", func(t *testing.T) {
		newRequest := originalRequest // copy
		// Contrived example, we never update the prompt.
		newRequest.Prompt = "Hello again! This is a much longer prompt now"
		count2, err := newRequest.GetPromptTokenCount(tk)
		require.NoError(t, err)
		assert.Equal(t, originalRequest.promptTokens.count, count2, "token count should be unchanged")
	})
}
