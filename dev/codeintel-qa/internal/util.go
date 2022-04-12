package internal

import (
	"fmt"
	"time"
)

// TImeSince returns the time since the given duration rounded down to the nearest second.
func TimeSince(start time.Time) time.Duration {
	return time.Since(start) / time.Second * time.Second
}

// MakeTestRepoName returns the given repo name as a fully qualified repository name in the
// sourcegraph-testing GitHub organization.
func MakeTestRepoName(repoName string) string {
	return fmt.Sprintf("github.com/%s/%s", "sourcegraph-testing", repoName)
}
