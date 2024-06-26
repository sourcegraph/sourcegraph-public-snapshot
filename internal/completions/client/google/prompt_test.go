package google

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
)

func TestGetPrompt(t *testing.T) {
	t.Run("invalid speaker", func(t *testing.T) {
		_, err := getGeminiPrompt([]types.Message{{Speaker: "invalid", Text: "hello"}})
		if err == nil {
			t.Errorf("expected error for invalid speaker, got nil")
		}
	})

	t.Run("empty text", func(t *testing.T) {
		_, err := getGeminiPrompt([]types.Message{{Speaker: types.HUMAN_MESSAGE_SPEAKER, Text: ""}})
		if err == nil {
			t.Errorf("expected error for empty text, got nil")
		}
	})

	t.Run("invalid prompt starts with human", func(t *testing.T) {
		_, err := getGeminiPrompt([]types.Message{
			{Speaker: types.HUMAN_MESSAGE_SPEAKER, Text: "human speaking"},
			{Speaker: types.HUMAN_MESSAGE_SPEAKER, Text: "hello"},
			{Speaker: types.ASSISTANT_MESSAGE_SPEAKER, Text: "assistant"},
		})
		if err == nil {
			t.Errorf("expected error for messages with repeated speaker, got nil")
		}
	})

	t.Run("valid prompt", func(t *testing.T) {
		messages := []types.Message{
			{Speaker: types.SYSTEM_MESSAGE_SPEAKER, Text: "system"},
			{Speaker: types.HUMAN_MESSAGE_SPEAKER, Text: "hello"},
			{Speaker: types.ASSISTANT_MESSAGE_SPEAKER, Text: "hi"},
		}
		prompt, err := getGeminiPrompt(messages)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		expected := []googleContentMessage{
			{Role: "model", Parts: []googleContentMessagePart{{Text: "system"}}},
			{Role: "user", Parts: []googleContentMessagePart{{Text: "hello"}}},
			{Role: "model", Parts: []googleContentMessagePart{{Text: "hi"}}},
		}
		if len(prompt) != len(expected) {
			t.Errorf("unexpected prompt length, got %d, want %d", len(prompt), len(expected))
		}
		for i := range prompt {
			if prompt[i].Parts[0].Text != expected[i].Parts[0].Text {
				t.Errorf("unexpected prompt message at index %d, got %v, want %v", i, prompt[i], expected[i])
			}
		}
	})

	t.Run("valid prompt with last empty message from assistnt should be removed", func(t *testing.T) {
		messages := []types.Message{
			{Speaker: types.SYSTEM_MESSAGE_SPEAKER, Text: "system"},
			{Speaker: types.HUMAN_MESSAGE_SPEAKER, Text: "hello"},
			{Speaker: types.ASSISTANT_MESSAGE_SPEAKER, Text: ""},
		}
		prompt, err := getGeminiPrompt(messages)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		expected := []googleContentMessage{
			{Role: "model", Parts: []googleContentMessagePart{{Text: "system"}}},
			{Role: "user", Parts: []googleContentMessagePart{{Text: "hello"}}},
		}
		if len(prompt) > len(expected) {
			t.Errorf("unexpected prompt message at last index")
		}
	})

	t.Run("invalid prompt ends with empty human message", func(t *testing.T) {
		_, err := getGeminiPrompt([]types.Message{
			{Speaker: types.SYSTEM_MESSAGE_SPEAKER, Text: "assistant"},
			{Speaker: types.HUMAN_MESSAGE_SPEAKER, Text: "hello"},
			{Speaker: types.ASSISTANT_MESSAGE_SPEAKER, Text: "assistant"},
			{Speaker: types.HUMAN_MESSAGE_SPEAKER, Text: ""},
		})
		if err == nil {
			t.Errorf("expected error for last human message to be empty, got nil")
		}
	})
}
