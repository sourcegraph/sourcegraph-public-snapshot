package executorqueue

import (
	"context"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type multiqueueCacheCleaner struct {
	queueNames []string
	cache      *rcache.Cache
	windowSize time.Duration
}

var _ goroutine.Handler = &multiqueueCacheCleaner{}

func NewMultiqueueCacheCleaner(queueNames []string, cache *rcache.Cache, windowSize time.Duration) goroutine.BackgroundRoutine {
	ctx := context.Background()
	return goroutine.NewPeriodicGoroutine(
		ctx,
		&multiqueueCacheCleaner{
			queueNames: queueNames,
			cache:      cache,
			windowSize: windowSize,
		},
		goroutine.WithName("executors.multiqueue-cache-cleaner"),
		goroutine.WithDescription("deletes entries from the dequeue cache older than the configured window"),
		goroutine.WithInterval(5*time.Second),
	)
}

func (m *multiqueueCacheCleaner) Handle(ctx context.Context) error {
	for _, queueName := range m.queueNames {
		all, err := m.cache.GetHashAll(queueName)
		if err != nil {
			return errors.Wrap(err, "multiqueue.cachecleaner")
		}
		for key := range all {
			timestampMillis, err := strconv.ParseInt(key, 10, 64)
			if err != nil {
				return err
			}
			t := time.Unix(0, timestampMillis*int64(time.Millisecond))
			if t.Before(time.Now().Add(-m.windowSize)) {
				// expired cache entry, delete
				deletedItems, err := m.cache.DeleteHashItem(queueName, key)
				if err != nil {
					return err
				}
				if deletedItems == 0 {
					return errors.Newf("failed to delete hash item %s for key %s: expected successful delete but redis deleted nothing", key, queueName)
				}
			}
		}
	}
	return nil
}
