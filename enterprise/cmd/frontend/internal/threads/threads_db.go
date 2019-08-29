package threads

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/actor"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	commentobjectdb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/nnz"
)

// DBThread describes a thread.
type DBThread struct {
	ID               int64
	RepositoryID     api.RepoID // the repository associated with this thread
	Title            string
	IsDraft          bool
	State            string
	Assignee         actor.DBColumns
	PrimaryCommentID int64
	CreatedAt        time.Time
	UpdatedAt        time.Time

	// Changeset
	BaseRef          string
	BaseRefOID       string
	HeadRepositoryID int32
	HeadRef          string
	HeadRefOID       string

	ImportedFromExternalServiceID int64
	ExternalID                    string
	ExternalMetadata              json.RawMessage
}

// errThreadNotFound occurs when a database operation expects a specific thread to exist but it does
// not exist.
var errThreadNotFound = errors.New("thread not found")

type dbThreads struct{}

const SelectColumns = "id, repository_id, title, is_draft, state, assignee_user_id, assignee_external_actor_username, assignee_external_actor_url, primary_comment_id, created_at, updated_at, base_ref, base_ref_oid, head_repository_id, head_ref, head_ref_oid, imported_from_external_service_id, external_id, external_metadata"

// Create creates a thread. The thread argument's (Thread).ID field is ignored. The new thread is
// returned.
func (dbThreads) Create(ctx context.Context, tx *sql.Tx, thread *DBThread, comment commentobjectdb.DBObjectCommentFields) (*DBThread, error) {
	if mocks.threads.Create != nil {
		return mocks.threads.Create(thread)
	}

	if thread.PrimaryCommentID != 0 {
		panic("thread.PrimaryCommentID must not be set")
	}

	now := time.Now()
	nowIfZeroTime := func(t time.Time) time.Time {
		if t.IsZero() {
			return now
		}
		return t
	}

	return thread, commentobjectdb.CreateCommentWithObject(ctx, tx, comment, func(ctx context.Context, tx *sql.Tx, commentID int64) (*types.CommentObject, error) {
		args := []interface{}{
			thread.RepositoryID,
			thread.Title,
			thread.IsDraft,
			thread.State,
			nnz.Int32(thread.Assignee.UserID),
			nnz.String(thread.Assignee.ExternalActorUsername),
			nnz.String(thread.Assignee.ExternalActorURL),
			commentID,
			nowIfZeroTime(thread.CreatedAt),
			nowIfZeroTime(thread.UpdatedAt),
			nnz.String(thread.BaseRef),
			nnz.String(thread.BaseRefOID),
			nnz.Int32(thread.HeadRepositoryID),
			nnz.String(thread.HeadRef),
			nnz.String(thread.HeadRefOID),
			nnz.Int64(thread.ImportedFromExternalServiceID),
			nnz.String(thread.ExternalID),
			nnz.JSON(thread.ExternalMetadata),
		}
		query := sqlf.Sprintf(
			`INSERT INTO threads(`+SelectColumns+`) VALUES(DEFAULT`+strings.Repeat(", %v", len(args))+`) RETURNING `+SelectColumns,
			args...,
		)
		var err error
		thread, err = dbThreads{}.scanRow(tx.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...))
		if err != nil {
			return nil, err
		}
		return &types.CommentObject{ThreadID: thread.ID}, nil
	})
}

// TODO!(sqs)
var Create = dbThreads{}.Create

type dbThreadUpdate struct {
	Title   *string
	IsDraft *bool
	State   *string
	BaseRef *string
	HeadRef *string
}

// Update updates a thread given its ID.
func (s dbThreads) Update(ctx context.Context, id int64, update dbThreadUpdate) (*DBThread, error) {
	if mocks.threads.Update != nil {
		return mocks.threads.Update(id, update)
	}

	var setFields []*sqlf.Query
	if update.Title != nil {
		setFields = append(setFields, sqlf.Sprintf("title=%s", *update.Title))
	}
	if update.IsDraft != nil {
		setFields = append(setFields, sqlf.Sprintf("is_draft=%s", *update.IsDraft))
	}
	if update.State != nil {
		setFields = append(setFields, sqlf.Sprintf("state=%s", *update.State))
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
func (s dbThreads) GetByID(ctx context.Context, id int64) (*DBThread, error) {
	if mocks.threads.GetByID != nil {
		return mocks.threads.GetByID(id)
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

// GetByExternal retrieves the thread (if any) given its external service ID information.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this thread.
func (s dbThreads) GetByExternal(ctx context.Context, importedFromExternalServiceID int64, externalID string) (*DBThread, error) {
	results, err := s.list(ctx, []*sqlf.Query{
		sqlf.Sprintf("imported_from_external_service_id=%d AND external_id=%s",
			importedFromExternalServiceID,
			externalID,
		),
	}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errThreadNotFound
	}
	return results[0], nil
}

// dbThreadsListOptions contains options for listing threads.
type dbThreadsListOptions struct {
	Query                         string       // only list threads matching this query (case-insensitively)
	RepositoryIDs                 []api.RepoID // only list threads in these repositories
	ThreadIDs                     []int64
	LabelNames                    []string
	States                        []string
	ImportedFromExternalServiceID int64
	*db.LimitOffset
}

func (o dbThreadsListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.Query != "" {
		conds = append(conds, sqlf.Sprintf("title ILIKE %s", "%"+o.Query+"%"))
	}
	if o.RepositoryIDs != nil {
		if len(o.RepositoryIDs) > 0 {
			conds = append(conds, sqlf.Sprintf("repository_id=ANY(%v)", pq.Array(o.RepositoryIDs)))
		} else {
			conds = append(conds, sqlf.Sprintf("FALSE"))
		}
	}
	if o.ThreadIDs != nil {
		if len(o.ThreadIDs) > 0 {
			conds = append(conds, sqlf.Sprintf("id=ANY(%v)", pq.Array(o.ThreadIDs)))
		} else {
			conds = append(conds, sqlf.Sprintf("FALSE"))
		}
	}
	if o.LabelNames != nil {
		if len(o.LabelNames) > 0 {
			conds = append(conds, sqlf.Sprintf("EXISTS (SELECT 1 FROM labels l LEFT JOIN labels_objects lo ON l.id=lo.label_id WHERE l.name=ANY(%v) AND lo.thread_id=threads.id)", pq.Array(o.LabelNames)))
		} else {
			conds = append(conds, sqlf.Sprintf("FALSE"))
		}
	}
	if o.States != nil {
		if len(o.States) > 0 {
			conds = append(conds, sqlf.Sprintf("state=ANY(%v)", pq.Array(o.States)))
		} else {
			conds = append(conds, sqlf.Sprintf("FALSE"))
		}
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
func (s dbThreads) List(ctx context.Context, opt dbThreadsListOptions) ([]*DBThread, error) {
	if mocks.threads.List != nil {
		return mocks.threads.List(opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s dbThreads) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*DBThread, error) {
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

func (dbThreads) Query(ctx context.Context, query *sqlf.Query) ([]*DBThread, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*DBThread
	for rows.Next() {
		t, err := dbThreads{}.scanRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, nil
}

func (dbThreads) scanRow(row interface {
	Scan(dest ...interface{}) error
}) (*DBThread, error) {
	var t DBThread
	if err := row.Scan(
		&t.ID,
		&t.RepositoryID,
		&t.Title,
		&t.IsDraft,
		&t.State,
		nnz.ToInt32(&t.Assignee.UserID),
		(*nnz.String)(&t.Assignee.ExternalActorUsername),
		(*nnz.String)(&t.Assignee.ExternalActorURL),
		&t.PrimaryCommentID,
		&t.CreatedAt,
		&t.UpdatedAt,
		(*nnz.String)(&t.BaseRef),
		(*nnz.String)(&t.BaseRefOID),
		nnz.ToInt32(&t.HeadRepositoryID),
		(*nnz.String)(&t.HeadRef),
		(*nnz.String)(&t.HeadRefOID),
		(*nnz.Int64)(&t.ImportedFromExternalServiceID),
		(*nnz.String)(&t.ExternalID),
		nnz.ToJSON(&t.ExternalMetadata),
	); err != nil {
		return nil, err
	}
	return &t, nil
}

// Count counts all threads that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the threads.
func (dbThreads) Count(ctx context.Context, opt dbThreadsListOptions) (int, error) {
	if mocks.threads.Count != nil {
		return mocks.threads.Count(opt)
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
func (dbThreads) Delete(ctx context.Context, tx *sql.Tx, opt dbThreadsListOptions) error {
	if mocks.threads.Delete != nil {
		return mocks.threads.Delete(opt)
	}

	query := sqlf.Sprintf(`DELETE FROM threads WHERE (%s)`, sqlf.Join(opt.sqlConditions(), ") AND ("))
	_, err := dbconn.TxOrGlobal(tx).ExecContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	return err
}

// DeleteByID deletes a thread given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to delete the thread.
func (s dbThreads) DeleteByID(ctx context.Context, id int64) error {
	if mocks.threads.DeleteByID != nil {
		return mocks.threads.DeleteByID(id)
	}
	return s.delete(ctx, sqlf.Sprintf("id=%d", id))
}

func (dbThreads) delete(ctx context.Context, cond *sqlf.Query) error {
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
	Update     func(int64, dbThreadUpdate) (*DBThread, error)
	GetByID    func(int64) (*DBThread, error)
	List       func(dbThreadsListOptions) ([]*DBThread, error)
	Count      func(dbThreadsListOptions) (int, error)
	Delete     func(dbThreadsListOptions) error
	DeleteByID func(int64) error
}

// TestCreateThread creates a thread in the DB, for use in tests only.
func TestCreateThread(ctx context.Context, title string, repositoryID api.RepoID, authorUserID int32) (id int64, err error) {
	thread, err := dbThreads{}.Create(ctx, nil,
		&DBThread{
			RepositoryID: repositoryID,
			Title:        title,
		},
		commentobjectdb.DBObjectCommentFields{Author: actor.DBColumns{UserID: authorUserID}},
	)
	if err != nil {
		return 0, err
	}
	return thread.ID, nil
}
