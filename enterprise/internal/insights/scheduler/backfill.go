package scheduler

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/scheduler/iterator"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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

func loadBackfill(ctx context.Context, store *basestore.Store, id int) (*SeriesBackfill, error) {
	q := "SELECT %s FROM insight_series_backfill WHERE id = %s"
	row := store.QueryRow(ctx, sqlf.Sprintf(q, backfillColumnsJoin, id))
	var tmp SeriesBackfill
	if err := row.Scan(
		&tmp.Id,
		&tmp.SeriesId,
		&tmp.repoIteratorId,
		&tmp.EstimatedCost,
		&tmp.State,
	); err != nil {
		return nil, err
	}
	return &tmp, nil
}

func NewBackfill(ctx context.Context, store basestore.Store, series types.InsightSeries, repos []int32) (_ *SeriesBackfill, err error) {
	tx, err := store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	itr, err := iterator.New(ctx, tx, repos)
	if err != nil {
		return nil, err
	}

	q := "INSERT INTO insight_series_backfill (series_id, estimated_cost, repo_iterator_id) VALUES(%s, %s, %s) RETURNING id;"
	cost := 0 // either calculate or take in cost here
	row := tx.QueryRow(ctx, sqlf.Sprintf(q, series.ID, cost, itr.Id))
	id, err := basestore.ScanInt(row)
	if err != nil {
		return nil, err
	}

	return loadBackfill(ctx, tx, id)
}

func (sb *SeriesBackfill) repoIterator(ctx context.Context, store *basestore.Store) (*iterator.PersistentRepoIterator, error) {
	if sb.repoIteratorId == 0 {
		return nil, errors.Newf("invalid repo_iterator_id on backfill_id: %d", sb.Id)
	}
	return iterator.Load(ctx, store, sb.repoIteratorId)
}

var backfillColumns = []*sqlf.Query{
	sqlf.Sprintf("insight_series_backfill.id"),
	sqlf.Sprintf("insight_series_backfill.series_id"),
	sqlf.Sprintf("insight_series_backfill.repo_iterator_id"),
	sqlf.Sprintf("insight_series_backfill.estimated_cost"),
	sqlf.Sprintf("insight_series_backfill.state"),
}

var backfillColumnsJoin = sqlf.Join(backfillColumns, ", ")
