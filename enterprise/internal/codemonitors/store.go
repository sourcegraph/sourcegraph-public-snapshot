package codemonitors

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

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

	CreateMonitor(ctx context.Context, args MonitorArgs) (*Monitor, error)
	UpdateMonitor(ctx context.Context, id int64, args MonitorArgs) (*Monitor, error)
	UpdateMonitorEnabled(ctx context.Context, id int64, enabled bool) (*Monitor, error)
	DeleteMonitor(ctx context.Context, id int64) error
	GetMonitor(ctx context.Context, monitorID int64) (*Monitor, error)
	ListMonitors(context.Context, ListMonitorsOpts) ([]*Monitor, error)
	CountMonitors(ctx context.Context, userID int32) (int32, error)

	CreateQueryTrigger(ctx context.Context, monitorID int64, query string) (*QueryTrigger, error)
	UpdateQueryTrigger(ctx context.Context, id int64, query string) error
	GetQueryTriggerForMonitor(ctx context.Context, monitorID int64) (*QueryTrigger, error)
	ResetQueryTriggerTimestamps(ctx context.Context, queryID int64) error
	SetQueryTriggerNextRun(ctx context.Context, triggerQueryID int64, next time.Time, latestResults time.Time) error
	GetQueryTriggerForJob(ctx context.Context, triggerJob int32) (*QueryTrigger, error)
	EnqueueQueryTriggerJobs(context.Context) ([]*TriggerJob, error)
	ListQueryTriggerJobs(context.Context, ListTriggerJobsOpts) ([]*TriggerJob, error)
	CountQueryTriggerJobs(ctx context.Context, queryID int64) (int32, error)

	DeleteObsoleteTriggerJobs(ctx context.Context) error
	UpdateTriggerJobWithResults(ctx context.Context, triggerJobID int32, queryString string, numResults int) error
	DeleteOldTriggerJobs(ctx context.Context, retentionInDays int) error

	UpdateEmailAction(_ context.Context, id int64, _ *EmailActionArgs) (*EmailAction, error)
	CreateEmailAction(ctx context.Context, monitorID int64, _ *EmailActionArgs) (*EmailAction, error)
	DeleteEmailActions(ctx context.Context, actionIDs []int64, monitorID int64) error
	GetEmailAction(ctx context.Context, emailID int64) (*EmailAction, error)
	ListEmailActions(context.Context, ListActionsOpts) ([]*EmailAction, error)

	UpdateWebhookAction(_ context.Context, id int64, enabled bool, url string) (*WebhookAction, error)
	CreateWebhookAction(ctx context.Context, monitorID int64, enabled bool, url string) (*WebhookAction, error)
	DeleteWebhookActions(ctx context.Context, monitorID int64, ids ...int64) error
	CountWebhookActions(ctx context.Context, monitorID int64) (int, error)
	GetWebhookAction(ctx context.Context, id int64) (*WebhookAction, error)
	ListWebhookActions(context.Context, ListActionsOpts) ([]*WebhookAction, error)

	UpdateSlackWebhookAction(_ context.Context, id int64, enabled bool, url string) (*SlackWebhookAction, error)
	CreateSlackWebhookAction(ctx context.Context, monitorID int64, enabled bool, url string) (*SlackWebhookAction, error)
	DeleteSlackWebhookActions(ctx context.Context, monitorID int64, ids ...int64) error
	CountSlackWebhookActions(ctx context.Context, monitorID int64) (int, error)
	GetSlackWebhookAction(ctx context.Context, id int64) (*SlackWebhookAction, error)
	ListSlackWebhookActions(context.Context, ListActionsOpts) ([]*SlackWebhookAction, error)

	CreateRecipient(ctx context.Context, emailID int64, userID, orgID *int32) (*Recipient, error)
	DeleteRecipients(ctx context.Context, emailID int64) error
	ListRecipients(context.Context, ListRecipientsOpts) ([]*Recipient, error)
	CountRecipients(ctx context.Context, emailID int64) (int32, error)

	ListActionJobs(context.Context, ListActionJobsOpts) ([]*ActionJob, error)
	CountActionJobs(context.Context, ListActionJobsOpts) (int, error)
	GetActionJobMetadata(ctx context.Context, jobID int32) (*ActionJobMetadata, error)
	GetActionJob(ctx context.Context, jobID int32) (*ActionJob, error)
	EnqueueActionJobsForMonitor(ctx context.Context, monitorID int64, triggerJob int32) ([]*ActionJob, error)
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
