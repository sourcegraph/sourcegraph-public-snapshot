package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/derision-test/glock"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/scheduler/iterator"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
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

type SeriesBackfillDebug struct {
	Info   BackfillDebugInfo
	Errors []iterator.IterationError
}

type BackfillDebugInfo struct {
	Id              int
	RepoIteratorId  int
	EstimatedCost   float64
	State           BackfillState
	StartedAt       *time.Time
	CompletedAt     *time.Time
	RuntimeDuration *int64
	PercentComplete *float64
	NumRepos        *int
}

func (s *BackfillStore) LoadSeriesBackfillsDebugInfo(ctx context.Context, seriesID int) ([]SeriesBackfillDebug, error) {
	backfills, err := s.LoadSeriesBackfills(ctx, seriesID)
	if err != nil {
		return nil, err
	}
	results := make([]SeriesBackfillDebug, 0, len(backfills))
	for _, backfill := range backfills {
		info := BackfillDebugInfo{
			Id:            backfill.Id,
			EstimatedCost: backfill.EstimatedCost,
			State:         backfill.State,
		}
		backfillErrors := []iterator.IterationError{}
		if backfill.repoIteratorId != 0 {
			it, err := iterator.Load(ctx, s.Store, backfill.repoIteratorId)
			if err != nil {
				return nil, err
			}
			info.RepoIteratorId = backfill.repoIteratorId
			info.StartedAt = &it.StartedAt
			info.CompletedAt = &it.CompletedAt
			info.PercentComplete = &it.PercentComplete
			info.NumRepos = &it.TotalCount
			backfillErrors = it.Errors()
		}
		results = append(results, SeriesBackfillDebug{
			Info:   info,
			Errors: backfillErrors,
		})

	}
	return results, nil
}

type BackfillQueueArgs struct {
	PaginationArgs *database.PaginationArgs
	States         *[]string
	TextSearch     *string
}
type BackfillQueueItem struct {
	ID                  int
	InsightTitle        string
	SeriesID            int
	SeriesLabel         string
	SeriesSearchQuery   string
	BackfillState       string
	PercentComplete     *int
	BackfillCost        *int
	RuntimeDuration     *time.Duration
	BackfillCreatedAt   *time.Time
	BackfillStartedAt   *time.Time
	BackfillCompletedAt *time.Time
	QueuePosition       *int
	Errors              *[]string
}

func (s *BackfillStore) GetBackfillQueueTotalCount(ctx context.Context, args BackfillQueueArgs) (int, error) {
	where := backfillWhere(args)
	query := sqlf.Sprintf(backfillCountSQL, sqlf.Sprintf("WHERE %s", sqlf.Join(where, " AND ")))
	count, _, err := basestore.ScanFirstInt(s.Query(ctx, query))
	return count, err
}

func backfillWhere(args BackfillQueueArgs) []*sqlf.Query {
	where := []*sqlf.Query{sqlf.Sprintf("s.deleted_at IS NULL")}
	if args.TextSearch != nil && len(*args.TextSearch) > 0 {
		likeStr := "%" + *args.TextSearch + "%"
		where = append(where, sqlf.Sprintf("(title LIKE %s OR label LIKE %s)", likeStr, likeStr))
	}

	if args.States != nil && len(*args.States) > 0 {
		states := make([]string, 0, len(*args.States))
		for _, s := range *args.States {
			states = append(states, fmt.Sprintf("'%s'", strings.ToLower(s)))
		}
		where = append(where, sqlf.Sprintf(fmt.Sprintf("state.backfill_state in (%s)", strings.Join(states, ","))))
	}
	return where
}

func (s *BackfillStore) GetBackfillQueueInfo(ctx context.Context, args BackfillQueueArgs) (results []BackfillQueueItem, err error) {
	where := backfillWhere(args)
	pagination := database.PaginationArgs{}
	if args.PaginationArgs != nil {
		pagination = *args.PaginationArgs
	}
	p, err := pagination.SQL()
	if err != nil {
		return nil, err
	}
	// Add in pagination where clause
	if p.Where != nil {
		where = append(where, p.Where)
	}
	query := sqlf.Sprintf(backfillQueueSQL, sqlf.Sprintf("WHERE %s", sqlf.Join(where, " AND ")))
	query = p.AppendOrderToQuery(query)
	query = p.AppendLimitToQuery(query)
	results, err = scanAllBackfillQueueItems(s.Query(ctx, query))
	if err != nil {
		return nil, err
	}
	return results, nil
}

func scanAllBackfillQueueItems(rows *sql.Rows, queryErr error) (_ []BackfillQueueItem, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var results []BackfillQueueItem
	for rows.Next() {
		var temp BackfillQueueItem
		var iteratorErrors []string
		if err := rows.Scan(
			&temp.ID,
			&temp.InsightTitle,
			&temp.SeriesID,
			&temp.SeriesLabel,
			&temp.SeriesSearchQuery,
			&temp.BackfillState,
			&temp.PercentComplete,
			&temp.BackfillCost,
			&temp.RuntimeDuration,
			&temp.BackfillCreatedAt,
			&temp.BackfillStartedAt,
			&temp.BackfillCompletedAt,
			&temp.QueuePosition,
			pq.Array(&iteratorErrors),
		); err != nil {
			return []BackfillQueueItem{}, err
		}
		if iteratorErrors != nil {
			temp.Errors = &iteratorErrors
		}

		results = append(results, temp)
	}
	return results, nil
}

var backfillCountSQL = `
WITH state as (
select isb.id, CASE
  WHEN ijbip.state IS NULL THEN isb.state
  ELSE ijbip.state
END backfill_state
    from insight_series_backfill isb
    left join insights_jobs_backfill_in_progress ijbip on isb.id = ijbip.backfill_id and ijbip.state = 'queued'
    )
select count(*)
from insight_series_backfill isb
    left join repo_iterator ri on isb.repo_iterator_id = ri.id
    join insight_view_series ivs on ivs.insight_series_id = isb.series_id
    join insight_series s on isb.series_id = s.id
    join insight_view iv on ivs.insight_view_id = iv.id
    join state  on isb.id = state.id
%s
`

type BackfillQueueColumn string

const (
	InsightTitle  BackfillQueueColumn = "title"
	SeriesLabel   BackfillQueueColumn = "label"
	State         BackfillQueueColumn = "state.backfill_state"
	BackfillID    BackfillQueueColumn = "isb.id"
	QueuePosition BackfillQueueColumn = "jq.queue_position"
)

var backfillQueueSQL = `
WITH job_queue as (
    select backfill_id, state, row_number() over () queue_position
    from insights_jobs_backfill_in_progress where state = 'queued'  order by cost_bucket
),
errors as (
    select repo_iterator_id, array_agg(err_msg) error_messages
    from repo_iterator_errors, unnest(error_message[:25]) err_msg
    group by  repo_iterator_id
),
state as (
select isb.id, CASE
  WHEN ijbip.state IS NULL THEN isb.state
  ELSE ijbip.state
END backfill_state
    from insight_series_backfill isb
    left join insights_jobs_backfill_in_progress ijbip on isb.id = ijbip.backfill_id and ijbip.state = 'queued'
    )
select isb.id,
       title,
       s.id,
       label,
       query,
       state.backfill_state,
       round(ri.percent_complete *100) percent_complete,
       isb.estimated_cost,
       ri.runtime_duration runtime_duration,
       ri.created_at backfill_created_at,
       ri.started_at backfill_started_at,
       ri.completed_at backfill_completed_at,
       jq.queue_position,
       e.error_messages
from insight_series_backfill isb
    left join repo_iterator ri on isb.repo_iterator_id = ri.id
    left join errors e on isb.repo_iterator_id = e.repo_iterator_id
    left join job_queue jq on jq.backfill_id = isb.id
    join insight_view_series ivs on ivs.insight_series_id = isb.series_id
    join insight_series s on isb.series_id = s.id
    join insight_view iv on ivs.insight_view_id = iv.id
    join state  on isb.id = state.id
	%s
`
