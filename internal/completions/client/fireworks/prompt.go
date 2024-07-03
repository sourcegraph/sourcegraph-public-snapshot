package fireworks

import (
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func getPrompt(messages []types.Message) (string, error) {
	if len(messages) != 1 {
		return "", errors.Errorf("expected to receive exactly one message with the prompt (got %d)", len(messages))
	}

	prompt := messages[0].Text
	if prompt == "" {
		return "", errors.New("Prompt message text is empty")
	}

	return prompt, nil
}
