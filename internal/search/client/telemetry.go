package client

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

// DetermineStatusForLogs determines the final status of a search for logging
// purposes.
func DetermineStatusForLogs(alert *search.Alert, stats streaming.Stats, err error) string {
	switch {
	case err == context.DeadlineExceeded:
		return "timeout"
	case err != nil:
		return "error"
	case stats.Status.All(search.RepoStatusTimedout) && stats.Status.Len() == len(stats.Repos):
		return "timeout"
	case stats.Status.Any(search.RepoStatusTimedout):
		return "partial_timeout"
	case alert != nil:
		return "alert"
	default:
		return "success"
	}
}
