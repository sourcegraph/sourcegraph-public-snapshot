package completions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
)

// getModelResult is a mock implementation of the getModelFn, one of the parameters
// to newCompletionsHandler.
type getModelResult struct {
	Model string
	Err   error
}

type mockGetModelFn struct {
	results []getModelResult
}

func (m *mockGetModelFn) PushResult(model string, err error) {
	result := getModelResult{
		Model: model,
		Err:   err,
	}
	m.results = append(m.results, result)
}

func (m *mockGetModelFn) ToFunc() func(context.Context, types.CodyCompletionRequestParameters, *conftypes.CompletionsConfig) (string, error) {
	return func(context.Context, types.CodyCompletionRequestParameters, *conftypes.CompletionsConfig) (string, error) {
		if len(m.results) == 0 {
			panic("no call registered to getModels")
		}
		v := m.results[0]
		m.results = m.results[1:]
		return v.Model, v.Err
	}
}

func TestGetCodeCompletionsModelFn(t *testing.T) {
	ctx := context.Background()

	validCompletionsConfig := conftypes.CompletionsConfig{
		ChatModel:       "chat-model-in-config",
		CompletionModel: "code-model-in-config",
		FastChatModel:   "fast-chat-model-in-config",
	}

	getModelFn := getCodeCompletionModelFn()

	t.Run("ErrorUnsupportedModel", func(t *testing.T) {

		reqParams := types.CodyCompletionRequestParameters{
			CompletionRequestParameters: types.CompletionRequestParameters{
				Model: "model-the-user-requested",
			},
		}
		_, err := getModelFn(ctx, reqParams, nil /* completionsConfig */)
		require.ErrorContains(t, err, `unsupported code completion model "model-the-user-requested"`)

		_, err2 := getModelFn(ctx, reqParams, &validCompletionsConfig)
		require.ErrorContains(t, err2, `unsupported code completion model "model-the-user-requested"`)
	})

	t.Run("OverrideSiteConfig", func(t *testing.T) {
		reqParams := types.CodyCompletionRequestParameters{
			// BUG: This is inconsistent with how user-requested models work for "chats", which
			// totally ignore user preferences. Here we _always_ honor the user's preference.
			//
			// We should reject requests to use models the calling user cannot access, or are not
			// available to the current "Cody Pro Subscription" or "Cody Enterprise config".
			CompletionRequestParameters: types.CompletionRequestParameters{
				Model: "google/gemini-pro",
			},
		}
		model, err := getModelFn(ctx, reqParams, nil)
		require.NoError(t, err)
		assert.Equal(t, "google/gemini-pro", model)
	})

	t.Run("Default", func(t *testing.T) {
		// For these tests, the Model field in the request body isn't set.
		t.Run("NoSiteConfig", func(t *testing.T) {
			reqParams := types.CodyCompletionRequestParameters{}
			_, err := getModelFn(ctx, reqParams, nil)
			assert.ErrorContains(t, err, "no completions config available")
		})
		t.Run("WithSiteConfig", func(t *testing.T) {
			reqParams := types.CodyCompletionRequestParameters{}
			model, err := getModelFn(ctx, reqParams, &validCompletionsConfig)
			require.NoError(t, err)
			assert.Equal(t, "code-model-in-config", model)
		})
	})
}

func TestGetChatModelFn(t *testing.T) {
	ctx := context.Background()
	mockDB := dbmocks.NewMockDB()

	validCompletionsConfig := conftypes.CompletionsConfig{
		ChatModel:       "chat-model-in-config",
		CompletionModel: "code-model-in-config",
		FastChatModel:   "fast-chat-model-in-config",
	}

	t.Run("IgnoreRequestUseConfig", func(t *testing.T) {
		t.Run("Chat", func(t *testing.T) {
			getModelFn := getChatModelFn(mockDB)

			reqParams := types.CodyCompletionRequestParameters{
				CompletionRequestParameters: types.CompletionRequestParameters{
					// For Cody Enterprise, this is totally ignored. Currently, only
					// Cody Pro users can configure the chat model used.
					// TODO(PRIME-283): Enable LLM model selection Cody Enterprise users.
					Model: "model-the-user-requested",
				},
			}
			model, err := getModelFn(ctx, reqParams, &validCompletionsConfig)

			require.NoError(t, err)
			assert.Equal(t, "chat-model-in-config", model)
		})

		t.Run("FastChat", func(t *testing.T) {
			getModelFn := getChatModelFn(mockDB)

			reqParams := types.CodyCompletionRequestParameters{
				CompletionRequestParameters: types.CompletionRequestParameters{
					// For Cody Enterprise, this is totally ignored. Currently, only
					// Cody Pro users can configure the chat model used.
					// TODO(PRIME-283): Enable LLM model selection Cody Enterprise users.
					Model: "model-the-user-requested",
				},
				// .. but again, for "fast" chats.
				Fast: true,
			}
			compConfig := conftypes.CompletionsConfig{
				ChatModel:       "chat-model-in-config",
				CompletionModel: "code-model-in-config",
				FastChatModel:   "fast-chat-model-in-config",
			}
			model, err := getModelFn(ctx, reqParams, &compConfig)

			require.NoError(t, err)
			assert.Equal(t, "fast-chat-model-in-config", model)
		})
	})

	// TODO(PRIME-283): As part of enabling model selection for Cody Enterprise users,
	// add more tests for the Cody Pro path as well. Where we only allow certain models
	// based on the calling user's subscription status, etc.
}
