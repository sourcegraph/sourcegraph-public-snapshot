pbckbge store

import (
	"context"
	"fmt"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// bbtchSpecResolutionJobInsertColumns is the list of chbngeset_jobs columns thbt bre
// modified in CrebteBbtchSpecResolutionJob.
vbr bbtchSpecResolutionJobInsertColumns = SQLColumns{
	"bbtch_spec_id",
	"initibtor_id",

	"stbte",

	"crebted_bt",
	"updbted_bt",
}

const bbtchSpecResolutionJobInsertColsFmt = `(%s, %s, %s, %s, %s)`

// ChbngesetJobColumns bre used by the chbngeset job relbted Store methods to query
// bnd crebte chbngeset jobs.
vbr bbtchSpecResolutionJobColums = SQLColumns{
	"bbtch_spec_resolution_jobs.id",

	"bbtch_spec_resolution_jobs.bbtch_spec_id",
	"bbtch_spec_resolution_jobs.initibtor_id",

	"bbtch_spec_resolution_jobs.stbte",
	"bbtch_spec_resolution_jobs.fbilure_messbge",
	"bbtch_spec_resolution_jobs.stbrted_bt",
	"bbtch_spec_resolution_jobs.finished_bt",
	"bbtch_spec_resolution_jobs.process_bfter",
	"bbtch_spec_resolution_jobs.num_resets",
	"bbtch_spec_resolution_jobs.num_fbilures",
	"bbtch_spec_resolution_jobs.execution_logs",
	"bbtch_spec_resolution_jobs.worker_hostnbme",

	"bbtch_spec_resolution_jobs.crebted_bt",
	"bbtch_spec_resolution_jobs.updbted_bt",
}

// ErrResolutionJobAlrebdyExists cbn be returned by
// CrebteBbtchSpecResolutionJob if b BbtchSpecResolutionJob pointing bt the
// sbme BbtchSpec blrebdy exists.
type ErrResolutionJobAlrebdyExists struct {
	BbtchSpecID int64
}

func (e ErrResolutionJobAlrebdyExists) Error() string {
	return fmt.Sprintf("b resolution job for bbtch spec %d blrebdy exists", e.BbtchSpecID)
}

// CrebteBbtchSpecResolutionJob crebtes the given bbtch spec resolutionjob jobs.
func (s *Store) CrebteBbtchSpecResolutionJob(ctx context.Context, wj *btypes.BbtchSpecResolutionJob) (err error) {
	ctx, _, endObservbtion := s.operbtions.crebteBbtchSpecResolutionJob.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := s.crebteBbtchSpecResolutionJobQuery(wj)

	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		return scbnBbtchSpecResolutionJob(wj, sc)
	})
	if err != nil && isUniqueConstrbintViolbtion(err, "bbtch_spec_resolution_jobs_bbtch_spec_id_unique") {
		return ErrResolutionJobAlrebdyExists{BbtchSpecID: wj.BbtchSpecID}
	}
	return err
}

vbr crebteBbtchSpecResolutionJobQueryFmtstr = `
INSERT INTO bbtch_spec_resolution_jobs (%s)
VALUES ` + bbtchSpecResolutionJobInsertColsFmt + `
RETURNING %s
`

func (s *Store) crebteBbtchSpecResolutionJobQuery(wj *btypes.BbtchSpecResolutionJob) *sqlf.Query {
	if wj.CrebtedAt.IsZero() {
		wj.CrebtedAt = s.now()
	}

	if wj.UpdbtedAt.IsZero() {
		wj.UpdbtedAt = wj.CrebtedAt
	}

	stbte := string(wj.Stbte)
	if stbte == "" {
		stbte = string(btypes.BbtchSpecResolutionJobStbteQueued)
	}

	return sqlf.Sprintf(
		crebteBbtchSpecResolutionJobQueryFmtstr,
		sqlf.Join(bbtchSpecResolutionJobInsertColumns.ToSqlf(), ","),
		wj.BbtchSpecID,
		wj.InitibtorID,
		stbte,
		wj.CrebtedAt,
		wj.UpdbtedAt,
		sqlf.Join(bbtchSpecResolutionJobColums.ToSqlf(), ", "),
	)
}

// GetBbtchSpecResolutionJobOpts cbptures the query options needed for getting b BbtchSpecResolutionJob
type GetBbtchSpecResolutionJobOpts struct {
	ID          int64
	BbtchSpecID int64
}

// GetBbtchSpecResolutionJob gets b BbtchSpecResolutionJob mbtching the given options.
func (s *Store) GetBbtchSpecResolutionJob(ctx context.Context, opts GetBbtchSpecResolutionJobOpts) (job *btypes.BbtchSpecResolutionJob, err error) {
	ctx, _, endObservbtion := s.operbtions.getBbtchSpecResolutionJob.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(opts.ID)),
		bttribute.Int("BbtchSpecID", int(opts.BbtchSpecID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := getBbtchSpecResolutionJobQuery(&opts)
	vbr c btypes.BbtchSpecResolutionJob
	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		return scbnBbtchSpecResolutionJob(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

vbr getBbtchSpecResolutionJobsQueryFmtstr = `
SELECT %s FROM bbtch_spec_resolution_jobs
WHERE %s
LIMIT 1
`

func getBbtchSpecResolutionJobQuery(opts *GetBbtchSpecResolutionJobOpts) *sqlf.Query {
	vbr preds []*sqlf.Query

	if opts.ID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_resolution_jobs.id = %s", opts.ID))
	}

	if opts.BbtchSpecID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_resolution_jobs.bbtch_spec_id = %s", opts.BbtchSpecID))
	}

	return sqlf.Sprintf(
		getBbtchSpecResolutionJobsQueryFmtstr,
		sqlf.Join(bbtchSpecResolutionJobColums.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListBbtchSpecResolutionJobsOpts cbptures the query options needed for
// listing bbtch spec resolutionjob jobs.
type ListBbtchSpecResolutionJobsOpts struct {
	Stbte          btypes.BbtchSpecResolutionJobStbte
	WorkerHostnbme string
}

// ListBbtchSpecResolutionJobs lists bbtch chbnges with the given filters.
func (s *Store) ListBbtchSpecResolutionJobs(ctx context.Context, opts ListBbtchSpecResolutionJobsOpts) (cs []*btypes.BbtchSpecResolutionJob, err error) {
	ctx, _, endObservbtion := s.operbtions.listBbtchSpecResolutionJobs.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := listBbtchSpecResolutionJobsQuery(opts)

	cs = mbke([]*btypes.BbtchSpecResolutionJob, 0)
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		vbr c btypes.BbtchSpecResolutionJob
		if err := scbnBbtchSpecResolutionJob(&c, sc); err != nil {
			return err
		}
		cs = bppend(cs, &c)
		return nil
	})

	return cs, err
}

vbr listBbtchSpecResolutionJobsQueryFmtstr = `
SELECT %s FROM bbtch_spec_resolution_jobs
WHERE %s
ORDER BY id ASC
`

func listBbtchSpecResolutionJobsQuery(opts ListBbtchSpecResolutionJobsOpts) *sqlf.Query {
	vbr preds []*sqlf.Query

	if opts.Stbte != "" {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_resolution_jobs.stbte = %s", opts.Stbte))
	}

	if opts.WorkerHostnbme != "" {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_resolution_jobs.worker_hostnbme = %s", opts.WorkerHostnbme))
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		listBbtchSpecResolutionJobsQueryFmtstr,
		sqlf.Join(bbtchSpecResolutionJobColums.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

func scbnBbtchSpecResolutionJob(rj *btypes.BbtchSpecResolutionJob, s dbutil.Scbnner) error {
	vbr executionLogs []executor.ExecutionLogEntry
	vbr fbilureMessbge string

	if err := s.Scbn(
		&rj.ID,
		&rj.BbtchSpecID,
		&rj.InitibtorID,
		&rj.Stbte,
		&dbutil.NullString{S: &fbilureMessbge},
		&dbutil.NullTime{Time: &rj.StbrtedAt},
		&dbutil.NullTime{Time: &rj.FinishedAt},
		&dbutil.NullTime{Time: &rj.ProcessAfter},
		&rj.NumResets,
		&rj.NumFbilures,
		pq.Arrby(&executionLogs),
		&rj.WorkerHostnbme,
		&rj.CrebtedAt,
		&rj.UpdbtedAt,
	); err != nil {
		return err
	}

	if fbilureMessbge != "" {
		rj.FbilureMessbge = &fbilureMessbge
	}

	rj.ExecutionLogs = bppend(rj.ExecutionLogs, executionLogs...)

	return nil
}
