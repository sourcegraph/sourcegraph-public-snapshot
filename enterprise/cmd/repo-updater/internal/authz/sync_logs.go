package authz

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

const syncJobsRecordsPrefix = "authz/sync-job-records"

const (
	syncJobStatusSucceeded = "SUCCESS"
	syncJobStatusError     = "ERROR"
)

// syncJobsRecords is used to record the results of recent permissions syncing jobs for
// diagnostic purposes.
type syncJobsRecordsStore struct {
	logger   log.Logger
	cacheTTL atomic.Int32

	mux sync.Mutex
	// cache is a replaceable abstraction over rcache.Cache.
	cache interface{ Set(key string, v []byte) }
}

type noopCache struct{}

func (noopCache) Set(string, []byte) {}

func newSyncJobsRecordsStore(logger log.Logger) *syncJobsRecordsStore {
	return &syncJobsRecordsStore{
		logger: logger.Scoped("jobRecords", "sync jobs records store"),
		cache:  noopCache{},
	}
}

func (r *syncJobsRecordsStore) Watch(c conftypes.WatchableSiteConfig) {
	c.Watch(func() {
		ttlMinutes := c.SiteConfig().AuthzSyncJobsLogsTTL
		if ttlMinutes == 0 {
			ttlMinutes = 30 // default documented
		}
		if !r.cacheTTL.CompareAndSwap(r.cacheTTL.Load(), int32(ttlMinutes)) {
			// unchanged
			return
		}

		// Update the cache
		r.mux.Lock()
		defer r.mux.Unlock()

		if ttlMinutes > 0 {
			ttlSeconds := ttlMinutes * 60
			r.cache = rcache.NewWithTTL(syncJobsRecordsPrefix, ttlSeconds)
			r.logger.Info("enabled records store cache")
		} else {
			r.cache = noopCache{}
			r.logger.Info("disabled records store cache")
		}
	})
}

// Add queues a record for the completion of a sync job.
func (r *syncJobsRecordsStore) Record(jobType string, jobID int32, providerStates []authz.SyncJobProviderState, err error) {
	completed := time.Now()

	r.mux.Lock()
	defer r.mux.Unlock()

	record := authz.SyncJobRecord{
		RequestType: jobType,
		RequestID:   jobID,
		Completed:   completed,

		// TODO export the providerState type
		// Providers:

		Status: syncJobStatusSucceeded,
	}
	if err != nil {
		record.Status = syncJobStatusError
		record.Message = err.Error()
	}

	val, err := json.Marshal(record)
	if err != nil {
		r.logger.Warn("failed to render entry",
			log.Int32("requestID", record.RequestID),
			log.Error(err))
		return
	}
	r.cache.Set(fmt.Sprintf("%s-%d", record.RequestType, record.RequestID), val)
}
