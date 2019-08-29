package events

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/nnz"
)

// dbEvent describes a event.
type dbEvent struct {
	ID int64
	Type
	ActorUserID           int32
	ExternalActorUsername string
	ExternalActorURL      string
	CreatedAt             time.Time
	Objects
	Data                          []byte
	ImportedFromExternalServiceID int64
}

// errEventNotFound occurs when a database operation expects a specific event to exist but it
// does not exist.
var errEventNotFound = errors.New("event not found")

type dbEvents struct{}

const selectColumns = `id, type, actor_user_id, external_actor_username, external_actor_url, created_at, data, thread_id, thread_diagnostic_edge_id, campaign_id, comment_id, rule_id, repository_id, user_id, organization_id, registry_extension_id, imported_from_external_service_id`

// Create creates a event. The event argument's (Event).ID field is ignored. The new event is
// returned.
func (dbEvents) Create(ctx context.Context, tx *sql.Tx, event *dbEvent) (*dbEvent, error) {
	if mocks.events.Create != nil {
		return mocks.events.Create(event)
	}

	if event.ID != 0 {
		panic("event.ID must not be set")
	}
	var data *[]byte
	if event.Data != nil {
		data = &event.Data
	}
	args := []interface{}{
		event.Type,
		nnz.Int32(event.ActorUserID),
		nnz.String(event.ExternalActorUsername),
		nnz.String(event.ExternalActorURL),
		event.CreatedAt,
		data,
		nnz.Int64(event.Objects.Thread),
		nnz.Int64(event.Objects.ThreadDiagnosticEdge),
		nnz.Int64(event.Objects.Campaign),
		nnz.Int64(event.Objects.Comment),
		nnz.Int64(event.Objects.Rule),
		nnz.Int32(event.Objects.Repository),
		nnz.Int32(event.Objects.User),
		nnz.Int32(event.Objects.Organization),
		nnz.Int32(event.Objects.RegistryExtension),
		nnz.Int64(event.ImportedFromExternalServiceID),
	}
	query := sqlf.Sprintf(
		`INSERT INTO events(`+selectColumns+`) VALUES(DEFAULT`+strings.Repeat(", %v", len(args))+`) RETURNING `+selectColumns,
		args...,
	)

	return dbEvents{}.scanRow(dbconn.TxOrGlobal(tx).QueryRowContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...))
}

// GetByID retrieves the event (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this event.
func (s dbEvents) GetByID(ctx context.Context, id int64) (*dbEvent, error) {
	if mocks.events.GetByID != nil {
		return mocks.events.GetByID(id)
	}

	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errEventNotFound
	}
	return results[0], nil
}

// dbEventsListOptions contains options for listing events.
type dbEventsListOptions struct {
	AfterDate  time.Time
	BeforeDate time.Time
	Types      []Type
	Objects
	ImportedFromExternalServiceID int64
	*db.LimitOffset
}

func (o dbEventsListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if !o.AfterDate.IsZero() {
		conds = append(conds, sqlf.Sprintf("created_at>=%v", o.AfterDate))
	}
	if !o.BeforeDate.IsZero() {
		conds = append(conds, sqlf.Sprintf("created_at<=%v", o.BeforeDate))
	}
	if o.Types != nil {
		conds = append(conds, sqlf.Sprintf("type = ANY(%v)", o.Types))
	}
	addCondition := func(id int64, column string) {
		if id != 0 {
			conds = append(conds, sqlf.Sprintf(column+"=%d", id))
		}
	}
	addCondition(o.Objects.Thread, "thread_id")
	addCondition(o.Objects.ThreadDiagnosticEdge, "thread_diagnostic_edge_id")
	if o.Objects.Campaign != 0 {
		conds = append(conds, sqlf.Sprintf("campaign_id=%d OR thread_id IN (SELECT thread_id FROM campaigns_threads WHERE campaign_id=%d) OR comment_id IN (SELECT id FROM comments WHERE parent_comment_id IN (SELECT id FROM comments WHERE thread_id IN (SELECT thread_id FROM campaigns_threads WHERE campaign_id=%d)))", o.Objects.Campaign, o.Objects.Campaign, o.Objects.Campaign))
	}
	addCondition(o.Objects.Comment, "comment_id")
	addCondition(o.Objects.Rule, "rule_id")
	addCondition(int64(o.Objects.Repository), "repository_id")
	addCondition(int64(o.Objects.User), "user_id")
	addCondition(int64(o.Objects.Organization), "organization_id")
	addCondition(int64(o.Objects.RegistryExtension), "registry_extension_id")
	addCondition(o.ImportedFromExternalServiceID, "imported_from_external_service_id")
	return conds
}

// List lists all events that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s dbEvents) List(ctx context.Context, opt dbEventsListOptions) ([]*dbEvent, error) {
	if mocks.events.List != nil {
		return mocks.events.List(opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s dbEvents) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*dbEvent, error) {
	q := sqlf.Sprintf(`
SELECT `+selectColumns+` FROM events
WHERE (%s)
ORDER BY created_at ASC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)
	return s.query(ctx, q)
}

func (dbEvents) query(ctx context.Context, query *sqlf.Query) ([]*dbEvent, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*dbEvent
	for rows.Next() {
		t, err := dbEvents{}.scanRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, nil
}

func (dbEvents) scanRow(row interface {
	Scan(dest ...interface{}) error
}) (*dbEvent, error) {
	var t dbEvent
	if err := row.Scan(
		&t.ID,
		&t.Type,
		nnz.ToInt32(&t.ActorUserID),
		(*nnz.String)(&t.ExternalActorUsername),
		(*nnz.String)(&t.ExternalActorURL),
		&t.CreatedAt,
		&t.Data,
		(*nnz.Int64)(&t.Objects.Thread),
		(*nnz.Int64)(&t.Objects.ThreadDiagnosticEdge),
		(*nnz.Int64)(&t.Objects.Campaign),
		(*nnz.Int64)(&t.Objects.Comment),
		(*nnz.Int64)(&t.Objects.Rule),
		nnz.ToInt32(&t.Objects.Repository),
		nnz.ToInt32(&t.Objects.User),
		nnz.ToInt32(&t.Objects.Organization),
		nnz.ToInt32(&t.Objects.RegistryExtension),
		(*nnz.Int64)(&t.ImportedFromExternalServiceID),
	); err != nil {
		return nil, err
	}
	return &t, nil
}

// Count counts all events that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the events.
func (dbEvents) Count(ctx context.Context, opt dbEventsListOptions) (int, error) {
	if mocks.events.Count != nil {
		return mocks.events.Count(opt)
	}

	q := sqlf.Sprintf("SELECT COUNT(*) FROM events WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Delete deletes all events matching the criteria (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to delete the events.
func (dbEvents) Delete(ctx context.Context, tx *sql.Tx, opt dbEventsListOptions) error {
	if mocks.events.Delete != nil {
		return mocks.events.Delete(opt)
	}

	query := sqlf.Sprintf(`DELETE FROM events WHERE (%s)`, sqlf.Join(opt.sqlConditions(), ") AND ("))
	_, err := dbconn.TxOrGlobal(tx).ExecContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	return err
}

// mockEvents mocks the events-related DB operations.
type mockEvents struct {
	Create  func(*dbEvent) (*dbEvent, error)
	GetByID func(int64) (*dbEvent, error)
	List    func(dbEventsListOptions) ([]*dbEvent, error)
	Count   func(dbEventsListOptions) (int, error)
	Delete  func(dbEventsListOptions) error
}

type dbMocks struct {
	events mockEvents
}

var mocks dbMocks

func resetMocks() {
	mocks = dbMocks{}
}
