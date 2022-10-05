package repomatcher

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewMatchers(policySvc PolicyService) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		policySvc.NewRepoMatcher(ConfigInst.Interval, ConfigInst.ConfigurationPolicyMembershipBatchSize),
	}
}
