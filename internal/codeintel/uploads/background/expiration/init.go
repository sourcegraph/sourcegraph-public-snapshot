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
			log.String("dbStore", fmt.Sprint(store)),
			log.String("policyMatcher", fmt.Sprint(policyMatcher)),
			log.String("metrics", fmt.Sprint(metrics)),
		),
	})
}
