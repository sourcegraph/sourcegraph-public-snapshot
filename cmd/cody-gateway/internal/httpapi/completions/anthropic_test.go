package completions

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/regexp"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/tokenizer"
)

func TestIsFlaggedAnthropicRequest(t *testing.T) {
	validPreamble := "You are cody-gateway."

	tk, err := tokenizer.NewAnthropicClaudeTokenizer()
	require.NoError(t, err)

	t.Run("missing known preamble", func(t *testing.T) {
		ar := anthropicRequest{Model: "claude-2", Prompt: "some prompt without known preamble"}
		result, err := isFlaggedAnthropicRequest(tk, ar, []*regexp.Regexp{regexp.MustCompile(validPreamble)})
		require.NoError(t, err)
		require.True(t, result.IsFlagged())
		require.False(t, result.shouldBlock)
		require.Contains(t, result.reasons, "unknown_prompt")
	})

	t.Run("preamble not configured", func(t *testing.T) {
		ar := anthropicRequest{Model: "claude-2", Prompt: "some prompt without known preamble"}
		result, err := isFlaggedAnthropicRequest(tk, ar, []*regexp.Regexp{})
		require.NoError(t, err)
		require.False(t, result.IsFlagged())
	})

	t.Run("preamble not configured for claude-2.1", func(t *testing.T) {
		ar := anthropicRequest{Model: "claude-2.1", Prompt: "some prompt without known preamble"}
		result, err := isFlaggedAnthropicRequest(tk, ar, []*regexp.Regexp{})
		require.NoError(t, err)
		require.False(t, result.IsFlagged())
	})

	t.Run("high max tokens to sample", func(t *testing.T) {
		ar := anthropicRequest{Model: "claude-2", MaxTokensToSample: 10000, Prompt: validPreamble}
		result, err := isFlaggedAnthropicRequest(tk, ar, []*regexp.Regexp{})
		require.NoError(t, err)
		require.True(t, result.IsFlagged())
		require.True(t, result.shouldBlock)
		require.Contains(t, result.reasons, "high_max_tokens_to_sample")
		require.Equal(t, int32(result.maxTokensToSample), ar.MaxTokensToSample)
	})
	t.Run("high prompt token count (below block limit)", func(t *testing.T) {
		tokenLengths, err := tk.Tokenize(validPreamble)
		require.NoError(t, err)

		validPreambleTokens := len(tokenLengths)
		longPrompt := strings.Repeat("word ", promptTokenFlaggingLimit+1)
		ar := anthropicRequest{Model: "claude-2", Prompt: validPreamble + " " + longPrompt}
		result, err := isFlaggedAnthropicRequest(tk, ar, []*regexp.Regexp{regexp.MustCompile(validPreamble)})
		require.NoError(t, err)
		require.True(t, result.IsFlagged())
		require.False(t, result.shouldBlock)
		require.Contains(t, result.reasons, "high_prompt_token_count")
		require.Equal(t, result.promptTokenCount, validPreambleTokens+1+promptTokenFlaggingLimit+1)
	})

	t.Run("high prompt token count (below block limit)", func(t *testing.T) {
		tokenLengths, err := tk.Tokenize(validPreamble)
		require.NoError(t, err)

		validPreambleTokens := len(tokenLengths)
		longPrompt := strings.Repeat("word ", promptTokenBlockingLimit+1)
		ar := anthropicRequest{Model: "claude-2", Prompt: validPreamble + " " + longPrompt}
		result, err := isFlaggedAnthropicRequest(tk, ar, []*regexp.Regexp{regexp.MustCompile(validPreamble)})
		require.NoError(t, err)
		require.True(t, result.IsFlagged())
		require.True(t, result.shouldBlock)
		require.Contains(t, result.reasons, "high_prompt_token_count")
		require.Equal(t, result.promptTokenCount, validPreambleTokens+1+promptTokenBlockingLimit+1)
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

func TestActor_IsDotComActor(t *testing.T) {
	t.Run("with dotcom actor", func(t *testing.T) {
		actor := &actor.Actor{
			ID: "d3d2b638-d0a2-4539-a099-b36860b09819",
		}

		isDotCom := actor.IsDotComActor()

		require.True(t, isDotCom)
	})

	t.Run("with nondotcom actor", func(t *testing.T) {
		actor := &actor.Actor{
			ID: "NOT_DOTCOM",
		}

		isDotCom := actor.IsDotComActor()

		require.False(t, isDotCom)
	})

	t.Run("with nil actor", func(t *testing.T) {
		var actor *actor.Actor = nil

		isDotCom := actor.IsDotComActor()

		require.False(t, isDotCom)
	})
}
