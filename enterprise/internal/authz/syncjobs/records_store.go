package syncjobs

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

const syncJobsRecordsKey = "authz/sync-job-records"

// default documented in site.schema.json
const defaultSyncJobsRecordsLimit = 100

// RecordsStore is used to record the results of recent permissions syncing jobs for
// diagnostic purposes.
type RecordsStore struct {
	logger log.Logger
	now    func() time.Time

	mux sync.Mutex
	// cache is a replaceable abstraction over rcache.FIFOList.
	cache interface {
		Insert(v []byte) error
	}
}

func NewRecordsStore(logger log.Logger, c conftypes.SiteConfigQuerier) *RecordsStore {
	return &RecordsStore{
		logger: logger,
		cache: rcache.NewFIFOListDynamic(syncJobsRecordsKey, func() int {
			return syncJobsRecordsLimit(c)
		}),
		now: time.Now,
	}
}

func syncJobsRecordsLimit(c conftypes.SiteConfigQuerier) int {
	recordsLimit := c.SiteConfig().AuthzSyncJobsRecordsLimit
	if recordsLimit == 0 {
		return defaultSyncJobsRecordsLimit
	}
	return recordsLimit
}

// Record inserts a record for this job's outcome into the records store.
func (r *RecordsStore) Record(jobType string, jobID int32, providerStates []ProviderStatus, err error) {
	completed := r.now()

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
	r.cache.Insert(val)
}
