package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dineshappavoo/basex"
	"github.com/jackc/pgconn"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

// SQLColumns is a slice of column names, that can be converted to a slice of
// *sqlf.Query.
type SQLColumns []string

// ToSqlf returns all the columns wrapped in a *sqlf.Query.
func (s SQLColumns) ToSqlf() []*sqlf.Query {
	columns := []*sqlf.Query{}
	for _, col := range s {
		columns = append(columns, sqlf.Sprintf(col))
	}
	return columns
}

// seededRand is used in RandomID() to generate a "random" number.
var seededRand = rand.New(rand.NewSource(timeutil.Now().UnixNano()))

// ErrNoResults is returned by Store method calls that found no results.
var ErrNoResults = errors.New("no results")

// RandomID generates a random ID to be used for identifiers in the database.
func RandomID() (string, error) {
	return basex.Encode(strconv.Itoa(seededRand.Int()))
}

// Store exposes methods to read and write batches domain models
// from persistent storage.
type Store struct {
	*basestore.Store
	key                encryption.Key
	now                func() time.Time
	operations         *operations
	observationContext *observation.Context
}

// New returns a new Store backed by the given database.
func New(db dbutil.DB, observationContext *observation.Context, key encryption.Key) *Store {
	return NewWithClock(db, observationContext, key, timeutil.Now)
}

// NewWithClock returns a new Store backed by the given database and
// clock for timestamps.
func NewWithClock(db dbutil.DB, observationContext *observation.Context, key encryption.Key, clock func() time.Time) *Store {
	return &Store{
		Store:              basestore.NewWithDB(db, sql.TxOptions{}),
		key:                key,
		now:                clock,
		operations:         newOperations(observationContext),
		observationContext: observationContext,
	}
}

// ObservationContext returns the observation context wrapped in this store.
func (s *Store) ObservationContext() *observation.Context {
	return s.observationContext
}

// Clock returns the clock used by the Store.
func (s *Store) Clock() func() time.Time { return s.now }

// DB returns the underlying dbutil.DB that this Store was
// instantiated with.
// It's here for legacy reason to pass the dbutil.DB to a repos.Store while
// repos.Store doesn't accept a basestore.TransactableHandle yet.
func (s *Store) DB() dbutil.DB { return s.Handle().DB() }

var _ basestore.ShareableStore = &Store{}

// Handle returns the underlying transactable database handle.
// Needed to implement the ShareableStore interface.
func (s *Store) Handle() *basestore.TransactableHandle { return s.Store.Handle() }

// With creates a new Store with the given basestore.Shareable store as the
// underlying basestore.Store.
// Needed to implement the basestore.Store interface
func (s *Store) With(other basestore.ShareableStore) *Store {
	return &Store{
		Store:              s.Store.With(other),
		key:                s.key,
		operations:         s.operations,
		observationContext: s.observationContext,
		now:                s.now,
	}
}

// Transact creates a new transaction.
// It's required to implement this method and wrap the Transact method of the
// underlying basestore.Store.
func (s *Store) Transact(ctx context.Context) (*Store, error) {
	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &Store{
		Store:              txBase,
		key:                s.key,
		operations:         s.operations,
		observationContext: s.observationContext,
		now:                s.now,
	}, nil
}

// Repos returns a database.RepoStore using the same connection as this store.
func (s *Store) Repos() *database.RepoStore {
	return database.ReposWith(s)
}

// ExternalServices returns a database.ExternalServiceStore using the same connection as this store.
func (s *Store) ExternalServices() *database.ExternalServiceStore {
	return database.ExternalServicesWith(s)
}

// UserCredentials returns a database.UserCredentialsStore using the same connection as this store.
func (s *Store) UserCredentials() *database.UserCredentialsStore {
	return database.UserCredentialsWith(s, s.key)
}

func (s *Store) query(ctx context.Context, q *sqlf.Query, sc scanFunc) error {
	rows, err := s.Store.Query(ctx, q)
	if err != nil {
		return err
	}
	return scanAll(rows, sc)
}

func (s *Store) queryCount(ctx context.Context, q *sqlf.Query) (int, error) {
	count, ok, err := basestore.ScanFirstInt(s.Query(ctx, q))
	if err != nil || !ok {
		return count, err
	}
	return count, nil
}

type operations struct {
	createBatchChange      *observation.Operation
	updateBatchChange      *observation.Operation
	deleteBatchChange      *observation.Operation
	countBatchChanges      *observation.Operation
	getBatchChange         *observation.Operation
	getBatchChangeDiffStat *observation.Operation
	getRepoDiffStat        *observation.Operation
	listBatchChanges       *observation.Operation

	createBatchSpecExecution *observation.Operation
	getBatchSpecExecution    *observation.Operation
	cancelBatchSpecExecution *observation.Operation
	listBatchSpecExecutions  *observation.Operation

	createBatchSpec         *observation.Operation
	updateBatchSpec         *observation.Operation
	deleteBatchSpec         *observation.Operation
	countBatchSpecs         *observation.Operation
	getBatchSpec            *observation.Operation
	getNewestBatchSpec      *observation.Operation
	listBatchSpecs          *observation.Operation
	deleteExpiredBatchSpecs *observation.Operation

	getBulkOperation        *observation.Operation
	listBulkOperations      *observation.Operation
	countBulkOperations     *observation.Operation
	listBulkOperationErrors *observation.Operation

	getChangesetEvent     *observation.Operation
	listChangesetEvents   *observation.Operation
	countChangesetEvents  *observation.Operation
	upsertChangesetEvents *observation.Operation

	createChangesetJob *observation.Operation
	getChangesetJob    *observation.Operation

	createChangesetSpec                      *observation.Operation
	updateChangesetSpec                      *observation.Operation
	deleteChangesetSpec                      *observation.Operation
	countChangesetSpecs                      *observation.Operation
	getChangesetSpec                         *observation.Operation
	listChangesetSpecs                       *observation.Operation
	deleteExpiredChangesetSpecs              *observation.Operation
	getRewirerMappings                       *observation.Operation
	listChangesetSpecsWithConflictingHeadRef *observation.Operation
	deleteChangesetSpecs                     *observation.Operation

	createChangeset                   *observation.Operation
	deleteChangeset                   *observation.Operation
	countChangesets                   *observation.Operation
	getChangeset                      *observation.Operation
	listChangesetSyncData             *observation.Operation
	listChangesets                    *observation.Operation
	enqueueChangeset                  *observation.Operation
	updateChangeset                   *observation.Operation
	updateChangesetBatchChanges       *observation.Operation
	updateChangesetUIPublicationState *observation.Operation
	updateChangesetCodeHostState      *observation.Operation
	getChangesetExternalIDs           *observation.Operation
	cancelQueuedBatchChangeChangesets *observation.Operation
	enqueueChangesetsToClose          *observation.Operation
	getChangesetsStats                *observation.Operation
	getRepoChangesetsStats            *observation.Operation
	enqueueNextScheduledChangeset     *observation.Operation
	getChangesetPlaceInSchedulerQueue *observation.Operation

	listCodeHosts         *observation.Operation
	getExternalServiceIDs *observation.Operation

	createSiteCredential *observation.Operation
	deleteSiteCredential *observation.Operation
	getSiteCredential    *observation.Operation
	listSiteCredentials  *observation.Operation
	updateSiteCredential *observation.Operation

	createBatchSpecWorkspace       *observation.Operation
	getBatchSpecWorkspace          *observation.Operation
	listBatchSpecWorkspaces        *observation.Operation
	markSkippedBatchSpecWorkspaces *observation.Operation

	createBatchSpecWorkspaceExecutionJobs *observation.Operation
	getBatchSpecWorkspaceExecutionJob     *observation.Operation
	listBatchSpecWorkspaceExecutionJobs   *observation.Operation
	cancelBatchSpecWorkspaceExecutionJobs *observation.Operation

	createBatchSpecResolutionJob *observation.Operation
	getBatchSpecResolutionJob    *observation.Operation
	listBatchSpecResolutionJobs  *observation.Operation

	setBatchSpecWorkspaceExecutionJobAccessToken   *observation.Operation
	resetBatchSpecWorkspaceExecutionJobAccessToken *observation.Operation
}

var (
	singletonOperations *operations
	operationsOnce      sync.Once
)

// newOperations generates a singleton of the operations struct.
// TODO: We should create one per observationContext.
func newOperations(observationContext *observation.Context) *operations {
	operationsOnce.Do(func() {
		m := metrics.NewOperationMetrics(
			observationContext.Registerer,
			"batches_dbstore",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)

		op := func(name string) *observation.Operation {
			return observationContext.Operation(observation.Op{
				Name:              fmt.Sprintf("batches.dbstore.%s", name),
				MetricLabelValues: []string{name},
				Metrics:           m,
				ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
					if errors.Is(err, ErrNoResults) {
						return observation.EmitForNone
					}
					return observation.EmitForAll
				},
			})
		}

		singletonOperations = &operations{
			createBatchChange:      op("CreateBatchChange"),
			updateBatchChange:      op("UpdateBatchChange"),
			deleteBatchChange:      op("DeleteBatchChange"),
			countBatchChanges:      op("CountBatchChanges"),
			listBatchChanges:       op("ListBatchChanges"),
			getBatchChange:         op("GetBatchChange"),
			getBatchChangeDiffStat: op("GetBatchChangeDiffStat"),
			getRepoDiffStat:        op("GetRepoDiffStat"),

			createBatchSpecExecution: op("CreateBatchSpecExecution"),
			getBatchSpecExecution:    op("GetBatchSpecExecution"),
			cancelBatchSpecExecution: op("CancelBatchSpecExecution"),
			listBatchSpecExecutions:  op("ListBatchSpecExecutions"),

			createBatchSpec:         op("CreateBatchSpec"),
			updateBatchSpec:         op("UpdateBatchSpec"),
			deleteBatchSpec:         op("DeleteBatchSpec"),
			countBatchSpecs:         op("CountBatchSpecs"),
			getBatchSpec:            op("GetBatchSpec"),
			getNewestBatchSpec:      op("GetNewestBatchSpec"),
			listBatchSpecs:          op("ListBatchSpecs"),
			deleteExpiredBatchSpecs: op("DeleteExpiredBatchSpecs"),

			getBulkOperation:        op("GetBulkOperation"),
			listBulkOperations:      op("ListBulkOperations"),
			countBulkOperations:     op("CountBulkOperations"),
			listBulkOperationErrors: op("ListBulkOperationErrors"),

			getChangesetEvent:     op("GetChangesetEvent"),
			listChangesetEvents:   op("ListChangesetEvents"),
			countChangesetEvents:  op("CountChangesetEvents"),
			upsertChangesetEvents: op("UpsertChangesetEvents"),

			createChangesetJob: op("CreateChangesetJob"),
			getChangesetJob:    op("GetChangesetJob"),

			createChangesetSpec:                      op("CreateChangesetSpec"),
			updateChangesetSpec:                      op("UpdateChangesetSpec"),
			deleteChangesetSpec:                      op("DeleteChangesetSpec"),
			countChangesetSpecs:                      op("CountChangesetSpecs"),
			getChangesetSpec:                         op("GetChangesetSpec"),
			listChangesetSpecs:                       op("ListChangesetSpecs"),
			deleteExpiredChangesetSpecs:              op("DeleteExpiredChangesetSpecs"),
			deleteChangesetSpecs:                     op("DeleteChangesetSpecs"),
			getRewirerMappings:                       op("GetRewirerMappings"),
			listChangesetSpecsWithConflictingHeadRef: op("ListChangesetSpecsWithConflictingHeadRef"),

			createChangeset:                   op("CreateChangeset"),
			deleteChangeset:                   op("DeleteChangeset"),
			countChangesets:                   op("CountChangesets"),
			getChangeset:                      op("GetChangeset"),
			listChangesetSyncData:             op("ListChangesetSyncData"),
			listChangesets:                    op("ListChangesets"),
			enqueueChangeset:                  op("EnqueueChangeset"),
			updateChangeset:                   op("UpdateChangeset"),
			updateChangesetBatchChanges:       op("UpdateChangesetBatchChanges"),
			updateChangesetUIPublicationState: op("UpdateChangesetUIPublicationState"),
			updateChangesetCodeHostState:      op("UpdateChangesetCodeHostState"),
			getChangesetExternalIDs:           op("GetChangesetExternalIDs"),
			cancelQueuedBatchChangeChangesets: op("CancelQueuedBatchChangeChangesets"),
			enqueueChangesetsToClose:          op("EnqueueChangesetsToClose"),
			getChangesetsStats:                op("GetChangesetsStats"),
			getRepoChangesetsStats:            op("GetRepoChangesetsStats"),
			enqueueNextScheduledChangeset:     op("EnqueueNextScheduledChangeset"),
			getChangesetPlaceInSchedulerQueue: op("GetChangesetPlaceInSchedulerQueue"),

			listCodeHosts:         op("ListCodeHosts"),
			getExternalServiceIDs: op("GetExternalServiceIDs"),

			createSiteCredential: op("CreateSiteCredential"),
			deleteSiteCredential: op("DeleteSiteCredential"),
			getSiteCredential:    op("GetSiteCredential"),
			listSiteCredentials:  op("ListSiteCredentials"),
			updateSiteCredential: op("UpdateSiteCredential"),

			createBatchSpecWorkspace:       op("CreateBatchSpecWorkspace"),
			getBatchSpecWorkspace:          op("GetBatchSpecWorkspace"),
			listBatchSpecWorkspaces:        op("ListBatchSpecWorkspaces"),
			markSkippedBatchSpecWorkspaces: op("MarkSkippedBatchSpecWorkspaces"),

			createBatchSpecWorkspaceExecutionJobs: op("CreateBatchSpecWorkspaceExecutionJobs"),
			getBatchSpecWorkspaceExecutionJob:     op("GetBatchSpecWorkspaceExecutionJob"),
			listBatchSpecWorkspaceExecutionJobs:   op("ListBatchSpecWorkspaceExecutionJobs"),
			cancelBatchSpecWorkspaceExecutionJobs: op("CancelBatchSpecWorkspaceExecutionJobs"),

			createBatchSpecResolutionJob: op("CreateBatchSpecResolutionJob"),
			getBatchSpecResolutionJob:    op("GetBatchSpecResolutionJob"),
			listBatchSpecResolutionJobs:  op("ListBatchSpecResolutionJobs"),

			setBatchSpecWorkspaceExecutionJobAccessToken:   op("SetBatchSpecWorkspaceExecutionJobAccessToken"),
			resetBatchSpecWorkspaceExecutionJobAccessToken: op("ResetBatchSpecWorkspaceExecutionJobAccessToken"),
		}
	})

	return singletonOperations
}

// a scanFunc scans one or more rows from a dbutil.Scanner, returning
// the last id column scanned and the count of scanned rows.
type scanFunc func(dbutil.Scanner) (err error)

func scanAll(rows *sql.Rows, scan scanFunc) (err error) {
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err = scan(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

func jsonbColumn(metadata interface{}) (msg json.RawMessage, err error) {
	switch m := metadata.(type) {
	case nil:
		msg = json.RawMessage("{}")
	case string:
		msg = json.RawMessage(m)
	case []byte:
		msg = m
	case json.RawMessage:
		msg = m
	default:
		msg, err = json.MarshalIndent(m, "        ", "    ")
	}
	return
}

func nullInt32Column(n int32) *int32 {
	if n == 0 {
		return nil
	}
	return &n
}

func nullInt64Column(n int64) *int64 {
	if n == 0 {
		return nil
	}
	return &n
}

func nullTimeColumn(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func nullStringColumn(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type LimitOpts struct {
	Limit int
}

func (o LimitOpts) DBLimit() int {
	if o.Limit == 0 {
		return o.Limit
	}
	// We always request one item more than actually requested, to determine the next ID for pagination.
	// The store should make sure to strip the last element in a result set, if len(rs) == o.DBLimit().
	return o.Limit + 1
}

func (o LimitOpts) ToDB() string {
	var limitClause string
	if o.Limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", o.DBLimit())
	}
	return limitClause
}

func isUniqueConstraintViolation(err error, constraintName string) bool {
	var e *pgconn.PgError
	return errors.As(err, &e) && e.Code == "23505" && e.ConstraintName == constraintName
}
