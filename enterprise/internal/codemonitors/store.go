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
	Monitors(ctx context.Context, userID int32, args *graphqlbackend.ListMonitorsArgs) ([]*Monitor, error)
	CountMonitors(ctx context.Context, userID int32) (int32, error)

	UpdateEmailAction(ctx context.Context, monitorID int64, action *graphqlbackend.EditActionArgs) (*MonitorEmail, error)
	CreateEmailAction(ctx context.Context, monitorID int64, action *graphqlbackend.CreateActionArgs) (*MonitorEmail, error)
	DeleteEmailActions(ctx context.Context, actionIDs []int64, monitorID int64) error
	CountEmailActions(ctx context.Context, monitorID int64) (int32, error)
	GetEmailAction(ctx context.Context, emailID int64) (*MonitorEmail, error)
	ListEmailActions(context.Context, ListActionsOpts) ([]*MonitorEmail, error)
	EnqueueActionEmailsForQueryIDInt64(ctx context.Context, queryID int64, triggerEventID int) (err error)

	ListActionJobs(context.Context, ListActionJobsOpts) ([]*ActionJob, error)
	CountActionJobs(context.Context, ListActionJobsOpts) (int, error)
	GetActionJobMetadata(ctx context.Context, recordID int) (*ActionJobMetadata, error)
	GetActionJob(ctx context.Context, recordID int) (*ActionJob, error)
	CreateActions(ctx context.Context, args []*graphqlbackend.CreateActionArgs, monitorID int64) error

	CreateQueryTrigger(ctx context.Context, monitorID int64, args *graphqlbackend.CreateTriggerArgs) (err error)
	UpdateQueryTrigger(ctx context.Context, args *graphqlbackend.UpdateCodeMonitorArgs) (err error)
	TriggerQueryByMonitorIDInt64(ctx context.Context, monitorID int64) (*QueryTrigger, error)
	ResetTriggerQueryTimestamps(ctx context.Context, queryID int64) error
	SetTriggerQueryNextRun(ctx context.Context, triggerQueryID int64, next time.Time, latestResults time.Time) error
	EnqueueTriggerQueries(ctx context.Context) (err error)

	GetQueryByRecordID(ctx context.Context, recordID int) (query *QueryTrigger, err error)

	CreateRecipients(ctx context.Context, recipients []graphql.ID, emailID int64) (err error)
	DeleteRecipients(ctx context.Context, emailID int64) (err error)
	RecipientsForEmailIDInt64(ctx context.Context, emailID int64, args *graphqlbackend.ListRecipientsArgs) ([]*Recipient, error)
	AllRecipientsForEmailIDInt64(ctx context.Context, emailID int64) (rs []*Recipient, err error)
	TotalCountRecipients(ctx context.Context, emailID int64) (count int32, err error)

	DeleteObsoleteJobLogs(ctx context.Context) error
	LogSearch(ctx context.Context, queryString string, numResults int, recordID int) error
	DeleteOldJobLogs(ctx context.Context, retentionInDays int) error

	GetEventsForQueryIDInt64(ctx context.Context, queryID int64, args *graphqlbackend.ListEventsArgs) ([]*TriggerJob, error)
	TotalCountEventsForQueryIDInt64(ctx context.Context, queryID int64) (totalCount int32, err error)
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
