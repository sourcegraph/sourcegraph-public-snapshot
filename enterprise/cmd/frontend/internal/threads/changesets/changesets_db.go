package changesets

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// dbChangeset describes a changeset.
type dbChangeset struct {
	threads.DBThreadCommon
	Status    graphqlbackend.ChangesetStatus
	IsPreview bool
	BaseRef   string
	HeadRef   string
}

// errChangesetNotFound occurs when a database operation expects a specific changeset to exist but it does
// not exist.
var errChangesetNotFound = errors.New("changeset not found")

type dbChangesets struct{}

const selectColumns = "is_preview, base_ref, head_ref"

// Create creates a changeset. The changeset argument's (Changeset).ID field is ignored. The database ID of
// the new changeset is returned.
func (dbChangesets) Create(ctx context.Context, changeset *dbChangeset) (*dbChangeset, error) {
	if mocks.changesets.Create != nil {
		return mocks.changesets.Create(changeset)
	}

	var id int64
	if err := dbconn.Global.QueryRowContext(ctx,
		`INSERT INTO changesets(`+selectColumns+`) VALUES(DEFAULT, $1, $2, $3, $4, $5, $6) RETURNING id`,
		changeset.RepositoryID,
		changeset.Title,
		changeset.ExternalURL,
		changeset.Status,
	).Scan(&id); err != nil {
		return nil, err
	}
	created := *changeset
	created.ID = id
	return &created, nil
}

type dbChangesetUpdate struct {
	Title       *string
	ExternalURL *string
	Status      *graphqlbackend.ChangesetStatus
}

// Update updates a changeset given its ID.
func (s dbChangesets) Update(ctx context.Context, id int64, update dbChangesetUpdate) (*dbChangeset, error) {
	if mocks.changesets.Update != nil {
		return mocks.changesets.Update(id, update)
	}

	var setFields []*sqlf.Query
	if update.Title != nil {
		setFields = append(setFields, sqlf.Sprintf("title=%s", *update.Title))
	}
	if update.ExternalURL != nil {
		// Treat empty string as meaning "set to null". Otherwise there is no way to express that
		// intent.
		var value *string
		if *update.ExternalURL != "" {
			value = update.ExternalURL
		}
		setFields = append(setFields, sqlf.Sprintf("external_url=%s", value))
	}
	if update.Status != nil {
		setFields = append(setFields, sqlf.Sprintf("status=%s", *update.Status))
	}

	if len(setFields) == 0 {
		return nil, nil
	}

	results, err := s.query(ctx, sqlf.Sprintf(`UPDATE changesets SET %v WHERE id=%s RETURNING `+selectColumns, sqlf.Join(setFields, ", "), id))
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errChangesetNotFound
	}
	return results[0], nil
}

// GetByID retrieves the changeset (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this changeset.
func (s dbChangesets) GetByID(ctx context.Context, id int64) (*dbChangeset, error) {
	if mocks.changesets.GetByID != nil {
		return mocks.changesets.GetByID(id)
	}

	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errChangesetNotFound
	}
	return results[0], nil
}

// dbChangesetsListOptions contains options for listing changesets.
type dbChangesetsListOptions struct {
	common threads.DBThreadsListOptionsCommon
	Status graphqlbackend.ChangesetStatus
}

func (o dbChangesetsListOptions) sqlConditions() []*sqlf.Query {
	conds := o.common.SQLConditions()
	if o.Status != "" {
		conds = append(conds, sqlf.Sprintf("status=%s", o.Status))
	}
	return conds
}

// List lists all changesets that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s dbChangesets) List(ctx context.Context, opt dbChangesetsListOptions) ([]*dbChangeset, error) {
	if mocks.changesets.List != nil {
		return mocks.changesets.List(opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s dbChangesets) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*dbChangeset, error) {
	q := sqlf.Sprintf(`
SELECT `+selectColumns+` FROM changesets
WHERE (%s)
ORDER BY title ASC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)
	return s.query(ctx, q)
}

func (dbChangesets) query(ctx context.Context, query *sqlf.Query) ([]*dbChangeset, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*dbChangeset
	for rows.Next() {
		var t dbChangeset
		if err := rows.Scan(
			&t.ID,
			&t.RepositoryID,
			&t.Title,
			&t.ExternalURL,
			&t.Settings,
			&t.Status,
			&t.Type,
		); err != nil {
			return nil, err
		}
		results = append(results, &t)
	}
	return results, nil
}

// Count counts all changesets that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the changesets.
func (dbChangesets) Count(ctx context.Context, opt dbChangesetsListOptions) (int, error) {
	if mocks.changesets.Count != nil {
		return mocks.changesets.Count(opt)
	}

	q := sqlf.Sprintf("SELECT COUNT(*) FROM changesets WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Delete deletes a changeset given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to delete the changeset.
func (s dbChangesets) DeleteByID(ctx context.Context, id int64) error {
	if mocks.changesets.DeleteByID != nil {
		return mocks.changesets.DeleteByID(id)
	}
	return s.delete(ctx, sqlf.Sprintf("id=%d", id))
}

func (dbChangesets) delete(ctx context.Context, cond *sqlf.Query) error {
	conds := []*sqlf.Query{cond, sqlf.Sprintf("TRUE")}
	q := sqlf.Sprintf("DELETE FROM changesets WHERE (%s)", sqlf.Join(conds, ") AND ("))

	res, err := dbconn.Global.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errChangesetNotFound
	}
	return nil
}

// mockChangesets mocks the changesets-related DB operations.
type mockChangesets struct {
	Create     func(*dbChangeset) (*dbChangeset, error)
	Update     func(int64, dbChangesetUpdate) (*dbChangeset, error)
	GetByID    func(int64) (*dbChangeset, error)
	List       func(dbChangesetsListOptions) ([]*dbChangeset, error)
	Count      func(dbChangesetsListOptions) (int, error)
	DeleteByID func(int64) error
}

// TestCreateChangeset creates a changeset in the DB, for use in tests only.
func TestCreateChangeset(ctx context.Context, title string, repositoryID api.RepoID) (id int64, err error) {
	changeset, err := dbChangesets{}.Create(ctx, &dbChangeset{
		RepositoryID: repositoryID,
		Title:        title,
	})
	if err != nil {
		return 0, err
	}
	return changeset.ID, nil
}
