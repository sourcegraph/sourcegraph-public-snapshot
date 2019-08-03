package events

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/types"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// dbEvent describes a event.
type dbEvent struct {
	ID              int64
	NamespaceUserID int32  // the user namespace where this event is defined
	NamespaceOrgID  int32  // the org namespace where this event is defined
	Name            string // the name (case-preserving)
	IsPreview       bool
	Rules           string // the JSON rules TODO!(sqs)

	PrimaryCommentID int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// errEventNotFound occurs when a database operation expects a specific event to exist but it
// does not exist.
var errEventNotFound = errors.New("event not found")

type dbEvents struct{}

const selectColumns = `id, namespace_user_id, namespace_org_id, name, is_preview, rules, primary_comment_id, created_at, updated_at`

// Create creates a event. The event argument's (Event).ID field is ignored. The new
// event is returned.
func (dbEvents) Create(ctx context.Context, event *dbEvent, comment commentobjectdb.DBObjectCommentFields) (*dbEvent, error) {
	if mocks.events.Create != nil {
		return mocks.events.Create(event)
	}

	if event.PrimaryCommentID != 0 {
		panic("event.PrimaryCommentID must not be set")
	}

	nilIfZero := func(v int32) *int32 {
		if v == 0 {
			return nil
		}
		return &v
	}

	return event, commentobjectdb.CreateCommentWithObject(ctx, comment, func(ctx context.Context, tx *sql.Tx, commentID int64) (*types.CommentObject, error) {
		var err error
		event, err = dbEvents{}.scanRow(tx.QueryRowContext(ctx,
			`INSERT INTO events(`+selectColumns+`) VALUES(DEFAULT, $1, $2, $3, $4, $5, $6, DEFAULT, DEFAULT) RETURNING `+selectColumns,
			nilIfZero(event.NamespaceUserID),
			nilIfZero(event.NamespaceOrgID),
			event.Name,
			event.IsPreview,
			event.Rules,
			commentID,
		))
		if err != nil {
			return nil, err
		}
		return &types.CommentObject{EventID: event.ID}, nil
	})
}

type dbEventUpdate struct {
	Name      *string
	IsPreview *bool
	Rules     *string
}

// Update updates a event given its ID.
func (s dbEvents) Update(ctx context.Context, id int64, update dbEventUpdate) (*dbEvent, error) {
	if mocks.events.Update != nil {
		return mocks.events.Update(id, update)
	}

	var setFields []*sqlf.Query
	if update.Name != nil {
		setFields = append(setFields, sqlf.Sprintf("name=%s", *update.Name))
	}
	if update.IsPreview != nil {
		setFields = append(setFields, sqlf.Sprintf("is_preview=%s", *update.IsPreview))
	}
	if update.Rules != nil {
		setFields = append(setFields, sqlf.Sprintf("rules=%s", *update.Rules))
	}

	if len(setFields) == 0 {
		return nil, nil
	}
	setFields = append(setFields, sqlf.Sprintf("updated_at=now()"))

	results, err := s.query(ctx, sqlf.Sprintf(`UPDATE events SET %v WHERE id=%s RETURNING `+selectColumns, sqlf.Join(setFields, ", "), id))
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errEventNotFound
	}
	return results[0], nil
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
	Query           string // only list events matching this query (case-insensitively)
	NamespaceUserID int32  // only list events in this user's namespace
	NamespaceOrgID  int32  // only list events in this org's namespace
	ObjectThreadID  int64
	*db.LimitOffset
}

func (o dbEventsListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.Query != "" {
		conds = append(conds, sqlf.Sprintf("name ILIKE %s", "%"+o.Query+"%"))
	}
	if o.NamespaceUserID != 0 {
		conds = append(conds, sqlf.Sprintf("namespace_user_id=%d", o.NamespaceUserID))
	}
	if o.NamespaceOrgID != 0 {
		conds = append(conds, sqlf.Sprintf("namespace_org_id=%d", o.NamespaceOrgID))
	}
	if o.ObjectThreadID != 0 {
		conds = append(conds, sqlf.Sprintf("id IN (SELECT DISTINCT event_id FROM events_threads WHERE thread_id=%d)", o.ObjectThreadID))
	}
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
ORDER BY name ASC
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
	var namespaceUserID, namespaceOrgID *int32
	if err := row.Scan(
		&t.ID,
		&namespaceUserID,
		&namespaceOrgID,
		&t.Name,
		&t.IsPreview,
		&t.Rules,
		&t.PrimaryCommentID,
		&t.CreatedAt,
		&t.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if namespaceUserID != nil {
		t.NamespaceUserID = *namespaceUserID
	}
	if namespaceOrgID != nil {
		t.NamespaceOrgID = *namespaceOrgID
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

// Delete deletes a event given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to delete the event.
func (s dbEvents) DeleteByID(ctx context.Context, id int64) error {
	if mocks.events.DeleteByID != nil {
		return mocks.events.DeleteByID(id)
	}
	return s.delete(ctx, sqlf.Sprintf("id=%d", id))
}

func (dbEvents) delete(ctx context.Context, cond *sqlf.Query) error {
	conds := []*sqlf.Query{cond, sqlf.Sprintf("TRUE")}
	q := sqlf.Sprintf("DELETE FROM events WHERE (%s)", sqlf.Join(conds, ") AND ("))

	res, err := dbconn.Global.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errEventNotFound
	}
	return nil
}

// mockEvents mocks the events-related DB operations.
type mockEvents struct {
	Create     func(*dbEvent) (*dbEvent, error)
	Update     func(int64, dbEventUpdate) (*dbEvent, error)
	GetByID    func(int64) (*dbEvent, error)
	List       func(dbEventsListOptions) ([]*dbEvent, error)
	Count      func(dbEventsListOptions) (int, error)
	DeleteByID func(int64) error
}
