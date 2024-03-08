package anthropic

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const HUMAN_PROMPT = "\n\nHuman:"
const ASSISTANT_PROMPT = "\n\nAssistant:"

func GetPrompt(messages []types.Message) (string, error) {
	prompt := make([]string, 0, len(messages))
	for idx, message := range messages {
		if idx > 0 && messages[idx-1].Speaker == message.Speaker {
			return "", errors.Newf("found consecutive messages with the same speaker '%s'", message.Speaker)
		}

		messagePrompt, err := message.GetPrompt(HUMAN_PROMPT, ASSISTANT_PROMPT)
		if err != nil {
			return "", err
		}
		prompt = append(prompt, messagePrompt)
	}
	return strings.Join(prompt, ""), nil
}

func ToAnthropicMessages(messages []types.Message) ([]anthropicMessage, error) {
	if len(messages) == 0 {
		return nil, errors.New("expected at least one message")
	}

	anthropicMessages := make([]anthropicMessage, 0, len(messages))
	systemRoleFound := false

	for _, message := range messages {
		speaker := message.Speaker
		text := message.Text

		if speaker == types.SYSTEM_MESSAGE_SPEAKER {
			if systemRoleFound || len(anthropicMessages) > 0 {
				return nil, errors.New("system role can only be used in the first message")
			}
			systemRoleFound = true
		} else if speaker != types.HUMAN_MESSAGE_SPEAKER && speaker != types.ASSISTANT_MESSAGE_SPEAKER {
			return nil, fmt.Errorf("unexpected role: %s", text)
		}

		if text == "" {
			return nil, errors.New("message content cannot be empty")
		}

		role := "user"
		if speaker == types.SYSTEM_MESSAGE_SPEAKER {
			role = "system"
		}
		if speaker == types.ASSISTANT_MESSAGE_SPEAKER {
			role = "assistant"
		}

		anthropicMessages = append(anthropicMessages, anthropicMessage{
			Role:    role,
			Content: []anthropicMessageContent{{Text: text, Type: "text"}},
		})
	}

	return anthropicMessages, nil
}
