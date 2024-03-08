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
	sanitizedMessages := sanitizeMessagesForCompleteApi(messages)
	prompt := make([]string, 0, len(sanitizedMessages))
	for idx, message := range sanitizedMessages {
		if idx > 0 && sanitizedMessages[idx-1].Speaker == message.Speaker {
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

func sanitizeMessagesForMessagesApi(messages []types.Message) []types.Message {
	sanitizedMessages := messages

	// 1. If the last message is from an `assistant` with no or empty `text`, omit it
	lastMessage := messages[len(messages)-1]
	truncateLastMessage := lastMessage.Speaker == types.ASSISTANT_MESSAGE_SPEAKER && lastMessage.Text == ""
	if truncateLastMessage {
		sanitizedMessages = messages[:len(messages)-1]
	}

	// 2. If there is any assistant message in the middle of the messages without a `text`, omit
	//    both the empty assistant message as well as the unanswered question from the `user`
	filteredMessages := make([]types.Message, 0)
	for i, message := range sanitizedMessages {
		// If the message is the last message, it is not a middle message
		if i >= len(sanitizedMessages)-1 {
			filteredMessages = append(filteredMessages, message)
			continue
		}

		// If the next message is an assistant message with no or empty `content`, omit the current and
		// the next one
		nextMessage := sanitizedMessages[i+1]
		if (nextMessage.Speaker == types.ASSISTANT_MESSAGE_SPEAKER && nextMessage.Text == "") ||
			(message.Speaker == types.ASSISTANT_MESSAGE_SPEAKER && message.Text == "") {
			continue
		}
		filteredMessages = append(filteredMessages, message)
	}
	sanitizedMessages = filteredMessages

	// 3. Final assistant content cannot end with trailing whitespace
	lastMessage = sanitizedMessages[len(sanitizedMessages)-1]
	if lastMessage.Speaker == types.ASSISTANT_MESSAGE_SPEAKER && lastMessage.Text != "" {
		lastMessage.Text = strings.TrimRight(lastMessage.Text, " \t\n\r")
	}

	return sanitizedMessages
}

func sanitizeMessagesForCompleteApi(messages []types.Message) []types.Message {
	sanitizedMessages := make([]types.Message, 0)
	for i, message := range messages {
		// 1. Convert the first message (which might be a `system` message) to a human message
		if i == 0 && message.Speaker == types.SYSTEM_MESSAGE_SPEAKER {
			message.Speaker = types.HUMAN_MESSAGE_SPEAKER
			sanitizedMessages = append(sanitizedMessages, message)
			sanitizedMessages = append(sanitizedMessages, types.Message{Speaker: types.ASSISTANT_MESSAGE_SPEAKER, Text: "Ok."})
		} else {
			sanitizedMessages = append(sanitizedMessages, message)
		}
	}

	// 2. The last message must be from from an `assistant`
	lastMessage := sanitizedMessages[len(sanitizedMessages)-1]
	if lastMessage.Speaker != types.ASSISTANT_MESSAGE_SPEAKER {
		sanitizedMessages = append(sanitizedMessages, types.Message{Speaker: types.ASSISTANT_MESSAGE_SPEAKER})
	}

	return sanitizedMessages
}
