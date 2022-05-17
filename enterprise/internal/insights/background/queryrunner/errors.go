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
		return fmt.Sprintf("compute streaming search: errors: %v", e.Messages)
	}
	return fmt.Sprintf("streaming search: errors: %v", e.Messages)
}

func (e StreamingError) NonRetryable() bool { return true }
