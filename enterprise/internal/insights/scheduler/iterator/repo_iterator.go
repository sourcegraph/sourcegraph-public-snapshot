package iterator

import (
	"context"
	"time"

	"github.com/derision-test/glock"

	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"

	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	"github.com/keegancsmith/sqlf"
)

type RepoIterator interface {
	NextWithFinish() (api.RepoID, bool, finishFunc)
}

var _ RepoIterator = &persistentRepoIterator{}

type finishFunc func(ctx context.Context, store *basestore.Store, maybeErr error) error

type persistentRepoIterator struct {
	id              int
	CreatedAt       time.Time
	StartedAt       time.Time
	CompletedAt     time.Time
	LastUpdatedAt   time.Time
	RuntimeDuration time.Duration
	PercentComplete float64
	TotalCount      int
	SuccessCount    int
	repos           []int32
	Cursor          int
	errors          errorMap

	itrStart time.Time // time the current iteration started
	itrEnd   time.Time // time the current iteration ended

	glock glock.Clock
}

type errorMap map[int32]*IterationError

type IterationError struct {
	id            int
	RepoId        int32
	FailureCount  int
	ErrorMessages []string
}

var repoIteratorCols = []*sqlf.Query{
	sqlf.Sprintf("repo_iterator.id"),
	sqlf.Sprintf("repo_iterator.created_at"),
	sqlf.Sprintf("repo_iterator.started_at"),
	sqlf.Sprintf("repo_iterator.completed_at"),
	sqlf.Sprintf("repo_iterator.last_updated_at"),
	sqlf.Sprintf("repo_iterator.runtime_duration"),
	sqlf.Sprintf("repo_iterator.percent_complete"),
	sqlf.Sprintf("repo_iterator.total_count"),
	sqlf.Sprintf("repo_iterator.success_count"),
	sqlf.Sprintf("repo_iterator.repos"),
	sqlf.Sprintf("repo_iterator.repo_cursor"),
}
var iteratorJoinCols = sqlf.Join(repoIteratorCols, ", ")

var repoIteratorErrorCols = []*sqlf.Query{
	sqlf.Sprintf("repo_iterator_errors.id"),
	sqlf.Sprintf("repo_iterator_errors.repo_id"),
	sqlf.Sprintf("repo_iterator_errors.error_message"),
	sqlf.Sprintf("repo_iterator_errors.failure_count"),
}
var errorJoinCols = sqlf.Join(repoIteratorErrorCols, ", ")

// New returns a new (durable) repo iterator starting from cursor position 0.
func New(ctx context.Context, store *basestore.Store, repos []int32) (*persistentRepoIterator, error) {
	return NewWithClock(ctx, store, glock.NewRealClock(), repos)
}

// NewWithClock returns a new (durable) repo iterator starting from cursor position 0 and optionally overrides the internal clock. Useful for tests.
func NewWithClock(ctx context.Context, store *basestore.Store, clock glock.Clock, repos []int32) (*persistentRepoIterator, error) {
	if len(repos) == 0 {
		return nil, errors.New("unable to construct a repo iterator for an empty set")
	}

	q := "insert into repo_iterator(repos, total_count) VALUES (%s, %s) returning id"
	id, err := basestore.ScanInt(store.QueryRow(ctx, sqlf.Sprintf(q, pq.Int32Array(repos), len(repos))))
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
func Load(ctx context.Context, store *basestore.Store, id int) (got *persistentRepoIterator, err error) {
	baseQuery := "select %s from repo_iterator where repo_iterator.id = %s"
	row := store.QueryRow(ctx, sqlf.Sprintf(baseQuery, iteratorJoinCols, id))
	var repos pq.Int32Array
	var tmp persistentRepoIterator
	if err = row.Scan(
		&tmp.id,
		&tmp.CreatedAt,
		&dbutil.NullTime{Time: &tmp.StartedAt},
		&dbutil.NullTime{Time: &tmp.CompletedAt},
		&tmp.LastUpdatedAt,
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
		return nil, errors.Newf("invalid repo iterator state id:%d cursor:%d length:%d", tmp.id, tmp.Cursor, len(repos))
	}

	tmp.errors, err = loadRepoIteratorErrors(ctx, store, &tmp)
	if err != nil {
		return nil, errors.Wrap(err, "loadRepoIteratorErrors")
	}

	tmp.glock = glock.NewRealClock()

	return &tmp, nil
}

// NextWithFinish will iterate the repository set from the current cursor position. If the iterator is marked complete
// or has no more repositories this will do nothing. The finish function returned is a mechanism to have atomic updates,
// callers will need to call the finish function when complete with work. Errors during work processing can be passed
// into the finish function and will be marked as errors on the repo iterator.
func (p *persistentRepoIterator) NextWithFinish() (api.RepoID, bool, finishFunc) {
	current, got := p.peek(p.Cursor + 1)
	if !p.CompletedAt.IsZero() || !got {
		return 0, false, func(ctx context.Context, store *basestore.Store, err error) error {
			return nil
		}
	}
	p.itrStart = p.glock.Now()
	return api.RepoID(current), true, func(ctx context.Context, store *basestore.Store, maybeErr error) error {
		p.itrEnd = p.glock.Now()
		if err := p.doFinish(ctx, store, maybeErr, current); err != nil {
			return err
		}
		return nil
	}
}

// MarkComplete will mark the repo iterator as complete. Once marked complete the iterator is no longer eligible for iteration.
// This can be called at any time to mark the iterator as complete, and does not require the cursor have passed all the way through the set.
func (p *persistentRepoIterator) MarkComplete(ctx context.Context, store *basestore.Store) error {
	now := p.glock.Now()
	err := store.Exec(ctx, sqlf.Sprintf("update repo_iterator set percent_complete = 1, completed_at = %s, last_updated_at = %s", now, now))
	if err != nil {
		return err
	}
	p.CompletedAt = now
	return nil
}

func stampStartedAt(ctx context.Context, store *basestore.Store, itrId int, stampTime time.Time) error {
	return store.Exec(ctx, sqlf.Sprintf("update repo_iterator set started_at = %s where id = %s", stampTime, itrId))
}

func (p *persistentRepoIterator) peek(offset int) (int32, bool) {
	if offset >= len(p.repos) {
		return 0, false
	}
	return p.repos[offset], true
}

func (p *persistentRepoIterator) insertIterationError(ctx context.Context, store *basestore.Store, repoId int32, msg string) (err error) {
	var query *sqlf.Query
	if p.id == 0 {
		return errors.New("invalid iterator to insert iterator error")
	}

	v, ok := p.errors[repoId]
	if !ok {
		query = sqlf.Sprintf("insert into repo_iterator_errors(repo_iterator_id, repo_id, error_message) VALUES (%s, %s, %s) RETURNING %s", p.id, repoId, pq.Array([]string{msg}), errorJoinCols)
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
		p.errors[int32(tmp.RepoId)] = &tmp
	} else {
		v.FailureCount += 1
		query = sqlf.Sprintf("update repo_iterator_errors set failure_count = %s, error_message = array_append(error_message, %s) where id = %s", v.FailureCount, msg, v.id)
		if err = store.Exec(ctx, query); err != nil {
			return errors.Wrap(err, "UpdateIterationError")
		}
	}
	return nil
}

func percent(success, total int) float64 {
	if total == 0 {
		return 1
	}
	return float64(success) / float64(total)
}

func (p *persistentRepoIterator) doFinish(ctx context.Context, store *basestore.Store, maybeErr error, cursorVal int32) (err error) {
	cursor := p.Cursor + 1
	success := p.SuccessCount
	pct := p.PercentComplete
	if maybeErr == nil {
		success = p.SuccessCount + 1
		pct = percent(success, p.TotalCount)
	}
	itrDuration := p.itrEnd.Sub(p.itrStart)

	tx, err := store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	q := "update repo_iterator set percent_complete = %s, success_count = %s, repo_cursor = %s, last_updated_at = %s, runtime_duration = runtime_duration + %s where id = %s"
	if err = tx.Exec(ctx, sqlf.Sprintf(q, pct, success, cursor, p.glock.Now(), itrDuration, p.id)); err != nil {
		return errors.Wrapf(err, "unable to update cursor on iteration success iteratorId: %d, new_cursor:%d", p.id, cursor)
	}
	if maybeErr != nil {
		if err = p.insertIterationError(ctx, tx, cursorVal, maybeErr.Error()); err != nil {
			return errors.Wrapf(err, "unable to upsert error iteratorId: %d, new_cursor:%d", p.id, cursor)
		}
	}
	if p.StartedAt.IsZero() {
		if err = stampStartedAt(ctx, tx, p.id, p.itrStart); err != nil {
			return errors.Wrap(err, "stampStartedAt")
		}
		p.StartedAt = p.itrStart
	}

	p.Cursor = cursor
	p.SuccessCount = success
	p.PercentComplete = pct
	p.RuntimeDuration += itrDuration
	p.itrStart = time.Time{}
	p.itrEnd = time.Time{}

	return nil
}

func loadRepoIteratorErrors(ctx context.Context, store *basestore.Store, iterator *persistentRepoIterator) (got errorMap, err error) {
	baseQuery := "select %s from repo_iterator_errors where repo_iterator_id = %s"
	rows, err := store.Query(ctx, sqlf.Sprintf(baseQuery, errorJoinCols, iterator.id))
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
