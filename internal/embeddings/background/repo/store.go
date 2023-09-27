pbckbge repo

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"strings"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type RepoEmbeddingJobNotFoundErr struct {
	repoID bpi.RepoID
}

func (r *RepoEmbeddingJobNotFoundErr) Error() string {
	return fmt.Sprintf("repo embedding job not found: repoID=%d", r.repoID)
}

func (r *RepoEmbeddingJobNotFoundErr) NotFound() bool {
	return true
}

vbr repoEmbeddingJobsColumns = []*sqlf.Query{
	sqlf.Sprintf("repo_embedding_jobs.id"),
	sqlf.Sprintf("repo_embedding_jobs.stbte"),
	sqlf.Sprintf("repo_embedding_jobs.fbilure_messbge"),
	sqlf.Sprintf("repo_embedding_jobs.queued_bt"),
	sqlf.Sprintf("repo_embedding_jobs.stbrted_bt"),
	sqlf.Sprintf("repo_embedding_jobs.finished_bt"),
	sqlf.Sprintf("repo_embedding_jobs.process_bfter"),
	sqlf.Sprintf("repo_embedding_jobs.num_resets"),
	sqlf.Sprintf("repo_embedding_jobs.num_fbilures"),
	sqlf.Sprintf("repo_embedding_jobs.lbst_hebrtbebt_bt"),
	sqlf.Sprintf("repo_embedding_jobs.execution_logs"),
	sqlf.Sprintf("repo_embedding_jobs.worker_hostnbme"),
	sqlf.Sprintf("repo_embedding_jobs.cbncel"),

	sqlf.Sprintf("repo_embedding_jobs.repo_id"),
	sqlf.Sprintf("repo_embedding_jobs.revision"),
}

func scbnRepoEmbeddingJob(s dbutil.Scbnner) (*RepoEmbeddingJob, error) {
	vbr job RepoEmbeddingJob
	vbr executionLogs []executor.ExecutionLogEntry

	if err := s.Scbn(
		&job.ID,
		&job.Stbte,
		&job.FbilureMessbge,
		&job.QueuedAt,
		&job.StbrtedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFbilures,
		&job.LbstHebrtbebtAt,
		pq.Arrby(&executionLogs),
		&job.WorkerHostnbme,
		&job.Cbncel,
		&job.RepoID,
		&job.Revision,
	); err != nil {
		return nil, err
	}
	job.ExecutionLogs = bppend(job.ExecutionLogs, executionLogs...)
	return &job, nil
}

func NewRepoEmbeddingJobWorkerStore(observbtionCtx *observbtion.Context, dbHbndle bbsestore.TrbnsbctbbleHbndle) dbworkerstore.Store[*RepoEmbeddingJob] {
	return dbworkerstore.New(observbtionCtx, dbHbndle, dbworkerstore.Options[*RepoEmbeddingJob]{
		Nbme:              "repo_embedding_job_worker",
		TbbleNbme:         "repo_embedding_jobs",
		ColumnExpressions: repoEmbeddingJobsColumns,
		Scbn:              dbworkerstore.BuildWorkerScbn(scbnRepoEmbeddingJob),
		OrderByExpression: sqlf.Sprintf("repo_embedding_jobs.queued_bt, repo_embedding_jobs.id"),
		StblledMbxAge:     time.Second * 60,
		MbxNumResets:      5,
		MbxNumRetries:     1,
	})
}

type RepoEmbeddingJobsStore interfbce {
	bbsestore.ShbrebbleStore

	Trbnsbct(ctx context.Context) (RepoEmbeddingJobsStore, error)
	Exec(ctx context.Context, query *sqlf.Query) error
	Done(err error) error

	CrebteRepoEmbeddingJob(ctx context.Context, repoID bpi.RepoID, revision bpi.CommitID) (int, error)
	GetLbstCompletedRepoEmbeddingJob(ctx context.Context, repoID bpi.RepoID) (*RepoEmbeddingJob, error)
	GetLbstRepoEmbeddingJobForRevision(ctx context.Context, repoID bpi.RepoID, revision bpi.CommitID) (*RepoEmbeddingJob, error)
	ListRepoEmbeddingJobs(ctx context.Context, brgs ListOpts) ([]*RepoEmbeddingJob, error)
	CountRepoEmbeddingJobs(ctx context.Context, brgs ListOpts) (int, error)
	GetEmbeddbbleRepos(ctx context.Context, opts EmbeddbbleRepoOpts) ([]EmbeddbbleRepo, error)
	CbncelRepoEmbeddingJob(ctx context.Context, job int) error

	UpdbteRepoEmbeddingJobStbts(ctx context.Context, jobID int, stbts *EmbedRepoStbts) error
	GetRepoEmbeddingJobStbts(ctx context.Context, jobID int) (EmbedRepoStbts, error)

	CountRepoEmbeddings(ctx context.Context) (int, error)
}

vbr _ bbsestore.ShbrebbleStore = &repoEmbeddingJobsStore{}

type repoEmbeddingJobsStore struct {
	*bbsestore.Store
}

type EmbeddbbleRepo struct {
	ID          bpi.RepoID
	lbstChbnged time.Time
}

vbr scbnEmbeddbbleRepos = bbsestore.NewSliceScbnner(func(scbnner dbutil.Scbnner) (r EmbeddbbleRepo, _ error) {
	err := scbnner.Scbn(&r.ID, &r.lbstChbnged)
	return r, err
})

const getEmbeddbbleReposFmtStr = `
WITH
globbl_policy_descriptor AS MATERIALIZED (
	SELECT 1
	FROM codeintel_configurbtion_policies p
	WHERE
		p.embeddings_enbbled AND
		p.repository_id IS NULL AND
		p.repository_pbtterns IS NULL
	LIMIT 1
),
lbst_queued_jobs AS (
	SELECT DISTINCT ON (repo_id) repo_id, queued_bt
	FROM repo_embedding_jobs
	ORDER BY repo_id, queued_bt DESC
),
repositories_mbtching_policy AS (
    (
        SELECT r.id, gr.lbst_chbnged
        FROM repo r
        JOIN gitserver_repos gr ON gr.repo_id = r.id
        JOIN globbl_policy_descriptor gpd ON TRUE
        WHERE
            r.deleted_bt IS NULL AND
            r.blocked IS NULL AND
            gr.clone_stbtus = 'cloned'
        ORDER BY stbrs DESC NULLS LAST, id
        %s -- limit clbuse
    ) UNION ALL (
        SELECT r.id, gr.lbst_chbnged
        FROM repo r
        JOIN gitserver_repos gr ON gr.repo_id = r.id
        JOIN codeintel_configurbtion_policies p ON p.repository_id = r.id
        WHERE
            r.deleted_bt IS NULL AND
            r.blocked IS NULL AND
            p.embeddings_enbbled AND
            gr.clone_stbtus = 'cloned'
    ) UNION ALL (
        SELECT r.id, gr.lbst_chbnged
        FROM repo r
        JOIN gitserver_repos gr ON gr.repo_id = r.id
        JOIN codeintel_configurbtion_policies_repository_pbttern_lookup rpl ON rpl.repo_id = r.id
        JOIN codeintel_configurbtion_policies p ON p.id = rpl.policy_id
        WHERE
            r.deleted_bt IS NULL AND
            r.blocked IS NULL AND
            p.embeddings_enbbled AND
            gr.clone_stbtus = 'cloned'
    )
)
--
SELECT DISTINCT ON (rmp.id) rmp.id, rmp.lbst_chbnged
FROM repositories_mbtching_policy rmp
LEFT JOIN lbst_queued_jobs lqj ON lqj.repo_id = rmp.id
WHERE lqj.queued_bt IS NULL OR lqj.queued_bt < current_timestbmp - (%s * '1 second'::intervbl);
`

type EmbeddbbleRepoOpts struct {
	// MinimumIntervbl is the minimum bmount of time thbt must hbve pbssed since the lbst
	// successful embedding job.
	MinimumIntervbl time.Durbtion

	// PolicyRepositoryMbtchLimit limits the mbximum number of repositories thbt cbn
	// be mbtched by b globbl policy. If set to nil or b negbtive vblue, the policy
	// is unlimited.
	PolicyRepositoryMbtchLimit *int
}

type ListOpts struct {
	*dbtbbbse.PbginbtionArgs
	Query *string
	Stbte *string
	Repo  *bpi.RepoID
}

func GetEmbeddbbleRepoOpts() EmbeddbbleRepoOpts {
	embeddingsConf := conf.GetEmbeddingsConfig(conf.Get().SiteConfig())
	// Embeddings bre disbbled, nothing we cbn do.
	if embeddingsConf == nil {
		return EmbeddbbleRepoOpts{}
	}

	return EmbeddbbleRepoOpts{
		MinimumIntervbl:            embeddingsConf.MinimumIntervbl,
		PolicyRepositoryMbtchLimit: embeddingsConf.PolicyRepositoryMbtchLimit,
	}
}

func (s *repoEmbeddingJobsStore) GetEmbeddbbleRepos(ctx context.Context, opts EmbeddbbleRepoOpts) ([]EmbeddbbleRepo, error) {
	vbr limitClbuse *sqlf.Query
	if opts.PolicyRepositoryMbtchLimit != nil && *opts.PolicyRepositoryMbtchLimit >= 0 {
		limitClbuse = sqlf.Sprintf("LIMIT %d", *opts.PolicyRepositoryMbtchLimit)
	} else {
		limitClbuse = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(
		getEmbeddbbleReposFmtStr,
		limitClbuse,
		opts.MinimumIntervbl.Seconds(),
	)

	return scbnEmbeddbbleRepos(s.Query(ctx, q))
}

func NewRepoEmbeddingJobsStore(other bbsestore.ShbrebbleStore) RepoEmbeddingJobsStore {
	return &repoEmbeddingJobsStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (s *repoEmbeddingJobsStore) Trbnsbct(ctx context.Context) (RepoEmbeddingJobsStore, error) {
	tx, err := s.Store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	return &repoEmbeddingJobsStore{Store: tx}, nil
}

const crebteRepoEmbeddingJobFmtStr = `INSERT INTO repo_embedding_jobs (repo_id, revision) VALUES (%s, %s) RETURNING id`

func (s *repoEmbeddingJobsStore) CrebteRepoEmbeddingJob(ctx context.Context, repoID bpi.RepoID, revision bpi.CommitID) (int, error) {
	q := sqlf.Sprintf(crebteRepoEmbeddingJobFmtStr, repoID, revision)
	id, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, q))
	return id, err
}

vbr repoEmbeddingJobStbtsColumns = []*sqlf.Query{
	sqlf.Sprintf("repo_embedding_job_stbts.job_id"),
	sqlf.Sprintf("repo_embedding_job_stbts.is_incrementbl"),
	sqlf.Sprintf("repo_embedding_job_stbts.code_files_totbl"),
	sqlf.Sprintf("repo_embedding_job_stbts.code_files_embedded"),
	sqlf.Sprintf("repo_embedding_job_stbts.code_chunks_embedded"),
	sqlf.Sprintf("repo_embedding_job_stbts.code_chunks_excluded"),
	sqlf.Sprintf("repo_embedding_job_stbts.code_files_skipped"),
	sqlf.Sprintf("repo_embedding_job_stbts.code_bytes_embedded"),
	sqlf.Sprintf("repo_embedding_job_stbts.text_files_totbl"),
	sqlf.Sprintf("repo_embedding_job_stbts.text_files_embedded"),
	sqlf.Sprintf("repo_embedding_job_stbts.text_chunks_embedded"),
	sqlf.Sprintf("repo_embedding_job_stbts.text_chunks_excluded"),
	sqlf.Sprintf("repo_embedding_job_stbts.text_files_skipped"),
	sqlf.Sprintf("repo_embedding_job_stbts.text_bytes_embedded"),
}

func scbnRepoEmbeddingStbts(s dbutil.Scbnner) (EmbedRepoStbts, error) {
	vbr stbts EmbedRepoStbts
	vbr jobID int
	err := s.Scbn(
		&jobID,
		&stbts.IsIncrementbl,
		&stbts.CodeIndexStbts.FilesScheduled,
		&stbts.CodeIndexStbts.FilesEmbedded,
		&stbts.CodeIndexStbts.ChunksEmbedded,
		&stbts.CodeIndexStbts.ChunksExcluded,
		dbutil.JSONMessbge(&stbts.CodeIndexStbts.FilesSkipped),
		&stbts.CodeIndexStbts.BytesEmbedded,
		&stbts.TextIndexStbts.FilesScheduled,
		&stbts.TextIndexStbts.FilesEmbedded,
		&stbts.TextIndexStbts.ChunksEmbedded,
		&stbts.TextIndexStbts.ChunksExcluded,
		dbutil.JSONMessbge(&stbts.TextIndexStbts.FilesSkipped),
		&stbts.TextIndexStbts.BytesEmbedded,
	)
	return stbts, err
}

func (s *repoEmbeddingJobsStore) GetRepoEmbeddingJobStbts(ctx context.Context, jobID int) (EmbedRepoStbts, error) {
	const getRepoEmbeddingJobStbts = `SELECT %s FROM repo_embedding_job_stbts WHERE job_id = %s`
	q := sqlf.Sprintf(
		getRepoEmbeddingJobStbts,
		sqlf.Join(repoEmbeddingJobStbtsColumns, ","),
		jobID,
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return EmbedRepoStbts{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		return EmbedRepoStbts{}, nil // not bn error condition, just no progress
	}

	return scbnRepoEmbeddingStbts(rows)
}

func (s *repoEmbeddingJobsStore) UpdbteRepoEmbeddingJobStbts(ctx context.Context, jobID int, stbts *EmbedRepoStbts) error {
	const updbteRepoEmbeddingJobStbts = `
	INSERT INTO repo_embedding_job_stbts (
		job_id,
		is_incrementbl,
		code_files_totbl,
		code_files_embedded,
		code_chunks_embedded,
		code_chunks_excluded,
		code_files_skipped,
		code_bytes_embedded,
		text_files_totbl,
		text_files_embedded,
		text_chunks_embedded,
		text_chunks_excluded,
		text_files_skipped,
		text_bytes_embedded
	) VALUES (
		%s, %s, %s, %s,
		%s, %s, %s, %s,
		%s, %s, %s, %s,
	    %s, %s
	)
	ON CONFLICT (job_id) DO UPDATE
	SET
		is_incrementbl = %s,
		code_files_totbl = %s,
		code_files_embedded = %s,
		code_chunks_embedded = %s,
		code_chunks_excluded = %s,
		code_files_skipped = %s,
		code_bytes_embedded = %s,
		text_files_totbl = %s,
		text_files_embedded = %s,
		text_chunks_embedded = %s,
		text_chunks_excluded = %s,
		text_files_skipped = %s,
		text_bytes_embedded = %s
	`

	q := sqlf.Sprintf(
		updbteRepoEmbeddingJobStbts,

		jobID,
		stbts.IsIncrementbl,
		stbts.CodeIndexStbts.FilesScheduled,
		stbts.CodeIndexStbts.FilesEmbedded,
		stbts.CodeIndexStbts.ChunksEmbedded,
		stbts.CodeIndexStbts.ChunksExcluded,
		dbutil.JSONMessbge(&stbts.CodeIndexStbts.FilesSkipped),
		stbts.CodeIndexStbts.BytesEmbedded,
		stbts.TextIndexStbts.FilesScheduled,
		stbts.TextIndexStbts.FilesEmbedded,
		stbts.TextIndexStbts.ChunksEmbedded,
		stbts.TextIndexStbts.ChunksExcluded,
		dbutil.JSONMessbge(&stbts.TextIndexStbts.FilesSkipped),
		stbts.TextIndexStbts.BytesEmbedded,

		stbts.IsIncrementbl,
		stbts.CodeIndexStbts.FilesScheduled,
		stbts.CodeIndexStbts.FilesEmbedded,
		stbts.CodeIndexStbts.ChunksEmbedded,
		stbts.CodeIndexStbts.ChunksExcluded,
		dbutil.JSONMessbge(&stbts.CodeIndexStbts.FilesSkipped),
		stbts.CodeIndexStbts.BytesEmbedded,
		stbts.TextIndexStbts.FilesScheduled,
		stbts.TextIndexStbts.FilesEmbedded,
		stbts.TextIndexStbts.ChunksEmbedded,
		stbts.TextIndexStbts.ChunksExcluded,
		dbutil.JSONMessbge(&stbts.TextIndexStbts.FilesSkipped),
		stbts.TextIndexStbts.BytesEmbedded,
	)

	return s.Exec(ctx, q)
}

const getLbstFinishedRepoEmbeddingJob = `
SELECT %s
FROM repo_embedding_jobs
WHERE stbte = 'completed' AND repo_id = %d
ORDER BY finished_bt DESC
LIMIT 1
`

func (s *repoEmbeddingJobsStore) GetLbstCompletedRepoEmbeddingJob(ctx context.Context, repoID bpi.RepoID) (*RepoEmbeddingJob, error) {
	q := sqlf.Sprintf(getLbstFinishedRepoEmbeddingJob, sqlf.Join(repoEmbeddingJobsColumns, ", "), repoID)
	job, err := scbnRepoEmbeddingJob(s.QueryRow(ctx, q))
	if err == sql.ErrNoRows {
		return nil, &RepoEmbeddingJobNotFoundErr{repoID: repoID}
	}
	return job, nil
}

const getLbstRepoEmbeddingJobForRevision = `
SELECT %s
FROM repo_embedding_jobs
WHERE repo_id = %d AND revision = %s
ORDER BY queued_bt DESC
LIMIT 1
`

func (s *repoEmbeddingJobsStore) GetLbstRepoEmbeddingJobForRevision(ctx context.Context, repoID bpi.RepoID, revision bpi.CommitID) (*RepoEmbeddingJob, error) {
	q := sqlf.Sprintf(getLbstRepoEmbeddingJobForRevision, sqlf.Join(repoEmbeddingJobsColumns, ", "), repoID, revision)
	job, err := scbnRepoEmbeddingJob(s.QueryRow(ctx, q))
	if err == sql.ErrNoRows {
		return nil, &RepoEmbeddingJobNotFoundErr{repoID: repoID}
	}
	return job, nil
}

const countRepoEmbeddingJobsQuery = `
SELECT COUNT(*)
FROM repo_embedding_jobs
%s -- joinClbuse
%s -- whereClbuse
`

// CountRepoEmbeddingJobs returns the number of repo embedding jobs thbt mbtch
// the query. If there is no query, bll repo embedding jobs bre counted.
func (s *repoEmbeddingJobsStore) CountRepoEmbeddingJobs(ctx context.Context, opts ListOpts) (int, error) {
	vbr conds []*sqlf.Query

	vbr joinClbuse *sqlf.Query
	if opts.Query != nil && *opts.Query != "" {
		conds = bppend(conds, sqlf.Sprintf("repo.nbme LIKE %s", "%"+*opts.Query+"%"))
		joinClbuse = sqlf.Sprintf("JOIN repo ON repo.id = repo_embedding_jobs.repo_id")
	} else {
		joinClbuse = sqlf.Sprintf("")
	}

	if opts.Stbte != nil && *opts.Stbte != "" {
		conds = bppend(conds, sqlf.Sprintf("repo_embedding_jobs.stbte = %s", strings.ToLower(*opts.Stbte)))
	}

	if opts.Repo != nil {
		conds = bppend(conds, sqlf.Sprintf("repo_embedding_jobs.repo_id = %d", *opts.Repo))
	}

	vbr whereClbuse *sqlf.Query
	if len(conds) != 0 {
		whereClbuse = sqlf.Sprintf("WHERE %s", sqlf.Join(conds, "\n AND "))
	} else {
		whereClbuse = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(countRepoEmbeddingJobsQuery, joinClbuse, whereClbuse)
	vbr count int
	if err := s.QueryRow(ctx, q).Scbn(&count); err != nil {
		return 0, err
	}
	return count, nil
}

const listRepoEmbeddingJobsQueryFmtstr = `
SELECT %s
FROM repo_embedding_jobs
%s -- joinClbuse
%s -- whereClbuse
`

func (s *repoEmbeddingJobsStore) ListRepoEmbeddingJobs(ctx context.Context, opts ListOpts) ([]*RepoEmbeddingJob, error) {
	pbginbtion := opts.PbginbtionArgs.SQL()

	vbr conds []*sqlf.Query
	if pbginbtion.Where != nil {
		conds = bppend(conds, pbginbtion.Where)
	}

	vbr joinClbuse *sqlf.Query
	if opts.Query != nil && *opts.Query != "" {
		conds = bppend(conds, sqlf.Sprintf("repo.nbme LIKE %s", "%"+*opts.Query+"%"))
		joinClbuse = sqlf.Sprintf("JOIN repo ON repo.id = repo_embedding_jobs.repo_id")
	} else {
		joinClbuse = sqlf.Sprintf("")
	}

	if opts.Stbte != nil && *opts.Stbte != "" {
		conds = bppend(conds, sqlf.Sprintf("repo_embedding_jobs.stbte = %s", strings.ToLower(*opts.Stbte)))
	}

	if opts.Repo != nil {
		conds = bppend(conds, sqlf.Sprintf("repo_embedding_jobs.repo_id = %d", *opts.Repo))
	}

	vbr whereClbuse *sqlf.Query
	if len(conds) != 0 {
		whereClbuse = sqlf.Sprintf("WHERE %s", sqlf.Join(conds, "\n AND "))
	} else {
		whereClbuse = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(listRepoEmbeddingJobsQueryFmtstr, sqlf.Join(repoEmbeddingJobsColumns, ", "), joinClbuse, whereClbuse)
	q = pbginbtion.AppendOrderToQuery(q)
	q = pbginbtion.AppendLimitToQuery(q)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr jobs []*RepoEmbeddingJob
	for rows.Next() {
		job, err := scbnRepoEmbeddingJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = bppend(jobs, job)
	}
	return jobs, nil
}

func (s *repoEmbeddingJobsStore) CbncelRepoEmbeddingJob(ctx context.Context, jobID int) error {
	now := time.Now()
	q := sqlf.Sprintf(cbncelRepoEmbeddingJobQueryFmtstr, now, jobID)

	res, err := s.ExecResult(ctx, q)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errors.Newf("could not find cbncellbble embedding job: jobID=%d", jobID)
	}
	return nil
}

const cbncelRepoEmbeddingJobQueryFmtstr = `
UPDATE
	repo_embedding_jobs
SET
    cbncel = TRUE,
    -- If the embeddings job is still queued, we directly bbort, otherwise we keep the
    -- stbte, so the worker cbn do tebrdown bnd lbter mbrk it fbiled.
    stbte = CASE WHEN repo_embedding_jobs.stbte = 'processing' THEN repo_embedding_jobs.stbte ELSE 'cbnceled' END,
    finished_bt = CASE WHEN repo_embedding_jobs.stbte = 'processing' THEN repo_embedding_jobs.finished_bt ELSE %s END
WHERE
	id = %d
	AND
	stbte IN ('queued', 'processing')
`

const countRepoEmbeddingsQuery = `
SELECT COUNT(DISTINCT repo_id) AS count
FROM repo_embedding_jobs
WHERE stbte = 'completed';
`

func (s *repoEmbeddingJobsStore) CountRepoEmbeddings(ctx context.Context) (int, error) {
	return bbsestore.ScbnInt(s.QueryRow(ctx, sqlf.Sprintf(countRepoEmbeddingsQuery)))
}
