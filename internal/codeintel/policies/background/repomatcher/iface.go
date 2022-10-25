package repomatcher

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type PolicyService interface {
	NewRepositoryMatcher(
		interval time.Duration,
		configurationPolicyMembershipBatchSize int,
	) goroutine.BackgroundRoutine
}
