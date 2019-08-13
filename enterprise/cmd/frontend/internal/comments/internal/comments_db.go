package internal

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/actor"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/types"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/nnz"
)

// DBComment describes a comment.
type DBComment struct {
	ID        int64
	Object    types.CommentObject
	Author    actor.DBColumns
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ErrCommentNotFound occurs when a database operation expects a specific comment to exist but it
// does not exist.
var ErrCommentNotFound = errors.New("comment not found")

type DBComments struct{}

const selectColumns = `id, author_user_id, author_external_actor_username, author_external_actor_url, body, created_at, updated_at, parent_comment_id, thread_id, campaign_id`

// Create creates a comment. The comment argument's (Comment).ID field is ignored. The new comment
// is returned.
func (DBComments) Create(ctx context.Context, tx *sql.Tx, comment *DBComment) (*DBComment, error) {
	if Mocks.Comments.Create != nil {
		return Mocks.Comments.Create(comment)
	}

	now := time.Now()
	nowIfZeroTime := func(t time.Time) time.Time {
		if t.IsZero() {
			return now
		}
		return t
	}
	args := []interface{}{
		nnz.Int32(comment.Author.UserID),
		nnz.String(comment.Author.ExternalActorUsername),
		nnz.String(comment.Author.ExternalActorURL),
		comment.Body,
		nowIfZeroTime(comment.CreatedAt),
		nowIfZeroTime(comment.UpdatedAt),
		nnz.Int64(comment.Object.ParentCommentID),
		nnz.Int64(comment.Object.ThreadID),
		nnz.Int64(comment.Object.CampaignID),
	}
	query := sqlf.Sprintf(
		`INSERT INTO comments(`+selectColumns+`) VALUES(DEFAULT`+strings.Repeat(", %v", len(args))+`) RETURNING `+selectColumns,
		args...,
	)
	return DBComments{}.scanRow(dbconn.TxOrGlobal(tx).QueryRowContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...))
}

func nilIfZero(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

func nilIfZero32(v int32) *int32 {
	if v == 0 {
		return nil
	}
	return &v
}

type DBCommentUpdate struct {
	Body *string
}

// Update updates a comment given its ID.
func (s DBComments) Update(ctx context.Context, id int64, update DBCommentUpdate) (*DBComment, error) {
	if Mocks.Comments.Update != nil {
		return Mocks.Comments.Update(id, update)
	}

	var setFields []*sqlf.Query
	if update.Body != nil {
		setFields = append(setFields, sqlf.Sprintf("body=%s", *update.Body))
	}

	if len(setFields) == 0 {
		return nil, nil
	}
	setFields = append(setFields, sqlf.Sprintf("updated_at=now()"))

	results, err := s.query(ctx, sqlf.Sprintf(`UPDATE comments SET %v WHERE id=%s RETURNING `+selectColumns, sqlf.Join(setFields, ", "), id))
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, ErrCommentNotFound
	}
	return results[0], nil
}

// GetByID retrieves the comment (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this comment.
func (DBComments) GetByID(ctx context.Context, id int64) (*DBComment, error) {
	if Mocks.Comments.GetByID != nil {
		return Mocks.Comments.GetByID(id)
	}

	results, err := DBComments{}.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, ErrCommentNotFound
	}
	return results[0], nil
}

// DBCommentsListOptions contains options for listing comments.
type DBCommentsListOptions struct {
	Query                string // only list comments matching this query (case-insensitively)
	Object               types.CommentObject
	ObjectPrimaryComment bool // only return the object's primary comment
	*db.LimitOffset
}

func (o DBCommentsListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.Query != "" {
		conds = append(conds, sqlf.Sprintf("body ILIKE %s", "%"+o.Query+"%"))
	}
	if o.Object.ParentCommentID != 0 {
		conds = append(conds, sqlf.Sprintf("parent_comment_id=%d", o.Object.ParentCommentID))
	}

	addObjectCondition := func(id int64, column, otherTable string) {
		if id != 0 {
			objectConds := []*sqlf.Query{sqlf.Sprintf(column+"=%d", id)}
			if !o.ObjectPrimaryComment {
				objectConds = append(objectConds, sqlf.Sprintf("parent_comment_id=(SELECT primary_comment_id FROM "+otherTable+" WHERE id=%d)", id))
			}
			conds = append(conds, sqlf.Join(objectConds, " OR "))
		}
	}
	addObjectCondition(o.Object.ThreadID, "thread_id", "threads")
	addObjectCondition(o.Object.CampaignID, "campaign_id", "campaigns")

	return conds
}

// List lists all comments that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s DBComments) List(ctx context.Context, opt DBCommentsListOptions) ([]*DBComment, error) {
	if Mocks.Comments.List != nil {
		return Mocks.Comments.List(opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s DBComments) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*DBComment, error) {
	q := sqlf.Sprintf(`
SELECT `+selectColumns+` FROM comments
WHERE (%s)
ORDER BY id ASC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)
	return s.query(ctx, q)
}

func (DBComments) query(ctx context.Context, query *sqlf.Query) ([]*DBComment, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*DBComment
	for rows.Next() {
		t, err := DBComments{}.scanRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, nil
}

func (DBComments) scanRow(row interface {
	Scan(dest ...interface{}) error
}) (*DBComment, error) {
	var t DBComment
	if err := row.Scan(
		&t.ID,
		nnz.ToInt32(&t.Author.UserID),
		(*nnz.String)(&t.Author.ExternalActorUsername),
		(*nnz.String)(&t.Author.ExternalActorURL),
		&t.Body,
		&t.CreatedAt,
		&t.UpdatedAt,
		(*nnz.Int64)(&t.Object.ParentCommentID),
		(*nnz.Int64)(&t.Object.ThreadID),
		(*nnz.Int64)(&t.Object.CampaignID),
	); err != nil {
		return nil, err
	}
	return &t, nil
}

// Count counts all comments that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the comments.
func (DBComments) Count(ctx context.Context, opt DBCommentsListOptions) (int, error) {
	if Mocks.Comments.Count != nil {
		return Mocks.Comments.Count(opt)
	}

	q := sqlf.Sprintf("SELECT COUNT(*) FROM comments WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Delete deletes a comment given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to delete the comment.
func (s DBComments) DeleteByID(ctx context.Context, id int64) error {
	if Mocks.Comments.DeleteByID != nil {
		return Mocks.Comments.DeleteByID(id)
	}
	return s.delete(ctx, sqlf.Sprintf("id=%d", id))
}

func (DBComments) delete(ctx context.Context, cond *sqlf.Query) error {
	conds := []*sqlf.Query{cond, sqlf.Sprintf("TRUE")}
	q := sqlf.Sprintf("DELETE FROM comments WHERE (%s)", sqlf.Join(conds, ") AND ("))

	res, err := dbconn.Global.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return ErrCommentNotFound
	}
	return nil
}

// mockComments Mocks the comments-related DB operations.
type mockComments struct {
	Create     func(*DBComment) (*DBComment, error)
	Update     func(int64, DBCommentUpdate) (*DBComment, error)
	GetByID    func(int64) (*DBComment, error)
	List       func(DBCommentsListOptions) ([]*DBComment, error)
	Count      func(DBCommentsListOptions) (int, error)
	DeleteByID func(int64) error
}
