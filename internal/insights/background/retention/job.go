package retention

import (
	"strconv"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/executor"
)

type DataRetentionJob struct {
	ID              int
	State           string
	FailureMessage  *string
	QueuedAt        time.Time
	StartedAt       *time.Time
	FinishedAt      *time.Time
	ProcessAfter    *time.Time
	NumResets       int
	NumFailures     int
	LastHeartbeatAt time.Time
	ExecutionLogs   []executor.ExecutionLogEntry
	WorkerHostname  string
	Cancel          bool

	InsightSeriesID int
	SeriesID        string
}

var dataRetentionJobColumns = []*sqlf.Query{
	sqlf.Sprintf("insights_data_retention_jobs.series_id"),
	sqlf.Sprintf("insights_data_retention_jobs.series_id_string"),

	sqlf.Sprintf("id"),
	sqlf.Sprintf("state"),
	sqlf.Sprintf("failure_message"),
	sqlf.Sprintf("started_at"),
	sqlf.Sprintf("finished_at"),
	sqlf.Sprintf("process_after"),
	sqlf.Sprintf("num_resets"),
	sqlf.Sprintf("num_failures"),
	sqlf.Sprintf("execution_logs"),
}

func (j *DataRetentionJob) RecordID() int {
	return j.ID
}

func (j *DataRetentionJob) RecordUID() string {
	return strconv.Itoa(j.ID)
}

func scanDataRetentionJob(s dbutil.Scanner) (*DataRetentionJob, error) {
	var job DataRetentionJob
	var executionLogs []executor.ExecutionLogEntry

	if err := s.Scan(
		&job.InsightSeriesID,
		&job.SeriesID,

		&job.ID,
		&job.State,
		&job.FailureMessage,
		&job.StartedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFailures,
		pq.Array(&job.ExecutionLogs),
	); err != nil {
		return nil, err
	}

	job.ExecutionLogs = append(job.ExecutionLogs, executionLogs...)

	return &job, nil
}
