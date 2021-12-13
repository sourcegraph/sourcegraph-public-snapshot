package resolvers

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	cm "github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors/background"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// NewResolver returns a new Resolver that uses the given database
func NewResolver(db database.DB) graphqlbackend.CodeMonitorsResolver {
	return &Resolver{store: cm.NewStore(db)}
}

// newResolverWithClock is used in tests to set the clock manually.
func newResolverWithClock(db database.DB, clock func() time.Time) graphqlbackend.CodeMonitorsResolver {
	return &Resolver{store: cm.NewStoreWithClock(db, clock)}
}

type Resolver struct {
	store cm.CodeMonitorStore
}

func (r *Resolver) Now() time.Time {
	return r.store.Now()
}

func (r *Resolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		MonitorKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.MonitorByID(ctx, id)
		},
		// TODO: These kinds are currently not implemented, but need a node resolver.
		// monitorTriggerQueryKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
		// 	return r.MonitorTriggerQueryByID(ctx, id)
		// },
		// monitorTriggerEventKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
		// 	return r.MonitorTriggerEventByID(ctx, id)
		// },
		// monitorActionEmailKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
		// 	return r.MonitorActionEmailByID(ctx, id)
		// },
		// monitorActionEventKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
		// 	return r.MonitorActionEventByID(ctx, id)
		// },
	}
}

func (r *Resolver) Monitors(ctx context.Context, userID int32, args *graphqlbackend.ListMonitorsArgs) (graphqlbackend.MonitorConnectionResolver, error) {
	// Request one extra to determine if there are more pages
	newArgs := *args
	newArgs.First += 1

	after, err := unmarshalAfter(args.After)
	if err != nil {
		return nil, err
	}

	ms, err := r.store.ListMonitors(ctx, cm.ListMonitorsOpts{
		UserID: &userID,
		First:  intPtr(int(newArgs.First)),
		After:  intPtrToInt64Ptr(after),
	})
	if err != nil {
		return nil, err
	}

	totalCount, err := r.store.CountMonitors(ctx, userID)
	if err != nil {
		return nil, err
	}

	hasNextPage := false
	if len(ms) == int(args.First)+1 {
		hasNextPage = true
		ms = ms[:len(ms)-1]
	}

	mrs := make([]graphqlbackend.MonitorResolver, 0, len(ms))
	for _, m := range ms {
		mrs = append(mrs, &monitor{
			Resolver: r,
			Monitor:  m,
		})
	}

	return &monitorConnection{Resolver: r, monitors: mrs, totalCount: totalCount, hasNextPage: hasNextPage}, nil
}

func (r *Resolver) MonitorByID(ctx context.Context, id graphql.ID) (graphqlbackend.MonitorResolver, error) {
	err := r.isAllowedToEdit(ctx, id)
	if err != nil {
		return nil, err
	}
	var monitorID int64
	err = relay.UnmarshalSpec(id, &monitorID)
	if err != nil {
		return nil, err
	}
	mo, err := r.store.GetMonitor(ctx, monitorID)
	if err != nil {
		return nil, err
	}
	return &monitor{
		Resolver: r,
		Monitor:  mo,
	}, nil
}

func (r *Resolver) CreateCodeMonitor(ctx context.Context, args *graphqlbackend.CreateCodeMonitorArgs) (graphqlbackend.MonitorResolver, error) {
	if err := r.isAllowedToCreate(ctx, args.Monitor.Namespace); err != nil {
		return nil, err
	}

	// Start transaction.
	tx, err := r.transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.store.Done(err) }()

	userID, orgID, err := graphqlbackend.UnmarshalNamespaceToIDs(args.Monitor.Namespace)
	if err != nil {
		return nil, err
	}

	// Create monitor.
	m, err := tx.store.CreateMonitor(ctx, cm.MonitorArgs{
		Description:     args.Monitor.Description,
		Enabled:         args.Monitor.Enabled,
		NamespaceUserID: userID,
		NamespaceOrgID:  orgID,
	})
	if err != nil {
		return nil, err
	}

	// Create trigger.
	_, err = tx.store.CreateQueryTrigger(ctx, m.ID, args.Trigger.Query)
	if err != nil {
		return nil, err
	}

	// Create actions.
	err = tx.createActions(ctx, m.ID, args.Actions)
	if err != nil {
		return nil, err
	}
	return &monitor{
		Resolver: r,
		Monitor:  m,
	}, nil
}

func (r *Resolver) ToggleCodeMonitor(ctx context.Context, args *graphqlbackend.ToggleCodeMonitorArgs) (graphqlbackend.MonitorResolver, error) {
	err := r.isAllowedToEdit(ctx, args.Id)
	if err != nil {
		return nil, errors.Errorf("UpdateMonitorEnabled: %w", err)
	}
	var monitorID int64
	if err := relay.UnmarshalSpec(args.Id, &monitorID); err != nil {
		return nil, err
	}

	mo, err := r.store.UpdateMonitorEnabled(ctx, monitorID, args.Enabled)
	if err != nil {
		return nil, err
	}
	return &monitor{r, mo}, nil
}

func (r *Resolver) DeleteCodeMonitor(ctx context.Context, args *graphqlbackend.DeleteCodeMonitorArgs) (*graphqlbackend.EmptyResponse, error) {
	err := r.isAllowedToEdit(ctx, args.Id)
	if err != nil {
		return nil, errors.Errorf("DeleteCodeMonitor: %w", err)
	}

	var monitorID int64
	if err := relay.UnmarshalSpec(args.Id, &monitorID); err != nil {
		return nil, err
	}

	if err := r.store.DeleteMonitor(ctx, monitorID); err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) UpdateCodeMonitor(ctx context.Context, args *graphqlbackend.UpdateCodeMonitorArgs) (graphqlbackend.MonitorResolver, error) {
	err := r.isAllowedToEdit(ctx, args.Monitor.Id)
	if err != nil {
		return nil, errors.Errorf("UpdateCodeMonitor: %w", err)
	}

	err = r.isAllowedToCreate(ctx, args.Monitor.Update.Namespace)
	if err != nil {
		return nil, errors.Errorf("update namespace: %w", err)
	}

	var monitorID int64
	err = relay.UnmarshalSpec(args.Monitor.Id, &monitorID)
	if err != nil {
		return nil, err
	}

	// Get all action IDs of the monitor.
	actionIDs, err := r.actionIDsForMonitorIDInt64(ctx, monitorID)
	if err != nil {
		return nil, err
	}

	toCreate, toDelete, err := splitActionIDs(ctx, args, actionIDs)
	if len(toDelete) == len(actionIDs) {
		return nil, errors.Errorf("you tried to delete all actions, but every monitor must be connected to at least 1 action")
	}

	// Run all queries within a transaction.
	tx, err := r.transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.store.Done(err) }()

	err = tx.store.DeleteEmailActions(ctx, toDelete, monitorID)
	if err != nil {
		return nil, err
	}
	err = tx.createActions(ctx, monitorID, toCreate)
	if err != nil {
		return nil, err
	}
	m, err := tx.updateCodeMonitor(ctx, args)
	if err != nil {
		return nil, err
	}
	// Hydrate monitor with Resolver.
	m.(*monitor).Resolver = r
	return m, nil
}

func (r *Resolver) createActions(ctx context.Context, monitorID int64, args []*graphqlbackend.CreateActionArgs) error {
	for _, a := range args {
		if a.Email != nil {
			e, err := r.store.CreateEmailAction(ctx, monitorID, &cm.EmailActionArgs{
				Enabled:  a.Email.Enabled,
				Priority: a.Email.Priority,
				Header:   a.Email.Header,
			})
			if err != nil {
				return err
			}

			if err := r.createRecipients(ctx, e.ID, a.Email.Recipients); err != nil {
				return err
			}
		}
		// TODO(camdencheek): add other action types (webhooks) here
	}
	return nil
}

func (r *Resolver) createRecipients(ctx context.Context, emailID int64, recipients []graphql.ID) error {
	for _, recipient := range recipients {
		userID, orgID, err := graphqlbackend.UnmarshalNamespaceToIDs(recipient)
		if err != nil {
			return errors.Wrap(err, "UnmarshalNamespaceID")
		}

		_, err = r.store.CreateRecipient(ctx, emailID, userID, orgID)
		if err != nil {
			return err
		}
	}
	return nil
}

// ResetTriggerQueryTimestamps is a convenience function which resets the
// timestamps `next_run` and `last_result` with the purpose to trigger associated
// actions (emails, webhooks) immediately. This is useful during development and
// troubleshooting. Only site admins can call this functions.
func (r *Resolver) ResetTriggerQueryTimestamps(ctx context.Context, args *graphqlbackend.ResetTriggerQueryTimestampsArgs) (*graphqlbackend.EmptyResponse, error) {
	err := backend.CheckCurrentUserIsSiteAdmin(ctx, database.NewDB(r.store.Handle().DB()))
	if err != nil {
		return nil, err
	}
	var queryIDInt64 int64
	err = relay.UnmarshalSpec(args.Id, &queryIDInt64)
	if err != nil {
		return nil, err
	}
	err = r.store.ResetQueryTriggerTimestamps(ctx, queryIDInt64)
	if err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) TriggerTestEmailAction(ctx context.Context, args *graphqlbackend.TriggerTestEmailActionArgs) (*graphqlbackend.EmptyResponse, error) {
	err := r.isAllowedToCreate(ctx, args.Namespace)
	if err != nil {
		return nil, err
	}

	for _, recipient := range args.Email.Recipients {
		if err := sendTestEmail(ctx, recipient, args.Description); err != nil {
			return nil, err
		}
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func sendTestEmail(ctx context.Context, recipient graphql.ID, description string) error {
	var (
		userID int32
		orgID  int32
	)
	err := graphqlbackend.UnmarshalNamespaceID(recipient, &userID, &orgID)
	if err != nil {
		return err
	}
	// TODO: Send test email to org members.
	if orgID != 0 {
		return nil
	}
	data := background.NewTestTemplateDataForNewSearchResults(ctx, description)
	return background.SendEmailForNewSearchResult(ctx, userID, data)
}

func (r *Resolver) actionIDsForMonitorIDInt64(ctx context.Context, monitorID int64) ([]graphql.ID, error) {
	emailActions, err := r.store.ListEmailActions(ctx, cm.ListActionsOpts{
		MonitorID: &monitorID,
	})
	if err != nil {
		return nil, err
	}
	ids := make([]graphql.ID, len(emailActions))
	for i, emailAction := range emailActions {
		ids[i] = (&monitorEmail{EmailAction: emailAction}).ID()
	}
	return ids, nil
}

// splitActionIDs splits actions into three buckets: create, delete and update.
// Note: args is mutated. After splitActionIDs, args only contains actions to be updated.
func splitActionIDs(ctx context.Context, args *graphqlbackend.UpdateCodeMonitorArgs, actionIDs []graphql.ID) (toCreate []*graphqlbackend.CreateActionArgs, toDelete []int64, err error) {
	aMap := make(map[graphql.ID]struct{}, len(actionIDs))
	for _, id := range actionIDs {
		aMap[id] = struct{}{}
	}
	var toUpdateActions []*graphqlbackend.EditActionArgs
	for _, a := range args.Actions {
		if a.Email.Id == nil {
			toCreate = append(toCreate, &graphqlbackend.CreateActionArgs{Email: a.Email.Update})
			continue
		}
		if _, ok := aMap[*a.Email.Id]; !ok {
			return nil, nil, errors.Errorf("unknown ID=%s for action", *a.Email.Id)
		}
		toUpdateActions = append(toUpdateActions, a)
		delete(aMap, *a.Email.Id)
	}
	var actionID int64
	for k := range aMap {
		err = relay.UnmarshalSpec(k, &actionID)
		if err != nil {
			return nil, nil, err
		}
		toDelete = append(toDelete, actionID)
	}
	args.Actions = toUpdateActions
	return toCreate, toDelete, nil
}

func (r *Resolver) updateCodeMonitor(ctx context.Context, args *graphqlbackend.UpdateCodeMonitorArgs) (graphqlbackend.MonitorResolver, error) {
	// Update monitor.
	var monitorID int64
	if err := relay.UnmarshalSpec(args.Monitor.Id, &monitorID); err != nil {
		return nil, err
	}

	userID, orgID, err := graphqlbackend.UnmarshalNamespaceToIDs(args.Monitor.Update.Namespace)
	if err != nil {
		return nil, err
	}

	mo, err := r.store.UpdateMonitor(ctx, monitorID, cm.MonitorArgs{
		Description:     args.Monitor.Update.Description,
		Enabled:         args.Monitor.Update.Enabled,
		NamespaceUserID: userID,
		NamespaceOrgID:  orgID,
	})
	if err != nil {
		return nil, err
	}

	var triggerID int64
	if err := relay.UnmarshalSpec(args.Trigger.Id, &triggerID); err != nil {
		return nil, err
	}

	// Update trigger.
	err = r.store.UpdateQueryTrigger(ctx, triggerID, args.Trigger.Update.Query)
	if err != nil {
		return nil, err
	}
	// Update actions.
	if len(args.Actions) == 0 {
		return &monitor{
			Resolver: r,
			Monitor:  mo,
		}, nil
	}
	for i, action := range args.Actions {
		if action.Email == nil {
			return nil, errors.Errorf("missing email object for action %d", i)
		}
		var emailID int64
		err = relay.UnmarshalSpec(*action.Email.Id, &emailID)
		if err != nil {
			return nil, err
		}
		err = r.store.DeleteRecipients(ctx, emailID)
		if err != nil {
			return nil, err
		}

		e, err := r.store.UpdateEmailAction(ctx, emailID, &cm.EmailActionArgs{
			Enabled:  action.Email.Update.Enabled,
			Priority: action.Email.Update.Priority,
			Header:   action.Email.Update.Header,
		})
		if err != nil {
			return nil, err
		}
		err = r.createRecipients(ctx, e.ID, action.Email.Update.Recipients)
		if err != nil {
			return nil, err
		}
	}
	return &monitor{
		Resolver: r,
		Monitor:  mo,
	}, nil
}

func (r *Resolver) transact(ctx context.Context) (*Resolver, error) {
	txStore, err := r.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &Resolver{
		store: txStore,
	}, nil
}

// isAllowedToEdit checks whether an actor is allowed to edit a given monitor.
func (r *Resolver) isAllowedToEdit(ctx context.Context, monitorID graphql.ID) error {
	var monitorIDInt64 int64
	err := relay.UnmarshalSpec(monitorID, &monitorIDInt64)
	if err != nil {
		return err
	}
	owner, err := r.ownerForID64(ctx, monitorIDInt64)
	if err != nil {
		return err
	}
	return r.isAllowedToCreate(ctx, owner)
}

// isAllowedToCreate compares the owner of a monitor (user or org) to the actor of
// the request. A user can create a monitor if either of the following statements
// is true:
// - she is the owner
// - she is a member of the organization which is the owner of the monitor
// - she is a site-admin
func (r *Resolver) isAllowedToCreate(ctx context.Context, owner graphql.ID) error {
	var ownerInt32 int32
	err := relay.UnmarshalSpec(owner, &ownerInt32)
	if err != nil {
		return err
	}
	switch kind := relay.UnmarshalKind(owner); kind {
	case "User":
		return backend.CheckSiteAdminOrSameUser(ctx, database.NewDB(r.store.Handle().DB()), ownerInt32)
	case "Org":
		return errors.Errorf("creating a code monitor with an org namespace is no longer supported")
	default:
		return errors.Errorf("provided ID is not a namespace")
	}
}

func (r *Resolver) ownerForID64(ctx context.Context, monitorID int64) (graphql.ID, error) {
	monitor, err := r.store.GetMonitor(ctx, monitorID)
	if err != nil {
		return "", err
	}

	return graphqlbackend.MarshalUserID(monitor.UserID), nil
}

//
// MonitorConnection
//
type monitorConnection struct {
	*Resolver
	monitors    []graphqlbackend.MonitorResolver
	totalCount  int32
	hasNextPage bool
}

func (m *monitorConnection) Nodes() []graphqlbackend.MonitorResolver {
	return m.monitors
}

func (m *monitorConnection) TotalCount() int32 {
	return m.totalCount
}

func (m *monitorConnection) PageInfo() *graphqlutil.PageInfo {
	if len(m.monitors) == 0 || !m.hasNextPage {
		return graphqlutil.HasNextPage(false)
	}
	return graphqlutil.NextPageCursor(string(m.monitors[len(m.monitors)-1].ID()))
}

//
// Monitor
//
type monitor struct {
	*Resolver
	*cm.Monitor
}

const (
	MonitorKind                     = "CodeMonitor"
	monitorTriggerQueryKind         = "CodeMonitorTriggerQuery"
	monitorTriggerEventKind         = "CodeMonitorTriggerEvent"
	monitorActionEmailKind          = "CodeMonitorActionEmail"
	monitorActionEventKind          = "CodeMonitorActionEmailEvent"
	monitorActionEmailRecipientKind = "CodeMonitorActionEmailRecipient"
)

func (m *monitor) ID() graphql.ID {
	return relay.MarshalID(MonitorKind, m.Monitor.ID)
}

func (m *monitor) CreatedBy(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	return graphqlbackend.UserByIDInt32(ctx, database.NewDB(m.store.Handle().DB()), m.Monitor.CreatedBy)
}

func (m *monitor) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: m.Monitor.CreatedAt}
}

func (m *monitor) Description() string {
	return m.Monitor.Description
}

func (m *monitor) Enabled() bool {
	return m.Monitor.Enabled
}

func (m *monitor) Owner(ctx context.Context) (graphqlbackend.NamespaceResolver, error) {
	n, err := graphqlbackend.UserByIDInt32(ctx, database.NewDB(m.store.Handle().DB()), m.UserID)
	return graphqlbackend.NamespaceResolver{Namespace: n}, err
}

func (m *monitor) Trigger(ctx context.Context) (graphqlbackend.MonitorTrigger, error) {
	t, err := m.store.GetQueryTriggerForMonitor(ctx, m.Monitor.ID)
	if err != nil {
		return nil, err
	}
	return &monitorTrigger{&monitorQuery{m.Resolver, t}}, nil
}

func (m *monitor) Actions(ctx context.Context, args *graphqlbackend.ListActionArgs) (graphqlbackend.MonitorActionConnectionResolver, error) {
	return m.actionConnectionResolverWithTriggerID(ctx, nil, m.Monitor.ID, args)
}

func (r *Resolver) actionConnectionResolverWithTriggerID(ctx context.Context, triggerEventID *int32, monitorID int64, args *graphqlbackend.ListActionArgs) (graphqlbackend.MonitorActionConnectionResolver, error) {
	after, err := unmarshalAfter(args.After)
	if err != nil {
		return nil, err
	}
	// For now, we only support emails as actions. Once we add other actions such as
	// webhooks, we have to query those tables here too.
	es, err := r.store.ListEmailActions(ctx, cm.ListActionsOpts{
		MonitorID: &monitorID,
		After:     after,
		First:     intPtr(int(args.First)),
	})
	if err != nil {
		return nil, err
	}

	totalCount, err := r.store.CountEmailActions(ctx, monitorID)
	if err != nil {
		return nil, err
	}
	actions := make([]graphqlbackend.MonitorAction, 0, len(es))
	for _, e := range es {
		actions = append(actions, &action{
			email: &monitorEmail{
				Resolver:       r,
				EmailAction:    e,
				triggerEventID: triggerEventID,
			},
		})
	}
	return &monitorActionConnection{actions: actions, totalCount: totalCount}, nil
}

//
// MonitorTrigger <<UNION>>
//
type monitorTrigger struct {
	query graphqlbackend.MonitorQueryResolver
}

func (t *monitorTrigger) ToMonitorQuery() (graphqlbackend.MonitorQueryResolver, bool) {
	return t.query, t.query != nil
}

//
// Query
//
type monitorQuery struct {
	*Resolver
	*cm.QueryTrigger
}

func (q *monitorQuery) ID() graphql.ID {
	return relay.MarshalID(monitorTriggerQueryKind, q.QueryTrigger.ID)
}

func (q *monitorQuery) Query() string {
	return q.QueryString
}

func (q *monitorQuery) Events(ctx context.Context, args *graphqlbackend.ListEventsArgs) (graphqlbackend.MonitorTriggerEventConnectionResolver, error) {
	after, err := unmarshalAfter(args.After)
	if err != nil {
		return nil, err
	}
	es, err := q.store.ListQueryTriggerJobs(ctx, cm.ListTriggerJobsOpts{
		QueryID: &q.QueryTrigger.ID,
		First:   intPtr(int(args.First)),
		After:   intPtrToInt64Ptr(after),
	})
	if err != nil {
		return nil, err
	}
	totalCount, err := q.store.CountQueryTriggerJobs(ctx, q.QueryTrigger.ID)
	if err != nil {
		return nil, err
	}
	events := make([]graphqlbackend.MonitorTriggerEventResolver, 0, len(es))
	for _, e := range es {
		events = append(events, graphqlbackend.MonitorTriggerEventResolver(&monitorTriggerEvent{
			Resolver:   q.Resolver,
			monitorID:  q.Monitor,
			TriggerJob: e,
		}))
	}
	return &monitorTriggerEventConnection{Resolver: q.Resolver, events: events, totalCount: totalCount}, nil
}

//
// MonitorTriggerEventConnection
//
type monitorTriggerEventConnection struct {
	*Resolver
	events     []graphqlbackend.MonitorTriggerEventResolver
	totalCount int32
}

func (a *monitorTriggerEventConnection) Nodes() []graphqlbackend.MonitorTriggerEventResolver {
	return a.events
}

func (a *monitorTriggerEventConnection) TotalCount() int32 {
	return a.totalCount
}

func (a *monitorTriggerEventConnection) PageInfo() *graphqlutil.PageInfo {
	if len(a.events) == 0 {
		return graphqlutil.HasNextPage(false)
	}
	return graphqlutil.NextPageCursor(string(a.events[len(a.events)-1].ID()))
}

//
// MonitorTriggerEvent
//
type monitorTriggerEvent struct {
	*Resolver
	*cm.TriggerJob
	monitorID int64
}

func (m *monitorTriggerEvent) ID() graphql.ID {
	return relay.MarshalID(monitorTriggerEventKind, m.TriggerJob.ID)
}

// stateToStatus maps the state of the dbworker job to the public GraphQL status of
// events.
var stateToStatus = map[string]string{
	"completed":  "SUCCESS",
	"queued":     "PENDING",
	"processing": "PENDING",
	"errored":    "ERROR",
	"failed":     "ERROR",
}

func (m *monitorTriggerEvent) Status() (string, error) {
	if v, ok := stateToStatus[m.State]; ok {
		return v, nil
	}
	return "", errors.Errorf("unknown status: %s", m.State)
}

func (m *monitorTriggerEvent) Message() *string {
	return m.FailureMessage
}

func (m *monitorTriggerEvent) Timestamp() (graphqlbackend.DateTime, error) {
	if m.FinishedAt == nil {
		return graphqlbackend.DateTime{Time: m.store.Now()}, nil
	}
	return graphqlbackend.DateTime{Time: *m.FinishedAt}, nil
}

func (m *monitorTriggerEvent) Actions(ctx context.Context, args *graphqlbackend.ListActionArgs) (graphqlbackend.MonitorActionConnectionResolver, error) {
	return m.actionConnectionResolverWithTriggerID(ctx, &m.TriggerJob.ID, m.monitorID, args)
}

// ActionConnection
//
type monitorActionConnection struct {
	actions    []graphqlbackend.MonitorAction
	totalCount int32
}

func (a *monitorActionConnection) Nodes() []graphqlbackend.MonitorAction {
	return a.actions
}

func (a *monitorActionConnection) TotalCount() int32 {
	return a.totalCount
}

func (a *monitorActionConnection) PageInfo() *graphqlutil.PageInfo {
	if len(a.actions) == 0 {
		return graphqlutil.HasNextPage(false)
	}
	last := a.actions[len(a.actions)-1]
	if email, ok := last.ToMonitorEmail(); ok {
		return graphqlutil.NextPageCursor(string(email.ID()))
	}
	panic("found non-email monitor action")
}

//
// Action <<UNION>>
//
type action struct {
	email graphqlbackend.MonitorEmailResolver
}

func (a *action) ToMonitorEmail() (graphqlbackend.MonitorEmailResolver, bool) {
	return a.email, a.email != nil
}

//
// Email
//
type monitorEmail struct {
	*Resolver
	*cm.EmailAction

	// If triggerEventID == nil, all events of this action will be returned.
	// Otherwise, only those events of this action which are related to the specified
	// trigger event will be returned.
	triggerEventID *int32
}

func (m *monitorEmail) Recipients(ctx context.Context, args *graphqlbackend.ListRecipientsArgs) (graphqlbackend.MonitorActionEmailRecipientsConnectionResolver, error) {
	after, err := unmarshalAfter(args.After)
	if err != nil {
		return nil, err
	}
	ms, err := m.store.ListRecipients(ctx, cm.ListRecipientsOpts{
		EmailID: &m.EmailAction.ID,
		First:   intPtr(int(args.First)),
		After:   intPtrToInt64Ptr(after),
	})
	if err != nil {
		return nil, err
	}
	ns := make([]graphqlbackend.NamespaceResolver, 0, len(ms))
	for _, r := range ms {
		n := graphqlbackend.NamespaceResolver{}
		if r.NamespaceOrgID == nil {
			n.Namespace, err = graphqlbackend.UserByIDInt32(ctx, database.NewDB(m.store.Handle().DB()), *r.NamespaceUserID)
		} else {
			n.Namespace, err = graphqlbackend.OrgByIDInt32(ctx, database.NewDB(m.store.Handle().DB()), *r.NamespaceOrgID)
		}
		if err != nil {
			return nil, err
		}
		ns = append(ns, n)
	}

	// Since recipients can either be a user or an org it would be very tedious to
	// use the user-id or org-id of the last entry as a cursor for the next page. It
	// is easier to just use the id of the recipients table.
	var nextPageCursor string
	if len(ms) > 0 {
		nextPageCursor = string(relay.MarshalID(monitorActionEmailRecipientKind, ms[len(ms)-1].ID))
	}

	total, err := m.store.CountRecipients(ctx, m.EmailAction.ID)
	if err != nil {
		return nil, err
	}
	return &monitorActionEmailRecipientsConnection{ns, nextPageCursor, total}, nil
}

func (m *monitorEmail) Enabled() bool {
	return m.EmailAction.Enabled
}

func (m *monitorEmail) Priority() string {
	return m.EmailAction.Priority
}

func (m *monitorEmail) Header() string {
	return m.EmailAction.Header
}

func (m *monitorEmail) ID() graphql.ID {
	return relay.MarshalID(monitorActionEmailKind, m.EmailAction.ID)
}

func (m *monitorEmail) Events(ctx context.Context, args *graphqlbackend.ListEventsArgs) (graphqlbackend.MonitorActionEventConnectionResolver, error) {
	after, err := unmarshalAfter(args.After)
	if err != nil {
		return nil, err
	}

	ajs, err := m.store.ListActionJobs(ctx, cm.ListActionJobsOpts{
		EmailID:        intPtr(int(m.EmailAction.ID)),
		TriggerEventID: m.triggerEventID,
		First:          intPtr(int(args.First)),
		After:          after,
	})
	if err != nil {
		return nil, err
	}

	totalCount, err := m.store.CountActionJobs(ctx, cm.ListActionJobsOpts{
		EmailID:        intPtr(int(m.EmailAction.ID)),
		TriggerEventID: m.triggerEventID,
	})
	if err != nil {
		return nil, err
	}
	events := make([]graphqlbackend.MonitorActionEventResolver, len(ajs))
	for i, aj := range ajs {
		events[i] = &monitorActionEvent{Resolver: m.Resolver, ActionJob: aj}
	}
	return &monitorActionEventConnection{events: events, totalCount: int32(totalCount)}, nil
}

func intPtr(i int) *int { return &i }
func intPtrToInt64Ptr(i *int) *int64 {
	if i == nil {
		return nil
	}
	j := int64(*i)
	return &j
}

func unmarshalAfter(after *string) (*int, error) {
	if after == nil {
		return nil, nil
	}

	var a int
	err := relay.UnmarshalSpec(graphql.ID(*after), &a)
	return &a, err
}

//
// MonitorActionEmailRecipientConnection
//
type monitorActionEmailRecipientsConnection struct {
	recipients     []graphqlbackend.NamespaceResolver
	nextPageCursor string
	totalCount     int32
}

func (a *monitorActionEmailRecipientsConnection) Nodes() []graphqlbackend.NamespaceResolver {
	return a.recipients
}

func (a *monitorActionEmailRecipientsConnection) TotalCount() int32 {
	return a.totalCount
}

func (a *monitorActionEmailRecipientsConnection) PageInfo() *graphqlutil.PageInfo {
	if len(a.recipients) == 0 {
		return graphqlutil.HasNextPage(false)
	}
	return graphqlutil.NextPageCursor(a.nextPageCursor)
}

//
// MonitorActionEventConnection
//
type monitorActionEventConnection struct {
	events     []graphqlbackend.MonitorActionEventResolver
	totalCount int32
}

func (a *monitorActionEventConnection) Nodes() []graphqlbackend.MonitorActionEventResolver {
	return a.events
}

func (a *monitorActionEventConnection) TotalCount() int32 {
	return a.totalCount
}

func (a *monitorActionEventConnection) PageInfo() *graphqlutil.PageInfo {
	if len(a.events) == 0 {
		return graphqlutil.HasNextPage(false)
	}
	return graphqlutil.NextPageCursor(string(a.events[len(a.events)-1].ID()))
}

//
// MonitorEvent
//
type monitorActionEvent struct {
	*Resolver
	*cm.ActionJob
}

func (m *monitorActionEvent) ID() graphql.ID {
	return relay.MarshalID(monitorActionEventKind, m.ActionJob.ID)
}

func (m *monitorActionEvent) Status() (string, error) {
	status, ok := stateToStatus[m.State]
	if !ok {
		return "", errors.Errorf("unknown state: %s", m.State)
	}
	return status, nil
}

func (m *monitorActionEvent) Message() *string {
	return m.FailureMessage
}

func (m *monitorActionEvent) Timestamp() graphqlbackend.DateTime {
	if m.FinishedAt == nil {
		return graphqlbackend.DateTime{Time: m.store.Now()}
	}
	return graphqlbackend.DateTime{Time: *m.FinishedAt}
}
