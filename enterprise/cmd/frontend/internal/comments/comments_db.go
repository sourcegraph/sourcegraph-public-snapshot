package comments

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// dbComment describes a comment.
type dbComment struct {
	ID              int64
	NamespaceUserID int32  // the user namespace where this comment is defined
	NamespaceOrgID  int32  // the org namespace where this comment is defined
	Name            string // the name (case-preserving)
	Description     *string
	IsPreview       bool
	Rules           string // the JSON rules TODO!(sqs)
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// errCommentNotFound occurs when a database operation expects a specific comment to exist but it
// does not exist.
var errCommentNotFound = errors.New("comment not found")

type dbComments struct{}

const selectColumns = `id, namespace_user_id, namespace_org_id, name, description, is_preview, rules, created_at, updated_at`

// Create creates a comment. The comment argument's (Comment).ID field is ignored. The database
// ID of the new comment is returned.
func (dbComments) Create(ctx context.Context, comment *dbComment) (*dbComment, error) {
	if mocks.comments.Create != nil {
		return mocks.comments.Create(comment)
	}

	nilIfZero := func(v int32) *int32 {
		if v == 0 {
			return nil
		}
		return &v
	}

	return dbComments{}.scanRow(dbconn.Global.QueryRowContext(ctx,
		`INSERT INTO comments(`+selectColumns+`) VALUES(DEFAULT, $1, $2, $3, $4, $5, $6, DEFAULT, DEFAULT) RETURNING `+selectColumns,
		nilIfZero(comment.NamespaceUserID),
		nilIfZero(comment.NamespaceOrgID),
		comment.Name,
		comment.Description,
		comment.IsPreview,
		comment.Rules,
	))
}

type dbCommentUpdate struct {
	Name        *string
	Description *string
	IsPreview   *bool
	Rules       *string
}

// Update updates a comment given its ID.
func (s dbComments) Update(ctx context.Context, id int64, update dbCommentUpdate) (*dbComment, error) {
	if mocks.comments.Update != nil {
		return mocks.comments.Update(id, update)
	}

	var setFields []*sqlf.Query
	if update.Name != nil {
		setFields = append(setFields, sqlf.Sprintf("name=%s", *update.Name))
	}
	if update.Description != nil {
		// Treat empty string as meaning "set to null". Otherwise there is no way to express that
		// intent.
		var value *string
		if *update.Description != "" {
			value = update.Description
		}
		setFields = append(setFields, sqlf.Sprintf("description=%s", value))
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

	results, err := s.query(ctx, sqlf.Sprintf(`UPDATE comments SET %v WHERE id=%s RETURNING `+selectColumns, sqlf.Join(setFields, ", "), id))
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errCommentNotFound
	}
	return results[0], nil
}

// GetByID retrieves the comment (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this comment.
func (s dbComments) GetByID(ctx context.Context, id int64) (*dbComment, error) {
	if mocks.comments.GetByID != nil {
		return mocks.comments.GetByID(id)
	}

	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)}, nil)
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
	Query           string // only list comments matching this query (case-insensitively)
	NamespaceUserID int32  // only list comments in this user's namespace
	NamespaceOrgID  int32  // only list comments in this org's namespace
	*db.LimitOffset
}

func (o dbCommentsListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.Query != "" {
		conds = append(conds, sqlf.Sprintf("name LIKE %s", "%"+o.Query+"%"))
	}
	if o.NamespaceUserID != 0 {
		conds = append(conds, sqlf.Sprintf("namespace_user_id=%d", o.NamespaceUserID))
	}
	if o.NamespaceOrgID != 0 {
		conds = append(conds, sqlf.Sprintf("namespace_org_id=%d", o.NamespaceOrgID))
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
ORDER BY name ASC
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
	var namespaceUserID, namespaceOrgID *int32
	if err := row.Scan(
		&t.ID,
		&namespaceUserID,
		&namespaceOrgID,
		&t.Name,
		&t.Description,
		&t.IsPreview,
		&t.Rules,
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
