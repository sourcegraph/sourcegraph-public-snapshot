package campaigns

import (
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type CampaignApplyJob struct {
	ID             int        `json:"id"`
	QueuedAt       time.Time  `json:"queuedAt"`
	State          string     `json:"state"`
	FailureMessage *string    `json:"failureMessage"`
	StartedAt      *time.Time `json:"startedAt"`
	FinishedAt     *time.Time `json:"finishedAt"`
	ProcessAfter   *time.Time `json:"processAfter"`
	NumResets      int        `json:"numResets"`
	NumFailures    int        `json:"numFailures"`
	Spec           string     `json:"spec"`
	LogContents    string     `json:"logContents"`
}

var campaignApplyJobColumns = []*sqlf.Query{
	sqlf.Sprintf("j.id"),
	sqlf.Sprintf("j.queued_at"),
	sqlf.Sprintf("j.state"),
	sqlf.Sprintf("j.failure_message"),
	sqlf.Sprintf("j.started_at"),
	sqlf.Sprintf("j.finished_at"),
	sqlf.Sprintf("j.process_after"),
	sqlf.Sprintf("j.num_resets"),
	sqlf.Sprintf("j.num_failures"),
	sqlf.Sprintf("j.spec"),
	sqlf.Sprintf("j.log_contents"),
}

func (i CampaignApplyJob) RecordID() int {
	return i.ID
}

func scanFirstCampaignApplyJobRecord(rows *sql.Rows, queryErr error) (_ workerutil.Record, _ bool, err error) {
	if queryErr != nil {
		return nil, false, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if rows.Next() {
		var job CampaignApplyJob
		if err := rows.Scan(
			&job.ID,
			&job.QueuedAt,
			&job.State,
			&job.FailureMessage,
			&job.StartedAt,
			&job.FinishedAt,
			&job.ProcessAfter,
			&job.NumResets,
			&job.NumFailures,
			&job.Spec,
			&job.LogContents,
		); err != nil {
			return nil, false, err
		}
	}

	return nil, false, nil
}
