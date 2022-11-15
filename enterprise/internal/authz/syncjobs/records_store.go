package syncjobs

import (
	"encoding/json"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

const syncJobsRecordsPrefix = "authz/sync-job-records"

// default documented in site.schema.json
const defaultSyncJobsRecordsTTLMinutes = 15

// RecordsStore is used to record the results of recent permissions syncing jobs for
// diagnostic purposes.
type RecordsStore struct {
	logger   log.Logger
	cacheTTL atomic.Int32
	now      func() time.Time

	mux sync.Mutex
	// cache is a replaceable abstraction over rcache.Cache.
	cache interface{ Set(key string, v []byte) }
}

type noopCache struct{}

func (noopCache) Set(string, []byte) {}

func NewRecordsStore(logger log.Logger) *RecordsStore {
	return &RecordsStore{
		logger: logger,
		cache:  noopCache{},
		now:    time.Now,
	}
}

func (r *RecordsStore) Watch(c conftypes.WatchableSiteConfig) {
	c.Watch(func() {
		ttlMinutes := c.SiteConfig().AuthzSyncJobsLogsTTL
		if ttlMinutes == 0 {
			ttlMinutes = defaultSyncJobsRecordsTTLMinutes
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
			r.logger.Info("enabled records store cache", log.Int("ttlSeconds", ttlSeconds))
		} else {
			r.cache = noopCache{}
			r.logger.Info("disabled records store cache")
		}
	})
}

// Record inserts a record for this job's outcome into the records store.
func (r *RecordsStore) Record(jobType string, jobID int32, providerStates []authz.SyncJobProviderStatus, err error) {
	completed := r.now()

	r.mux.Lock()
	defer r.mux.Unlock()

	record := authz.SyncJobStatus{
		RequestType: jobType,
		RequestID:   jobID,
		Completed:   completed,
		Status:      "SUCCESS",
		Providers:   providerStates,
	}
	if err != nil {
		record.Status = "ERROR"
		record.Message = err.Error()
	}

	val, err := json.Marshal(record)
	if err != nil {
		r.logger.Warn("failed to render entry",
			log.Int32("requestID", record.RequestID),
			log.Error(err))
		return
	}

	// Key by timestamp for sorting
	r.cache.Set(strconv.FormatInt(record.Completed.UnixNano(), 10), val)
}
