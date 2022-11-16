package iterator

import (
	"context"
	"time"

	"github.com/derision-test/glock"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type FinishFunc func(ctx context.Context, store *basestore.Store, maybeErr error) error

// PersistentRepoIterator represents a durable (persisted) iterator over a set of repositories. This iteration is not
// concurrency safe and only one consumer should have access to this resource at a time.
type PersistentRepoIterator struct {
	Id              int
	CreatedAt       time.Time
	StartedAt       time.Time
	CompletedAt     time.Time
	RuntimeDuration time.Duration
	PercentComplete float64
	TotalCount      int
	SuccessCount    int
	repos           []int32
	Cursor          int

	errors         errorMap
	terminalErrors errorMap
	retryRepos     []int32
	retryCursor    int

	itrStart time.Time // time the current iteration started
	itrEnd   time.Time // time the current iteration ended

	glock glock.Clock
}

type errorMap map[int32]*IterationError

func (em errorMap) FailureCount(repo int32) int {
	v, ok := em[repo]
	if !ok {
		return 0
	}
	return v.FailureCount
}

type IterationError struct {
	id            int
	RepoId        int32
	FailureCount  int
	ErrorMessages []string
}

var repoIteratorCols = []*sqlf.Query{
	sqlf.Sprintf("repo_iterator.Id"),
	sqlf.Sprintf("repo_iterator.created_at"),
	sqlf.Sprintf("repo_iterator.started_at"),
	sqlf.Sprintf("repo_iterator.completed_at"),
	sqlf.Sprintf("repo_iterator.runtime_duration"),
	sqlf.Sprintf("repo_iterator.percent_complete"),
	sqlf.Sprintf("repo_iterator.total_count"),
	sqlf.Sprintf("repo_iterator.success_count"),
	sqlf.Sprintf("repo_iterator.repos"),
	sqlf.Sprintf("repo_iterator.repo_cursor"),
}
var iteratorJoinCols = sqlf.Join(repoIteratorCols, ", ")

var repoIteratorErrorCols = []*sqlf.Query{
	sqlf.Sprintf("repo_iterator_errors.Id"),
	sqlf.Sprintf("repo_iterator_errors.repo_id"),
	sqlf.Sprintf("repo_iterator_errors.error_message"),
	sqlf.Sprintf("repo_iterator_errors.failure_count"),
}
var errorJoinCols = sqlf.Join(repoIteratorErrorCols, ", ")

// New returns a new (durable) repo iterator starting from cursor position 0.
func New(ctx context.Context, store *basestore.Store, repos []int32) (*PersistentRepoIterator, error) {
	return NewWithClock(ctx, store, glock.NewRealClock(), repos)
}

// NewWithClock returns a new (durable) repo iterator starting from cursor position 0 and optionally overrides the internal clock. Useful for tests.
func NewWithClock(ctx context.Context, store *basestore.Store, clock glock.Clock, repos []int32) (*PersistentRepoIterator, error) {
	if len(repos) == 0 {
		return nil, errors.New("unable to construct a repo iterator for an empty set")
	}

	q := "INSERT INTO repo_iterator(repos, total_count, created_at) VALUES (%S, %S, %S) RETURNING Id"
	id, err := basestore.ScanInt(store.QueryRow(ctx, sqlf.Sprintf(q, pq.Int32Array(repos), len(repos), clock.Now())))
	if err != nil {
		return nil, err
	}

	loaded, err := Load(ctx, store, id)
	if err != nil {
		return nil, err
	}
	loaded.glock = clock
	return loaded, nil
}

// Load will load a repo iterator that has been persisted and prepare it at the current cursor state.
func Load(ctx context.Context, store *basestore.Store, id int) (got *PersistentRepoIterator, err error) {
	return LoadWithClock(ctx, store, id, glock.NewRealClock())
}

func LoadWithClock(ctx context.Context, store *basestore.Store, id int, clock glock.Clock) (_ *PersistentRepoIterator, err error) {
	baseQuery := "SELECT %S FROM repo_iterator WHERE repo_iterator.Id = %S"
	row := store.QueryRow(ctx, sqlf.Sprintf(baseQuery, iteratorJoinCols, id))
	var repos pq.Int32Array
	var tmp PersistentRepoIterator
	if err = row.Scan(
		&tmp.Id,
		&tmp.CreatedAt,
		&dbutil.NullTime{Time: &tmp.StartedAt},
		&dbutil.NullTime{Time: &tmp.CompletedAt},
		&tmp.RuntimeDuration,
		&tmp.PercentComplete,
		&tmp.TotalCount,
		&tmp.SuccessCount,
		&repos,
		&tmp.Cursor,
	); err != nil {
		return nil, errors.Wrap(err, "ScanRepoIterator")
	}
	tmp.repos = repos
	if tmp.Cursor > len(tmp.repos) {
		return nil, errors.Newf("invalid repo iterator state Id:%d cursor:%d length:%d", tmp.Id, tmp.Cursor, len(repos))
	}

	tmp.errors, err = loadRepoIteratorErrors(ctx, store, &tmp)
	if err != nil {
		return nil, errors.Wrap(err, "loadRepoIteratorErrors")
	}
	tmp.terminalErrors = make(errorMap)

	tmp.glock = clock
	return &tmp, nil
}

type IterationConfig struct {
	MaxFailures int
	OnTerminal  OnTerminalFunc
}

type OnTerminalFunc func(ctx context.Context, store *basestore.Store, repoId int32, terminalErr error) error

// NextWithFinish will iterate the repository set from the current cursor position. If the iterator is marked complete
// or has no more repositories this will do nothing. The finish function returned is a mechanism to have atomic updates,
// callers will need to call the finish function when complete with work. Errors during work processing can be passed
// into the finish function and will be marked as errors on the repo iterator. Calling NextWithFinish without calling the
// finish function will infinitely loop on the current cursor. This iteration for a given repo iterator is not
// concurrency safe and should only be called from a single thread. Care should be taken to ensure in a distributed
// environment only one consumer is able to access this resource at a time.
func (p *PersistentRepoIterator) NextWithFinish(config IterationConfig) (api.RepoID, bool, FinishFunc) {
	current, got := peek(p.Cursor, p.repos)
	if !p.CompletedAt.IsZero() || !got {
		return 0, false, func(ctx context.Context, store *basestore.Store, err error) error {
			return nil
		}
	}
	p.itrStart = p.glock.Now()
	return api.RepoID(current), true, func(ctx context.Context, store *basestore.Store, maybeErr error) error {
		p.itrEnd = p.glock.Now()
		if err := p.doFinish(ctx, store, maybeErr, current, false, config); err != nil {
			return err
		}
		return nil
	}
}

func (p *PersistentRepoIterator) NextRetryWithFinish(config IterationConfig) (api.RepoID, bool, FinishFunc) {
	if len(p.retryRepos) == 0 {
		p.resetRetry(config)
	}
	var current int32
	var got bool
	for {
		current, got = peek(p.retryCursor, p.retryRepos)
		if !p.CompletedAt.IsZero() || !got {
			return 0, false, func(ctx context.Context, store *basestore.Store, err error) error {
				return nil
			}
		} else if config.MaxFailures > 0 && p.errors.FailureCount(current) >= config.MaxFailures {
			// this repo has exceeded its retry count, skip it
			p.advanceRetry()
			p.setRepoTerminal(current)
			continue
		}
		break
	}

	p.itrStart = p.glock.Now()
	return api.RepoID(current), true, func(ctx context.Context, store *basestore.Store, maybeErr error) error {
		p.itrEnd = p.glock.Now()
		p.advanceRetry()
		if err := p.doFinish(ctx, store, maybeErr, current, true, config); err != nil {
			return err
		}
		return nil
	}
}

// MarkComplete will mark the repo iterator as complete. Once marked complete the iterator is no longer eligible for iteration.
// This can be called at any time to mark the iterator as complete, and does not require the cursor have passed all the way through the set.
func (p *PersistentRepoIterator) MarkComplete(ctx context.Context, store *basestore.Store) error {
	now := p.glock.Now()
	err := store.Exec(ctx, sqlf.Sprintf("UPDATE repo_iterator SET percent_complete = 1, completed_at = %S, last_updated_at = %S where id = %s", now, now, p.Id))
	if err != nil {
		return err
	}
	p.CompletedAt = now
	p.PercentComplete = 1
	return nil
}

func (p *PersistentRepoIterator) HasMore() bool {
	_, has := peek(p.Cursor, p.repos)
	return has
}

func (p *PersistentRepoIterator) HasErrors() bool {
	return len(p.errors) > 0
}

func (p *PersistentRepoIterator) HasTerminalErrors() bool {
	return len(p.errors) > 0
}

func (p *PersistentRepoIterator) ErroredRepos() int {
	return len(p.errors)
}

func (p *PersistentRepoIterator) TotalErrors() int {
	count := 0
	for _, iterationError := range p.errors {
		count += iterationError.FailureCount
	}
	for _, iterationError := range p.terminalErrors {
		count += iterationError.FailureCount
	}
	return count
}

func (p *PersistentRepoIterator) Errors() []IterationError {
	itErrors := []IterationError{}
	for _, iterationError := range p.errors {
		itErrors = append(itErrors, *iterationError)
	}
	return itErrors
}

func stampStartedAt(ctx context.Context, store *basestore.Store, itrId int, stampTime time.Time) error {
	return store.Exec(ctx, sqlf.Sprintf("UPDATE repo_iterator SET started_at = %S WHERE Id = %S", stampTime, itrId))
}

func peek(offset int, repos []int32) (int32, bool) {
	if offset >= len(repos) {
		return 0, false
	}
	return repos[offset], true
}

func (p *PersistentRepoIterator) advanceRetry() {
	p.retryCursor += 1
}

func (p *PersistentRepoIterator) insertIterationError(ctx context.Context, store *basestore.Store, repoId int32, msg string) (err error) {
	var query *sqlf.Query
	if p.Id == 0 {
		return errors.New("invalid iterator to insert iterator error")
	}

	v, ok := p.errors[repoId]
	if !ok {
		query = sqlf.Sprintf("INSERT INTO repo_iterator_errors(repo_iterator_id, repo_id, error_message) VALUES (%S, %S, %S) RETURNING %S", p.Id, repoId, pq.Array([]string{msg}), errorJoinCols)
		row := store.QueryRow(ctx, query)
		var tmp IterationError
		if err = row.Scan(
			&tmp.id,
			&tmp.RepoId,
			pq.Array(&tmp.ErrorMessages),
			&tmp.FailureCount,
		); err != nil {
			return errors.Wrap(err, "InsertIterationError")
		}
		p.errors[tmp.RepoId] = &tmp
	} else {
		v.FailureCount += 1
		query = sqlf.Sprintf("UPDATE repo_iterator_errors SET failure_count = %S, error_message = array_append(error_message, %S) WHERE Id = %S", v.FailureCount, msg, v.id)
		if err = store.Exec(ctx, query); err != nil {
			return errors.Wrap(err, "UpdateIterationError")
		}
	}
	return nil
}

func (p *PersistentRepoIterator) doFinish(ctx context.Context, store *basestore.Store, maybeErr error, currentRepo int32, isRetry bool, config IterationConfig) (err error) {
	didSucceed := 0
	cursorOffset := 1
	totalFailures := p.errors.FailureCount(currentRepo)

	if maybeErr == nil {
		// if this was for a success then we will want to update
		didSucceed = 1
	} else {
		totalFailures += 1
	}
	if isRetry {
		cursorOffset = 0
	}
	itrDuration := p.itrEnd.Sub(p.itrStart)

	tx, err := store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if p.StartedAt.IsZero() {
		if err = stampStartedAt(ctx, tx, p.Id, p.itrStart); err != nil {
			return errors.Wrap(err, "stampStartedAt")
		}
		p.StartedAt = p.itrStart
	}
	err = p.updateRepoIterator(ctx, tx, didSucceed, cursorOffset, itrDuration)
	if maybeErr != nil {
		if err = p.insertIterationError(ctx, tx, currentRepo, maybeErr.Error()); err != nil {
			return errors.Wrapf(err, "unable to upsert error for repo iterator id: %d", p.Id)
		}
		if config.MaxFailures != 0 && totalFailures >= config.MaxFailures {
			// the condition is if there was an error, and we have configured both a max attempts, and the total attempts exceeds the config
			if config.OnTerminal != nil {
				err = config.OnTerminal(ctx, tx, currentRepo, maybeErr)
				if err != nil {
					return errors.Wrap(err, "iterator.OnTerminal")
				}
				p.setRepoTerminal(currentRepo)
			}
		}
	} else if isRetry {
		// delete the error for this repo
		err = tx.Exec(ctx, sqlf.Sprintf(`DELETE FROM repo_iterator_errors WHERE id = %s`, p.errors[currentRepo].id))
		if err != nil {
			return errors.Wrap(err, "deleteIteratorError")
		}
		delete(p.errors, currentRepo)
	}

	return nil
}

// setRepoTerminal sets a repository to a terminal error state
func (p *PersistentRepoIterator) setRepoTerminal(repoId int32) {
	p.terminalErrors[repoId] = p.errors[repoId]
	delete(p.errors, repoId)
}

func (p *PersistentRepoIterator) updateRepoIterator(ctx context.Context, store *basestore.Store, successCount, cursorOffset int, duration time.Duration) error {
	updateQ := `UPDATE repo_iterator
    SET percent_complete = COALESCE(((%s + success_count)::float / NULLIF(total_count, 0)::float), 0),
    success_count    = success_count + %s,
    repo_cursor      = repo_cursor + %s,
    last_updated_at  = NOW(),
    runtime_duration = runtime_duration + %s
    WHERE id = %s RETURNING percent_complete, success_count, repo_cursor, runtime_duration;`

	q := sqlf.Sprintf(updateQ, successCount, successCount, cursorOffset, duration, p.Id)

	var pct float64
	var successCnt int
	var cursor int
	var runtime time.Duration

	row := store.QueryRow(ctx, q)
	if err := row.Scan(
		&pct,
		&successCnt,
		&cursor,
		&runtime,
	); err != nil {
		return errors.Wrapf(err, "unable to update repo iterator id: %d", p.Id)
	}

	p.Cursor = cursor
	p.SuccessCount = successCnt
	p.PercentComplete = pct
	p.RuntimeDuration = runtime
	p.itrStart = time.Time{}
	p.itrEnd = time.Time{}

	return nil
}

func loadRepoIteratorErrors(ctx context.Context, store *basestore.Store, iterator *PersistentRepoIterator) (got errorMap, err error) {
	baseQuery := "SELECT %S FROM repo_iterator_errors WHERE repo_iterator_id = %S"
	rows, err := store.Query(ctx, sqlf.Sprintf(baseQuery, errorJoinCols, iterator.Id))
	if err != nil {
		return nil, err
	}
	got = make(errorMap)
	for rows.Next() {
		var tmp IterationError
		if err := rows.Scan(
			&tmp.id,
			&tmp.RepoId,
			pq.Array(&tmp.ErrorMessages),
			&tmp.FailureCount,
		); err != nil {
			return nil, err
		}
		got[tmp.RepoId] = &tmp
	}

	return got, err
}

func (p *PersistentRepoIterator) resetRetry(config IterationConfig) {
	p.retryCursor = 0
	p.terminalErrors = make(errorMap)
	var retry []int32
	for repo, val := range p.errors {
		if config.MaxFailures > 0 && val.FailureCount >= config.MaxFailures {
			p.terminalErrors[repo] = val
			continue
		}
		retry = append(retry, repo)
	}
	p.retryRepos = retry
}
