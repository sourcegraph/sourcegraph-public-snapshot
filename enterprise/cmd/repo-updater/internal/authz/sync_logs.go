package authz

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

const syncJobsRecordsPrefix = "authz/sync-job-records"

const (
	syncJobStatusSucceeded = "SUCCESS"
	syncJobStatusError     = "ERROR"
)

// syncJobsRecords is used to record the results of recent permissions syncing jobs for
// diagnostic purposes.
//
// The volume of entries recorded can be very high, so implementations should commit the
// jobs to the backend in batches.
type syncJobsRecordsStore struct {
	logger      log.Logger
	recordQueue []authz.SyncJobRecord
	mux         sync.Mutex

	// cache is a replaceable abstraction over rcache.Cache.
	cache interface {
		SetMulti(...[2]string)
	}
}

type noopCache struct{}

func (noopCache) SetMulti(...[2]string) {}

func newSyncJobsRecordsStore(logger log.Logger, minutesTTL int) *syncJobsRecordsStore {
	s := &syncJobsRecordsStore{
		logger:      logger.Scoped("jobRecords", "sync jobs records store"),
		recordQueue: make([]authz.SyncJobRecord, 0, 10),
	}
	if minutesTTL > 0 {
		s.cache = rcache.NewWithTTL(syncJobsRecordsPrefix, minutesTTL*60)
		logger.Info("enabled records store cache")
	} else {
		s.cache = noopCache{}
		logger.Info("disabled records store cache")
	}

	return s
}

func (r *syncJobsRecordsStore) Start(ctx context.Context) {
	go r.commitRecordsRoutine(ctx)
}

// Add queues a record for the completion of a sync job. It is non-blocking.
func (r *syncJobsRecordsStore) Record(jobType string, jobID int32, providerStates []providerState, err error) {
	record := authz.SyncJobRecord{
		RequestType: jobType,
		RequestID:   jobID,
		Completed:   time.Now(),

		// TODO export the providerState type
		// Providers:

		Status: syncJobStatusSucceeded,
	}
	if err != nil {
		record.Status = syncJobStatusError
		record.Message = err.Error()
	}

	go func() {
		r.mux.Lock()
		r.recordQueue = append(r.recordQueue, record)
		r.mux.Unlock()
	}()
}

func (r *syncJobsRecordsStore) commitRecordsRoutine(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		var done bool
		select {
		case <-ctx.Done():
			done = true
		case <-ticker.C:
		}

		// Start work
		r.mux.Lock()

		// Generate entries
		entries := make([][2]string, len(r.recordQueue))
		for i, record := range r.recordQueue {
			val, err := json.Marshal(record)
			if err != nil {
				r.logger.Warn("failed to render entry",
					log.Int32("requestID", record.RequestID),
					log.Error(err))
				continue
			}
			entries[i] = [2]string{
				fmt.Sprintf("%s-%d", record.RequestType, record.RequestID),
				string(val),
			}
		}
		// Commit entries
		r.logger.Debug("committing entries", log.Int("entries", len(entries)))
		r.cache.SetMulti(entries...)
		// Reset entries
		r.recordQueue = make([]authz.SyncJobRecord, 0, len(r.recordQueue))

		r.mux.Unlock()
		if done {
			break
		}
	}
}
