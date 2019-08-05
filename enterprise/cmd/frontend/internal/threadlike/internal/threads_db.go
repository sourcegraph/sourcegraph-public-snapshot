package internal

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	commentobjectdb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

type DBThreadType string

const (
	DBThreadTypeThread    DBThreadType = "THREAD"
	DBThreadTypeIssue                  = "ISSUE"
	DBThreadTypeChangeset              = "CHANGESET"
)

// DBThread describes a thread.
type DBThread struct {
	ID           int64
	Type         DBThreadType
	RepositoryID api.RepoID // the repository associated with this thread
	Title        string
	State        string
	IsPreview    bool

	PrimaryCommentID int64
	CreatedAt        time.Time
	UpdatedAt        time.Time

	// Changeset
	BaseRef string
	HeadRef string

	ImportedFromExternalServiceID int64
	ExternalID                    string
}

// errThreadNotFound occurs when a database operation expects a specific thread to exist but it does
// not exist.
var errThreadNotFound = errors.New("thread not found")

type DBThreads struct{}

const SelectColumns = "id, type, repository_id, title, state, is_preview, primary_comment_id, created_at, updated_at, base_ref, head_ref, imported_from_external_service_id, external_id"

// Create creates a thread. The thread argument's (Thread).ID field is ignored. The new thread is
// returned.
func (DBThreads) Create(ctx context.Context, tx *sql.Tx, thread *DBThread, comment commentobjectdb.DBObjectCommentFields) (*DBThread, error) {
	if Mocks.Threads.Create != nil {
		return Mocks.Threads.Create(thread)
	}

	if thread.PrimaryCommentID != 0 {
		panic("thread.PrimaryCommentID must not be set")
	}

	nilIfZero := func(v int64) *int64 {
		if v == 0 {
			return nil
		}
		return &v
	}
	nilIfEmpty := func(v string) *string {
		if v == "" {
			return nil
		}
		return &v
	}
	now := time.Now()
	nowIfZeroTime := func(t time.Time) time.Time {
		if t.IsZero() {
			return now
		}
		return t
	}

	return thread, commentobjectdb.CreateCommentWithObject(ctx, tx, comment, func(ctx context.Context, tx *sql.Tx, commentID int64) (*types.CommentObject, error) {
		var err error
		thread, err = DBThreads{}.scanRow(tx.QueryRowContext(ctx,
			`INSERT INTO threads(`+SelectColumns+`) VALUES(DEFAULT, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING `+SelectColumns,
			thread.Type,
			thread.RepositoryID,
			thread.Title,
			thread.State,
			thread.IsPreview,
			commentID,
			nowIfZeroTime(thread.CreatedAt),
			nowIfZeroTime(thread.UpdatedAt),
			nilIfEmpty(thread.BaseRef),
			nilIfEmpty(thread.HeadRef),
			nilIfZero(thread.ImportedFromExternalServiceID),
			nilIfEmpty(thread.ExternalID),
		))
		if err != nil {
			return nil, err
		}
		return &types.CommentObject{ThreadID: thread.ID}, nil
	})
}

type DBThreadUpdate struct {
	Title     *string
	State     *string
	IsPreview *bool
	BaseRef   *string
	HeadRef   *string
}

// Update updates a thread given its ID.
func (s DBThreads) Update(ctx context.Context, id int64, update DBThreadUpdate) (*DBThread, error) {
	if Mocks.Threads.Update != nil {
		return Mocks.Threads.Update(id, update)
	}

	var setFields []*sqlf.Query
	if update.Title != nil {
		setFields = append(setFields, sqlf.Sprintf("title=%s", *update.Title))
	}
	if update.State != nil {
		setFields = append(setFields, sqlf.Sprintf("state=%s", *update.State))
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

	results, err := s.Query(ctx, sqlf.Sprintf(`UPDATE threads SET %v WHERE id=%s RETURNING `+SelectColumns, sqlf.Join(setFields, ", "), id))
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
	Query                         string     // only list threads matching this query (case-insensitively)
	RepositoryID                  api.RepoID // only list threads in this repository
	ThreadIDs                     []int64
	State                         string
	ImportedFromExternalServiceID int64
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
	if o.State != "" {
		conds = append(conds, sqlf.Sprintf("state=%s", o.State))
	}
	if o.ImportedFromExternalServiceID != 0 {
		conds = append(conds, sqlf.Sprintf("imported_from_external_service_id=%d", o.ImportedFromExternalServiceID))
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
SELECT `+SelectColumns+` FROM threads
WHERE (%s)
ORDER BY title ASC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)
	return s.Query(ctx, q)
}

func (DBThreads) Query(ctx context.Context, query *sqlf.Query) ([]*DBThread, error) {
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
	var (
		baseRef, headRef              *string
		importedFromExternalServiceID *int64
		externalID                    *string

		t DBThread
	)
	if err := row.Scan(
		&t.ID,
		&t.Type,
		&t.RepositoryID,
		&t.Title,
		&t.State,
		&t.IsPreview,
		&t.PrimaryCommentID,
		&t.CreatedAt,
		&t.UpdatedAt,
		&baseRef,
		&headRef,
		&importedFromExternalServiceID,
		&externalID,
	); err != nil {
		return nil, err
	}
	if baseRef != nil {
		t.BaseRef = *baseRef
	}
	if headRef != nil {
		t.HeadRef = *headRef
	}
	if importedFromExternalServiceID != nil {
		t.ImportedFromExternalServiceID = *importedFromExternalServiceID
	}
	if externalID != nil {
		t.ExternalID = *externalID
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

// Delete deletes all threads matching the criteria (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to delete the threads.
func (DBThreads) Delete(ctx context.Context, tx *sql.Tx, opt DBThreadsListOptions) error {
	if Mocks.Threads.Delete != nil {
		return Mocks.Threads.Delete(opt)
	}

	query := sqlf.Sprintf(`DELETE FROM threads WHERE (%s)`, sqlf.Join(opt.sqlConditions(), ") AND ("))
	_, err := dbconn.TxOrGlobal(tx).ExecContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	return err
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
	Delete     func(DBThreadsListOptions) error
	DeleteByID func(int64) error
}
