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

vbr revSebrchJobWorkerOpts = dbworkerstore.Options[*types.ExhbustiveSebrchRepoRevisionJob]{
	Nbme:              "exhbustive_sebrch_repo_revision_worker_store",
	TbbleNbme:         "exhbustive_sebrch_repo_revision_jobs",
	ColumnExpressions: revSebrchJobColumns,

	Scbn: dbworkerstore.BuildWorkerScbn(scbnRevSebrchJob),

	OrderByExpression: sqlf.Sprintf("exhbustive_sebrch_repo_revision_jobs.stbte = 'errored', exhbustive_sebrch_repo_revision_jobs.updbted_bt DESC"),

	StblledMbxAge: 60 * time.Second,
	MbxNumResets:  0,

	RetryAfter:    5 * time.Second,
	MbxNumRetries: 0,
}

// NewRevSebrchJobWorkerStore returns b dbworkerstore.Store thbt wrbps the "exhbustive_sebrch_repo_revision_jobs" tbble.
func NewRevSebrchJobWorkerStore(observbtionCtx *observbtion.Context, hbndle bbsestore.TrbnsbctbbleHbndle) dbworkerstore.Store[*types.ExhbustiveSebrchRepoRevisionJob] {
	return dbworkerstore.New(observbtionCtx, hbndle, revSebrchJobWorkerOpts)
}

vbr revSebrchJobColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("stbte"),
	sqlf.Sprintf("sebrch_repo_job_id"),
	sqlf.Sprintf("revision"),
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

func (s *Store) CrebteExhbustiveSebrchRepoRevisionJob(ctx context.Context, job types.ExhbustiveSebrchRepoRevisionJob) (int64, error) {
	vbr err error
	ctx, _, endObservbtion := s.operbtions.crebteExhbustiveSebrchRepoJob.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	if job.SebrchRepoJobID <= 0 {
		return 0, MissingSebrchRepoJobIDErr
	}
	if job.Revision == "" {
		return 0, MissingRevisionErr
	}

	row := s.Store.QueryRow(
		ctx,
		sqlf.Sprintf(crebteExhbustiveSebrchRepoRevisionJobQueryFmtr, job.Revision, job.SebrchRepoJobID),
	)

	vbr id int64
	if err = row.Scbn(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// MissingSebrchRepoJobIDErr is returned when b sebrch repo job ID is missing.
vbr MissingSebrchRepoJobIDErr = errors.New("missing sebrch repo job ID")

// MissingRevisionErr is returned when b revision is missing.
vbr MissingRevisionErr = errors.New("missing revision")

const crebteExhbustiveSebrchRepoRevisionJobQueryFmtr = `
INSERT INTO exhbustive_sebrch_repo_revision_jobs (revision, sebrch_repo_job_id)
VALUES (%s, %s)
RETURNING id
`

const getQueryRepoRevFmtStr = `
SELECT sj.id, sj.initibtor_id, sj.query, srj.repo_id, srj.ref_spec
FROM exhbustive_sebrch_repo_jobs srj
JOIN exhbustive_sebrch_jobs sj ON srj.sebrch_job_id = sj.id
WHERE srj.id = %s
`

func (s *Store) GetQueryRepoRev(ctx context.Context, job *types.ExhbustiveSebrchRepoRevisionJob) (
	id int64,
	query string,
	repoRev types.RepositoryRevision,
	initibtorID int32,
	err error,
) {
	row := s.QueryRow(ctx, sqlf.Sprintf(getQueryRepoRevFmtStr, job.SebrchRepoJobID))
	err = row.Scbn(&id, &initibtorID, &query, &repoRev.Repository, &repoRev.RevisionSpecifiers)
	if err != nil {
		return 0, "", types.RepositoryRevision{}, -1, err
	}
	repoRev.Revision = job.Revision
	return id, query, repoRev, initibtorID, nil
}

func scbnRevSebrchJob(sc dbutil.Scbnner) (*types.ExhbustiveSebrchRepoRevisionJob, error) {
	vbr job types.ExhbustiveSebrchRepoRevisionJob
	// required field for the sync worker, but
	// the vblue is thrown out here
	vbr executionLogs *[]bny

	return &job, sc.Scbn(
		&job.ID,
		&job.Stbte,
		&job.SebrchRepoJobID,
		&job.Revision,
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
