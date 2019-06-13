package projects

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// dbProject describes a project for a discussion thread.
type dbProject struct {
	ID              int64
	NamespaceUserID int32  // the user namespace where this project is defined
	NamespaceOrgID  int32  // the org namespace where this project is defined
	Name            string // the name (case-preserving)
}

// errProjectNotFound occurs when a database operation expects a specific project to exist but it
// does not exist.
var errProjectNotFound = errors.New("project not found")

type dbProjects struct{}

// Create creates a project. The project argument's (Project).ID field is ignored. The database ID
// of the new project is returned.
func (dbProjects) Create(ctx context.Context, project *dbProject) (*dbProject, error) {
	if mocks.projects.Create != nil {
		return mocks.projects.Create(project)
	}

	nilIfZero := func(v int32) *int32 {
		if v == 0 {
			return nil
		}
		return &v
	}

	var id int64
	if err := dbconn.Global.QueryRowContext(ctx,
		`INSERT INTO projects(namespace_user_id, namespace_org_id, name) VALUES($1, $2, $3) RETURNING id`,
		nilIfZero(project.NamespaceUserID), nilIfZero(project.NamespaceOrgID), project.Name,
	).Scan(&id); err != nil {
		return nil, err
	}
	created := *project
	created.ID = id
	return &created, nil
}

type dbProjectUpdate struct {
	Name *string
}

// Update updates a project given its ID.
func (s dbProjects) Update(ctx context.Context, id int64, update dbProjectUpdate) (*dbProject, error) {
	if mocks.projects.Update != nil {
		return mocks.projects.Update(id, update)
	}

	var setFields []*sqlf.Query
	if update.Name != nil {
		setFields = append(setFields, sqlf.Sprintf("name=%s", *update.Name))
	}

	if len(setFields) == 0 {
		return nil, nil
	}

	results, err := s.query(ctx, sqlf.Sprintf(`UPDATE projects SET %v WHERE id=%s RETURNING id, namespace_user_id, namespace_org_id, name`, sqlf.Join(setFields, ", "), id))
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errProjectNotFound
	}
	return results[0], nil
}

// GetByID retrieves the project (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this project.
func (s dbProjects) GetByID(ctx context.Context, id int64) (*dbProject, error) {
	if mocks.projects.GetByID != nil {
		return mocks.projects.GetByID(id)
	}

	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errProjectNotFound
	}
	return results[0], nil
}

// dbProjectsListOptions contains options for listing projects.
type dbProjectsListOptions struct {
	Query           string // only list projects matching this query (case-insensitively)
	NamespaceUserID int32  // only list projects in this user's namespace
	NamespaceOrgID  int32  // only list projects in this org's namespace
	*db.LimitOffset
}

func (o dbProjectsListOptions) sqlConditions() []*sqlf.Query {
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

// List lists all projects that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s dbProjects) List(ctx context.Context, opt dbProjectsListOptions) ([]*dbProject, error) {
	if mocks.projects.List != nil {
		return mocks.projects.List(opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s dbProjects) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*dbProject, error) {
	q := sqlf.Sprintf(`
SELECT id, namespace_user_id, namespace_org_id, name FROM projects
WHERE (%s)
ORDER BY name ASC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)
	return s.query(ctx, q)
}

func (dbProjects) query(ctx context.Context, query *sqlf.Query) ([]*dbProject, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*dbProject
	for rows.Next() {
		var t dbProject
		var namespaceUserID, namespaceOrgID *int32
		if err := rows.Scan(&t.ID, &namespaceUserID, &namespaceOrgID, &t.Name); err != nil {
			return nil, err
		}
		if namespaceUserID != nil {
			t.NamespaceUserID = *namespaceUserID
		}
		if namespaceOrgID != nil {
			t.NamespaceOrgID = *namespaceOrgID
		}
		results = append(results, &t)
	}
	return results, nil
}

// Count counts all projects that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the projects.
func (dbProjects) Count(ctx context.Context, opt dbProjectsListOptions) (int, error) {
	if mocks.projects.Count != nil {
		return mocks.projects.Count(opt)
	}

	q := sqlf.Sprintf("SELECT COUNT(*) FROM projects WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Delete deletes a project given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to delete the project.
func (s dbProjects) DeleteByID(ctx context.Context, id int64) error {
	if mocks.projects.DeleteByID != nil {
		return mocks.projects.DeleteByID(id)
	}
	return s.delete(ctx, sqlf.Sprintf("id=%d", id))
}

func (dbProjects) delete(ctx context.Context, cond *sqlf.Query) error {
	conds := []*sqlf.Query{cond, sqlf.Sprintf("TRUE")}
	q := sqlf.Sprintf("DELETE FROM projects WHERE (%s)", sqlf.Join(conds, ") AND ("))

	res, err := dbconn.Global.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errProjectNotFound
	}
	return nil
}

// mockProjects mocks the projects-related DB operations.
type mockProjects struct {
	Create     func(*dbProject) (*dbProject, error)
	Update     func(int64, dbProjectUpdate) (*dbProject, error)
	GetByID    func(int64) (*dbProject, error)
	List       func(dbProjectsListOptions) ([]*dbProject, error)
	Count      func(dbProjectsListOptions) (int, error)
	DeleteByID func(int64) error
}

// TestCreateProject creates a project in the DB, for use in tests only.
func TestCreateProject(ctx context.Context, name string, namespaceUserID, namespaceOrgID int32) (id int64, err error) {
	project, err := dbProjects{}.Create(ctx, &dbProject{
		Name:            name,
		NamespaceUserID: namespaceUserID,
		NamespaceOrgID:  namespaceOrgID,
	})
	if err != nil {
		return 0, err
	}
	return project.ID, nil
}
