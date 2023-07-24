package queryrunner

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type TerminalStreamingError struct {
	Type     types.GenerationMethod
	Messages []string
}

func (e TerminalStreamingError) Error() string {
	return stringifyStreamingError(e.Messages, e.Type, true)
}

func (e TerminalStreamingError) NonRetryable() bool { return true }

func stringifyStreamingError(messages []string, streamingType types.GenerationMethod, terminal bool) string {
	retryable := ""
	if terminal {
		retryable = " terminal"
	}
	if streamingType == types.SearchCompute {
		return fmt.Sprintf("compute streaming search:%s errors: %v", retryable, messages)
	}
	return fmt.Sprintf("streaming search:%s errors: %v", retryable, messages)
}

func classifiedError(messages []string, streamingType types.GenerationMethod) error {
	for _, m := range messages {
		if strings.Contains(m, "invalid query") {
			return TerminalStreamingError{Type: streamingType, Messages: messages}
		}
	}
	return errors.Errorf(stringifyStreamingError(messages, streamingType, false))
}

var SearchTimeoutError = errors.New("search timeout")
