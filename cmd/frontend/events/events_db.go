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
)

// dbEvent describes a event.
type dbEvent struct {
	ID int64
	Type
	ActorUserID           int32
	ExternalActorUsername sql.NullString
	ExternalActorURL      sql.NullString
	CreatedAt             time.Time
	Objects
	Data                          []byte
	ImportedFromExternalServiceID int64
}

// errEventNotFound occurs when a database operation expects a specific event to exist but it
// does not exist.
var errEventNotFound = errors.New("event not found")

type dbEvents struct{}

const selectColumns = `id, type, actor_user_id, external_actor_username, external_actor_url, created_at, data, thread_id, campaign_id, comment_id, rule_id, repository_id, user_id, organization_id, registry_extension_id, imported_from_external_service_id`

// Create creates a event. The event argument's (Event).ID field is ignored. The new event is
// returned.
func (dbEvents) Create(ctx context.Context, tx *sql.Tx, event *dbEvent) (*dbEvent, error) {
	if mocks.events.Create != nil {
		return mocks.events.Create(event)
	}

	if event.ID != 0 {
		panic("event.ID must not be set")
	}
	nilIfZero32 := func(v int32) *int32 {
		if v == 0 {
			return nil
		}
		return &v
	}
	nilIfZero64 := func(v int64) *int64 {
		if v == 0 {
			return nil
		}
		return &v
	}
	var data *[]byte
	if event.Data != nil {
		data = &event.Data
	}
	args := []interface{}{
		event.Type,
		nilIfZero32(event.ActorUserID),
		event.ExternalActorUsername,
		event.ExternalActorURL,
		event.CreatedAt,
		data,
		nilIfZero64(event.Objects.Thread),
		nilIfZero64(event.Objects.Campaign),
		nilIfZero64(event.Objects.Comment),
		nilIfZero64(event.Objects.Rule),
		nilIfZero32(event.Objects.Repository),
		nilIfZero32(event.Objects.User),
		nilIfZero32(event.Objects.Organization),
		nilIfZero32(event.Objects.RegistryExtension),
		nilIfZero64(event.ImportedFromExternalServiceID),
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
	if o.Objects.Campaign != 0 {
		conds = append(conds, sqlf.Sprintf("campaign_id=%d OR thread_id IN (SELECT thread_id FROM campaigns_threads WHERE campaign_id=%d)", o.Objects.Campaign, o.Objects.Campaign))
	}
	addCondition(o.Objects.Comment, "comment_id")
	addCondition(o.Objects.Rule, "rule_id")
	addCondition(int64(o.Objects.Repository), "repository_id")
	addCondition(int64(o.Objects.User), "user_id")
	addCondition(int64(o.Objects.Organization), "organization_id")
	addCondition(int64(o.Objects.RegistryExtension), "registry_extension_id")
	addCondition(int64(o.ImportedFromExternalServiceID), "imported_from_external_service_id")
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
	var actorUserID *int32
	var threadID, campaignID, commentID, ruleID, importedFromExternalServiceID *int64
	var repositoryID, userID, organizationID, registryExtensionID *int32
	if err := row.Scan(
		&t.ID,
		&t.Type,
		&actorUserID,
		&t.ExternalActorUsername,
		&t.ExternalActorURL,
		&t.CreatedAt,
		&t.Data,
		&threadID,
		&campaignID,
		&commentID,
		&ruleID,
		&repositoryID,
		&userID,
		&organizationID,
		&registryExtensionID,
		&importedFromExternalServiceID,
	); err != nil {
		return nil, err
	}
	if actorUserID != nil {
		t.ActorUserID = *actorUserID
	}
	if threadID != nil {
		t.Thread = *threadID
	}
	if campaignID != nil {
		t.Campaign = *campaignID
	}
	if commentID != nil {
		t.Comment = *commentID
	}
	if ruleID != nil {
		t.Rule = *ruleID
	}
	if repositoryID != nil {
		t.Repository = *repositoryID
	}
	if userID != nil {
		t.User = *userID
	}
	if organizationID != nil {
		t.Organization = *organizationID
	}
	if registryExtensionID != nil {
		t.RegistryExtension = *registryExtensionID
	}
	if importedFromExternalServiceID != nil {
		t.ImportedFromExternalServiceID = *importedFromExternalServiceID
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
