package syncjobs

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

// keep in sync with consumer in enterprise/cmd/frontend/internal/authz/resolvers/resolver.go
const syncJobsRecordsPrefix = "authz/sync-job-records"

// default documented in site.schema.json
const defaultSyncJobsRecordsTTLMinutes = 5

// RecordsStore is used to record the results of recent permissions syncing jobs for
// diagnostic purposes.
type RecordsStore struct {
	logger log.Logger
	now    func() time.Time

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
		r.mux.Lock()
		defer r.mux.Unlock()

		ttlMinutes := c.SiteConfig().AuthzSyncJobsRecordsTTL
		if ttlMinutes == 0 {
			ttlMinutes = defaultSyncJobsRecordsTTLMinutes
		}

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
func (r *RecordsStore) Record(jobType string, jobID int32, providerStates []ProviderStatus, err error) {
	completed := r.now()

	r.mux.Lock()
	defer r.mux.Unlock()

	record := Status{
		JobType:   jobType,
		JobID:     jobID,
		Completed: completed,
		Status:    "SUCCESS",
		Providers: providerStates,
	}
	if err != nil {
		record.Status = "ERROR"
		record.Message = err.Error()
	}

	val, err := json.Marshal(record)
	if err != nil {
		r.logger.Warn("failed to render entry",
			log.Int32("requestID", record.JobID),
			log.Error(err))
		return
	}

	// Key by timestamp for sorting
	r.cache.Set(strconv.FormatInt(record.Completed.UTC().UnixNano(), 10), val)
}
