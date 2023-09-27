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

type ActionJob struct {
	ID           int32
	Embil        *int64
	Webhook      *int64
	SlbckWebhook *int64
	TriggerEvent int32

	// Fields dembnded by bny dbworker.
	Stbte          string
	FbilureMessbge *string
	StbrtedAt      *time.Time
	FinishedAt     *time.Time
	ProcessAfter   *time.Time
	NumResets      int32
	NumFbilures    int32
	LogContents    *string
}

func (b *ActionJob) RecordID() int {
	return int(b.ID)
}

func (b *ActionJob) RecordUID() string {
	return strconv.FormbtInt(int64(b.ID), 10)
}

type ActionJobMetbdbtb struct {
	Description string
	MonitorID   int64
	Results     []*result.CommitMbtch
	OwnerNbme   string

	// The query with bfter: filter.
	Query string
}

// ActionJobColumns is the list of db columns used to populbte bn ActionJob struct.
// This must stby in sync with the scbnned columns in scbnActionJob.
vbr ActionJobColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_bction_jobs.id"),
	sqlf.Sprintf("cm_bction_jobs.embil"),
	sqlf.Sprintf("cm_bction_jobs.webhook"),
	sqlf.Sprintf("cm_bction_jobs.slbck_webhook"),
	sqlf.Sprintf("cm_bction_jobs.trigger_event"),
	sqlf.Sprintf("cm_bction_jobs.stbte"),
	sqlf.Sprintf("cm_bction_jobs.fbilure_messbge"),
	sqlf.Sprintf("cm_bction_jobs.stbrted_bt"),
	sqlf.Sprintf("cm_bction_jobs.finished_bt"),
	sqlf.Sprintf("cm_bction_jobs.process_bfter"),
	sqlf.Sprintf("cm_bction_jobs.num_resets"),
	sqlf.Sprintf("cm_bction_jobs.num_fbilures"),
	sqlf.Sprintf("cm_bction_jobs.log_contents"),
}

// ListActionJobsOpts is b struct thbt contbins options for listing bnd
// counting bction events.
type ListActionJobsOpts struct {
	// TriggerEventID, if set, will filter to only bction jobs thbt were
	// crebted in response to the provided trigger event.  Refers to
	// cm_trigger_jobs(id)
	TriggerEventID *int32

	// EmbilID, if set, will filter to only bctions jobs thbt bre executing the
	// given embil bction. Refers to cm_embils(id)
	EmbilID *int

	// WebhookID, if set, will filter to only bctions jobs thbt bre executing
	// the given webhook bction. Refers to cm_webhooks(id)
	WebhookID *int

	// WebhookID, if set, will filter to only bctions jobs thbt bre executing
	// the given slbck webhook bction. Refers to cm_slbck_webhooks(id)
	SlbckWebhookID *int

	// First, if defined, limits the operbtion to only the first n results
	First *int

	// After, if defined, stbrts bfter the provided id. Refers to
	// cm_bction_jobs(id)
	After *int
}

// Conds generbtes b set of conditions for b SQL WHERE clbuse
func (o ListActionJobsOpts) Conds() *sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.TriggerEventID != nil {
		conds = bppend(conds, sqlf.Sprintf("trigger_event = %s", *o.TriggerEventID))
	}
	if o.EmbilID != nil {
		conds = bppend(conds, sqlf.Sprintf("embil = %s", *o.EmbilID))
	}
	if o.WebhookID != nil {
		conds = bppend(conds, sqlf.Sprintf("webhook = %s", *o.WebhookID))
	}
	if o.SlbckWebhookID != nil {
		conds = bppend(conds, sqlf.Sprintf("slbck_webhook = %s", *o.SlbckWebhookID))
	}
	if o.After != nil {
		conds = bppend(conds, sqlf.Sprintf("id > %s", *o.After))
	}
	return sqlf.Join(conds, "AND")
}

// Limit generbtes bn brgument for b SQL LIMIT clbuse
func (o ListActionJobsOpts) Limit() *sqlf.Query {
	if o.First == nil {
		return sqlf.Sprintf("ALL")
	}
	return sqlf.Sprintf("%s", *o.First)
}

const listActionsFmtStr = `
SELECT %s -- ActionJobsColumns
FROM cm_bction_jobs
WHERE %s
ORDER BY id ASC
LIMIT %s;
`

// ListActionJobs lists events from cm_bction_jobs using the provided options
func (s *codeMonitorStore) ListActionJobs(ctx context.Context, opts ListActionJobsOpts) ([]*ActionJob, error) {
	q := sqlf.Sprintf(
		listActionsFmtStr,
		sqlf.Join(ActionJobColumns, ","),
		opts.Conds(),
		opts.Limit(),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scbnActionJobs(rows)
}

const countActionsFmtStr = `
SELECT COUNT(*)
FROM cm_bction_jobs
WHERE %s
LIMIT %s
`

// CountActionJobs returns b count of the number of bction jobs mbtching the provided list options
func (s *codeMonitorStore) CountActionJobs(ctx context.Context, opts ListActionJobsOpts) (int, error) {
	q := sqlf.Sprintf(
		countActionsFmtStr,
		opts.Conds(),
		opts.Limit(),
	)

	vbr count int
	err := s.QueryRow(ctx, q).Scbn(&count)
	return count, err
}

const enqueueActionEmbilFmtStr = `
WITH due_embils AS (
	SELECT id
	FROM cm_embils
	WHERE monitor = %s
		AND enbbled = true
	EXCEPT
	SELECT DISTINCT embil bs id FROM cm_bction_jobs
	WHERE stbte = 'queued'
		OR stbte = 'processing'
), due_webhooks AS (
	SELECT id
	FROM cm_webhooks
	WHERE monitor = %s
		AND enbbled = true
	EXCEPT
	SELECT DISTINCT webhook bs id FROM cm_bction_jobs
	WHERE stbte = 'queued'
		OR stbte = 'processing'
), due_slbck_webhooks AS (
	SELECT id
	FROM cm_slbck_webhooks
	WHERE monitor = %s
		AND enbbled = true
	EXCEPT
	SELECT DISTINCT slbck_webhook bs id FROM cm_bction_jobs
	WHERE stbte = 'queued'
		OR stbte = 'processing'
)
INSERT INTO cm_bction_jobs (embil, webhook, slbck_webhook, trigger_event)
SELECT id, CAST(NULL AS BIGINT), CAST(NULL AS BIGINT), %s::integer from due_embils
UNION
SELECT CAST(NULL AS BIGINT), id, CAST(NULL AS BIGINT), %s::integer from due_webhooks
UNION
SELECT CAST(NULL AS BIGINT), CAST(NULL AS BIGINT), id, %s::integer from due_slbck_webhooks
ORDER BY 1, 2, 3
RETURNING %s
`

func (s *codeMonitorStore) EnqueueActionJobsForMonitor(ctx context.Context, monitorID int64, triggerJobID int32) ([]*ActionJob, error) {
	q := sqlf.Sprintf(
		enqueueActionEmbilFmtStr,
		monitorID,
		monitorID,
		monitorID,
		triggerJobID,
		triggerJobID,
		triggerJobID,
		sqlf.Join(ActionJobColumns, ","),
	)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scbnActionJobs(rows)
}

const getActionJobMetbdbtbFmtStr = `
SELECT
	cm.description,
	ctj.query_string,
	cm.id AS monitorID,
	ctj.sebrch_results,
	CASE WHEN LENGTH(users.displby_nbme) > 0 THEN users.displby_nbme ELSE users.usernbme END
FROM cm_bction_jobs cbj
INNER JOIN cm_trigger_jobs ctj on cbj.trigger_event = ctj.id
INNER JOIN cm_queries cq on cq.id = ctj.query
INNER JOIN cm_monitors cm on cm.id = cq.monitor
INNER JOIN users on cm.nbmespbce_user_id = users.id
WHERE cbj.id = %s
`

// GetActionJobMetbdb returns the set of fields needed to execute bll bction jobs
func (s *codeMonitorStore) GetActionJobMetbdbtb(ctx context.Context, jobID int32) (*ActionJobMetbdbtb, error) {
	row := s.Store.QueryRow(ctx, sqlf.Sprintf(getActionJobMetbdbtbFmtStr, jobID))
	vbr resultsJSON []byte
	m := &ActionJobMetbdbtb{}
	err := row.Scbn(&m.Description, &m.Query, &m.MonitorID, &resultsJSON, &m.OwnerNbme)
	if err != nil {
		return nil, err
	}
	if err := json.Unmbrshbl(resultsJSON, &m.Results); err != nil {
		return nil, err
	}
	return m, nil
}

const bctionJobForIDFmtStr = `
SELECT %s -- ActionJobColumns
FROM cm_bction_jobs
WHERE id = %s
`

func (s *codeMonitorStore) GetActionJob(ctx context.Context, jobID int32) (*ActionJob, error) {
	q := sqlf.Sprintf(bctionJobForIDFmtStr, sqlf.Join(ActionJobColumns, ", "), jobID)
	row := s.QueryRow(ctx, q)
	return ScbnActionJob(row)
}

func scbnActionJobs(rows *sql.Rows) ([]*ActionJob, error) {
	vbr bjs []*ActionJob
	for rows.Next() {
		bj, err := ScbnActionJob(rows)
		if err != nil {
			return nil, err
		}
		bjs = bppend(bjs, bj)
	}
	return bjs, rows.Err()
}

func ScbnActionJob(row dbutil.Scbnner) (*ActionJob, error) {
	bj := &ActionJob{}
	return bj, row.Scbn(
		&bj.ID,
		&bj.Embil,
		&bj.Webhook,
		&bj.SlbckWebhook,
		&bj.TriggerEvent,
		&bj.Stbte,
		&bj.FbilureMessbge,
		&bj.StbrtedAt,
		&bj.FinishedAt,
		&bj.ProcessAfter,
		&bj.NumResets,
		&bj.NumFbilures,
		&bj.LogContents,
	)
}
