package internal

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// DBThread describes a thread.
type DBThread struct {
	ID           int64
	Type         graphqlbackend.ThreadlikeType
	RepositoryID api.RepoID // the repository associated with this thread
	Title        string
	ExternalURL  *string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Comment *comments.DBComment // Primary comment

	// Changeset
	IsPreview bool
	BaseRef   string
	HeadRef   string
}

// errThreadNotFound occurs when a database operation expects a specific thread to exist but it does
// not exist.
var errThreadNotFound = errors.New("thread not found")

type DBThreads struct{}

const selectColumns = "id, type, repository_id, title, external_url, status, is_preview, base_ref, head_ref"

// Create creates a thread. The thread argument's (Thread).ID field is ignored. The new thread is
// returned.
func (DBThreads) Create(ctx context.Context, thread *DBThread) (*DBThread, error) {
	if Mocks.Threads.Create != nil {
		return Mocks.Threads.Create(thread)
	}

	return DBThreads{}.scanRow(dbconn.Global.QueryRowContext(ctx,
		`INSERT INTO threads(`+selectColumns+`) VALUES(DEFAULT, $1, $2, $3, $4, $5, $6, $7, $8) RETURNING `+selectColumns,
		thread.Type,
		thread.RepositoryID,
		thread.Title,
		thread.ExternalURL,
		thread.Status,
		thread.IsPreview,
		thread.BaseRef,
		thread.HeadRef,
	))
}

type DBThreadUpdate struct {
	Title       *string
	ExternalURL *string
	Status      *string
	IsPreview   *bool
	BaseRef     *string
	HeadRef     *string
}

// Update updates a thread given its ID.
func (s DBThreads) Update(ctx context.Context, id int64, update DBThreadUpdate) (*DBThread, error) {
	if Mocks.Threads.Update != nil {
		return Mocks.Threads.Update(id, update)
	}

	// Treat empty string as meaning "set to null". Otherwise there is no way to express that
	// intent for some of the fields below.
	emptyStringAsNil := func(s *string) (value *string, isSet bool) {
		if s == nil {
			return nil, false
		}
		if *s == "" {
			return nil, true
		}
		return s, true
	}

	var setFields []*sqlf.Query
	if update.Title != nil {
		setFields = append(setFields, sqlf.Sprintf("title=%s", *update.Title))
	}
	if v, isSet := emptyStringAsNil(update.ExternalURL); isSet {
		setFields = append(setFields, sqlf.Sprintf("external_url=%s", v))
	}
	if update.Status != nil {
		setFields = append(setFields, sqlf.Sprintf("status=%s", *update.Status))
	}
	if update.IsPreview != nil {
		setFields = append(setFields, sqlf.Sprintf("is_preview=%s", *update.IsPreview))
	}
	if update.BaseRef != nil {
		setFields = append(setFields, sqlf.Sprintf("base_ref=%s", *update.BaseRef))
	}
	if update.HeadRef != nil {
		setFields = append(setFields, sqlf.Sprintf("head_ref=%s", *update.HeadRef))
	}

	if len(setFields) == 0 {
		return nil, nil
	}
	// TODO!(sqs): need to reset updated_at of the thread's corresponding comment
	// setFields = append(setFields, sqlf.Sprintf("updated_at=now()"))

	results, err := s.query(ctx, sqlf.Sprintf(`UPDATE threads SET %v WHERE id=%s RETURNING `+selectColumns, sqlf.Join(setFields, ", "), id))
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errThreadNotFound
	}
	return results[0], nil
}

// GetByID retrieves the thread (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this thread.
func (s DBThreads) GetByID(ctx context.Context, id int64) (*DBThread, error) {
	if Mocks.Threads.GetByID != nil {
		return Mocks.Threads.GetByID(id)
	}

	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errThreadNotFound
	}
	return results[0], nil
}

// DBThreadsListOptions contains options for listing threads.
type DBThreadsListOptions struct {
	Query        string     // only list threads matching this query (case-insensitively)
	RepositoryID api.RepoID // only list threads in this repository
	ThreadIDs    []int64
	Status       string
	*db.LimitOffset
}

func (o DBThreadsListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.Query != "" {
		conds = append(conds, sqlf.Sprintf("title ILIKE %s", "%"+o.Query+"%"))
	}
	if o.RepositoryID != 0 {
		conds = append(conds, sqlf.Sprintf("repository_id=%d", o.RepositoryID))
	}
	if o.ThreadIDs != nil {
		if len(o.ThreadIDs) > 0 {
			conds = append(conds, sqlf.Sprintf("id=ANY(%v)", pq.Array(o.ThreadIDs)))
		} else {
			conds = append(conds, sqlf.Sprintf("FALSE"))
		}
	}
	if o.Status != "" {
		conds = append(conds, sqlf.Sprintf("status=%s", o.Status))
	}
	return conds
}

// List lists all threads that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s DBThreads) List(ctx context.Context, opt DBThreadsListOptions) ([]*DBThread, error) {
	if Mocks.Threads.List != nil {
		return Mocks.Threads.List(opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s DBThreads) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*DBThread, error) {
	q := sqlf.Sprintf(`
SELECT `+selectColumns+` FROM threads
WHERE (%s)
ORDER BY title ASC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)
	return s.query(ctx, q)
}

func (DBThreads) query(ctx context.Context, query *sqlf.Query) ([]*DBThread, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*DBThread
	for rows.Next() {
		t, err := DBThreads{}.scanRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, nil
}

func (DBThreads) scanRow(row interface {
	Scan(dest ...interface{}) error
}) (*DBThread, error) {
	var t DBThread
	if err := row.Scan(
		&t.ID,
		&t.Type,
		&t.RepositoryID,
		&t.Title,
		&t.ExternalURL,
		&t.Status,
		&t.IsPreview,
		&t.BaseRef,
		&t.HeadRef,
	); err != nil {
		return nil, err
	}
	return &t, nil
}

// Count counts all threads that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the threads.
func (DBThreads) Count(ctx context.Context, opt DBThreadsListOptions) (int, error) {
	if Mocks.Threads.Count != nil {
		return Mocks.Threads.Count(opt)
	}

	q := sqlf.Sprintf("SELECT COUNT(*) FROM threads WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// DeleteByID deletes a thread given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to delete the thread.
func (s DBThreads) DeleteByID(ctx context.Context, id int64) error {
	if Mocks.Threads.DeleteByID != nil {
		return Mocks.Threads.DeleteByID(id)
	}
	return s.delete(ctx, sqlf.Sprintf("id=%d", id))
}

func (DBThreads) delete(ctx context.Context, cond *sqlf.Query) error {
	conds := []*sqlf.Query{cond, sqlf.Sprintf("TRUE")}
	q := sqlf.Sprintf("DELETE FROM threads WHERE (%s)", sqlf.Join(conds, ") AND ("))

	res, err := dbconn.Global.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errThreadNotFound
	}
	return nil
}

// mockThreads mocks the threads-related DB operations.
type mockThreads struct {
	Create     func(*DBThread) (*DBThread, error)
	Update     func(int64, DBThreadUpdate) (*DBThread, error)
	GetByID    func(int64) (*DBThread, error)
	List       func(DBThreadsListOptions) ([]*DBThread, error)
	Count      func(DBThreadsListOptions) (int, error)
	DeleteByID func(int64) error
}
