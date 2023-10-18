package executors

import (
	"context"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type multiqueueCacheCleaner struct {
	queueNames []string
	cache      *rcache.Cache
	windowSize time.Duration
	logger     log.Logger
}

var _ goroutine.Handler = &multiqueueCacheCleaner{}

// NewMultiqueueCacheCleaner returns a PeriodicGoroutine that will check the cache for entries that are older than the configured
// window size. A cache key is represented by a queue name; the value is a hash containing timestamps as the field key and the
// job ID as the field value (which is not used for anything currently).
func NewMultiqueueCacheCleaner(queueNames []string, cache *rcache.Cache, windowSize time.Duration, cleanupInterval time.Duration) goroutine.BackgroundRoutine {
	logger := log.Scoped("multiqueue-cache-cleaner")
	observationCtx := observation.NewContext(logger)
	handler := &multiqueueCacheCleaner{
		queueNames: queueNames,
		cache:      cache,
		windowSize: windowSize,
		logger:     logger,
	}
	for _, queue := range queueNames {
		handler.initMetrics(observationCtx, queue, map[string]string{"queue": queue})
	}
	ctx := context.Background()
	return goroutine.NewPeriodicGoroutine(
		ctx,
		handler,
		goroutine.WithName("executors.multiqueue-cache-cleaner"),
		goroutine.WithDescription("deletes entries from the dequeue cache older than the configured window"),
		goroutine.WithInterval(cleanupInterval),
	)
}

// Handle loops over the configured queue names and deletes stale entries.
func (m *multiqueueCacheCleaner) Handle(ctx context.Context) error {
	for _, queueName := range m.queueNames {
		all, err := m.cache.GetHashAll(queueName)
		if err != nil {
			if errors.Is(err, redis.ErrNil) {
				return nil
			}
			return errors.Wrap(err, "multiqueue.cachecleaner")
		}

		for key := range all {
			keyAsUnixNano, err := strconv.ParseInt(key, 10, 64)
			if err != nil {
				return err
			}
			t := time.Unix(0, keyAsUnixNano)
			maxAge := timeNow().Add(-m.windowSize)
			if t.Before(maxAge) {
				// expired cache entry, delete
				deletedItems, err := m.cache.DeleteHashItem(queueName, key)
				if err != nil {
					return err
				}
				if deletedItems == 0 {
					return errors.Newf("failed to delete hash item %s for key %s: expected successful delete but redis deleted nothing", key, queueName)
				}
				m.logger.Debug("Deleted stale dequeue cache key", log.String("queue", queueName), log.String("key", key), log.String("dateTime", t.GoString()), log.String("maxAge", maxAge.GoString()))
			} else {
				m.logger.Debug("Preserved dequeue cache key", log.String("queue", queueName), log.String("key", key), log.String("dateTime", t.GoString()), log.String("maxAge", maxAge.GoString()))
			}
		}
	}
	return nil
}

var timeNow = time.Now

func (m *multiqueueCacheCleaner) initMetrics(observationCtx *observation.Context, queue string, constLabels prometheus.Labels) {
	logger := observationCtx.Logger.Scoped("multiqueue.cachecleaner.metrics")
	observationCtx.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "multiqueue_executor_dequeue_cache_size",
		Help:        "Current size of the executor dequeue cache",
		ConstLabels: constLabels,
	}, func() float64 {
		all, err := m.cache.GetHashAll(queue)
		if err != nil && !errors.Is(err, redis.ErrNil) {
			logger.Error("Failed to get cache size", log.String("queue", queue), log.Error(err))
			return 0
		}

		return float64(len(all))
	}))
}
