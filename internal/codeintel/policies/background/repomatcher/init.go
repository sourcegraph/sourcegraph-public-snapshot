package repomatcher

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewRepositoryMatcher(policySvc PolicyService) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		policySvc.NewRepositoryMatcher(
			ConfigInst.Interval,
			ConfigInst.ConfigurationPolicyMembershipBatchSize,
		),
	}
}
