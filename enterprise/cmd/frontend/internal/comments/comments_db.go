package comments

import (
	"context"
	"database/sql"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// dbComment describes a comment.
type dbComment struct {
	ID           int64
	Object       CommentObject
	AuthorUserID int32
	Body         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CommentObject stores the object that the comment is associated with. Exactly 1 field is nonzero.
type CommentObject struct {
	ThreadID   int64
	CampaignID int64
}

// errCommentNotFound occurs when a database operation expects a specific comment to exist but it
// does not exist.
var errCommentNotFound = errors.New("comment not found")

type dbComments struct{}

const selectColumns = `id, author_user_id, body, created_at, updated_at, thread_id, campaign_id`

// Create creates a comment. The comment argument's (Comment).ID field is ignored. The new comment
// is returned.
func (dbComments) Create(ctx context.Context, comment *dbComment) (*dbComment, error) {
	if mocks.comments.Create != nil {
		return mocks.comments.Create(comment)
	}

	return dbComments{}.create(ctx, dbconn.Global, comment)
}

type QueryRowContexter interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func nilIfZero(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

func (dbComments) create(ctx context.Context, dbh QueryRowContexter, comment *dbComment) (*dbComment, error) {
	return dbComments{}.scanRow(dbh.QueryRowContext(ctx,
		`INSERT INTO comments(`+selectColumns+`) VALUES(DEFAULT, $1, $2, DEFAULT, DEFAULT, $3, $4) RETURNING `+selectColumns,
		comment.AuthorUserID,
		comment.Body,
		nilIfZero(comment.Object.ThreadID),
		nilIfZero(comment.Object.CampaignID),
	))
}

type insertRelatedObjectFunc func(ctx context.Context, tx QueryRowContexter, commentID int64) (*CommentObject, error)

// DBObjectCommentFields contains the subset of fields required when creating a comment that is
// related to another object.
type DBObjectCommentFields struct {
	AuthorUserID int32
	Body         string
}

// CreateWithObject creates a comment and its related object (such as a thread or campaign) in a
// transaction.
//
// The insertRelatedObject func is called with the comment ID, in case the related object needs to
// store the comment ID. After the related object is inserted, the comment row is updated to refer
// to the object by ID (e.g., the thread ID or campaign ID).
func CreateWithObject(ctx context.Context, comment DBObjectCommentFields, insertRelatedObject insertRelatedObjectFunc) (err error) {
	tx, err := dbconn.Global.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			rollErr := tx.Rollback()
			if rollErr != nil {
				err = multierror.Append(err, rollErr)
			}
			return
		}
		err = tx.Commit()
	}()

	insertedComment, err := dbComments{}.create(ctx, tx, &dbComment{AuthorUserID: comment.AuthorUserID, Body: comment.Body})
	if err != nil {
		return err
	}

	object, err := insertRelatedObject(ctx, tx, insertedComment.ID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE comments SET thread_id=$1, campaign_id=$2 WHERE id=$3`,
		nilIfZero(object.ThreadID),
		nilIfZero(object.CampaignID),
		insertedComment.ID,
	)
	return err
}

type dbCommentUpdate struct {
	Body *string
}

// Update updates a comment given its ID.
func (s dbComments) Update(ctx context.Context, id int64, update dbCommentUpdate) (*dbComment, error) {
	if mocks.comments.Update != nil {
		return mocks.comments.Update(id, update)
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
		return nil, errCommentNotFound
	}
	return results[0], nil
}

var DBGetByID = (dbComments{}).GetByID

// GetByID retrieves the comment (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this comment.
func (dbComments) GetByID(ctx context.Context, id int64) (*dbComment, error) {
	if mocks.comments.GetByID != nil {
		return mocks.comments.GetByID(id)
	}

	results, err := dbComments{}.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errCommentNotFound
	}
	return results[0], nil
}

// dbCommentsListOptions contains options for listing comments.
type dbCommentsListOptions struct {
	Query  string // only list comments matching this query (case-insensitively)
	Object CommentObject
	*db.LimitOffset
}

func (o dbCommentsListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.Query != "" {
		conds = append(conds, sqlf.Sprintf("body ILIKE %s", "%"+o.Query+"%"))
	}
	if o.Object.ThreadID != 0 {
		conds = append(conds, sqlf.Sprintf("thread_id=%d", o.Object.ThreadID))
	}
	if o.Object.CampaignID != 0 {
		conds = append(conds, sqlf.Sprintf("campaign_id=%d", o.Object.CampaignID))
	}
	return conds
}

// List lists all comments that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s dbComments) List(ctx context.Context, opt dbCommentsListOptions) ([]*dbComment, error) {
	if mocks.comments.List != nil {
		return mocks.comments.List(opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s dbComments) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*dbComment, error) {
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

func (dbComments) query(ctx context.Context, query *sqlf.Query) ([]*dbComment, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*dbComment
	for rows.Next() {
		t, err := dbComments{}.scanRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, nil
}

func (dbComments) scanRow(row interface {
	Scan(dest ...interface{}) error
}) (*dbComment, error) {
	var t dbComment
	var threadID, campaignID *int64
	if err := row.Scan(
		&t.ID,
		&t.AuthorUserID,
		&t.Body,
		&t.CreatedAt,
		&t.UpdatedAt,
		&threadID,
		&campaignID,
	); err != nil {
		return nil, err
	}
	if threadID != nil {
		t.Object.ThreadID = *threadID
	}
	if campaignID != nil {
		t.Object.CampaignID = *campaignID
	}
	return &t, nil
}

// Count counts all comments that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the comments.
func (dbComments) Count(ctx context.Context, opt dbCommentsListOptions) (int, error) {
	if mocks.comments.Count != nil {
		return mocks.comments.Count(opt)
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
func (s dbComments) DeleteByID(ctx context.Context, id int64) error {
	if mocks.comments.DeleteByID != nil {
		return mocks.comments.DeleteByID(id)
	}
	return s.delete(ctx, sqlf.Sprintf("id=%d", id))
}

func (dbComments) delete(ctx context.Context, cond *sqlf.Query) error {
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
		return errCommentNotFound
	}
	return nil
}

// mockComments mocks the comments-related DB operations.
type mockComments struct {
	Create     func(*dbComment) (*dbComment, error)
	Update     func(int64, dbCommentUpdate) (*dbComment, error)
	GetByID    func(int64) (*dbComment, error)
	List       func(dbCommentsListOptions) ([]*dbComment, error)
	Count      func(dbCommentsListOptions) (int, error)
	DeleteByID func(int64) error
}
