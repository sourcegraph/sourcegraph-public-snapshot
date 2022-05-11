package queryrunner

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
)

type StreamingError struct {
	Type     types.GenerationMethod
	Messages []string
}

func (e StreamingError) Error() string {
	if e.Type == types.SearchCompute {
		return fmt.Sprintf("encountered error(s) while running a streaming compute search: %v", e.Messages)
	}
	return fmt.Sprintf("encountered error(s) while running a streaming search: %v", e.Messages)
}

func (e StreamingError) NonRetryable() bool { return true }
