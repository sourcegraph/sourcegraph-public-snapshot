package expiration

import (
	"context"
	"fmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewExpirer(store DBStore, policyMatcher PolicyMatcher, metrics *metrics) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &expirer{
		dbStore:       store,
		policyMatcher: policyMatcher,
		metrics:       metrics,
		logger: log.Scoped("NewExpirer", "").With(
			log.String("dbStore", fmt.Sprintf("%v", store)),
			log.String("policyMatcher", fmt.Sprintf("%v", policyMatcher)),
			log.String("metrics", fmt.Sprintf("%v", metrics)),
		),
	})
}
