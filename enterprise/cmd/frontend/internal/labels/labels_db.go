package labels

import (
	"context"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// dbLabel describes a label for a discussion thread.
type dbLabel struct {
	ID           int64
	RepositoryID int64  // the repository that defines the label
	Name         string // the name (case-preserving)
	Description  *string
	Color        string // the hex color code (omitting the '#' prefix)
}

// errLabelNotFound occurs when a database operation expects a specific label to exist but it does
// not exist.
var errLabelNotFound = errors.New("label not found")

type dbLabels struct{}

const selectColumns = `id, repository_id, name, description, color`

// Create creates a label. The label argument's (Label).ID field is ignored.
func (dbLabels) Create(ctx context.Context, label *dbLabel) (*dbLabel, error) {
	if mocks.labels.Create != nil {
		return mocks.labels.Create(label)
	}

	args := []interface{}{
		label.RepositoryID,
		label.Name,
		label.Description,
		label.Color,
	}
	query := sqlf.Sprintf(
		`INSERT INTO labels(`+selectColumns+`) VALUES(DEFAULT`+strings.Repeat(", %v", len(args))+`) RETURNING `+selectColumns,
		args...,
	)
	return dbLabels{}.scanRow(dbconn.Global.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...))
}

type dbLabelUpdate struct {
	Name        *string
	Description *string
	Color       *string
}

// Update updates a label given its ID.
func (s dbLabels) Update(ctx context.Context, id int64, update dbLabelUpdate) (*dbLabel, error) {
	if mocks.labels.Update != nil {
		return mocks.labels.Update(id, update)
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
	if update.Color != nil {
		setFields = append(setFields, sqlf.Sprintf("color=%s", *update.Color))
	}

	if len(setFields) == 0 {
		return nil, nil
	}

	results, err := s.query(ctx, sqlf.Sprintf(`UPDATE labels SET %v WHERE id=%s RETURNING `+selectColumns, sqlf.Join(setFields, ", "), id))
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errLabelNotFound
	}
	return results[0], nil
}

// GetByID retrieves the label (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this label.
func (s dbLabels) GetByID(ctx context.Context, id int64) (*dbLabel, error) {
	if mocks.labels.GetByID != nil {
		return mocks.labels.GetByID(id)
	}

	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errLabelNotFound
	}
	return results[0], nil
}

// dbLabelsListOptions contains options for listing labels.
type dbLabelsListOptions struct {
	Query        string // only list labels matching this query (case-insensitively)
	RepositoryID int64  // only list labels defined in this repository
	*db.LimitOffset
}

func (o dbLabelsListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.Query != "" {
		conds = append(conds, sqlf.Sprintf("name LIKE %s", "%"+o.Query+"%"))
	}
	if o.RepositoryID != 0 {
		conds = append(conds, sqlf.Sprintf("repository_id=%d", o.RepositoryID))
	}
	return conds
}

// List lists all labels that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s dbLabels) List(ctx context.Context, opt dbLabelsListOptions) ([]*dbLabel, error) {
	if mocks.labels.List != nil {
		return mocks.labels.List(opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s dbLabels) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*dbLabel, error) {
	q := sqlf.Sprintf(`
SELECT `+selectColumns+` FROM labels
WHERE (%s)
ORDER BY name ASC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)
	return s.query(ctx, q)
}

func (dbLabels) query(ctx context.Context, query *sqlf.Query) ([]*dbLabel, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*dbLabel
	for rows.Next() {
		t, err := dbLabels{}.scanRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, nil
}

func (dbLabels) scanRow(row interface {
	Scan(dest ...interface{}) error
}) (*dbLabel, error) {
	var t dbLabel
	if err := row.Scan(
		&t.ID,
		&t.RepositoryID,
		&t.Name,
		&t.Description,
		&t.Color,
	); err != nil {
		return nil, err
	}
	return &t, nil
}

// Count counts all labels that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the labels.
func (dbLabels) Count(ctx context.Context, opt dbLabelsListOptions) (int, error) {
	if mocks.labels.Count != nil {
		return mocks.labels.Count(opt)
	}

	q := sqlf.Sprintf("SELECT COUNT(*) FROM labels WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Delete deletes a label given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to delete the label.
func (s dbLabels) DeleteByID(ctx context.Context, id int64) error {
	if mocks.labels.DeleteByID != nil {
		return mocks.labels.DeleteByID(id)
	}
	return s.delete(ctx, sqlf.Sprintf("id=%d", id))
}

func (dbLabels) delete(ctx context.Context, cond *sqlf.Query) error {
	conds := []*sqlf.Query{cond, sqlf.Sprintf("TRUE")}
	q := sqlf.Sprintf("DELETE FROM labels WHERE (%s)", sqlf.Join(conds, ") AND ("))

	res, err := dbconn.Global.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errLabelNotFound
	}
	return nil
}

// mockLabels mocks the labels-related DB operations.
type mockLabels struct {
	Create     func(*dbLabel) (*dbLabel, error)
	Update     func(int64, dbLabelUpdate) (*dbLabel, error)
	GetByID    func(int64) (*dbLabel, error)
	List       func(dbLabelsListOptions) ([]*dbLabel, error)
	Count      func(dbLabelsListOptions) (int, error)
	DeleteByID func(int64) error
}
