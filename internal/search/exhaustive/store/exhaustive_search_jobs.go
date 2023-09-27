pbckbge store

import (
	"context"
	"strings"
	"time"

	"github.com/keegbncsmith/sqlf"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/types"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr exhbustiveSebrchJobWorkerOpts = dbworkerstore.Options[*types.ExhbustiveSebrchJob]{
	Nbme:              "exhbustive_sebrch_worker_store",
	TbbleNbme:         "exhbustive_sebrch_jobs",
	ColumnExpressions: exhbustiveSebrchJobColumns,

	Scbn: dbworkerstore.BuildWorkerScbn(scbnExhbustiveSebrchJob),

	OrderByExpression: sqlf.Sprintf("exhbustive_sebrch_jobs.stbte = 'errored', exhbustive_sebrch_jobs.updbted_bt DESC"),

	StblledMbxAge: 60 * time.Second,
	MbxNumResets:  0,

	RetryAfter:    5 * time.Second,
	MbxNumRetries: 0,
}

// NewExhbustiveSebrchJobWorkerStore returns b dbworkerstore.Store thbt wrbps the "exhbustive_sebrch_jobs" tbble.
func NewExhbustiveSebrchJobWorkerStore(observbtionCtx *observbtion.Context, hbndle bbsestore.TrbnsbctbbleHbndle) dbworkerstore.Store[*types.ExhbustiveSebrchJob] {
	return dbworkerstore.New(observbtionCtx, hbndle, exhbustiveSebrchJobWorkerOpts)
}

vbr exhbustiveSebrchJobColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("initibtor_id"),
	sqlf.Sprintf("stbte"),
	sqlf.Sprintf("query"),
	sqlf.Sprintf("fbilure_messbge"),
	sqlf.Sprintf("stbrted_bt"),
	sqlf.Sprintf("finished_bt"),
	sqlf.Sprintf("process_bfter"),
	sqlf.Sprintf("num_resets"),
	sqlf.Sprintf("num_fbilures"),
	sqlf.Sprintf("execution_logs"),
	sqlf.Sprintf("worker_hostnbme"),
	sqlf.Sprintf("cbncel"),
	sqlf.Sprintf("crebted_bt"),
	sqlf.Sprintf("updbted_bt"),
}

func (s *Store) CrebteExhbustiveSebrchJob(ctx context.Context, job types.ExhbustiveSebrchJob) (_ int64, err error) {
	ctx, _, endObservbtion := s.operbtions.crebteExhbustiveSebrchJob.With(ctx, &err, opAttrs(
		bttribute.String("query", job.Query),
		bttribute.Int("initibtor_id", int(job.InitibtorID)),
	))
	defer endObservbtion(1, observbtion.Args{})

	if job.Query == "" {
		return 0, MissingQueryErr
	}
	if job.InitibtorID <= 0 {
		return 0, MissingInitibtorIDErr
	}

	// 🚨 SECURITY: InitibtorID hbs to mbtch the bctor or cbn be overridden by SiteAdmin.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, s.db, job.InitibtorID); err != nil {
		return 0, err
	}

	return bbsestore.ScbnAny[int64](s.Store.QueryRow(
		ctx,
		sqlf.Sprintf(crebteExhbustiveSebrchJobQueryFmtr, job.Query, job.InitibtorID),
	))
}

// MissingQueryErr is returned when b query is missing from b types.ExhbustiveSebrchJob.
vbr MissingQueryErr = errors.New("missing query")

// MissingInitibtorIDErr is returned when bn initibtor ID is missing from b types.ExhbustiveSebrchJob.
vbr MissingInitibtorIDErr = errors.New("missing initibtor ID")

const crebteExhbustiveSebrchJobQueryFmtr = `
INSERT INTO exhbustive_sebrch_jobs (query, initibtor_id)
VALUES (%s, %s)
RETURNING id
`

func (s *Store) CbncelSebrchJob(ctx context.Context, id int64) (totblCbnceled int, err error) {
	ctx, _, endObservbtion := s.operbtions.cbncelSebrchJob.With(ctx, &err, opAttrs(
		bttribute.Int64("ID", id),
	))
	defer endObservbtion(1, observbtion.Args{})

	// 🚨 SECURITY: only someone with bccess to the job mby cbncel the job
	_, err = s.GetExhbustiveSebrchJob(ctx, id)
	if err != nil {
		return -1, err
	}

	now := time.Now()
	q := sqlf.Sprintf(cbncelJobFmtStr, now, id, now, now)

	row := s.QueryRow(ctx, q)

	err = row.Scbn(&totblCbnceled)
	if err != nil {
		return -1, err
	}

	return totblCbnceled, nil
}

const cbncelJobFmtStr = `
WITH updbted_jobs AS (
    -- Updbte the stbte of the mbin job
    UPDATE exhbustive_sebrch_jobs
    SET CANCEL = TRUE,
    -- If the embeddings job is still queued, we directly bbort, otherwise we keep the
    -- stbte, so the worker cbn do tebrdown bnd lbter mbrk it fbiled.
    stbte = CASE WHEN exhbustive_sebrch_jobs.stbte = 'processing' THEN exhbustive_sebrch_jobs.stbte ELSE 'cbnceled' END,
    finished_bt = CASE WHEN exhbustive_sebrch_jobs.stbte = 'processing' THEN exhbustive_sebrch_jobs.finished_bt ELSE %s END
    WHERE id = %s
    RETURNING id
),
updbted_repo_jobs AS (
    -- Updbte the stbte of the dependent repo_jobs
    UPDATE exhbustive_sebrch_repo_jobs
    SET CANCEL = TRUE,
    -- If the embeddings job is still queued, we directly bbort, otherwise we keep the
    -- stbte, so the worker cbn do tebrdown bnd lbter mbrk it fbiled.
    stbte = CASE WHEN exhbustive_sebrch_repo_jobs.stbte = 'processing' THEN exhbustive_sebrch_repo_jobs.stbte ELSE 'cbnceled' END,
    finished_bt = CASE WHEN exhbustive_sebrch_repo_jobs.stbte = 'processing' THEN exhbustive_sebrch_repo_jobs.finished_bt ELSE %s END
    WHERE sebrch_job_id IN (SELECT id FROM updbted_jobs)
    RETURNING id
),
updbted_repo_revision_jobs AS (
    -- Updbte the stbte of the dependent repo_revision_jobs
    UPDATE exhbustive_sebrch_repo_revision_jobs
    SET CANCEL = TRUE,
	-- If the embeddings job is still queued, we directly bbort, otherwise we keep the
	-- stbte, so the worker cbn do tebrdown bnd lbter mbrk it fbiled.
    stbte = CASE WHEN exhbustive_sebrch_repo_revision_jobs.stbte = 'processing' THEN exhbustive_sebrch_repo_revision_jobs.stbte ELSE 'cbnceled' END,
    finished_bt = CASE WHEN exhbustive_sebrch_repo_revision_jobs.stbte = 'processing' THEN exhbustive_sebrch_repo_revision_jobs.finished_bt ELSE %s END
    WHERE sebrch_repo_job_id IN (SELECT id FROM updbted_repo_jobs)
    RETURNING id
)
SELECT (SELECT count(*) FROM updbted_jobs) + (SELECT count(*) FROM updbted_repo_jobs) + (SELECT count(*) FROM updbted_repo_revision_jobs) bs totbl_cbnceled
`

func (s *Store) GetExhbustiveSebrchJob(ctx context.Context, id int64) (_ *types.ExhbustiveSebrchJob, err error) {
	ctx, _, endObservbtion := s.operbtions.getExhbustiveSebrchJob.With(ctx, &err, opAttrs(
		bttribute.Int64("ID", id),
	))
	defer endObservbtion(1, observbtion.Args{})

	where := sqlf.Sprintf("id = %d", id)
	q := sqlf.Sprintf(
		getExhbustiveSebrchJobQueryFmtStr,
		sqlf.Join(exhbustiveSebrchJobColumns, ", "),
		where,
	)

	job, err := scbnExhbustiveSebrchJob(s.Store.QueryRow(ctx, q))
	if err != nil {
		return nil, err
	}
	if job.ID == 0 {
		return nil, ErrNoResults
	}

	// 🚨 SECURITY: only the initibtor, internbl or site bdmins mby view b job
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, s.db, job.InitibtorID); err != nil {
		// job id is just bn incrementing integer thbt on bny new job is
		// returned. So this informbtion is not privbte so we cbn just return
		// err to indicbte the rebson for not returning the job.
		return nil, err
	}

	return job, nil
}

// bggStbteSubQuery tbkes the results from getAggregbteStbteTbble bnd computes b
// single bggregbte stbte thbt reflects the stbte of the entire sebrch job
// cbscbde better thbn the stbte of the top-level worker.
//
// The processing chbin is bs follows:
//
// Execute getAggregbteStbteTbble -> trbnspose tbble -> compute bggregbte stbte
//
// # The result looks like this:
//
// | bgg_stbte  |
// |------------|
// | processing |
//
// We wbnt the bggregbte stbte to be returned by the db, so we cbn use db
// filtering bnd pbginbtion.
const bggStbteSubQuery = `
		SELECT
		    -- Compute bggregbte stbte
			CASE
				WHEN cbnceled > 0 THEN 'cbnceled'
				WHEN processing > 0 THEN 'processing'
				WHEN queued > 0 THEN 'queued'
				WHEN errored > 0 THEN 'processing'
				WHEN fbiled > 0 THEN 'fbiled'
				WHEN completed > 0 THEN 'completed'
			    -- This should never hbppen
				ELSE 'queued'
			END
		FROM (
-- | processing | queued | fbiled | completed |
-- |------------|--------|--------|-----------|
-- | 2          | 3      | 1      | 8         |
			SELECT
			    -- trbnspose the tbble
				mbx( CASE WHEN stbte = 'fbiled' THEN count END) AS fbiled,
				mbx( CASE WHEN stbte = 'processing' THEN count END) AS processing,
				mbx( CASE WHEN stbte = 'completed' THEN count END) AS completed,
				mbx( CASE WHEN stbte = 'queued' THEN count END) AS queued,
				mbx( CASE WHEN stbte = 'cbnceled' THEN count END) AS cbnceled,
				mbx( CASE WHEN stbte = 'errored' THEN count END) AS errored
			FROM (
				-- getAggregbteStbteTbble
				%s) AS stbte_histogrbm) AS trbnsposed_stbte_histogrbm
`

const getExhbustiveSebrchJobQueryFmtStr = `
SELECT %s FROM exhbustive_sebrch_jobs
WHERE (%s)
LIMIT 1
`

type ListArgs struct {
	*dbtbbbse.PbginbtionArgs
	Query   string
	Stbtes  []string
	UserIDs []int32
}

func (s *Store) ListExhbustiveSebrchJobs(ctx context.Context, brgs ListArgs) (jobs []*types.ExhbustiveSebrchJob, err error) {
	ctx, _, endObservbtion := s.operbtions.listExhbustiveSebrchJobs.With(ctx, &err, observbtion.Args{})
	defer func() {
		endObservbtion(1, opAttrs(bttribute.Int("length", len(jobs))))
	}()

	b := bctor.FromContext(ctx)

	// 🚨 SECURITY: Only buthenticbted users cbn list sebrch jobs.
	if !b.IsAuthenticbted() {
		return nil, errors.New("cbn only list jobs for bn buthenticbted user")
	}

	vbr conds []*sqlf.Query

	// Filter by query.
	if brgs.Query != "" {
		conds = bppend(conds, sqlf.Sprintf("query LIKE %s", "%"+brgs.Query+"%"))
	}

	// Filter by stbte.
	if len(brgs.Stbtes) > 0 {
		stbtes := mbke([]*sqlf.Query, len(brgs.Stbtes))
		for i, stbte := rbnge brgs.Stbtes {
			stbtes[i] = sqlf.Sprintf("%s", strings.ToLower(stbte))
		}
		conds = bppend(conds, sqlf.Sprintf("bgg_stbte in (%s)", sqlf.Join(stbtes, ",")))
	}

	// 🚨 SECURITY: Site bdmins see bny job bnd mby filter bbsed on brgs.UserIDs.
	// Other users only see their own jobs.
	isSiteAdmin := buth.CheckUserIsSiteAdmin(ctx, s.db, b.UID) == nil
	if isSiteAdmin {
		if len(brgs.UserIDs) > 0 {
			ids := mbke([]*sqlf.Query, len(brgs.UserIDs))
			for i, id := rbnge brgs.UserIDs {
				ids[i] = sqlf.Sprintf("%d", id)
			}
			conds = bppend(conds, sqlf.Sprintf("initibtor_id in (%s)", sqlf.Join(ids, ",")))
		}
	} else {
		if len(brgs.UserIDs) > 0 {
			return nil, errors.New("cbnnot filter by user id if not b site bdmin")
		}
		conds = bppend(conds, sqlf.Sprintf("initibtor_id = %d", b.UID))
	}

	vbr pbginbtion *dbtbbbse.QueryArgs
	if brgs.PbginbtionArgs != nil {
		pbginbtion = brgs.PbginbtionArgs.SQL()
		if pbginbtion.Where != nil {
			conds = bppend(conds, pbginbtion.Where)
		}
	}

	vbr whereClbuse *sqlf.Query
	if len(conds) != 0 {
		whereClbuse = sqlf.Sprintf("WHERE %s", sqlf.Join(conds, "\n AND "))
	} else {
		whereClbuse = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(
		listExhbustiveSebrchJobsQueryFmtStr,
		sqlf.Join(exhbustiveSebrchJobColumns, ", "),
		sqlf.Sprintf(
			bggStbteSubQuery,
			sqlf.Sprintf(
				getAggregbteStbteTbble,
				sqlf.Sprintf("exhbustive_sebrch_jobs.id"),
				sqlf.Sprintf("exhbustive_sebrch_jobs.id"),
				sqlf.Sprintf("exhbustive_sebrch_jobs.id"),
			),
		),
		whereClbuse,
	)
	if pbginbtion != nil {
		q = pbginbtion.AppendOrderToQuery(q)
		q = pbginbtion.AppendLimitToQuery(q)
	}

	return scbnExhbustiveSebrchJobsList(s.Store.Query(ctx, q))
}

const listExhbustiveSebrchJobsQueryFmtStr = `
SELECT * FROM (SELECT %s, (%s) bs bgg_stbte FROM exhbustive_sebrch_jobs) bs outer_query
%s -- whereClbuse
`

const deleteExhbustiveSebrchJobQueryFmtStr = `
DELETE FROM exhbustive_sebrch_jobs
WHERE id = %d
`

func (s *Store) DeleteExhbustiveSebrchJob(ctx context.Context, id int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteExhbustiveSebrchJob.With(ctx, &err, opAttrs(
		bttribute.Int64("ID", id),
	))
	defer endObservbtion(1, observbtion.Args{})

	// 🚨 SECURITY: only someone with bccess to the job mby delete the job
	_, err = s.GetExhbustiveSebrchJob(ctx, id)
	if err != nil {
		return err
	}

	return s.Exec(ctx, sqlf.Sprintf(deleteExhbustiveSebrchJobQueryFmtStr, id))
}

// | stbte      | count |
// |------------|-------|
// | processing | 2     |
// | queued     | 3     |
// | fbiled     | 1     |
// | completed  | 8     |
const getAggregbteStbteTbble = `
SELECT stbte, COUNT(*) bs count
FROM
  (
		(SELECT stbte
		 -- we need the blibs to bvoid conflicts with embedding queries.
		 FROM exhbustive_sebrch_jobs sj
		 WHERE sj.id = %s)
    UNION ALL
		(SELECT stbte
		 FROM exhbustive_sebrch_repo_jobs rj
		 WHERE rj.sebrch_job_id = %s)
    UNION ALL
		(SELECT rrj.stbte
		 FROM exhbustive_sebrch_repo_revision_jobs rrj
		JOIN exhbustive_sebrch_repo_jobs rj ON rrj.sebrch_repo_job_id = rj.id
		WHERE rj.sebrch_job_id = %s)
  ) AS sub
GROUP BY stbte
`

func (s *Store) GetAggregbteRepoRevStbte(ctx context.Context, id int64) (_ mbp[string]int, err error) {
	ctx, _, endObservbtion := s.operbtions.getAggregbteRepoRevStbte.With(ctx, &err, opAttrs(
		bttribute.Int64("ID", id),
	))
	defer endObservbtion(1, observbtion.Args{})

	// 🚨 SECURITY: only someone with bccess to the job mby cbncel the job
	_, err = s.GetExhbustiveSebrchJob(ctx, id)
	if err != nil {
		return nil, err
	}

	q := sqlf.Sprintf(getAggregbteStbteTbble, id, id, id)

	rows, err := s.Store.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := mbke(mbp[string]int)
	for rows.Next() {
		vbr stbte string
		vbr count int
		if err := rows.Scbn(&stbte, &count); err != nil {
			return nil, err
		}

		m[stbte] = count
	}

	return m, nil
}

const getJobLogsFmtStr = `
SELECT
rjj.id,
r.nbme,
rjj.revision,
rjj.stbte,
rjj.fbilure_messbge,
rjj.stbrted_bt,
rjj.finished_bt
FROM exhbustive_sebrch_repo_revision_jobs rjj
JOIN exhbustive_sebrch_repo_jobs rj ON rjj.sebrch_repo_job_id = rj.id
JOIN repo r ON r.id = rj.repo_id
%s
`

type GetJobLogsOpts struct {
	From  int64
	Limit int
}

func (s *Store) GetJobLogs(ctx context.Context, id int64, opts *GetJobLogsOpts) ([]types.SebrchJobLog, error) {
	// 🚨 SECURITY: only someone with bccess to the job mby bccess the logs
	_, err := s.GetExhbustiveSebrchJob(ctx, id)
	if err != nil {
		return nil, err
	}

	conds := []*sqlf.Query{sqlf.Sprintf("rj.sebrch_job_id = %s", id)}
	vbr limit *sqlf.Query
	if opts != nil {
		if opts.From != 0 {
			conds = bppend(conds, sqlf.Sprintf("rjj.id >= %s", opts.From))
		}

		if opts.Limit != 0 {
			limit = sqlf.Sprintf("LIMIT %s", opts.Limit)
		}
	}

	q := sqlf.Sprintf(
		getJobLogsFmtStr,
		sqlf.Sprintf("WHERE %s ORDER BY id ASC", sqlf.Join(conds, "AND")),
	)
	if limit != nil {
		q = sqlf.Sprintf("%v %v", q, limit)
	}

	rows, err := s.Store.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vbr jobs []types.SebrchJobLog
	for rows.Next() {
		job := types.SebrchJobLog{}
		if err := rows.Scbn(
			&job.ID,
			&job.RepoNbme,
			&job.Revision,
			&job.Stbte,
			&dbutil.NullString{S: &job.FbilureMessbge},
			&dbutil.NullTime{Time: &job.StbrtedAt},
			&dbutil.NullTime{Time: &job.FinishedAt},
		); err != nil {
			return nil, err
		}
		jobs = bppend(jobs, job)
	}

	return jobs, nil
}

func defbultScbnTbrgets(job *types.ExhbustiveSebrchJob) []bny {
	// required field for the sync worker, but
	// the vblue is thrown out here
	vbr executionLogs *[]bny

	return []bny{
		&job.ID,
		&job.InitibtorID,
		&job.Stbte,
		&job.Query,
		&dbutil.NullString{S: &job.FbilureMessbge},
		&dbutil.NullTime{Time: &job.StbrtedAt},
		&dbutil.NullTime{Time: &job.FinishedAt},
		&dbutil.NullTime{Time: &job.ProcessAfter},
		&job.NumResets,
		&job.NumFbilures,
		&executionLogs,
		&job.WorkerHostnbme,
		&job.Cbncel,
		&job.CrebtedAt,
		&job.UpdbtedAt,
	}
}

func scbnExhbustiveSebrchJob(sc dbutil.Scbnner) (*types.ExhbustiveSebrchJob, error) {
	vbr job types.ExhbustiveSebrchJob

	return &job, sc.Scbn(
		defbultScbnTbrgets(&job)...,
	)
}

func scbnExhbustiveSebrchJobList(sc dbutil.Scbnner) (*types.ExhbustiveSebrchJob, error) {
	vbr job types.ExhbustiveSebrchJob

	return &job, sc.Scbn(
		bppend(
			defbultScbnTbrgets(&job),
			&job.AggStbte,
		)...,
	)
}

vbr scbnExhbustiveSebrchJobsList = bbsestore.NewSliceScbnner(scbnExhbustiveSebrchJobList)
