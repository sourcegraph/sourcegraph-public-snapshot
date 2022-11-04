package scheduler

import (
	"context"
	"database/sql"

	"github.com/derision-test/glock"
	"github.com/keegancsmith/sqlf"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/scheduler/iterator"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type BackfillStore struct {
	*basestore.Store
	clock glock.Clock
}

func NewBackfillStore(edb edb.InsightsDB) *BackfillStore {
	return newBackfillStoreWithClock(edb, glock.NewRealClock())
}
func newBackfillStoreWithClock(edb edb.InsightsDB, clock glock.Clock) *BackfillStore {
	return &BackfillStore{Store: basestore.NewWithHandle(edb.Handle()), clock: clock}
}

func (s *BackfillStore) With(other basestore.ShareableStore) *BackfillStore {
	return &BackfillStore{Store: s.Store.With(other), clock: s.clock}
}

func (s *BackfillStore) Transact(ctx context.Context) (*BackfillStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &BackfillStore{Store: txBase, clock: s.clock}, err
}

type SeriesBackfill struct {
	Id             int
	SeriesId       int
	repoIteratorId int
	EstimatedCost  float64
	State          BackfillState
}

type BackfillState string

const (
	BackfillStateNew        BackfillState = "new"
	BackfillStateProcessing BackfillState = "processing"
	BackfillStateCompleted  BackfillState = "completed"
	BackfillStateFailed     BackfillState = "failed"
)

func (s *BackfillStore) NewBackfill(ctx context.Context, series types.InsightSeries) (_ *SeriesBackfill, err error) {
	q := "INSERT INTO insight_series_backfill (series_id, state) VALUES(%s, %s) RETURNING %s;"
	row := s.QueryRow(ctx, sqlf.Sprintf(q, series.ID, string(BackfillStateNew), backfillColumnsJoin))
	return scanBackfill(row)
}

func (s *BackfillStore) loadBackfill(ctx context.Context, id int) (*SeriesBackfill, error) {
	q := "SELECT %s FROM insight_series_backfill WHERE id = %s"
	row := s.QueryRow(ctx, sqlf.Sprintf(q, backfillColumnsJoin, id))
	return scanBackfill(row)
}

func (s *BackfillStore) LoadSeriesBackfills(ctx context.Context, seriesID int) ([]SeriesBackfill, error) {
	q := "SELECT %s FROM insight_series_backfill where series_id = %s"
	return scanAllBackfills(s.Query(ctx, sqlf.Sprintf(q, backfillColumnsJoin, seriesID)))
}

func scanBackfill(scanner dbutil.Scanner) (*SeriesBackfill, error) {
	var tmp SeriesBackfill
	var cost *float64
	if err := scanner.Scan(
		&tmp.Id,
		&tmp.SeriesId,
		&dbutil.NullInt{N: &tmp.repoIteratorId},
		&cost,
		&tmp.State,
	); err != nil {
		return nil, err
	}
	if cost != nil {
		tmp.EstimatedCost = *cost
	}
	return &tmp, nil
}

func (b *SeriesBackfill) SetScope(ctx context.Context, store *BackfillStore, repos []int32, cost float64) (*SeriesBackfill, error) {
	if b == nil || b.Id == 0 {
		return nil, errors.New("invalid series backfill")
	}

	tx, err := store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	itr, err := iterator.NewWithClock(ctx, tx.Store, store.clock, repos)
	if err != nil {
		return nil, errors.Wrap(err, "iterator.New")
	}

	q := "UPDATE insight_series_backfill set repo_iterator_id = %s, estimated_cost = %s, state = %s where id = %s RETURNING %s"
	row := tx.QueryRow(ctx, sqlf.Sprintf(q, itr.Id, cost, string(BackfillStateProcessing), b.Id, backfillColumnsJoin))
	return scanBackfill(row)
}

func (b *SeriesBackfill) SetCompleted(ctx context.Context, store *BackfillStore) error {
	return b.setState(ctx, store, BackfillStateCompleted)
}

func (b *SeriesBackfill) SetFailed(ctx context.Context, store *BackfillStore) error {
	return b.setState(ctx, store, BackfillStateFailed)
}

func (b *SeriesBackfill) setState(ctx context.Context, store *BackfillStore, newState BackfillState) error {
	err := store.Exec(ctx, sqlf.Sprintf("update insight_series_backfill set state = %s where id = %s;", string(newState), b.Id))
	if err != nil {
		return err
	}
	b.State = newState
	return nil
}

func (b *SeriesBackfill) IsTerminalState() bool {
	return b.State == BackfillStateCompleted || b.State == BackfillStateFailed
}

func (sb *SeriesBackfill) repoIterator(ctx context.Context, store *BackfillStore) (*iterator.PersistentRepoIterator, error) {
	if sb.repoIteratorId == 0 {
		return nil, errors.Newf("invalid repo_iterator_id on backfill_id: %d", sb.Id)
	}
	return iterator.LoadWithClock(ctx, store.Store, sb.repoIteratorId, store.clock)
}

var backfillColumns = []*sqlf.Query{
	sqlf.Sprintf("insight_series_backfill.id"),
	sqlf.Sprintf("insight_series_backfill.series_id"),
	sqlf.Sprintf("insight_series_backfill.repo_iterator_id"),
	sqlf.Sprintf("insight_series_backfill.estimated_cost"),
	sqlf.Sprintf("insight_series_backfill.state"),
}

var backfillColumnsJoin = sqlf.Join(backfillColumns, ", ")

func scanAllBackfills(rows *sql.Rows, queryErr error) (_ []SeriesBackfill, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var results []SeriesBackfill
	for rows.Next() {
		var cost *float64
		var temp SeriesBackfill
		if err := rows.Scan(
			&temp.Id,
			&temp.SeriesId,
			&dbutil.NullInt{N: &temp.repoIteratorId},
			&cost,
			&temp.State,
		); err != nil {
			return []SeriesBackfill{}, err
		}
		if cost != nil {
			temp.EstimatedCost = *cost
		}
		results = append(results, temp)
	}
	return results, nil
}
