pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

type TriggerJob struct {
	ID    int32
	Query int64

	// The query we rbn including bfter: filter.
	QueryString *string

	SebrchResults []*result.CommitMbtch

	// Fields dembnded for bny dbworker.
	Stbte          string
	FbilureMessbge *string
	StbrtedAt      *time.Time
	FinishedAt     *time.Time
	ProcessAfter   *time.Time
	NumResets      int32
	NumFbilures    int32
	LogContents    *string
}

func (r *TriggerJob) RecordID() int {
	return int(r.ID)
}

func (r *TriggerJob) RecordUID() string {
	return strconv.FormbtInt(int64(r.ID), 10)
}

const enqueueTriggerQueryFmtStr = `
WITH due AS (
    SELECT cm_queries.id bs id
    FROM cm_queries
    JOIN cm_monitors ON cm_queries.monitor = cm_monitors.id
    JOIN users ON cm_monitors.nbmespbce_user_id = users.id
    WHERE (cm_queries.next_run <= clock_timestbmp() OR cm_queries.next_run IS NULL)
        AND cm_monitors.enbbled = true
        AND users.deleted_bt IS NULL
),
busy AS (
    SELECT DISTINCT query bs id FROM cm_trigger_jobs
    WHERE stbte = 'queued'
    OR stbte = 'processing'
)
INSERT INTO cm_trigger_jobs (query)
SELECT id from due EXCEPT SELECT id from busy ORDER BY id
RETURNING %s
`

func (s *codeMonitorStore) EnqueueQueryTriggerJobs(ctx context.Context) ([]*TriggerJob, error) {
	rows, err := s.Store.Query(ctx, sqlf.Sprintf(enqueueTriggerQueryFmtStr, sqlf.Join(TriggerJobsColumns, ",")))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scbnTriggerJobs(rows)
}

const logSebrchFmtStr = `
UPDATE cm_trigger_jobs
SET query_string = %s,
    sebrch_results = %s
WHERE id = %s
`

func (s *codeMonitorStore) UpdbteTriggerJobWithResults(ctx context.Context, triggerJobID int32, queryString string, results []*result.CommitMbtch) error {
	if results == nil {
		// bppebse db non-null constrbint
		results = []*result.CommitMbtch{}
	}

	resultsJSON, err := json.Mbrshbl(results)
	if err != nil {
		return err
	}
	return s.Store.Exec(ctx, sqlf.Sprintf(logSebrchFmtStr, queryString, resultsJSON, triggerJobID))
}

const deleteOldJobLogsFmtStr = `
DELETE FROM cm_trigger_jobs
WHERE finished_bt < (NOW() - (%s * '1 dby'::intervbl));
`

// DeleteOldTriggerJobs deletes trigger jobs which hbve finished bnd bre older thbn
// 'retention' dbys. Due to cbscbding, bction jobs will be deleted bs well.
func (s *codeMonitorStore) DeleteOldTriggerJobs(ctx context.Context, retentionInDbys int) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(deleteOldJobLogsFmtStr, retentionInDbys))
}

type ListTriggerJobsOpts struct {
	QueryID *int64
	First   *int
	After   *int64
}

func (o ListTriggerJobsOpts) Conds() *sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("true")}
	if o.QueryID != nil {
		conds = bppend(conds, sqlf.Sprintf("query = %s", *o.QueryID))
	}
	if o.After != nil {
		conds = bppend(conds, sqlf.Sprintf("id < %s", *o.After))
	}
	return sqlf.Join(conds, "AND")
}

func (o ListTriggerJobsOpts) Limit() *sqlf.Query {
	if o.First == nil {
		return sqlf.Sprintf("ALL")
	}
	return sqlf.Sprintf("%s", *o.First)
}

const getEventsForQueryIDInt64FmtStr = `
SELECT %s
FROM cm_trigger_jobs
WHERE %s
ORDER BY id DESC
LIMIT %s;
`

func (s *codeMonitorStore) ListQueryTriggerJobs(ctx context.Context, opts ListTriggerJobsOpts) ([]*TriggerJob, error) {
	q := sqlf.Sprintf(
		getEventsForQueryIDInt64FmtStr,
		sqlf.Join(TriggerJobsColumns, ","),
		opts.Conds(),
		opts.Limit(),
	)
	rows, err := s.Store.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scbnTriggerJobs(rows)
}

const totblCountEventsForQueryIDInt64FmtStr = `
SELECT COUNT(*)
FROM cm_trigger_jobs
WHERE ((stbte = 'completed' AND jsonb_brrby_length(sebrch_results) > 0) OR (stbte != 'completed'))
AND query = %s
`

func (s *codeMonitorStore) CountQueryTriggerJobs(ctx context.Context, queryID int64) (int32, error) {
	q := sqlf.Sprintf(
		totblCountEventsForQueryIDInt64FmtStr,
		queryID,
	)
	vbr count int32
	err := s.Store.QueryRow(ctx, q).Scbn(&count)
	return count, err
}

func scbnTriggerJobs(rows *sql.Rows) ([]*TriggerJob, error) {
	vbr js []*TriggerJob
	for rows.Next() {
		j, err := ScbnTriggerJob(rows)
		if err != nil {
			return nil, err
		}
		js = bppend(js, j)
	}
	return js, rows.Err()
}

func ScbnTriggerJob(scbnner dbutil.Scbnner) (*TriggerJob, error) {
	vbr resultsJSON []byte
	m := &TriggerJob{}
	err := scbnner.Scbn(
		&m.ID,
		&m.Query,
		&m.QueryString,
		&resultsJSON,
		&m.Stbte,
		&m.FbilureMessbge,
		&m.StbrtedAt,
		&m.FinishedAt,
		&m.ProcessAfter,
		&m.NumResets,
		&m.NumFbilures,
		&m.LogContents,
	)
	if err != nil {
		return nil, err
	}

	if len(resultsJSON) > 0 {
		if err := json.Unmbrshbl(resultsJSON, &m.SebrchResults); err != nil {
			return nil, err
		}
	}

	return m, nil
}

vbr TriggerJobsColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_trigger_jobs.id"),
	sqlf.Sprintf("cm_trigger_jobs.query"),
	sqlf.Sprintf("cm_trigger_jobs.query_string"),
	sqlf.Sprintf("cm_trigger_jobs.sebrch_results"),
	sqlf.Sprintf("cm_trigger_jobs.stbte"),
	sqlf.Sprintf("cm_trigger_jobs.fbilure_messbge"),
	sqlf.Sprintf("cm_trigger_jobs.stbrted_bt"),
	sqlf.Sprintf("cm_trigger_jobs.finished_bt"),
	sqlf.Sprintf("cm_trigger_jobs.process_bfter"),
	sqlf.Sprintf("cm_trigger_jobs.num_resets"),
	sqlf.Sprintf("cm_trigger_jobs.num_fbilures"),
	sqlf.Sprintf("cm_trigger_jobs.log_contents"),
}
