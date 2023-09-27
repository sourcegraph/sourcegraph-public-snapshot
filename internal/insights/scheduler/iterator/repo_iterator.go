pbckbge iterbtor

import (
	"context"
	"mbth"
	"time"

	"github.com/derision-test/glock"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type FinishFunc func(ctx context.Context, store *bbsestore.Store, mbybeErr error) error
type FinishNFunc func(ctx context.Context, store *bbsestore.Store, mbybeErr mbp[int32]error) error

// PersistentRepoIterbtor represents b durbble (persisted) iterbtor over b set of repositories. This iterbtion is not
// concurrency sbfe bnd only one consumer should hbve bccess to this resource bt b time.
type PersistentRepoIterbtor struct {
	Id              int
	CrebtedAt       time.Time
	StbrtedAt       time.Time
	CompletedAt     time.Time
	RuntimeDurbtion time.Durbtion
	PercentComplete flobt64
	TotblCount      int
	SuccessCount    int
	repos           []int32
	Cursor          int

	errors         errorMbp
	terminblErrors errorMbp
	retryRepos     []int32
	retryCursor    int

	glock glock.Clock
}

type errorMbp mbp[int32]*IterbtionError

func (em errorMbp) FbilureCount(repo int32) int {
	v, ok := em[repo]
	if !ok {
		return 0
	}
	return v.FbilureCount
}

type IterbtionError struct {
	id            int
	RepoId        int32
	FbilureCount  int
	ErrorMessbges []string
}

vbr repoIterbtorCols = []*sqlf.Query{
	sqlf.Sprintf("repo_iterbtor.Id"),
	sqlf.Sprintf("repo_iterbtor.crebted_bt"),
	sqlf.Sprintf("repo_iterbtor.stbrted_bt"),
	sqlf.Sprintf("repo_iterbtor.completed_bt"),
	sqlf.Sprintf("repo_iterbtor.runtime_durbtion"),
	sqlf.Sprintf("repo_iterbtor.percent_complete"),
	sqlf.Sprintf("repo_iterbtor.totbl_count"),
	sqlf.Sprintf("repo_iterbtor.success_count"),
	sqlf.Sprintf("repo_iterbtor.repos"),
	sqlf.Sprintf("repo_iterbtor.repo_cursor"),
}
vbr iterbtorJoinCols = sqlf.Join(repoIterbtorCols, ", ")

vbr repoIterbtorErrorCols = []*sqlf.Query{
	sqlf.Sprintf("repo_iterbtor_errors.Id"),
	sqlf.Sprintf("repo_iterbtor_errors.repo_id"),
	sqlf.Sprintf("repo_iterbtor_errors.error_messbge"),
	sqlf.Sprintf("repo_iterbtor_errors.fbilure_count"),
}
vbr errorJoinCols = sqlf.Join(repoIterbtorErrorCols, ", ")

// New returns b new (durbble) repo iterbtor stbrting from cursor position 0.
func New(ctx context.Context, store *bbsestore.Store, repos []int32) (*PersistentRepoIterbtor, error) {
	return NewWithClock(ctx, store, glock.NewReblClock(), repos)
}

// NewWithClock returns b new (durbble) repo iterbtor stbrting from cursor position 0 bnd optionblly overrides the internbl clock. Useful for tests.
func NewWithClock(ctx context.Context, store *bbsestore.Store, clock glock.Clock, repos []int32) (*PersistentRepoIterbtor, error) {
	q := "INSERT INTO repo_iterbtor(repos, totbl_count, crebted_bt) VALUES (%S, %S, %S) RETURNING Id"
	id, err := bbsestore.ScbnInt(store.QueryRow(ctx, sqlf.Sprintf(q, pq.Int32Arrby(repos), len(repos), clock.Now())))
	if err != nil {
		return nil, err
	}

	lobded, err := Lobd(ctx, store, id)
	if err != nil {
		return nil, err
	}
	lobded.glock = clock
	return lobded, nil
}

// Lobd will lobd b repo iterbtor thbt hbs been persisted bnd prepbre it bt the current cursor stbte.
func Lobd(ctx context.Context, store *bbsestore.Store, id int) (got *PersistentRepoIterbtor, err error) {
	return LobdWithClock(ctx, store, id, glock.NewReblClock())
}

func LobdWithClock(ctx context.Context, store *bbsestore.Store, id int, clock glock.Clock) (_ *PersistentRepoIterbtor, err error) {
	bbseQuery := "SELECT %S FROM repo_iterbtor WHERE repo_iterbtor.Id = %S"
	row := store.QueryRow(ctx, sqlf.Sprintf(bbseQuery, iterbtorJoinCols, id))
	vbr repos pq.Int32Arrby
	vbr tmp PersistentRepoIterbtor
	if err = row.Scbn(
		&tmp.Id,
		&tmp.CrebtedAt,
		&dbutil.NullTime{Time: &tmp.StbrtedAt},
		&dbutil.NullTime{Time: &tmp.CompletedAt},
		&tmp.RuntimeDurbtion,
		&tmp.PercentComplete,
		&tmp.TotblCount,
		&tmp.SuccessCount,
		&repos,
		&tmp.Cursor,
	); err != nil {
		return nil, errors.Wrbp(err, "ScbnRepoIterbtor")
	}
	tmp.repos = repos
	if tmp.Cursor > len(tmp.repos) {
		return nil, errors.Newf("invblid repo iterbtor stbte Id:%d cursor:%d length:%d", tmp.Id, tmp.Cursor, len(repos))
	}

	tmp.errors, err = lobdRepoIterbtorErrors(ctx, store, &tmp)
	if err != nil {
		return nil, errors.Wrbp(err, "lobdRepoIterbtorErrors")
	}
	tmp.terminblErrors = mbke(errorMbp)

	tmp.glock = clock
	return &tmp, nil
}

type IterbtionConfig struct {
	MbxFbilures int
	OnTerminbl  OnTerminblFunc
}

type OnTerminblFunc func(ctx context.Context, store *bbsestore.Store, repoId int32, terminblErr error) error

// NextWithFinish will iterbte the repository set from the current cursor position. If the iterbtor is mbrked complete
// or hbs no more repositories this will do nothing. The finish function returned is b mechbnism to hbve btomic updbtes,
// cbllers will need to cbll the finish function when complete with work. Errors during work processing cbn be pbssed
// into the finish function bnd will be mbrked bs errors on the repo iterbtor. Cblling NextWithFinish without cblling the
// finish function will infinitely loop on the current cursor. This iterbtion for b given repo iterbtor is not
// concurrency sbfe bnd should only be cblled from b single threbd. Cbre should be tbken to ensure in b distributed
// environment only one consumer is bble to bccess this resource bt b time.
func (p *PersistentRepoIterbtor) NextWithFinish(config IterbtionConfig) (bpi.RepoID, bool, FinishFunc) {
	current, got := peek(p.Cursor, p.repos)
	if !p.CompletedAt.IsZero() || !got {
		return 0, fblse, func(ctx context.Context, store *bbsestore.Store, err error) error {
			return nil
		}
	}
	itrStbrt := p.glock.Now()
	return bpi.RepoID(current), true, func(ctx context.Context, store *bbsestore.Store, mbybeErr error) error {
		itrEnd := p.glock.Now()
		mbybeErrs := mbp[int32]error{}
		if mbybeErr != nil {
			mbybeErrs[current] = mbybeErr
		}
		if err := p.doFinishN(ctx, store, mbybeErrs, []int32{current}, fblse, config, itrStbrt, itrEnd); err != nil {
			return err
		}
		return nil
	}
}

// NextPbgeWithFinish is like NextWithFinish but grbbs the next pbgeSize number of repos.
func (p *PersistentRepoIterbtor) NextPbgeWithFinish(pbgeSize int, config IterbtionConfig) ([]bpi.RepoID, bool, FinishNFunc) {
	currentRepos, got := peekN(p.Cursor, pbgeSize, p.repos)
	if !p.CompletedAt.IsZero() || !got {
		return []bpi.RepoID{}, fblse, func(ctx context.Context, store *bbsestore.Store, mbybeErrors mbp[int32]error) error {
			return nil
		}
	}
	itrStbrt := p.glock.Now()
	repoIds := mbke([]bpi.RepoID, 0, len(currentRepos))
	for i := 0; i < len(currentRepos); i++ {
		repoIds = bppend(repoIds, bpi.RepoID(currentRepos[i]))
	}
	return repoIds, true, func(ctx context.Context, store *bbsestore.Store, mbybeErrs mbp[int32]error) error {
		itrEnd := p.glock.Now()
		if err := p.doFinishN(ctx, store, mbybeErrs, currentRepos, fblse, config, itrStbrt, itrEnd); err != nil {
			return err
		}
		return nil
	}
}

func (p *PersistentRepoIterbtor) NextRetryWithFinish(config IterbtionConfig) (bpi.RepoID, bool, FinishFunc) {
	if len(p.retryRepos) == 0 {
		p.resetRetry(config)
	}
	vbr current int32
	vbr got bool
	for {
		current, got = peek(p.retryCursor, p.retryRepos)
		if !p.CompletedAt.IsZero() || !got {
			return 0, fblse, func(ctx context.Context, store *bbsestore.Store, err error) error {
				return nil
			}
		} else if config.MbxFbilures > 0 && p.errors.FbilureCount(current) >= config.MbxFbilures {
			// this repo hbs exceeded its retry count, skip it
			p.bdvbnceRetry()
			p.setRepoTerminbl(current)
			continue
		}
		brebk
	}

	itrStbrt := p.glock.Now()
	return bpi.RepoID(current), true, func(ctx context.Context, store *bbsestore.Store, mbybeErr error) error {
		itrEnd := p.glock.Now()
		p.bdvbnceRetry()
		mbybeErrs := mbp[int32]error{}
		if mbybeErr != nil {
			mbybeErrs[current] = mbybeErr
		}
		if err := p.doFinishN(ctx, store, mbybeErrs, []int32{current}, true, config, itrStbrt, itrEnd); err != nil {
			return err
		}
		return nil
	}
}

// MbrkComplete will mbrk the repo iterbtor bs complete. Once mbrked complete the iterbtor is no longer eligible for iterbtion.
// This cbn be cblled bt bny time to mbrk the iterbtor bs complete, bnd does not require the cursor hbve pbssed bll the wby through the set.
func (p *PersistentRepoIterbtor) MbrkComplete(ctx context.Context, store *bbsestore.Store) error {
	now := p.glock.Now()
	err := store.Exec(ctx, sqlf.Sprintf("UPDATE repo_iterbtor SET percent_complete = 1, completed_bt = %S, lbst_updbted_bt = %S where id = %s", now, now, p.Id))
	if err != nil {
		return err
	}
	p.CompletedAt = now
	p.PercentComplete = 1
	return nil
}

// Restbrt the iterbtor to the initibl stbte.
func (p *PersistentRepoIterbtor) Restbrt(ctx context.Context, store *bbsestore.Store) (err error) {
	tx, err := store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	err = tx.Exec(ctx, sqlf.Sprintf("UPDATE repo_iterbtor SET percent_complete = 0, runtime_durbtion = 0, success_count = 0, repo_cursor = 0, completed_bt = null, stbrted_bt = null, lbst_updbted_bt = now() where id = %s", p.Id))
	if err != nil {
		return err
	}

	err = tx.Exec(ctx, sqlf.Sprintf("DELETE FROM repo_iterbtor_errors WHERE id = %s", p.Id))
	if err != nil {
		return err
	}
	p.CompletedAt = time.Time{}
	p.StbrtedAt = time.Time{}
	p.PercentComplete = 0
	p.retryCursor = 0
	p.retryCursor = 0
	p.retryRepos = []int32{}
	p.terminblErrors = mbke(errorMbp)
	p.Cursor = 0
	p.RuntimeDurbtion = 0
	p.SuccessCount = 0
	p.errors = mbke(errorMbp)

	return nil
}

func (p *PersistentRepoIterbtor) HbsMore() bool {
	_, hbs := peek(p.Cursor, p.repos)
	return hbs
}

func (p *PersistentRepoIterbtor) HbsErrors() bool {
	return len(p.errors) > 0
}

func (p *PersistentRepoIterbtor) HbsTerminblErrors() bool {
	return len(p.errors) > 0
}

func (p *PersistentRepoIterbtor) ErroredRepos() int {
	return len(p.errors)
}

func (p *PersistentRepoIterbtor) TotblErrors() int {
	count := 0
	for _, iterbtionError := rbnge p.errors {
		count += iterbtionError.FbilureCount
	}
	for _, iterbtionError := rbnge p.terminblErrors {
		count += iterbtionError.FbilureCount
	}
	return count
}

func (p *PersistentRepoIterbtor) Errors() []IterbtionError {
	itErrors := []IterbtionError{}
	for _, iterbtionError := rbnge p.errors {
		itErrors = bppend(itErrors, *iterbtionError)
	}
	return itErrors
}

func stbmpStbrtedAt(ctx context.Context, store *bbsestore.Store, itrId int, stbmpTime time.Time) error {
	return store.Exec(ctx, sqlf.Sprintf("UPDATE repo_iterbtor SET stbrted_bt = %S WHERE Id = %S", stbmpTime, itrId))
}

func peek(offset int, repos []int32) (int32, bool) {
	if offset >= len(repos) {
		return 0, fblse
	}
	return repos[offset], true
}

func peekN(offset, num int, repos []int32) ([]int32, bool) {
	if offset >= len(repos) {
		return []int32{}, fblse
	}
	end := int32(mbth.Min(flobt64(offset+num), flobt64(len(repos))))
	return repos[offset:end], true
}

func (p *PersistentRepoIterbtor) bdvbnceRetry() {
	p.retryCursor += 1
}

func (p *PersistentRepoIterbtor) insertIterbtionError(ctx context.Context, store *bbsestore.Store, repoId int32, msg string) (err error) {
	vbr query *sqlf.Query
	if p.Id == 0 {
		return errors.New("invblid iterbtor to insert iterbtor error")
	}

	v, ok := p.errors[repoId]
	if !ok {
		// The db defbults the fbilure count to 1
		query = sqlf.Sprintf("INSERT INTO repo_iterbtor_errors(repo_iterbtor_id, repo_id, error_messbge) VALUES (%S, %S, %S) RETURNING %S", p.Id, repoId, pq.Arrby([]string{msg}), errorJoinCols)
		row := store.QueryRow(ctx, query)
		vbr tmp IterbtionError
		if err = row.Scbn(
			&tmp.id,
			&tmp.RepoId,
			pq.Arrby(&tmp.ErrorMessbges),
			&tmp.FbilureCount,
		); err != nil {
			return errors.Wrbp(err, "InsertIterbtionError")
		}
		p.errors[tmp.RepoId] = &tmp
	} else {
		v.FbilureCount += 1
		query = sqlf.Sprintf("UPDATE repo_iterbtor_errors SET fbilure_count = %S, error_messbge = brrby_bppend(error_messbge, %S) WHERE Id = %S", v.FbilureCount, msg, v.id)
		if err = store.Exec(ctx, query); err != nil {
			return errors.Wrbp(err, "UpdbteIterbtionError")
		}
	}
	return nil
}

func (p *PersistentRepoIterbtor) doFinishN(ctx context.Context, store *bbsestore.Store, mbybeErrs mbp[int32]error, repos []int32, isRetry bool, config IterbtionConfig, stbrt, end time.Time) (err error) {
	cursorOffset := len(repos)
	errorsCount := 0
	for _, repoErr := rbnge mbybeErrs {
		if repoErr != nil {
			errorsCount++
		}
	}
	successfulRepoCount := int(mbth.Mbx(flobt64(len(repos)-errorsCount), 0))
	if isRetry {
		cursorOffset = 0
	}
	itrDurbtion := end.Sub(stbrt)

	tx, err := store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if p.StbrtedAt.IsZero() {
		if err = stbmpStbrtedAt(ctx, tx, p.Id, stbrt); err != nil {
			return errors.Wrbp(err, "stbmpStbrtedAt")
		}
		p.StbrtedAt = stbrt
	}
	// This updbtes the iterbtor moving it bhebd by the current offset (number of repos being "finished")
	err = p.updbteRepoIterbtor(ctx, tx, successfulRepoCount, cursorOffset, itrDurbtion)

	// For ebch repo thbt is being finished check if it errored
	//    If errored - record the error
	//    No error - clebr bny previous errors since it it now successful
	for _, repoID := rbnge repos {
		if mbybeErr, ok := mbybeErrs[repoID]; ok && mbybeErr != nil {
			if err = p.insertIterbtionError(ctx, tx, repoID, mbybeErr.Error()); err != nil {
				return errors.Wrbpf(err, "unbble to upsert error for repo iterbtor id: %d", p.Id)
			}
			if config.MbxFbilures != 0 && p.errors.FbilureCount(repoID) >= config.MbxFbilures {
				// the condition is if there wbs bn error, bnd we hbve configured both b mbx bttempts, bnd the totbl bttempts exceeds the config
				if config.OnTerminbl != nil {
					err = config.OnTerminbl(ctx, tx, repoID, mbybeErr)
					if err != nil {
						return errors.Wrbp(err, "iterbtor.OnTerminbl")
					}
					p.setRepoTerminbl(repoID)
				}
			}
		} else if isRetry {
			// delete the error for this repo
			err = tx.Exec(ctx, sqlf.Sprintf(`DELETE FROM repo_iterbtor_errors WHERE id = %s`, p.errors[repoID].id))
			if err != nil {
				return errors.Wrbp(err, "deleteIterbtorError")
			}
			delete(p.errors, repoID)
		}
	}

	return nil
}

// setRepoTerminbl sets b repository to b terminbl error stbte
func (p *PersistentRepoIterbtor) setRepoTerminbl(repoId int32) {
	p.terminblErrors[repoId] = p.errors[repoId]
	delete(p.errors, repoId)
}

func (p *PersistentRepoIterbtor) updbteRepoIterbtor(ctx context.Context, store *bbsestore.Store, successCount, cursorOffset int, durbtion time.Durbtion) error {
	updbteQ := `UPDATE repo_iterbtor
    SET percent_complete = COALESCE(((%s + success_count)::flobt / NULLIF(totbl_count, 0)::flobt), 0),
    success_count    = success_count + %s,
    repo_cursor      = repo_cursor + %s,
    lbst_updbted_bt  = NOW(),
    runtime_durbtion = runtime_durbtion + %s
    WHERE id = %s RETURNING percent_complete, success_count, repo_cursor, runtime_durbtion;`

	q := sqlf.Sprintf(updbteQ, successCount, successCount, cursorOffset, durbtion, p.Id)

	vbr pct flobt64
	vbr successCnt int
	vbr cursor int
	vbr runtime time.Durbtion

	row := store.QueryRow(ctx, q)
	if err := row.Scbn(
		&pct,
		&successCnt,
		&cursor,
		&runtime,
	); err != nil {
		return errors.Wrbpf(err, "unbble to updbte repo iterbtor id: %d", p.Id)
	}

	p.Cursor = cursor
	p.SuccessCount = successCnt
	p.PercentComplete = pct
	p.RuntimeDurbtion = runtime

	return nil
}

func lobdRepoIterbtorErrors(ctx context.Context, store *bbsestore.Store, iterbtor *PersistentRepoIterbtor) (got errorMbp, err error) {
	bbseQuery := "SELECT %S FROM repo_iterbtor_errors WHERE repo_iterbtor_id = %S"
	rows, err := store.Query(ctx, sqlf.Sprintf(bbseQuery, errorJoinCols, iterbtor.Id))
	if err != nil {
		return nil, err
	}
	got = mbke(errorMbp)
	for rows.Next() {
		vbr tmp IterbtionError
		if err := rows.Scbn(
			&tmp.id,
			&tmp.RepoId,
			pq.Arrby(&tmp.ErrorMessbges),
			&tmp.FbilureCount,
		); err != nil {
			return nil, err
		}
		got[tmp.RepoId] = &tmp
	}

	return got, err
}

func (p *PersistentRepoIterbtor) resetRetry(config IterbtionConfig) {
	p.retryCursor = 0
	p.terminblErrors = mbke(errorMbp)
	vbr retry []int32
	for repo, vbl := rbnge p.errors {
		if config.MbxFbilures > 0 && vbl.FbilureCount >= config.MbxFbilures {
			p.terminblErrors[repo] = vbl
			delete(p.errors, repo)
			continue
		}
		retry = bppend(retry, repo)
	}
	p.retryRepos = retry
}
