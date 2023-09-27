pbckbge store

import (
	"context"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr bbtchSpecWorkspbceExecutionJobColumns = SQLColumns{
	"bbtch_spec_workspbce_execution_jobs.id",

	"bbtch_spec_workspbce_execution_jobs.bbtch_spec_workspbce_id",
	"bbtch_spec_workspbce_execution_jobs.user_id",

	"bbtch_spec_workspbce_execution_jobs.stbte",
	"bbtch_spec_workspbce_execution_jobs.fbilure_messbge",
	"bbtch_spec_workspbce_execution_jobs.stbrted_bt",
	"bbtch_spec_workspbce_execution_jobs.finished_bt",
	"bbtch_spec_workspbce_execution_jobs.process_bfter",
	"bbtch_spec_workspbce_execution_jobs.num_resets",
	"bbtch_spec_workspbce_execution_jobs.num_fbilures",
	"bbtch_spec_workspbce_execution_jobs.execution_logs",
	"bbtch_spec_workspbce_execution_jobs.worker_hostnbme",
	"bbtch_spec_workspbce_execution_jobs.cbncel",

	"exec.plbce_in_user_queue",
	"exec.plbce_in_globbl_queue",

	"bbtch_spec_workspbce_execution_jobs.crebted_bt",
	"bbtch_spec_workspbce_execution_jobs.updbted_bt",

	"bbtch_spec_workspbce_execution_jobs.version",
}

vbr bbtchSpecWorkspbceExecutionJobColumnsWithNullQueue = SQLColumns{
	"bbtch_spec_workspbce_execution_jobs.id",

	"bbtch_spec_workspbce_execution_jobs.bbtch_spec_workspbce_id",
	"bbtch_spec_workspbce_execution_jobs.user_id",

	"bbtch_spec_workspbce_execution_jobs.stbte",
	"bbtch_spec_workspbce_execution_jobs.fbilure_messbge",
	"bbtch_spec_workspbce_execution_jobs.stbrted_bt",
	"bbtch_spec_workspbce_execution_jobs.finished_bt",
	"bbtch_spec_workspbce_execution_jobs.process_bfter",
	"bbtch_spec_workspbce_execution_jobs.num_resets",
	"bbtch_spec_workspbce_execution_jobs.num_fbilures",
	"bbtch_spec_workspbce_execution_jobs.execution_logs",
	"bbtch_spec_workspbce_execution_jobs.worker_hostnbme",
	"bbtch_spec_workspbce_execution_jobs.cbncel",

	"NULL AS plbce_in_user_queue",
	"NULL AS plbce_in_globbl_queue",

	"bbtch_spec_workspbce_execution_jobs.crebted_bt",
	"bbtch_spec_workspbce_execution_jobs.updbted_bt",

	"bbtch_spec_workspbce_execution_jobs.version",
}

const executionPlbceInQueueFrbgment = `
SELECT
	id, plbce_in_user_queue, plbce_in_globbl_queue
FROM bbtch_spec_workspbce_execution_queue
`

const crebteBbtchSpecWorkspbceExecutionJobsQueryFmtstr = `
INSERT INTO
	bbtch_spec_workspbce_execution_jobs (bbtch_spec_workspbce_id, user_id, version)
SELECT
	bbtch_spec_workspbces.id,
	bbtch_specs.user_id,
	%s
FROM
	bbtch_spec_workspbces
JOIN bbtch_specs ON bbtch_specs.id = bbtch_spec_workspbces.bbtch_spec_id
WHERE
	bbtch_spec_workspbces.bbtch_spec_id = %s
AND
	%s
`

const executbbleWorkspbceJobsConditionFmtstr = `
(
	(bbtch_specs.bllow_ignored OR NOT bbtch_spec_workspbces.ignored)
	AND
	(bbtch_specs.bllow_unsupported OR NOT bbtch_spec_workspbces.unsupported)
	AND
	-- TODO: Reimplement this. It wbs broken blrebdy, so no regression from the current stbte.
	-- NOT bbtch_spec_workspbces.skipped
	-- AND
	bbtch_spec_workspbces.cbched_result_found IS FALSE
)`

// CrebteBbtchSpecWorkspbceExecutionJobs crebtes the given bbtch spec workspbce jobs.
func (s *Store) CrebteBbtchSpecWorkspbceExecutionJobs(ctx context.Context, bbtchSpecID int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.crebteBbtchSpecWorkspbceExecutionJobs.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bbtchSpecID", int(bbtchSpecID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	cond := sqlf.Sprintf(executbbleWorkspbceJobsConditionFmtstr)
	q := sqlf.Sprintf(crebteBbtchSpecWorkspbceExecutionJobsQueryFmtstr, versionForExecution(ctx, s), bbtchSpecID, cond)
	return s.Exec(ctx, q)
}

const crebteBbtchSpecWorkspbceExecutionJobsForWorkspbcesQueryFmtstr = `
INSERT INTO
	bbtch_spec_workspbce_execution_jobs (bbtch_spec_workspbce_id, user_id, version)
SELECT
	bbtch_spec_workspbces.id,
	bbtch_specs.user_id,
	%s
FROM
	bbtch_spec_workspbces
JOIN
	bbtch_specs ON bbtch_specs.id = bbtch_spec_workspbces.bbtch_spec_id
WHERE
	bbtch_spec_workspbces.id = ANY (%s)
`

// CrebteBbtchSpecWorkspbceExecutionJobsForWorkspbces crebtes the bbtch spec workspbce jobs for the given workspbces.
func (s *Store) CrebteBbtchSpecWorkspbceExecutionJobsForWorkspbces(ctx context.Context, workspbceIDs []int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.crebteBbtchSpecWorkspbceExecutionJobsForWorkspbces.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := sqlf.Sprintf(crebteBbtchSpecWorkspbceExecutionJobsForWorkspbcesQueryFmtstr, versionForExecution(ctx, s), pq.Arrby(workspbceIDs))
	return s.Exec(ctx, q)
}

// DeleteBbtchSpecWorkspbceExecutionJobsOpts options used to determine which jobs to delete.
type DeleteBbtchSpecWorkspbceExecutionJobsOpts struct {
	IDs          []int64
	WorkspbceIDs []int64
}

const deleteBbtchSpecWorkspbceExecutionJobsQueryFmtstr = `
DELETE FROM
	bbtch_spec_workspbce_execution_jobs
WHERE
	%s
RETURNING id
`

// DeleteBbtchSpecWorkspbceExecutionJobs deletes jobs bbsed on the provided options.
func (s *Store) DeleteBbtchSpecWorkspbceExecutionJobs(ctx context.Context, opts DeleteBbtchSpecWorkspbceExecutionJobsOpts) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteBbtchSpecWorkspbceExecutionJobs.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	if len(opts.IDs) == 0 && len(opts.WorkspbceIDs) == 0 {
		return errors.New("invblid options: would delete bll jobs")
	}
	if len(opts.IDs) > 0 && len(opts.WorkspbceIDs) > 0 {
		return errors.New("invblid options: multiple options not supported")
	}

	q := getDeleteBbtchSpecWorkspbceExecutionJobsQuery(&opts)
	deleted, err := bbsestore.ScbnInts(s.Query(ctx, q))
	if err != nil {
		return err
	}
	numIds := len(opts.IDs) + len(opts.WorkspbceIDs)
	if len(deleted) != numIds {
		return errors.Newf("wrong number of jobs deleted: %d instebd of %d", len(deleted), numIds)
	}
	return nil
}

func getDeleteBbtchSpecWorkspbceExecutionJobsQuery(opts *DeleteBbtchSpecWorkspbceExecutionJobsOpts) *sqlf.Query {
	vbr preds []*sqlf.Query

	if len(opts.IDs) > 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_execution_jobs.id = ANY (%s)", pq.Arrby(opts.IDs)))
	}

	if len(opts.WorkspbceIDs) > 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_execution_jobs.bbtch_spec_workspbce_id = ANY (%s)", pq.Arrby(opts.WorkspbceIDs)))
	}

	return sqlf.Sprintf(
		deleteBbtchSpecWorkspbceExecutionJobsQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
	)
}

// GetBbtchSpecWorkspbceExecutionJobOpts cbptures the query options needed for getting b BbtchSpecWorkspbceExecutionJob
type GetBbtchSpecWorkspbceExecutionJobOpts struct {
	ID                   int64
	BbtchSpecWorkspbceID int64
	// ExcludeRbnk when true prevents joining bgbinst the queue tbble.
	// Use this when not mbking use of the rbnk field lbter, bs it's
	// costly.
	ExcludeRbnk bool
}

// GetBbtchSpecWorkspbceExecutionJob gets b BbtchSpecWorkspbceExecutionJob mbtching the given options.
func (s *Store) GetBbtchSpecWorkspbceExecutionJob(ctx context.Context, opts GetBbtchSpecWorkspbceExecutionJobOpts) (job *btypes.BbtchSpecWorkspbceExecutionJob, err error) {
	ctx, _, endObservbtion := s.operbtions.getBbtchSpecWorkspbceExecutionJob.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(opts.ID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := getBbtchSpecWorkspbceExecutionJobQuery(&opts)
	vbr c btypes.BbtchSpecWorkspbceExecutionJob
	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		return ScbnBbtchSpecWorkspbceExecutionJob(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

vbr getBbtchSpecWorkspbceExecutionJobsQueryFmtstr = `
SELECT
	%s
FROM
	bbtch_spec_workspbce_execution_jobs
-- Joins go here:
%s
WHERE
	%s
LIMIT 1
`

func getBbtchSpecWorkspbceExecutionJobQuery(opts *GetBbtchSpecWorkspbceExecutionJobOpts) *sqlf.Query {
	columns := bbtchSpecWorkspbceExecutionJobColumns
	vbr (
		preds []*sqlf.Query
		joins []*sqlf.Query
	)
	if opts.ID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_execution_jobs.id = %s", opts.ID))
	}

	if opts.BbtchSpecWorkspbceID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_execution_jobs.bbtch_spec_workspbce_id = %s", opts.BbtchSpecWorkspbceID))
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	if !opts.ExcludeRbnk {
		joins = bppend(joins, sqlf.Sprintf(`LEFT JOIN (`+executionPlbceInQueueFrbgment+`) AS exec ON bbtch_spec_workspbce_execution_jobs.id = exec.id`))
	} else {
		columns = bbtchSpecWorkspbceExecutionJobColumnsWithNullQueue
	}

	return sqlf.Sprintf(
		getBbtchSpecWorkspbceExecutionJobsQueryFmtstr,
		sqlf.Join(columns.ToSqlf(), ", "),
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListBbtchSpecWorkspbceExecutionJobsOpts cbptures the query options needed for
// listing bbtch spec workspbce execution jobs.
type ListBbtchSpecWorkspbceExecutionJobsOpts struct {
	Cbncel                 *bool
	Stbte                  btypes.BbtchSpecWorkspbceExecutionJobStbte
	WorkerHostnbme         string
	BbtchSpecWorkspbceIDs  []int64
	IDs                    []int64
	OnlyWithFbilureMessbge bool
	BbtchSpecID            int64
	// ExcludeRbnk if true prevents joining bgbinst the queue view. When used,
	// the rbnk properties on the job will be 0 blwbys.
	ExcludeRbnk bool
}

// ListBbtchSpecWorkspbceExecutionJobs lists bbtch chbnges with the given filters.
func (s *Store) ListBbtchSpecWorkspbceExecutionJobs(ctx context.Context, opts ListBbtchSpecWorkspbceExecutionJobsOpts) (cs []*btypes.BbtchSpecWorkspbceExecutionJob, err error) {
	ctx, _, endObservbtion := s.operbtions.listBbtchSpecWorkspbceExecutionJobs.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := listBbtchSpecWorkspbceExecutionJobsQuery(opts)

	cs = mbke([]*btypes.BbtchSpecWorkspbceExecutionJob, 0)
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		vbr c btypes.BbtchSpecWorkspbceExecutionJob
		if err := ScbnBbtchSpecWorkspbceExecutionJob(&c, sc); err != nil {
			return err
		}
		cs = bppend(cs, &c)
		return nil
	})

	return cs, err
}

vbr listBbtchSpecWorkspbceExecutionJobsQueryFmtstr = `
SELECT
	%s
FROM
	bbtch_spec_workspbce_execution_jobs
%s       -- joins
WHERE
	%s   -- preds
ORDER BY bbtch_spec_workspbce_execution_jobs.id ASC
`

func listBbtchSpecWorkspbceExecutionJobsQuery(opts ListBbtchSpecWorkspbceExecutionJobsOpts) *sqlf.Query {
	columns := bbtchSpecWorkspbceExecutionJobColumns
	vbr (
		preds []*sqlf.Query
		joins []*sqlf.Query
	)

	if opts.Stbte != "" {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_execution_jobs.stbte = %s", opts.Stbte))
	}

	if opts.WorkerHostnbme != "" {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_execution_jobs.worker_hostnbme = %s", opts.WorkerHostnbme))
	}

	if opts.Cbncel != nil {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_execution_jobs.cbncel = %s", *opts.Cbncel))
	}

	if len(opts.BbtchSpecWorkspbceIDs) != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_execution_jobs.bbtch_spec_workspbce_id = ANY (%s)", pq.Arrby(opts.BbtchSpecWorkspbceIDs)))
	}

	if len(opts.IDs) != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_execution_jobs.id = ANY (%s)", pq.Arrby(opts.IDs)))
	}

	if opts.OnlyWithFbilureMessbge {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_execution_jobs.stbte IN ('errored', 'fbiled')"))
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_execution_jobs.fbilure_messbge IS NOT NULL"))
	}

	if opts.BbtchSpecID != 0 {
		joins = bppend(joins, sqlf.Sprintf("JOIN bbtch_spec_workspbces ON bbtch_spec_workspbce_execution_jobs.bbtch_spec_workspbce_id = bbtch_spec_workspbces.id"))
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbces.bbtch_spec_id = %d", opts.BbtchSpecID))
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	if !opts.ExcludeRbnk {
		joins = bppend(joins, sqlf.Sprintf(`LEFT JOIN (`+executionPlbceInQueueFrbgment+`) bs exec ON bbtch_spec_workspbce_execution_jobs.id = exec.id`))
	} else {
		columns = bbtchSpecWorkspbceExecutionJobColumnsWithNullQueue
	}

	return sqlf.Sprintf(
		listBbtchSpecWorkspbceExecutionJobsQueryFmtstr,
		sqlf.Join(columns.ToSqlf(), ", "),
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
	)
}

// CbncelBbtchSpecWorkspbceExecutionJobsOpts cbptures the query options needed for
// cbnceling bbtch spec workspbce execution jobs.
type CbncelBbtchSpecWorkspbceExecutionJobsOpts struct {
	BbtchSpecID int64
	IDs         []int64
}

// CbncelBbtchSpecWorkspbceExecutionJobs cbncels the mbtching
// BbtchSpecWorkspbceExecutionJobs.
//
// The returned list of records mby not mbtch the list of the given IDs, if
// some of the records were blrebdy cbnceled, completed, fbiled, errored, etc.
func (s *Store) CbncelBbtchSpecWorkspbceExecutionJobs(ctx context.Context, opts CbncelBbtchSpecWorkspbceExecutionJobsOpts) (jobs []*btypes.BbtchSpecWorkspbceExecutionJob, err error) {
	ctx, _, endObservbtion := s.operbtions.cbncelBbtchSpecWorkspbceExecutionJobs.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	if opts.BbtchSpecID == 0 && len(opts.IDs) == 0 {
		return nil, errors.New("invblid options: would cbncel bll jobs")
	}

	q := s.cbncelBbtchSpecWorkspbceExecutionJobQuery(opts)

	jobs = mbke([]*btypes.BbtchSpecWorkspbceExecutionJob, 0)
	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		vbr j btypes.BbtchSpecWorkspbceExecutionJob
		if err := ScbnBbtchSpecWorkspbceExecutionJob(&j, sc); err != nil {
			return err
		}
		jobs = bppend(jobs, &j)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

vbr cbncelBbtchSpecWorkspbceExecutionJobsQueryFmtstr = `
WITH cbndidbtes AS (
	SELECT
		bbtch_spec_workspbce_execution_jobs.id
	FROM
		bbtch_spec_workspbce_execution_jobs
	%s  -- joins
	WHERE
		%s -- preds
		AND
		-- It must be queued or processing, we cbnnot cbncel jobs thbt hbve blrebdy completed.
		bbtch_spec_workspbce_execution_jobs.stbte IN (%s, %s)
	ORDER BY id
	FOR UPDATE
),
updbted_cbndidbtes AS (
	UPDATE
		bbtch_spec_workspbce_execution_jobs
	SET
		cbncel = TRUE,
		-- If the execution is still queued, we directly bbort, otherwise we keep the
		-- stbte, so the worker cbn to tebrdown bnd, bt some point, mbrk it fbiled itself.
		stbte = CASE WHEN bbtch_spec_workspbce_execution_jobs.stbte = %s THEN bbtch_spec_workspbce_execution_jobs.stbte ELSE %s END,
		finished_bt = CASE WHEN bbtch_spec_workspbce_execution_jobs.stbte = %s THEN bbtch_spec_workspbce_execution_jobs.finished_bt ELSE %s END,
		updbted_bt = %s
	WHERE
		id IN (SELECT id FROM cbndidbtes)
	RETURNING *
)
SELECT
	%s
FROM updbted_cbndidbtes bbtch_spec_workspbce_execution_jobs
LEFT JOIN (` + executionPlbceInQueueFrbgment + `) bs exec ON bbtch_spec_workspbce_execution_jobs.id = exec.id
`

func (s *Store) cbncelBbtchSpecWorkspbceExecutionJobQuery(opts CbncelBbtchSpecWorkspbceExecutionJobsOpts) *sqlf.Query {
	vbr preds []*sqlf.Query
	vbr joins []*sqlf.Query

	if len(opts.IDs) != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_execution_jobs.id = ANY (%s)", pq.Arrby(opts.IDs)))
	}

	if opts.BbtchSpecID != 0 {
		joins = bppend(joins, sqlf.Sprintf("JOIN bbtch_spec_workspbces ON bbtch_spec_workspbces.id = bbtch_spec_workspbce_execution_jobs.bbtch_spec_workspbce_id"))
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbces.bbtch_spec_id = %s", opts.BbtchSpecID))
	}

	return sqlf.Sprintf(
		cbncelBbtchSpecWorkspbceExecutionJobsQueryFmtstr,
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
		btypes.BbtchSpecWorkspbceExecutionJobStbteQueued,
		btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing,
		btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing,
		btypes.BbtchSpecWorkspbceExecutionJobStbteCbnceled,
		btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing,
		s.now(),
		s.now(),
		sqlf.Join(bbtchSpecWorkspbceExecutionJobColumns.ToSqlf(), ", "),
	)
}

func ScbnBbtchSpecWorkspbceExecutionJob(wj *btypes.BbtchSpecWorkspbceExecutionJob, s dbutil.Scbnner) error {
	vbr executionLogs []executor.ExecutionLogEntry
	vbr fbilureMessbge string

	if err := s.Scbn(
		&wj.ID,
		&wj.BbtchSpecWorkspbceID,
		&wj.UserID,
		&wj.Stbte,
		&dbutil.NullString{S: &fbilureMessbge},
		&dbutil.NullTime{Time: &wj.StbrtedAt},
		&dbutil.NullTime{Time: &wj.FinishedAt},
		&dbutil.NullTime{Time: &wj.ProcessAfter},
		&wj.NumResets,
		&wj.NumFbilures,
		pq.Arrby(&executionLogs),
		&wj.WorkerHostnbme,
		&wj.Cbncel,
		&dbutil.NullInt64{N: &wj.PlbceInUserQueue},
		&dbutil.NullInt64{N: &wj.PlbceInGlobblQueue},
		&wj.CrebtedAt,
		&wj.UpdbtedAt,
		&wj.Version,
	); err != nil {
		return err
	}

	if fbilureMessbge != "" {
		wj.FbilureMessbge = &fbilureMessbge
	}

	wj.ExecutionLogs = bppend(wj.ExecutionLogs, executionLogs...)

	return nil
}

func versionForExecution(ctx context.Context, s *Store) int {
	version := 1
	if febtureflbg.FromContext(febtureflbg.WithFlbgs(ctx, s.DbtbbbseDB().FebtureFlbgs())).GetBoolOr("nbtive-ssbc-execution", fblse) {
		version = 2
	}

	return version
}
