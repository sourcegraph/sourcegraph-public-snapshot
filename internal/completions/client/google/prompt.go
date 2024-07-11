package google

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// It returns an error if the input messages are invalid, such as an empty slice,
// the first message is not a non-empty assistant message, or any message content is empty
// (except for the last message if it's an assistant message).
func getAnthropicPrompt(messages []types.Message) ([]anthropicMessage, string, error) {
	if len(messages) == 0 {
		return nil, "", errors.New("messages cannot be empty")
	}
	var systemPrompt string
	anthropicMessages := make([]anthropicMessage, 0, len(messages))

	for i, message := range messages {
		var anthropicRole string

		switch message.Speaker {
		case types.SYSTEM_MESSAGE_SPEAKER:
			if i != 0 {
				return nil, "", errors.New("system role can only be used in the first message")
			}
			systemPrompt = message.Text
			continue
		case types.ASSISTANT_MESSAGE_SPEAKER:
			anthropicRole = "assistant"
			if i == 0 {
				anthropicRole = "system"
			}
		case types.HUMAN_MESSAGE_SPEAKER:
			anthropicRole = "user"
		default:
			return nil, "", errors.Errorf("unexpected role: %s", message.Speaker)
		}

		// Trim whitespace from the message text if it's the last message
		trimmedText := message.Text
		if i == len(messages)-1 {
			trimmedText = strings.TrimSpace(message.Text)
		}
		anthropicMessages = append(anthropicMessages, anthropicMessage{
			Role:    anthropicRole,
			Content: []anthropicMessagePart{{Text: trimmedText, Type: "text"}},
		})
	}

	return anthropicMessages, systemPrompt, nil
}

// getPrompt converts a slice of types.Message into a slice of googleContentMessage,
// which is the format expected by the Google Completions API. It ensures that the
// speaker roles are consistent and that the message content is not empty.
func getGeminiPrompt(messages []types.Message) ([]googleContentMessage, error) {
	googleMessages := make([]googleContentMessage, 0, len(messages))

	for i, message := range messages {
		var googleRole string

		switch message.Speaker {
		case types.SYSTEM_MESSAGE_SPEAKER:
			if i != 0 {
				return nil, errors.New("system role can only be used in the first message")
			}
			googleRole = "model"
		case types.ASSISTANT_MESSAGE_SPEAKER:
			googleRole = "model"
		case types.HUMAN_MESSAGE_SPEAKER:
			googleRole = "user"
		default:
			return nil, errors.Errorf("unexpected role: %s", message.Text)
		}

		if message.Text == "" {
			// Skip empty assistant messages only if it's the last message.
			if googleRole == "model" && i != 0 && i == len(messages)-1 {
				continue
			}
			return nil, errors.New("message content cannot be empty")
		}
		if len(googleMessages) > 0 {
			if googleMessages[i-1].Role == googleRole {
				return nil, errors.New("consistent speaker role is not allowed")
			}
		}

		googleMessages = append(googleMessages, googleContentMessage{
			Role:  googleRole,
			Parts: []googleContentMessagePart{{Text: message.Text}},
		})
	}

	return googleMessages, nil
}
