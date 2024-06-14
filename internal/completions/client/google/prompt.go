package google

import (
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// getPrompt converts a slice of types.Message into a slice of googleContentMessage,
// which is the format expected by the Google Completions API. It ensures that the
// speaker roles are consistent and that the message content is not empty.
func getPrompt(messages []types.Message) ([]googleContentMessage, error) {
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
