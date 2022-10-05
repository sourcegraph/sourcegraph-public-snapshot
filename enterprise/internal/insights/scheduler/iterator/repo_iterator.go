package iterator

import (
	"context"
	"time"

	"github.com/derision-test/glock"

	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"

	"github.com/lib/pq"

	"github.com/RoaringBitmap/roaring"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	"github.com/keegancsmith/sqlf"
)

type RepoIterator interface {
	NextWithFinish() (api.RepoID, bool, finishFunc)
}

var _ RepoIterator = &persistentRepoIterator{}

type finishFunc func(ctx context.Context) error

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
	repos           *roaring.Bitmap
	Cursor          int
	errors          map[uint32]*IterationError

	internalItr roaring.IntPeekable

	store basestore.Store
	glock glock.Clock
}

func (p *persistentRepoIterator) NextWithFinish() (api.RepoID, bool, finishFunc) {
	if !p.internalItr.HasNext() {
		return 0, false, func(ctx context.Context) error {}
	}
	current := p.internalItr.PeekNext()

	return api.RepoID(current), true, func(ctx context.Context) error {
		// update internal iterator
		// update cursor
		// write state change to db
		p.internalItr.Next()
		p.Cursor += 1

	}
}

type IterationError struct {
	id            int
	RepoId        int
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
	sqlf.Sprintf("repo_iterator.id"),
	sqlf.Sprintf("repo_iterator.repo_id"),
	sqlf.Sprintf("repo_iterator.error_message"),
	sqlf.Sprintf("repo_iterator.failure_count"),
}
var errorJoinCols = sqlf.Join(repoIteratorErrorCols, ", ")

func Load(ctx context.Context, store basestore.Store, id int) (got *persistentRepoIterator, err error) {
	baseQuery := "select %s from repo_iterator where repo_iterator.id = %s"
	row := store.QueryRow(ctx, sqlf.Sprintf(baseQuery, iteratorJoinCols, id))

	var tmp persistentRepoIterator
	var repos []uint32
	if err = row.Scan(
		&tmp.id,
		&tmp.CreatedAt,
		dbutil.NullTime{Time: &tmp.StartedAt},
		dbutil.NullTime{Time: &tmp.CompletedAt},
		&tmp.LastUpdatedAt,
		&tmp.RuntimeDuration,
		&tmp.PercentComplete,
		&tmp.TotalCount,
		&tmp.SuccessCount,
		pq.Array(&repos),
		&tmp.Cursor,
	); err != nil {
		return nil, errors.Wrap(err, "ScanRepoIterator")
	}
	if tmp.Cursor > len(repos) {
		return nil, errors.Newf("invalid repo iterator state cursor:%d length:%d", tmp.Cursor, len(repos))
	}

	tmp.errors, err = loadRepoIteratorErrors(ctx, store, &tmp)
	if err != nil {
		return nil, errors.Wrap(err, "loadRepoIteratorErrors")
	}

	tmp.repos = roaring.BitmapOf(repos[tmp.Cursor:]...)
	tmp.internalItr = tmp.repos.Iterator()

	tmp.glock = glock.NewRealClock()

	return &tmp, nil
}

func New(ctx context.Context, store basestore.Store, repos []int) (*persistentRepoIterator, error) {
	return NewWithClock(ctx, store, glock.NewRealClock(), repos)
}

func NewWithClock(ctx context.Context, store basestore.Store, clock glock.Clock, repos []int) (*persistentRepoIterator, error) {
	return &persistentRepoIterator{
		store: basestore.Store{},
		glock: nil,
	}, nil
}

func loadRepoIteratorErrors(ctx context.Context, store basestore.Store, iterator *persistentRepoIterator) (got map[uint32]*IterationError, err error) {
	baseQuery := "select %s from repo_iterator_errors where repo_iterator_id = %s"
	rows, err := store.Query(ctx, sqlf.Sprintf(baseQuery, errorJoinCols, iterator.id))
	if err != nil {
		return nil, err
	}
	got = make(map[uint32]*IterationError)
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
		got[uint32(tmp.RepoId)] = &tmp
	}

	return got, err
}

func (p *persistentRepoIterator) updateCursor(ctx context.Context, maybeErr error) (err error) {
	cursor := p.Cursor + 1
	success := p.SuccessCount
	pct := p.PercentComplete
	if maybeErr == nil {
		success = p.SuccessCount + 1
		pct = percent(success, p.TotalCount)
	}

	tx, err := p.store.Transact(ctx)
	if maybeErr != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	q := "update repo_iterator set percent_complete = %s, success_count = %s, repo_cursor = %s, last_updated_at = %s"
	if err = tx.Exec(ctx, sqlf.Sprintf(q, pct, success, cursor, p.glock.Now())); err != nil {
		return errors.Wrapf(err, "unable to update cursor on iteration success iteratorId: %d, new_cursor:%d", p.id, cursor)
	}
	// todo do update errors

	p.Cursor = cursor
	p.SuccessCount = success
	p.PercentComplete = pct
	return nil
}

func (p *persistentRepoIterator) insertIterationError(ctx context.Context, repoId int, msg string) (err error) {
	var query *sqlf.Query
	if p.id == 0 {
		return errors.New("invalid iterator to insert error")
	}

	v, ok := p.errors[uint32(repoId)]
	if !ok {
		query = sqlf.Sprintf("insert into repo_iterator_errors(repo_iterator_id, repo_id, error_message) VALUES (%s, %s, %s) RETURNING %s", p.id, repoId, pq.Array([]string{msg}), errorJoinCols)
		row, err := p.store.QueryRow(ctx, query)

	} else {
		v.FailureCount += 1
		query = sqlf.Sprintf("update repo_iterator_errors set failure_count = %s, error_message = array_append(error_message, %s) where id = %s", v.FailureCount, msg, v.id)
	}
}

func percent(success, total int) float64 {
	if total == 0 {
		return 1
	}
	return float64(success) / float64(total)
}

// operations
// create new
// update
// mark error
// peek
// next
