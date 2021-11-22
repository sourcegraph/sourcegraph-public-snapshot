package codemonitors

import (
	"context"
	"database/sql"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

// CodeMonitorStore is an interface for interacting with the code monitor tables in the database
//go:generate ../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors -i CodeMonitorStore -o mock_store_test.go
type CodeMonitorStore interface {
	basestore.ShareableStore
	Transact(context.Context) (CodeMonitorStore, error)
	Done(error) error
	Now() time.Time
	Clock() func() time.Time
	Exec(ctx context.Context, query *sqlf.Query) error

	CreateCodeMonitor(ctx context.Context, args *graphqlbackend.CreateCodeMonitorArgs) (*Monitor, error)

	CreateMonitor(ctx context.Context, args *graphqlbackend.CreateMonitorArgs) (*Monitor, error)
	UpdateMonitor(ctx context.Context, args *graphqlbackend.UpdateCodeMonitorArgs) (*Monitor, error)
	ToggleMonitor(ctx context.Context, args *graphqlbackend.ToggleCodeMonitorArgs) (*Monitor, error)
	DeleteMonitor(ctx context.Context, args *graphqlbackend.DeleteCodeMonitorArgs) error
	GetMonitor(ctx context.Context, monitorID int64) (*Monitor, error)
	ListMonitors(ctx context.Context, userID int32, args *graphqlbackend.ListMonitorsArgs) ([]*Monitor, error)
	CountMonitors(ctx context.Context, userID int32) (int32, error)

	CreateQueryTrigger(ctx context.Context, monitorID int64, args *graphqlbackend.CreateTriggerArgs) error
	UpdateQueryTrigger(ctx context.Context, args *graphqlbackend.UpdateCodeMonitorArgs) error
	GetQueryTriggerForMonitor(ctx context.Context, monitorID int64) (*QueryTrigger, error)
	ResetQueryTriggerTimestamps(ctx context.Context, queryID int64) error
	SetQueryTriggerNextRun(ctx context.Context, triggerQueryID int64, next time.Time, latestResults time.Time) error
	GetQueryTriggerForJob(ctx context.Context, jobID int) (*QueryTrigger, error)

	DeleteObsoleteTriggerJobs(ctx context.Context) error
	UpdateTriggerJobWithResults(ctx context.Context, queryString string, numResults int, recordID int) error
	DeleteOldTriggerJobs(ctx context.Context, retentionInDays int) error

	EnqueueQueryTriggerJobs(ctx context.Context) error
	ListQueryTriggerJobs(ctx context.Context, queryID int64, args *graphqlbackend.ListEventsArgs) ([]*TriggerJob, error)
	CountQueryTriggerJobs(ctx context.Context, queryID int64) (int32, error)

	CreateActions(ctx context.Context, args []*graphqlbackend.CreateActionArgs, monitorID int64) error

	UpdateEmailAction(ctx context.Context, monitorID int64, action *graphqlbackend.EditActionArgs) (*EmailAction, error)
	CreateEmailAction(ctx context.Context, monitorID int64, action *graphqlbackend.CreateActionArgs) (*EmailAction, error)
	DeleteEmailActions(ctx context.Context, actionIDs []int64, monitorID int64) error
	CountEmailActions(ctx context.Context, monitorID int64) (int32, error)
	GetEmailAction(ctx context.Context, emailID int64) (*EmailAction, error)
	ListEmailActions(context.Context, ListActionsOpts) ([]*EmailAction, error)

	CreateRecipients(ctx context.Context, recipients []graphql.ID, emailID int64) error
	DeleteRecipients(ctx context.Context, emailID int64) error
	ListRecipientsForEmailAction(ctx context.Context, emailID int64, args *graphqlbackend.ListRecipientsArgs) ([]*Recipient, error)
	ListAllRecipientsForEmailAction(ctx context.Context, emailID int64) ([]*Recipient, error)
	CountRecipients(ctx context.Context, emailID int64) (int32, error)

	ListActionJobs(context.Context, ListActionJobsOpts) ([]*ActionJob, error)
	CountActionJobs(context.Context, ListActionJobsOpts) (int, error)
	GetActionJobMetadata(ctx context.Context, recordID int) (*ActionJobMetadata, error)
	GetActionJob(ctx context.Context, recordID int) (*ActionJob, error)
	EnqueueActionJobsForQuery(ctx context.Context, queryID int64, triggerEventID int) error
}

// codeMonitorStore exposes methods to read and write codemonitors domain models
// from persistent storage.
type codeMonitorStore struct {
	*basestore.Store
	now func() time.Time
}

var _ CodeMonitorStore = (*codeMonitorStore)(nil)

// NewStore returns a new Store backed by the given database.
func NewStore(db dbutil.DB) *codeMonitorStore {
	return NewStoreWithClock(db, timeutil.Now)
}

// NewStoreWithClock returns a new Store backed by the given database and
// clock for timestamps.
func NewStoreWithClock(db dbutil.DB, clock func() time.Time) *codeMonitorStore {
	return &codeMonitorStore{Store: basestore.NewWithDB(db, sql.TxOptions{}), now: clock}
}

// Clock returns the clock of the underlying store.
func (s *codeMonitorStore) Clock() func() time.Time {
	return s.now
}

func (s *codeMonitorStore) Now() time.Time {
	return s.now()
}

// Transact creates a new transaction.
// It's required to implement this method and wrap the Transact method of the
// underlying basestore.Store.
func (s *codeMonitorStore) Transact(ctx context.Context) (CodeMonitorStore, error) {
	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &codeMonitorStore{Store: txBase, now: s.now}, nil
}
