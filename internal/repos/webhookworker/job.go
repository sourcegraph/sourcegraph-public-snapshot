package webhookworker

import (
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type Job struct {
	// Webhook builder fields
	RepoID     int32
	RepoName   string
	Org        string
	ExtSvcID   int64
	ExtSvcKind string
	QueuedAt   *time.Time

	// Standard dbworker fields
	ID             int
	State          string
	FailureMessage *string
	StartedAt      *time.Time
	FinishedAt     *time.Time
	ProcessAfter   *time.Time
	NumResets      int32
	NumFailures    int32
	ExecutionLogs  []workerutil.ExecutionLogEntry
}

func (j *Job) RecordID() int {
	return j.ID
}

var jobColumns = []*sqlf.Query{
	sqlf.Sprintf("webhook_build_jobs.repo_id"),
	sqlf.Sprintf("webhook_build_jobs.repo_name"),
	sqlf.Sprintf("webhook_build_jobs.org"),
	sqlf.Sprintf("webhook_build_jobs.extsvc_id"),
	sqlf.Sprintf("webhook_build_jobs.extsvc_kind"),
	sqlf.Sprintf("webhook_build_jobs.queued_at"),
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
