package google

import (
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func getPrompt(messages []types.Message) ([]googleContentMessage, error) {
	googleMessages := make([]googleContentMessage, 0, len(messages))

	for i, message := range messages {
		speaker := message.Speaker
		text := message.Text

		googleRole := message.Speaker

		switch speaker {
		case types.SYSTEM_MESSAGE_SPEAKER:
			if i != 0 {
				return nil, errors.New("system role can only be used in the first message")
			}
		case types.ASSISTANT_MESSAGE_SPEAKER:
			if i == 0 {
				return nil, errors.New("assistant role can only be used in the first message")
			} else {
				googleRole = "model"
			}
		case types.HUMAN_MESSAGE_SPEAKER:
			googleRole = "user"
		default:
			return nil, errors.Errorf("unexpected role: %s", text)
		}

		if text == "" {
			return nil, errors.New("message content cannot be empty")
		}

		googleMessages = append(googleMessages, googleContentMessage{
			Role:  googleRole,
			Parts: []googleContentMessagePart{{Text: text}},
		})
	}

	return googleMessages, nil
}
