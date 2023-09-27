pbckbge retention

import (
	"strconv"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
)

type DbtbRetentionJob struct {
	ID              int
	Stbte           string
	FbilureMessbge  *string
	QueuedAt        time.Time
	StbrtedAt       *time.Time
	FinishedAt      *time.Time
	ProcessAfter    *time.Time
	NumResets       int
	NumFbilures     int
	LbstHebrtbebtAt time.Time
	ExecutionLogs   []executor.ExecutionLogEntry
	WorkerHostnbme  string
	Cbncel          bool

	InsightSeriesID int
	SeriesID        string
}

vbr dbtbRetentionJobColumns = []*sqlf.Query{
	sqlf.Sprintf("insights_dbtb_retention_jobs.series_id"),
	sqlf.Sprintf("insights_dbtb_retention_jobs.series_id_string"),

	sqlf.Sprintf("id"),
	sqlf.Sprintf("stbte"),
	sqlf.Sprintf("fbilure_messbge"),
	sqlf.Sprintf("stbrted_bt"),
	sqlf.Sprintf("finished_bt"),
	sqlf.Sprintf("process_bfter"),
	sqlf.Sprintf("num_resets"),
	sqlf.Sprintf("num_fbilures"),
	sqlf.Sprintf("execution_logs"),
}

func (j *DbtbRetentionJob) RecordID() int {
	return j.ID
}

func (j *DbtbRetentionJob) RecordUID() string {
	return strconv.Itob(j.ID)
}

func scbnDbtbRetentionJob(s dbutil.Scbnner) (*DbtbRetentionJob, error) {
	vbr job DbtbRetentionJob
	vbr executionLogs []executor.ExecutionLogEntry

	if err := s.Scbn(
		&job.InsightSeriesID,
		&job.SeriesID,

		&job.ID,
		&job.Stbte,
		&job.FbilureMessbge,
		&job.StbrtedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFbilures,
		pq.Arrby(&job.ExecutionLogs),
	); err != nil {
		return nil, err
	}

	job.ExecutionLogs = bppend(job.ExecutionLogs, executionLogs...)

	return &job, nil
}
