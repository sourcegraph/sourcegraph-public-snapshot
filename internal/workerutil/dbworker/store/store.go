package store

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/derision-test/glock"
	"github.com/grafana/regexp"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type HeartbeatOptions struct {
	// WorkerHostname, if set, enforces worker_hostname to be set to a specific value.
	WorkerHostname string
}

func (o *HeartbeatOptions) ToSQLConds(formatQuery func(query string, args ...any) *sqlf.Query) []*sqlf.Query {
	conds := []*sqlf.Query{}
	if o.WorkerHostname != "" {
		conds = append(conds, formatQuery("{worker_hostname} = %s", o.WorkerHostname))
	}
	return conds
}

type ExecutionLogEntryOptions struct {
	// WorkerHostname, if set, enforces worker_hostname to be set to a specific value.
	WorkerHostname string
	// State, if set, enforces state to be set to a specific value.
	State string
}

func (o *ExecutionLogEntryOptions) ToSQLConds(formatQuery func(query string, args ...any) *sqlf.Query) []*sqlf.Query {
	conds := []*sqlf.Query{}
	if o.WorkerHostname != "" {
		conds = append(conds, formatQuery("{worker_hostname} = %s", o.WorkerHostname))
	}
	if o.State != "" {
		conds = append(conds, formatQuery("{state} = %s", o.State))
	}
	return conds
}

type MarkFinalOptions struct {
	// WorkerHostname, if set, enforces worker_hostname to be set to a specific value.
	WorkerHostname string
}

func (o *MarkFinalOptions) ToSQLConds(formatQuery func(query string, args ...any) *sqlf.Query) []*sqlf.Query {
	conds := []*sqlf.Query{}
	if o.WorkerHostname != "" {
		conds = append(conds, formatQuery("{worker_hostname} = %s", o.WorkerHostname))
	}
	return conds
}

// ErrExecutionLogEntryNotUpdated is returned by AddExecutionLogEntry and UpdateExecutionLogEntry, when
// the log entry was not updated.
var ErrExecutionLogEntryNotUpdated = errors.New("execution log entry not updated")

// Store is the persistence layer for the dbworker package that handles worker-side operations backed by a Postgres
// database. See Options for details on the required shape of the database tables (e.g. table column names/types).
type Store[T workerutil.Record] interface {
	basestore.ShareableStore

	// With creates a new instance of Store using the underlying database
	// handle of the other ShareableStore.
	With(other basestore.ShareableStore) Store[T]

	// CountByState returns the number of records for all the states in
	// stateBitset. This method will trigger a full table scan due to
	// Postgres MVCC, so avoid calling this method frequently, especially
	// for potentially large queues.
	//
	// For example, when the auto-indexing queue on Sourcegraph.com
	// goes over 100K+ records, this method can take 1s+ to run.
	//
	// If possible, prefer using Exists over this function.
	//
	// Pre-condition: stateBitset must be one or more RecordState values or-ed together.
	CountByState(ctx context.Context, stateBitset RecordState) (int, error)

	// Exists checks if there is at least one record in one of the given states
	// in stateBitset.
	//
	// Pre-condition: stateBitset must be one or more RecordState values or-ed together.
	Exists(ctx context.Context, stateBitset RecordState) (bool, error)

	// MaxDurationInQueue returns the maximum age of queued records in this store. Returns 0 if there are no queued records.
	MaxDurationInQueue(ctx context.Context) (time.Duration, error)

	// Dequeue selects the first queued record matching the given conditions and updates the state to processing. If there
	// is such a record, it is returned. If there is no such unclaimed record, a nil record and a nil cancel function
	// will be returned along with a false-valued flag. This method must not be called from within a transaction.
	//
	// The supplied conditions may use the alias provided in `ViewName`, if one was supplied.
	Dequeue(ctx context.Context, workerHostname string, conditions []*sqlf.Query) (T, bool, error)

	// Heartbeat marks the given records as currently being processed and returns the list of records that are
	// still known to the database (to detect lost jobs) and jobs that are marked as to be canceled.
	Heartbeat(ctx context.Context, ids []string, options HeartbeatOptions) (knownIDs, cancelIDs []string, err error)

	// Requeue updates the state of the record with the given identifier to queued and adds a processing delay before
	// the next dequeue of this record can be performed.
	Requeue(ctx context.Context, id int, after time.Time) error

	// AddExecutionLogEntry adds an executor log entry to the record and returns the ID of the new entry (which can be
	// used with UpdateExecutionLogEntry) and a possible error. When the record is not found (due to options not matching
	// or the record being deleted), ErrExecutionLogEntryNotUpdated is returned.
	AddExecutionLogEntry(ctx context.Context, id int, entry executor.ExecutionLogEntry, options ExecutionLogEntryOptions) (entryID int, err error)

	// UpdateExecutionLogEntry updates the executor log entry with the given ID on the given record. When the record is not
	// found (due to options not matching or the record being deleted), ErrExecutionLogEntryNotUpdated is returned.
	UpdateExecutionLogEntry(ctx context.Context, recordID, entryID int, entry executor.ExecutionLogEntry, options ExecutionLogEntryOptions) error

	// MarkComplete attempts to update the state of the record to complete. If this record has already been moved from
	// the processing state to a terminal state, this method will have no effect. This method returns a boolean flag
	// indicating if the record was updated.
	MarkComplete(ctx context.Context, id int, options MarkFinalOptions) (bool, error)

	// MarkErrored attempts to update the state of the record to errored. This method will only have an effect
	// if the current state of the record is processing or completed. A requeued record or a record already marked
	// with an error will not be updated. This method returns a boolean flag indicating if the record was updated.
	MarkErrored(ctx context.Context, id int, failureMessage string, options MarkFinalOptions) (bool, error)

	// MarkFailed attempts to update the state of the record to failed. This method will only have an effect
	// if the current state of the record is processing or completed. A requeued record or a record already marked
	// with an error will not be updated. This method returns a boolean flag indicating if the record was updated.
	MarkFailed(ctx context.Context, id int, failureMessage string, options MarkFinalOptions) (bool, error)

	// ResetStalled moves all processing records that have not received a heartbeat within `StalledMaxAge` back to the
	// queued state. In order to prevent input that continually crashes worker instances, records that have been reset
	// more than `MaxNumResets` times will be marked as failed. This method returns a pair of maps from record
	// identifiers the age of the record's last heartbeat timestamp for each record reset to queued and failed states,
	// respectively.
	ResetStalled(ctx context.Context) (resetLastHeartbeatsByIDs, failedLastHeartbeatsByIDs map[int]time.Duration, err error)
}

type RecordState uint

const (
	StateQueued     RecordState = 1 << 0
	StateErrored    RecordState = 1 << 1
	StateProcessing RecordState = 1 << 2
)

func (bitset RecordState) toList() []*sqlf.Query {
	fragments := []*sqlf.Query{}
	if (bitset & StateQueued) != 0 {
		fragments = append(fragments, sqlf.Sprintf("%s", "queued"))
	}
	if (bitset & StateErrored) != 0 {
		fragments = append(fragments, sqlf.Sprintf("%s", "errored"))
	}
	if (bitset & StateProcessing) != 0 {
		fragments = append(fragments, sqlf.Sprintf("%s", "processing"))
	}
	return fragments
}

type store[T workerutil.Record] struct {
	*basestore.Store
	options                         Options[T]
	columnReplacer                  *strings.Replacer
	modifiedColumnExpressionMatches [][]MatchingColumnExpressions
	operations                      *operations
	logger                          log.Logger
}

var _ Store[workerutil.Record] = &store[workerutil.Record]{}

// Options configure the behavior of Store over a particular set of tables, columns, and expressions.
type Options[T workerutil.Record] struct {
	// Name denotes the name of the store used to distinguish log messages and emitted metrics. The
	// store constructor will fail if this field is not supplied.
	Name string

	// TableName is the name of the table containing work records.
	//
	// The target table (and the target view referenced by `ViewName`) must have the following columns
	// and types:
	//
	//   - id: integer primary key
	//   - state: text (may be updated to `queued`, `processing`, `errored`, or `failed`)
	//   - failure_message: text
	//   - queued_at: timestamp with time zone
	//   - started_at: timestamp with time zone
	//   - last_heartbeat_at: timestamp with time zone
	//   - finished_at: timestamp with time zone
	//   - process_after: timestamp with time zone
	//   - num_resets: integer not null
	//   - num_failures: integer not null
	//   - execution_logs: json[] (each entry has the form of `ExecutionLogEntry`)
	//   - worker_hostname: text
	//
	// The names of these columns may be customized based on the table name by adding a replacement
	// pair in the AlternateColumnNames mapping.
	//
	// It's recommended to put an index or (or partial index) on the state column for more efficient
	// dequeue operations.
	TableName string

	// AlternateColumnNames is a map from expected column names to actual column names in the target
	// table. This allows existing tables to be more easily retrofitted into the expected record
	// shape.
	AlternateColumnNames map[string]string

	// ViewName is an optional name of a view on top of the table containing work records to query when
	// selecting a candidate. If this value is not supplied, `TableName` will be used. The value supplied
	// may also indicate a table alias, which can be referenced in `OrderByExpression`, `ColumnExpressions`,
	// and the conditions supplied to `Dequeue`.
	//
	// The target of this column must be a view on top of the configured table with the same column
	// requirements as the base table described above.
	//
	// Example use case:
	// The processor for LSIF uploads supplies `lsif_uploads_with_repository_name`, a view on top of the
	// `lsif_uploads` table that joins work records with the `repo` table and adds an additional repository
	// name column. This allows `Dequeue` to return a record with additional data so that a second query
	// is not necessary by the caller.
	ViewName string

	// Scan is the function used to scan a resultset into a slice of the expected type.
	Scan ResultsetScanFn[T]

	// OrderByExpression is the SQL expression used to order candidate records when selecting the next
	// batch of work to perform. This expression may use the alias provided in `ViewName`, if one was
	// supplied.
	OrderByExpression *sqlf.Query

	// ColumnExpressions are the target columns provided to the query when selecting a job record. These
	// expressions may use the alias provided in `ViewName`, if one was supplied.
	ColumnExpressions []*sqlf.Query

	// StalledMaxAge is the maximum allowed duration between heartbeat updates of a job's last_heartbeat_at
	// field. An unmodified row that is marked as processing likely indicates that the worker that dequeued
	// the record has died.
	StalledMaxAge time.Duration

	// MaxNumResets is the maximum number of times a record can be implicitly reset back to the queued
	// state (via `ResetStalled`). If a record's reset attempts counter reaches this threshold, it will
	// be moved into the "failed" state rather than queued on its next reset to prevent an infinite retry
	// cycle of the same input.
	MaxNumResets int

	// ResetFailureMessage overrides the default failure message written to job records that have been
	// reset the maximum number of times.
	ResetFailureMessage string

	// RetryAfter determines whether the store dequeues jobs that have errored more than RetryAfter ago.
	// Setting this value to zero will disable retries entirely.
	//
	// If RetryAfter is a non-zero duration, the store dequeues records where:
	//
	//   - the state is 'errored'
	//   - the failed attempts counter hasn't reached MaxNumRetries
	//   - the finished_at timestamp was more than RetryAfter ago
	RetryAfter time.Duration

	// MaxNumRetries is the maximum number of times a record can be retried after an explicit failure.
	// Setting this value to zero will disable retries entirely.
	MaxNumRetries int

	// clock is used to mock out the wall clock used for heartbeat updates.
	clock glock.Clock
}

// ResultsetScanFn is a function that scans row values from a resultset into
// records. This function must close the rows value if the given error value is
// nil.
//
// See the `CloseRows` function in the store/base package for suggested
// implementation details.
type ResultsetScanFn[T workerutil.Record] func(rows *sql.Rows, err error) ([]T, error)

func New[T workerutil.Record](observationCtx *observation.Context, handle basestore.TransactableHandle, options Options[T]) Store[T] {
	return newStore(observationCtx, handle, options)
}

func newStore[T workerutil.Record](observationCtx *observation.Context, handle basestore.TransactableHandle, options Options[T]) *store[T] {
	logger := observationCtx.Logger
	if options.Name == "" {
		panic("no name supplied to github.com/sourcegraph/sourcegraph/internal/dbworker/store:newStore")
	}

	if options.ViewName == "" {
		options.ViewName = options.TableName
	}

	if options.clock == nil {
		options.clock = glock.NewRealClock()
	}

	alternateColumnNames := map[string]string{}
	for _, column := range columnNames {
		alternateColumnNames[column] = column
	}
	for k, v := range options.AlternateColumnNames {
		alternateColumnNames[k] = v
	}

	var replacements []string
	for k, v := range alternateColumnNames {
		replacements = append(replacements, fmt.Sprintf("{%s}", k), v)
	}

	modifiedColumnExpressionMatches := matchModifiedColumnExpressions(options.ViewName, options.ColumnExpressions, alternateColumnNames)

	for i, expression := range options.ColumnExpressions {
		for _, match := range modifiedColumnExpressionMatches[i] {
			if match.exact {
				continue
			}

			logger.Error(``+
				`dbworker store: column expression refers to a column modified by dequeue in a complex expression. `+
				`The given expression will currently evaluate to the OLD value of the row, and the associated handler `+
				`will not have a completely up-to-date record. Please refer to this column without a transform.`,
				log.Int("index", i),
				log.String("expression", expression.Query(sqlf.PostgresBindVar)),
				log.String("columnName", match.columnName),
				log.String("storeName", options.Name),
			)
		}
	}

	return &store[T]{
		Store:                           basestore.NewWithHandle(handle),
		options:                         options,
		columnReplacer:                  strings.NewReplacer(replacements...),
		modifiedColumnExpressionMatches: modifiedColumnExpressionMatches,
		operations:                      newOperations(observationCtx, options.Name),
		logger:                          logger,
	}
}

// With creates a new Store with the given basestore.Shareable store as the
// underlying basestore.Store.
func (s *store[T]) With(other basestore.ShareableStore) Store[T] {
	return &store[T]{
		Store:                           s.Store.With(other),
		options:                         s.options,
		columnReplacer:                  s.columnReplacer,
		modifiedColumnExpressionMatches: s.modifiedColumnExpressionMatches,
		operations:                      s.operations,
		logger:                          s.logger,
	}
}

// columnNames contain the names of the columns expected to be defined by the target table.
// Note: adding a new column to this list requires updating the worker documentation
// https://github.com/sourcegraph/sourcegraph/blob/main/doc/dev/background-information/workers.md#database-backed-stores
var columnNames = []string{
	"id",
	"state",
	"failure_message",
	"queued_at",
	"started_at",
	"last_heartbeat_at",
	"finished_at",
	"process_after",
	"num_resets",
	"num_failures",
	"execution_logs",
	"worker_hostname",
	"cancel",
}

var errAtLeastOneState = errors.New("pre-condition failure: statesBitset should contain at least one state")

func (s *store[T]) CountByState(ctx context.Context, statesBitset RecordState) (_ int, err error) {
	ctx, _, endObservation := s.operations.countByState.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	fragments := statesBitset.toList()
	if len(fragments) == 0 {
		return 0, errAtLeastOneState
	}

	count, _, err := basestore.ScanFirstInt(s.Query(ctx, s.formatQuery(
		countByStateQuery,
		quote(s.options.ViewName),
		sqlf.Join(fragments, ","),
	)))

	return count, err
}

const countByStateQuery = `
SELECT COUNT(*)
FROM %s
WHERE {state} IN (%s)
`

func (s *store[T]) Exists(ctx context.Context, statesBitset RecordState) (_ bool, err error) {
	ctx, _, endObservation := s.operations.exists.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	fragments := statesBitset.toList()
	if len(fragments) == 0 {
		return false, errAtLeastOneState
	}

	exists, _, err := basestore.ScanFirstBool(s.Query(ctx, s.formatQuery(
		existsQuery,
		quote(s.options.ViewName),
		sqlf.Join(fragments, ","),
	)))

	return exists, err
}

const existsQuery = `
SELECT EXISTS (
	SELECT *
	FROM %s
	WHERE {state} IN (%s)
)
`

// MaxDurationInQueue returns the longest duration for which a job associated with this store instance has
// been in the queued state (including errored records that can be retried in the future). This method returns
// a duration of zero if there are no jobs ready for processing.
//
// If records backed by this store do not have an initial state of 'queued', or if it is possible to requeue
// records outside of this package, manual care should be taken to set the queued_at column to the proper time.
// This method makes no guarantees otherwise.
//
// See https://github.com/sourcegraph/sourcegraph/issues/32624.
func (s *store[T]) MaxDurationInQueue(ctx context.Context) (_ time.Duration, err error) {
	ctx, _, endObservation := s.operations.maxDurationInQueue.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	now := s.now()
	retryAfter := int(s.options.RetryAfter / time.Second)

	ageInSeconds, ok, err := basestore.ScanFirstInt(s.Query(ctx, s.formatQuery(
		maxDurationInQueueQuery,
		// oldest_queued
		quote(s.options.TableName),
		now,
		// oldest_retryable
		retryAfter,
		quote(s.options.TableName),
		retryAfter,
		now,
		retryAfter,
	)))
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, nil
	}

	return time.Duration(ageInSeconds) * time.Second, nil
}

const maxDurationInQueueQuery = `
WITH
oldest_queued AS (
	SELECT
		-- Select when the record was most recently dequeueable
		GREATEST({queued_at}, {process_after}) AS last_queued_at
	FROM %s
	WHERE
		{state} = 'queued' AND
		({process_after} IS NULL OR {process_after} <= %s)
),
oldest_retryable AS (
	SELECT
		-- Select when the record was most recently dequeueable
		{finished_at} + (%s * '1 second'::interval) AS last_queued_at
	FROM %s
	WHERE
		%s > 0 AND
		{state} = 'errored' AND
		%s - {finished_at} > (%s * '1 second'::interval)
),
oldest_record AS (
	(
		SELECT last_queued_at FROM oldest_queued
		UNION
		SELECT last_queued_at FROM oldest_retryable
	)
	ORDER BY last_queued_at
	LIMIT 1
)
SELECT EXTRACT(EPOCH FROM NOW() - last_queued_at)::integer AS age FROM oldest_record
`

// columnsUpdatedByDequeue are the unmapped column names modified by the dequeue method.
var columnsUpdatedByDequeue = []string{
	"state",
	"started_at",
	"last_heartbeat_at",
	"finished_at",
	"failure_message",
	"execution_logs",
	"worker_hostname",
}

// Dequeue selects the first queued record matching the given conditions and updates the state to processing. If there
// is such a record, it is returned. If there is no such unclaimed record, a nil record and a nil cancel function
// will be returned along with a false-valued flag. This method must not be called from within a transaction.
//
// A background goroutine that continuously updates the record's last modified time will be started. The returned cancel
// function should be called once the record no longer needs to be locked from selection or reset by another process.
// Most often, this will be when the handler moves the record into a terminal state.
//
// The supplied conditions may use the alias provided in `ViewName`, if one was supplied.
func (s *store[T]) Dequeue(ctx context.Context, workerHostname string, conditions []*sqlf.Query) (ret T, _ bool, err error) {
	ctx, trace, endObservation := s.operations.dequeue.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if s.InTransaction() {
		return ret, false, ErrDequeueTransaction
	}

	now := s.now()
	retryAfter := int(s.options.RetryAfter / time.Second)

	var (
		processingExpr     = sqlf.Sprintf("%s", "processing")
		nowTimestampExpr   = sqlf.Sprintf("%s::timestamp", now)
		nullExpr           = sqlf.Sprintf("NULL")
		workerHostnameExpr = sqlf.Sprintf("%s", workerHostname)
	)

	// NOTE: Changes to this mapping should be reflected in the package variable
	// columnsUpdatedByDequeue, also defined in this file.
	updatedColumns := map[string]*sqlf.Query{
		s.columnReplacer.Replace("{state}"):             processingExpr,
		s.columnReplacer.Replace("{started_at}"):        nowTimestampExpr,
		s.columnReplacer.Replace("{last_heartbeat_at}"): nowTimestampExpr,
		s.columnReplacer.Replace("{finished_at}"):       nullExpr,
		s.columnReplacer.Replace("{failure_message}"):   nullExpr,
		s.columnReplacer.Replace("{execution_logs}"):    nullExpr,
		s.columnReplacer.Replace("{worker_hostname}"):   workerHostnameExpr,
	}

	records, err := s.options.Scan(s.Query(ctx, s.formatQuery(
		dequeueQuery,
		s.options.OrderByExpression,
		quote(s.options.ViewName),
		now,
		retryAfter,
		now,
		retryAfter,
		makeConditionSuffix(conditions),
		s.options.OrderByExpression,
		quote(s.options.TableName),
		quote(s.options.TableName),
		quote(s.options.TableName),
		sqlf.Join(s.makeDequeueUpdateStatements(updatedColumns), ", "),
		sqlf.Join(s.makeDequeueSelectExpressions(updatedColumns), ", "),
		quote(s.options.ViewName),
	)))
	if err != nil {
		return ret, false, err
	}
	if len(records) > 1 {
		return ret, false, errors.Newf("more than one record dequeued: %d", len(records))
	}
	if len(records) == 0 {
		return ret, false, nil
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("recordID", records[0].RecordID()))

	return records[0], true, nil
}

const dequeueQuery = `
WITH potential_candidates AS (
	SELECT
		{id} AS candidate_id,
		ROW_NUMBER() OVER (ORDER BY %s) AS order
	FROM %s
	WHERE
		(
			(
				{state} = 'queued' AND
				({process_after} IS NULL OR {process_after} <= %s)
			) OR (
				%s > 0 AND
				{state} = 'errored' AND
				%s - {finished_at} > (%s * '1 second'::interval)
			)
		)
		%s
	ORDER BY %s
	LIMIT 50
),
candidate AS (
	SELECT
		{id} FROM %s
	JOIN potential_candidates pc ON pc.candidate_id = {id}
	WHERE
		-- Recheck state.
		{state} IN ('queued', 'errored')
	ORDER BY pc.order
	FOR UPDATE OF %s SKIP LOCKED
	LIMIT 1
),
updated_record AS (
	UPDATE
		%s
	SET
		%s
	WHERE
		{id} IN (SELECT {id} FROM candidate)
)
SELECT
	%s
FROM
	%s
WHERE
	{id} IN (SELECT {id} FROM candidate)
`

// makeDequeueSelectExpressions constructs the ordered set of SQL expressions that are returned
// from the dequeue query. This method returns a copy of the configured column expressions slice
// where expressions referencing one of the column updated by dequeue are replaced by the updated
// value.
//
// Note that this method only considers select expressions like `alias.ColumnName` and not something
// more complex like `SomeFunction(alias.ColumnName) + 1`. We issue a warning on construction of a
// new store configured in this way to indicate this (probably) unwanted behavior.
func (s *store[T]) makeDequeueSelectExpressions(updatedColumns map[string]*sqlf.Query) []*sqlf.Query {
	selectExpressions := make([]*sqlf.Query, len(s.options.ColumnExpressions))
	copy(selectExpressions, s.options.ColumnExpressions)

	for i := range selectExpressions {
		for _, match := range s.modifiedColumnExpressionMatches[i] {
			if match.exact {
				selectExpressions[i] = updatedColumns[match.columnName]
				break
			}
		}
	}

	return selectExpressions
}

// makeDequeueUpdateStatements constructs the set of SQL statements that update values of the target table
// in the dequeue query.
func (s *store[T]) makeDequeueUpdateStatements(updatedColumns map[string]*sqlf.Query) []*sqlf.Query {
	updateStatements := make([]*sqlf.Query, 0, len(updatedColumns))
	for columnName, updateValue := range updatedColumns {
		updateStatements = append(updateStatements, sqlf.Sprintf(columnName+"=%s", updateValue))
	}

	return updateStatements
}

func (s *store[T]) Heartbeat(ctx context.Context, ids []string, options HeartbeatOptions) (knownIDs, cancelIDs []string, err error) {
	ctx, _, endObservation := s.operations.heartbeat.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if len(ids) == 0 {
		return []string{}, []string{}, nil
	}

	quotedTableName := quote(s.options.TableName)

	conds := []*sqlf.Query{
		s.formatQuery("{id} = ANY (%s)", pq.Array(ids)),
		s.formatQuery("{state} = 'processing'"),
	}
	conds = append(conds, options.ToSQLConds(s.formatQuery)...)

	scanner := basestore.NewMapScanner(func(scanner dbutil.Scanner) (id string, cancel bool, err error) {
		err = scanner.Scan(&id, &cancel)
		return
	})
	jobMap, err := scanner(s.Query(ctx, s.formatQuery(updateCandidateQuery, quotedTableName, sqlf.Join(conds, "AND"), quotedTableName, s.now())))
	if err != nil {
		return nil, nil, err
	}

	for id, cancel := range jobMap {
		knownIDs = append(knownIDs, id)
		if cancel {
			cancelIDs = append(cancelIDs, id)
		}
	}

	if len(knownIDs) != len(ids) {
	outer:
		for _, recordID := range ids {
			for _, test := range knownIDs {
				if test == recordID {
					continue outer
				}
			}

			var debug string
			intId, convErr := strconv.Atoi(recordID)
			if convErr != nil {
				debug = fmt.Sprintf("can't fetch debug information for job, failed to convert recordID to int: %s", convErr.Error())
			} else {
				var debugErr error
				debug, debugErr = s.fetchDebugInformationForJob(ctx, intId)
				if debugErr != nil {
					s.logger.Error("failed to fetch debug information for job",
						log.String("recordID", recordID),
						log.Error(debugErr),
					)
				}
			}
			s.logger.Error("heartbeat lost a job",
				log.String("recordID", recordID),
				log.String("debug", debug),
				log.String("options.workerHostname", options.WorkerHostname),
			)
		}
	}

	return knownIDs, cancelIDs, nil
}

const updateCandidateQuery = `
WITH alive_candidates AS (
	SELECT
		{id}
	FROM
		%s
	WHERE
		%s
	ORDER BY
		{id} ASC
	FOR UPDATE
)
UPDATE
	%s
SET
	{last_heartbeat_at} = %s
WHERE
	{id} IN (SELECT {id} FROM alive_candidates)
RETURNING {id}, {cancel}
`

// Requeue updates the state of the record with the given identifier to queued and adds a processing delay before
// the next dequeue of this record can be performed.
func (s *store[T]) Requeue(ctx context.Context, id int, after time.Time) (err error) {
	ctx, _, endObservation := s.operations.requeue.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("id", id),
		attribute.Stringer("after", after),
	}})
	defer endObservation(1, observation.Args{})

	return s.Exec(ctx, s.formatQuery(
		requeueQuery,
		quote(s.options.TableName),
		after,
		id,
	))
}

const requeueQuery = `
UPDATE %s
SET
	{state} = 'queued',
	{queued_at} = clock_timestamp(),
	{started_at} = null,
	{process_after} = %s,
	{cancel} = false
WHERE {id} = %s
`

// AddExecutionLogEntry adds an executor log entry to the record and returns the ID of the new entry (which can be
// used with UpdateExecutionLogEntry) and a possible error. When the record is not found (due to options not matching
// or the record being deleted), ErrExecutionLogEntryNotUpdated is returned.
func (s *store[T]) AddExecutionLogEntry(ctx context.Context, id int, entry executor.ExecutionLogEntry, options ExecutionLogEntryOptions) (entryID int, err error) {
	ctx, _, endObservation := s.operations.addExecutionLogEntry.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	conds := []*sqlf.Query{sqlf.Sprintf("%s", true)}
	conds = append(conds, options.ToSQLConds(s.formatQuery)...)

	entryID, ok, err := basestore.ScanFirstInt(s.Query(ctx, s.formatQuery(
		addExecutionLogEntryQuery,
		quote(s.options.TableName),
		id,
		quote(s.options.TableName),
		entry,
		sqlf.Join(conds, "AND"),
	)))
	if err != nil {
		return entryID, err
	}
	if !ok {
		debug, debugErr := s.fetchDebugInformationForJob(ctx, id)
		if debugErr != nil {
			s.logger.Error("failed to fetch debug information for job",
				log.Int("recordID", id),
				log.Error(debugErr),
			)
		}
		s.logger.Error("addExecutionLogEntry failed and didn't match rows",
			log.Int("recordID", id),
			log.String("debug", debug),
			log.String("options.workerHostname", options.WorkerHostname),
			log.String("options.state", options.State),
		)
		return entryID, ErrExecutionLogEntryNotUpdated
	}
	return entryID, nil
}

const addExecutionLogEntryQuery = `
WITH candidate AS (
  -- Directly using id = blah in WHERE clause would sometimes
  -- trigger use of the state index under high contention, so
  -- try forcing the use of pkey on id
  SELECT id FROM %s WHERE id = %s FOR UPDATE
)
UPDATE %s
SET {execution_logs} = {execution_logs} || %s::json
WHERE id IN (SELECT id FROM candidate) AND %s
RETURNING array_length({execution_logs}, 1)
`

// UpdateExecutionLogEntry updates the executor log entry with the given ID on the given record. When the record is not
// found (due to options not matching or the record being deleted), ErrExecutionLogEntryNotUpdated is returned.
func (s *store[T]) UpdateExecutionLogEntry(ctx context.Context, recordID, entryID int, entry executor.ExecutionLogEntry, options ExecutionLogEntryOptions) (err error) {
	ctx, _, endObservation := s.operations.updateExecutionLogEntry.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("recordID", recordID),
		attribute.Int("entryID", entryID),
	}})
	defer endObservation(1, observation.Args{})

	conds := []*sqlf.Query{
		s.formatQuery("{id} = %s", recordID),
		s.formatQuery("array_length({execution_logs}, 1) >= %s", entryID),
	}
	conds = append(conds, options.ToSQLConds(s.formatQuery)...)

	_, ok, err := basestore.ScanFirstInt(s.Query(ctx, s.formatQuery(
		updateExecutionLogEntryQuery,
		quote(s.options.TableName),
		entryID,
		entry,
		sqlf.Join(conds, "AND"),
	)))
	if err != nil {
		return err
	}
	if !ok {
		debug, debugErr := s.fetchDebugInformationForJob(ctx, recordID)
		if debugErr != nil {
			s.logger.Error("failed to fetch debug information for job",
				log.Int("recordID", recordID),
				log.Error(debugErr),
			)
		}
		s.logger.Error("updateExecutionLogEntry failed and didn't match rows",
			log.Int("recordID", recordID),
			log.String("debug", debug),
			log.String("options.workerHostname", options.WorkerHostname),
			log.String("options.state", options.State),
		)

		return ErrExecutionLogEntryNotUpdated
	}

	return nil
}

const updateExecutionLogEntryQuery = `
UPDATE
	%s
SET {execution_logs}[%s] = %s::json
WHERE
	%s
RETURNING
	array_length({execution_logs}, 1)
`

// MarkComplete attempts to update the state of the record to complete. If this record has already been moved from
// the processing state to a terminal state, this method will have no effect. This method returns a boolean flag
// indicating if the record was updated.
func (s *store[T]) MarkComplete(ctx context.Context, id int, options MarkFinalOptions) (_ bool, err error) {
	ctx, _, endObservation := s.operations.markComplete.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	conds := []*sqlf.Query{
		s.formatQuery("{id} = %s", id),
		s.formatQuery("{state} = 'processing'"),
	}
	conds = append(conds, options.ToSQLConds(s.formatQuery)...)

	_, ok, err := basestore.ScanFirstInt(s.Query(ctx, s.formatQuery(markCompleteQuery, quote(s.options.TableName), sqlf.Join(conds, "AND"))))
	return ok, err
}

const markCompleteQuery = `
UPDATE %s
SET {state} = 'completed', {finished_at} = clock_timestamp()
WHERE %s
RETURNING {id}
`

// MarkErrored attempts to update the state of the record to errored. This method will only have an effect
// if the current state of the record is processing. A requeued record or a record already marked with an
// error will not be updated. This method returns a boolean flag indicating if the record was updated.
func (s *store[T]) MarkErrored(ctx context.Context, id int, failureMessage string, options MarkFinalOptions) (_ bool, err error) {
	ctx, _, endObservation := s.operations.markErrored.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	conds := []*sqlf.Query{
		s.formatQuery("{id} = %s", id),
		s.formatQuery("{state} = 'processing'"),
	}
	conds = append(conds, options.ToSQLConds(s.formatQuery)...)

	q := s.formatQuery(markErroredQuery, quote(s.options.TableName), s.options.MaxNumRetries, failureMessage, sqlf.Join(conds, "AND"))
	_, ok, err := basestore.ScanFirstInt(s.Query(ctx, q))
	return ok, err
}

const markErroredQuery = `
UPDATE %s
SET {state} = CASE WHEN {cancel} THEN 'canceled' WHEN {num_failures} + 1 >= %d THEN 'failed' ELSE 'errored' END,
	{finished_at} = clock_timestamp(),
	{failure_message} = %s,
	{num_failures} = CASE WHEN {cancel} THEN {num_failures} ELSE {num_failures} + 1 END
WHERE %s
RETURNING {id}
`

// MarkFailed attempts to update the state of the record to failed. This method will only have an effect
// if the current state of the record is processing. A requeued record or a record already marked with an
// error will not be updated. This method returns a boolean flag indicating if the record was updated.
func (s *store[T]) MarkFailed(ctx context.Context, id int, failureMessage string, options MarkFinalOptions) (_ bool, err error) {
	ctx, _, endObservation := s.operations.markFailed.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	conds := []*sqlf.Query{
		s.formatQuery("{id} = %s", id),
		s.formatQuery("{state} = 'processing'"),
	}
	conds = append(conds, options.ToSQLConds(s.formatQuery)...)

	q := s.formatQuery(markFailedQuery, quote(s.options.TableName), failureMessage, sqlf.Join(conds, "AND"))
	_, ok, err := basestore.ScanFirstInt(s.Query(ctx, q))
	return ok, err
}

const markFailedQuery = `
UPDATE %s
SET {state} = CASE WHEN {cancel} THEN 'canceled' ELSE 'failed' END,
	{finished_at} = clock_timestamp(),
	{failure_message} = %s,
	{num_failures} = CASE WHEN {cancel} THEN {num_failures} ELSE {num_failures} + 1 END
WHERE %s
RETURNING {id}
`

const defaultResetFailureMessage = "job processor died while handling this message too many times"

// ResetStalled moves all processing records that have not received a heartbeat within `StalledMaxAge` back to the
// queued state. In order to prevent input that continually crashes worker instances, records that have been reset
// more than `MaxNumResets` times will be marked as failed. This method returns a pair of maps from record
// identifiers the age of the record's last heartbeat timestamp for each record reset to queued and failed states,
// respectively.
func (s *store[T]) ResetStalled(ctx context.Context) (resetLastHeartbeatsByIDs, failedLastHeartbeatsByIDs map[int]time.Duration, err error) {
	ctx, trace, endObservation := s.operations.resetStalled.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	now := s.now()
	scan := scanLastHeartbeatTimestampsFrom(now)

	resetLastHeartbeatsByIDs, err = scan(s.Query(
		ctx,
		s.formatQuery(
			resetStalledQuery,
			quote(s.options.TableName),
			now,
			int(s.options.StalledMaxAge/time.Second),
			s.options.MaxNumResets,
			quote(s.options.TableName),
		),
	))
	if err != nil {
		return resetLastHeartbeatsByIDs, failedLastHeartbeatsByIDs, err
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("numResetIDs", len(resetLastHeartbeatsByIDs)))

	resetFailureMessage := s.options.ResetFailureMessage
	if resetFailureMessage == "" {
		resetFailureMessage = defaultResetFailureMessage
	}

	failedLastHeartbeatsByIDs, err = scan(s.Query(
		ctx,
		s.formatQuery(
			resetStalledMaxResetsQuery,
			quote(s.options.TableName),
			now,
			int(s.options.StalledMaxAge/time.Second),
			s.options.MaxNumResets,
			quote(s.options.TableName),
			resetFailureMessage,
		),
	))
	if err != nil {
		return resetLastHeartbeatsByIDs, failedLastHeartbeatsByIDs, err
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("numErroredIDs", len(failedLastHeartbeatsByIDs)))

	return resetLastHeartbeatsByIDs, failedLastHeartbeatsByIDs, nil
}

func scanLastHeartbeatTimestampsFrom(now time.Time) func(rows *sql.Rows, queryErr error) (_ map[int]time.Duration, err error) {
	return func(rows *sql.Rows, queryErr error) (_ map[int]time.Duration, err error) {
		if queryErr != nil {
			return nil, queryErr
		}
		defer func() { err = basestore.CloseRows(rows, err) }()

		m := map[int]time.Duration{}
		for rows.Next() {
			var id int
			var lastHeartbeat time.Time
			if err := rows.Scan(&id, &lastHeartbeat); err != nil {
				return nil, err
			}

			m[id] = now.Sub(lastHeartbeat)
		}

		return m, nil
	}
}

const resetStalledQuery = `
WITH stalled AS (
	SELECT {id} FROM %s
	WHERE
		{state} = 'processing' AND
		%s - {last_heartbeat_at} > (%s * '1 second'::interval) AND
		{num_resets} < %s
	FOR UPDATE SKIP LOCKED
)
UPDATE %s
SET
	{state} = 'queued',
	{queued_at} = clock_timestamp(),
	{started_at} = null,
	{num_resets} = {num_resets} + 1
WHERE {id} IN (SELECT {id} FROM stalled)
RETURNING {id}, {last_heartbeat_at}
`

const resetStalledMaxResetsQuery = `
WITH stalled AS (
	SELECT {id} FROM %s
	WHERE
		{state} = 'processing' AND
		%s - {last_heartbeat_at} > (%s * '1 second'::interval) AND
		{num_resets} >= %s
	FOR UPDATE SKIP LOCKED
)
UPDATE %s
SET
	{state} = 'failed',
	{finished_at} = clock_timestamp(),
	{failure_message} = %s
WHERE {id} IN (SELECT {id} FROM stalled)
RETURNING {id}, {last_heartbeat_at}
`

func (s *store[T]) formatQuery(query string, args ...any) *sqlf.Query {
	return sqlf.Sprintf(s.columnReplacer.Replace(query), args...)
}

func (s *store[T]) now() time.Time {
	return s.options.clock.Now().UTC()
}

const fetchDebugInformationForJob = `
SELECT
	row_to_json(%s)
FROM
	%s
WHERE
	{id} = %s
`

func (s *store[T]) fetchDebugInformationForJob(ctx context.Context, recordID int) (debug string, err error) {
	debug, ok, err := basestore.ScanFirstNullString(s.Query(ctx, s.formatQuery(
		fetchDebugInformationForJob,
		quote(extractTableName(s.options.TableName)),
		quote(s.options.TableName),
		recordID,
	)))
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.Newf("fetching debug information for record %d didn't return rows", recordID)
	}
	return debug, nil
}

// quote wraps the given string in a *sqlf.Query so that it is not passed to the database
// as a parameter. It is necessary to quote things such as table names, columns, and other
// expressions that are not simple values.
func quote(s string) *sqlf.Query {
	return sqlf.Sprintf(s)
}

// makeConditionSuffix returns a *sqlf.Query containing "AND {c1 AND c2 AND ...}" when the
// given set of conditions is non-empty, and an empty string otherwise.
func makeConditionSuffix(conditions []*sqlf.Query) *sqlf.Query {
	if len(conditions) == 0 {
		return sqlf.Sprintf("")
	}

	var quotedConditions []*sqlf.Query
	for _, condition := range conditions {
		// Ensure everything is quoted in case the condition has an OR
		quotedConditions = append(quotedConditions, sqlf.Sprintf("(%s)", condition))
	}

	return sqlf.Sprintf("AND %s", sqlf.Join(quotedConditions, " AND "))
}

type MatchingColumnExpressions struct {
	columnName string
	exact      bool
}

// matchModifiedColumnExpressions returns a slice of columns to which each of the
// given column expressions refers. Column references that do not refer to a member
// of the columnsUpdatedByDequeue slice are ignored. Each match indicates the column
// name and whether or not the expression is an exact reference or a reference within
// a more complex expression (arithmetic, function call argument, etc.).
//
// The output slice has the same number of elements as the input column expressions
// and the results are ordered in parallel with the given column expressions.
func matchModifiedColumnExpressions(viewName string, columnExpressions []*sqlf.Query, alternateColumnNames map[string]string) [][]MatchingColumnExpressions {
	matches := make([][]MatchingColumnExpressions, len(columnExpressions))
	columnPrefixes := makeColumnPrefixes(viewName)

	for i, columnExpression := range columnExpressions {
		columnExpressionText := columnExpression.Query(sqlf.PostgresBindVar)

		for _, columnName := range columnsUpdatedByDequeue {
			match := false
			exact := false

			if name, ok := alternateColumnNames[columnName]; ok {
				columnName = name
			}

			for _, columnPrefix := range columnPrefixes {
				if regexp.MustCompile(fmt.Sprintf(`^%s%s$`, columnPrefix, columnName)).MatchString(columnExpressionText) {
					match = true
					exact = true
					break
				}

				if !match && regexp.MustCompile(fmt.Sprintf(`\b%s%s\b`, columnPrefix, columnName)).MatchString(columnExpressionText) {
					match = true
				}
			}

			if match {
				matches[i] = append(matches[i], MatchingColumnExpressions{columnName: columnName, exact: exact})
				break
			}
		}
	}

	return matches
}

// makeColumnPrefixes returns the set of prefixes of a column to indicate that the column belongs to a
// particular table or aliased table. The given name should be the table name  or the aliased table
// reference: `TableName` or `TableName alias`. The return slice always  includes an empty string for
// a bare column reference.
func makeColumnPrefixes(name string) []string {
	parts := strings.Split(name, " ")

	switch len(parts) {
	case 1:
		// name = TableName
		// prefixes = <empty> and `TableName.`
		return []string{"", parts[0] + "."}
	case 2:
		// name = TableName alias
		// prefixes = <empty>, `TableName.`, and `alias.`
		return []string{"", parts[0] + ".", parts[1] + "."}

	default:
		return []string{""}
	}
}

// extractTableName returns the alias if supplied (`Tablename alias`) and the tablename otherwise.
func extractTableName(name string) string {
	parts := strings.Split(name, " ")
	if len(parts) == 2 {
		return parts[1]
	}

	return parts[0]
}
