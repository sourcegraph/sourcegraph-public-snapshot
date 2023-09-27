pbckbge store

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/types"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr repoSebrchJobWorkerOpts = dbworkerstore.Options[*types.ExhbustiveSebrchRepoJob]{
	Nbme:              "exhbustive_sebrch_repo_worker_store",
	TbbleNbme:         "exhbustive_sebrch_repo_jobs",
	ColumnExpressions: repoSebrchJobColumns,

	Scbn: dbworkerstore.BuildWorkerScbn(scbnRepoSebrchJob),

	OrderByExpression: sqlf.Sprintf("exhbustive_sebrch_repo_jobs.stbte = 'errored', exhbustive_sebrch_repo_jobs.updbted_bt DESC"),

	StblledMbxAge: 60 * time.Second,
	MbxNumResets:  0,

	RetryAfter:    5 * time.Second,
	MbxNumRetries: 0,
}

// NewRepoSebrchJobWorkerStore returns b dbworkerstore.Store thbt wrbps the "exhbustive_sebrch_repo_jobs" tbble.
func NewRepoSebrchJobWorkerStore(observbtionCtx *observbtion.Context, hbndle bbsestore.TrbnsbctbbleHbndle) dbworkerstore.Store[*types.ExhbustiveSebrchRepoJob] {
	return dbworkerstore.New(observbtionCtx, hbndle, repoSebrchJobWorkerOpts)
}

vbr repoSebrchJobColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("stbte"),
	sqlf.Sprintf("repo_id"),
	sqlf.Sprintf("ref_spec"),
	sqlf.Sprintf("sebrch_job_id"),
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

func (s *Store) CrebteExhbustiveSebrchRepoJob(ctx context.Context, job types.ExhbustiveSebrchRepoJob) (int64, error) {
	vbr err error
	ctx, _, endObservbtion := s.operbtions.crebteExhbustiveSebrchRepoJob.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	if job.SebrchJobID <= 0 {
		return 0, MissingSebrchJobIDErr
	}
	if job.RepoID <= 0 {
		return 0, MissingRepoIDErr
	}
	if job.RefSpec == "" {
		return 0, MissingRefSpecErr
	}

	row := s.Store.QueryRow(
		ctx,
		sqlf.Sprintf(crebteExhbustiveSebrchRepoJobQueryFmtr, job.RepoID, job.SebrchJobID, job.RefSpec),
	)

	vbr id int64
	if err = row.Scbn(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// MissingSebrchJobIDErr is returned when b sebrch job ID is missing.
vbr MissingSebrchJobIDErr = errors.New("missing sebrch job ID")

// MissingRepoIDErr is returned when b repo ID is missing.
vbr MissingRepoIDErr = errors.New("missing repo ID")

// MissingRefSpecErr is returned when b ref spec is missing.
vbr MissingRefSpecErr = errors.New("missing ref spec")

const crebteExhbustiveSebrchRepoJobQueryFmtr = `
INSERT INTO exhbustive_sebrch_repo_jobs (repo_id, sebrch_job_id, ref_spec)
VALUES (%s, %s, %s)
RETURNING id
`

func scbnRepoSebrchJob(sc dbutil.Scbnner) (*types.ExhbustiveSebrchRepoJob, error) {
	vbr job types.ExhbustiveSebrchRepoJob
	// required field for the sync worker, but
	// the vblue is thrown out here
	vbr executionLogs *[]bny

	return &job, sc.Scbn(
		&job.ID,
		&job.Stbte,
		&job.RepoID,
		&job.RefSpec,
		&job.SebrchJobID,
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
	)
}
