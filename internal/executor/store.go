package executor

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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

type ExecutorStore interface {
	dbworkerstore.Store

	With(other basestore.ShareableStore) ExecutorStore

	// AddExecutionLogEntry adds an executor log entry to the record and returns the ID of the new entry (which can be
	// used with UpdateExecutionLogEntry) and a possible error. When the record is not found (due to options not matching
	// or the record being deleted), ErrExecutionLogEntryNotUpdated is returned.
	AddExecutionLogEntry(ctx context.Context, id int, entry ExecutionLogEntry, options ExecutionLogEntryOptions) (entryID int, err error)

	// UpdateExecutionLogEntry updates the executor log entry with the given ID on the given record. When the record is not
	// found (due to options not matching or the record being deleted), ErrExecutionLogEntryNotUpdated is returned.
	UpdateExecutionLogEntry(ctx context.Context, recordID, entryID int, entry ExecutionLogEntry, options ExecutionLogEntryOptions) error
}

// ExecutionLogEntry represents a command run by the executor.
type ExecutionLogEntry struct {
	Key        string    `json:"key"`
	Command    []string  `json:"command"`
	StartTime  time.Time `json:"startTime"`
	ExitCode   *int      `json:"exitCode,omitempty"`
	Out        string    `json:"out,omitempty"`
	DurationMs *int      `json:"durationMs,omitempty"`
}

func (e *ExecutionLogEntry) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.Errorf("value is not []byte: %T", value)
	}

	return json.Unmarshal(b, &e)
}

func (e ExecutionLogEntry) Value() (driver.Value, error) {
	return json.Marshal(e)
}

type operations struct {
	addExecutionLogEntry    *observation.Operation
	updateExecutionLogEntry *observation.Operation
}

type OnCompleteHook func(ctx context.Context, tx ExecutorStore, job workerutil.Record) error
type OnFailedHook func(ctx context.Context, tx ExecutorStore, job workerutil.Record) error

type ExecutorStoreOptions struct {
	dbworkerstore.Options
	OnComplete OnCompleteHook
	OnFailed   OnFailedHook
}

func NewExecutorStore() ExecutorStore {
	return &store{
		Store:      dbworkerstore.NewWithMetrics(nil, dbworkerstore.Options{}, nil),
		logger:     log.Scoped("ExecutorStore", ""),
		operations: &operations{},
	}
}

type store struct {
	baseStore *basestore.Store
	dbworkerstore.Store

	options dbworkerstore.Options

	logger log.Logger

	operations *operations
}

// AddExecutionLogEntry adds an executor log entry to the record and returns the ID of the new entry (which can be
// used with UpdateExecutionLogEntry) and a possible error. When the record is not found (due to options not matching
// or the record being deleted), ErrExecutionLogEntryNotUpdated is returned.
func (s *store) AddExecutionLogEntry(ctx context.Context, id int, entry ExecutionLogEntry, options ExecutionLogEntryOptions) (entryID int, err error) {
	ctx, _, endObservation := s.operations.addExecutionLogEntry.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	conds := []*sqlf.Query{
		s.formatQuery("{id} = %s", id),
	}
	conds = append(conds, options.ToSQLConds(s.formatQuery)...)

	entryID, ok, err := basestore.ScanFirstInt(s.baseStore.Query(ctx, s.formatQuery(
		addExecutionLogEntryQuery,
		sqlf.Sprintf(s.options.TableName),
		ExecutionLogEntry(entry),
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
UPDATE
	%s
SET {execution_logs} = {execution_logs} || %s::json
WHERE
	%s
RETURNING array_length({execution_logs}, 1)
`

// UpdateExecutionLogEntry updates the executor log entry with the given ID on the given record. When the record is not
// found (due to options not matching or the record being deleted), ErrExecutionLogEntryNotUpdated is returned.
func (s *store) UpdateExecutionLogEntry(ctx context.Context, recordID, entryID int, entry ExecutionLogEntry, options ExecutionLogEntryOptions) (err error) {
	ctx, _, endObservation := s.operations.updateExecutionLogEntry.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("recordID", recordID),
		otlog.Int("entryID", entryID),
	}})
	defer endObservation(1, observation.Args{})

	conds := []*sqlf.Query{
		s.formatQuery("{id} = %s", recordID),
		s.formatQuery("array_length({execution_logs}, 1) >= %s", entryID),
	}
	conds = append(conds, options.ToSQLConds(s.formatQuery)...)

	_, ok, err := basestore.ScanFirstInt(s.baseStore.Query(ctx, s.formatQuery(
		updateExecutionLogEntryQuery,
		sqlf.Sprintf(s.options.TableName),
		entryID,
		ExecutionLogEntry(entry),
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

// ErrExecutionLogEntryNotUpdated is returned by AddExecutionLogEntry and UpdateExecutionLogEntry, when
// the log entry was not updated.
var ErrExecutionLogEntryNotUpdated = errors.New("execution log entry not updated")
