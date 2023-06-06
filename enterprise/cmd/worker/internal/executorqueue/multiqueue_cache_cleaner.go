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
	queueName  string
	cache      *rcache.Cache
	windowSize time.Duration
}

var _ goroutine.Handler = &multiqueueCacheCleaner{}

func NewMultiqueueCacheCleaner(queueName string, cache *rcache.Cache, windowSize time.Duration) goroutine.BackgroundRoutine {
	ctx := context.Background()
	return goroutine.NewPeriodicGoroutine(
		ctx,
		&multiqueueCacheCleaner{
			queueName:  queueName,
			cache:      cache,
			windowSize: windowSize,
		},
		goroutine.WithName("executors.multiqueue-cache-cleaner"),
		goroutine.WithDescription("deletes entries from the dequeue cache older than the configured window"),
		goroutine.WithInterval(5*time.Second),
	)
}

func (m *multiqueueCacheCleaner) Handle(ctx context.Context) error {
	all, err := m.cache.GetHashAll(m.queueName)
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
			// TODO: need to add a HDEL function to rcache
			m.cache.Delete(key)
		}
	}
	return nil
}
