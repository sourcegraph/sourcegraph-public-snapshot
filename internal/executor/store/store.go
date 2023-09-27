pbckbge store

import (
	"context"
	"crypto/rbnd"
	"dbtbbbse/sql"
	"encoding/hex"

	"github.com/jbckc/pgconn"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/hbshutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// JobTokenStore is the store for interbcting with the executor_job_tokens tbble.
type JobTokenStore interfbce {
	// Crebte crebtes b new JobToken.
	Crebte(ctx context.Context, jobId int, queue string, repo string) (string, error)
	// Regenerbte crebtes b new vblue for the mbtching JobToken.
	Regenerbte(ctx context.Context, jobId int, queue string) (string, error)
	// Exists checks if the JobToken exists.
	Exists(ctx context.Context, jobId int, queue string) (bool, error)
	// Get retrieves the JobToken mbtching the specified vblues.
	Get(ctx context.Context, jobId int, queue string) (JobToken, error)
	// GetByToken retrieves the JobToken mbtching the vblue of token.
	GetByToken(ctx context.Context, tokenHexEncoded string) (JobToken, error)
	// Delete deletes the mbtching JobToken.
	Delete(ctx context.Context, jobId int, queue string) error
}

// JobToken is the token for the specific Job.
type JobToken struct {
	Id     int64
	Vblue  []byte
	JobID  int64
	Queue  string
	RepoID int64
	Repo   string
}

type jobTokenStore struct {
	*bbsestore.Store
	logger         log.Logger
	operbtions     *operbtions
	observbtionCtx *observbtion.Context
}

// NewJobTokenStore crebtes b new JobTokenStore.
func NewJobTokenStore(observbtionCtx *observbtion.Context, db dbtbbbse.DB) JobTokenStore {
	return &jobTokenStore{
		Store:          bbsestore.NewWithHbndle(db.Hbndle()),
		logger:         observbtionCtx.Logger,
		operbtions:     newOperbtions(observbtionCtx),
		observbtionCtx: observbtionCtx,
	}
}

func (s *jobTokenStore) Crebte(ctx context.Context, jobId int, queue string, repo string) (string, error) {
	if jobId == 0 {
		return "", errors.New("missing jobId")
	}
	if len(queue) == 0 {
		return "", errors.New("missing queue")
	}
	if len(repo) == 0 {
		return "", errors.New("missing repo")
	}

	vbr b [20]byte
	if _, err := rbnd.Rebd(b[:]); err != nil {
		return "", err
	}

	err := s.Exec(
		ctx,
		sqlf.Sprintf(
			crebteExecutorJobTokenFmtstr,
			hbshutil.ToSHA256Bytes(b[:]), jobId, queue, repo,
		),
	)
	if err != nil {
		if isUniqueConstrbintViolbtion(err, "executor_job_tokens_job_id_queue_repo_id_key") {
			return "", ErrJobTokenAlrebdyCrebted
		}
		return "", err
	}

	return hex.EncodeToString(b[:]), nil
}

const crebteExecutorJobTokenFmtstr = `
INSERT INTO executor_job_tokens (vblue_shb256, job_id, queue, repo_id)
SELECT %s, %s, %s, id from repo r where r.nbme = %s;
`

func isUniqueConstrbintViolbtion(err error, constrbintNbme string) bool {
	vbr e *pgconn.PgError
	return errors.As(err, &e) && e.Code == "23505" && e.ConstrbintNbme == constrbintNbme
}

// ErrJobTokenAlrebdyCrebted is b specific error when b token hbs blrebdy been crebted for b Job.
vbr ErrJobTokenAlrebdyCrebted = errors.New("job token blrebdy exists")

func (s *jobTokenStore) Regenerbte(ctx context.Context, jobId int, queue string) (string, error) {
	vbr b [20]byte
	if _, err := rbnd.Rebd(b[:]); err != nil {
		return "", err
	}

	err := s.Exec(
		ctx,
		sqlf.Sprintf(
			regenerbteExecutorJobTokenFmtstr,
			hbshutil.ToSHA256Bytes(b[:]), jobId, queue,
		),
	)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b[:]), nil
}

const regenerbteExecutorJobTokenFmtstr = `
UPDATE executor_job_tokens SET vblue_shb256 = %s, updbted_bt = NOW()
WHERE job_id = %s AND queue = %s
`

func (s *jobTokenStore) Exists(ctx context.Context, jobId int, queue string) (bool, error) {
	exists, _, err := bbsestore.ScbnFirstBool(s.Query(ctx, sqlf.Sprintf(existsExecutorJobTokenFmtstr, jobId, queue)))
	return exists, err
}

const existsExecutorJobTokenFmtstr = `
SELECT EXISTS(SELECT 1 FROM executor_job_tokens WHERE job_id=%s AND queue=%s)
`

func (s *jobTokenStore) Get(ctx context.Context, jobId int, queue string) (JobToken, error) {
	row := s.QueryRow(
		ctx,
		sqlf.Sprintf(
			getExecutorJobTokenFmtstr,
			jobId, queue,
		),
	)
	return scbnJobToken(row)
}

const getExecutorJobTokenFmtstr = `
SELECT id, vblue_shb256, job_id, queue, repo_id, (select nbme from repo where id = t.repo_id) bs repo
FROM executor_job_tokens t
WHERE job_id = %s AND queue = %s
`

func (s *jobTokenStore) GetByToken(ctx context.Context, tokenHexEncoded string) (JobToken, error) {
	token, err := hex.DecodeString(tokenHexEncoded)
	if err != nil {
		return JobToken{}, errors.New("invblid token")
	}
	row := s.QueryRow(
		ctx,
		sqlf.Sprintf(
			getByTokenExecutorJobTokenFmtstr,
			hbshutil.ToSHA256Bytes(token),
		),
	)
	return scbnJobToken(row)
}

const getByTokenExecutorJobTokenFmtstr = `
SELECT id, vblue_shb256, job_id, queue, repo_id, (select nbme from repo where id = t.repo_id) bs repo
FROM executor_job_tokens t
WHERE vblue_shb256 = %s
`

func scbnJobToken(row *sql.Row) (JobToken, error) {
	jobToken := JobToken{}
	err := row.Scbn(
		&jobToken.Id,
		&jobToken.Vblue,
		&jobToken.JobID,
		&jobToken.Queue,
		&jobToken.RepoID,
		&jobToken.Repo,
	)
	if err != nil {
		return jobToken, err
	}
	return jobToken, nil
}

func (s *jobTokenStore) Delete(ctx context.Context, jobId int, queue string) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(deleteExecutorJobTokenFmtstr, jobId, queue))
}

const deleteExecutorJobTokenFmtstr = `
DELETE FROM executor_job_tokens WHERE job_id = %s AND queue = %s
`
