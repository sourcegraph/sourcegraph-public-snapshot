package campaigns

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// dbChangesetCampaign describes a changeset campaign.
type dbChangesetCampaign struct {
	ID          int64
	ProjectID   int64  // the project that defines the changeset campaign
	Name        string // the name (case-preserving)
	Description *string
}

// errChangesetCampaignNotFound occurs when a database operation expects a specific changeset
// campaign to exist but it does not exist.
var errChangesetCampaignNotFound = errors.New("changesetCampaign not found")

type dbChangesetCampaigns struct{}

// Create creates a changeset campaign. The changesetCampaign argument's (ChangesetCampaign).ID
// field is ignored. The database ID of the new changeset campaign is returned.
func (dbChangesetCampaigns) Create(ctx context.Context, changesetCampaign *dbChangesetCampaign) (*dbChangesetCampaign, error) {
	if mocks.campaigns.Create != nil {
		return mocks.campaigns.Create(changesetCampaign)
	}

	var id int64
	if err := dbconn.Global.QueryRowContext(ctx,
		`INSERT INTO changeset_campaigns(project_id, name, description) VALUES($1, $2, $3, $4) RETURNING id`,
		changesetCampaign.ProjectID, changesetCampaign.Name, changesetCampaign.Description,
	).Scan(&id); err != nil {
		return nil, err
	}
	created := *changesetCampaign
	created.ID = id
	return &created, nil
}

type dbChangesetCampaignUpdate struct {
	Name        *string
	Description *string
}

// Update updates a changeset campaign given its ID.
func (s dbChangesetCampaigns) Update(ctx context.Context, id int64, update dbChangesetCampaignUpdate) (*dbChangesetCampaign, error) {
	if mocks.campaigns.Update != nil {
		return mocks.campaigns.Update(id, update)
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

	if len(setFields) == 0 {
		return nil, nil
	}

	results, err := s.query(ctx, sqlf.Sprintf(`UPDATE campaigns SET %v WHERE id=%s RETURNING id, project_id, name, description`, sqlf.Join(setFields, ", "), id))
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errChangesetCampaignNotFound
	}
	return results[0], nil
}

// GetByID retrieves the changesetCampaign (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this changesetCampaign.
func (s dbChangesetCampaigns) GetByID(ctx context.Context, id int64) (*dbChangesetCampaign, error) {
	if mocks.campaigns.GetByID != nil {
		return mocks.campaigns.GetByID(id)
	}

	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errChangesetCampaignNotFound
	}
	return results[0], nil
}

// dbChangesetCampaignsListOptions contains options for listing campaigns.
type dbChangesetCampaignsListOptions struct {
	Query     string // only list campaigns matching this query (case-insensitively)
	ProjectID int64  // only list campaigns defined in this project
	*db.LimitOffset
}

func (o dbChangesetCampaignsListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.Query != "" {
		conds = append(conds, sqlf.Sprintf("name LIKE %s", "%"+o.Query+"%"))
	}
	if o.ProjectID != 0 {
		conds = append(conds, sqlf.Sprintf("project_id=%d", o.ProjectID))
	}
	return conds
}

// List lists all changeset campaigns that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s dbChangesetCampaigns) List(ctx context.Context, opt dbChangesetCampaignsListOptions) ([]*dbChangesetCampaign, error) {
	if mocks.campaigns.List != nil {
		return mocks.campaigns.List(opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s dbChangesetCampaigns) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*dbChangesetCampaign, error) {
	q := sqlf.Sprintf(`
SELECT id, project_id, name, description FROM campaigns
WHERE (%s)
ORDER BY name ASC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)
	return s.query(ctx, q)
}

func (dbChangesetCampaigns) query(ctx context.Context, query *sqlf.Query) ([]*dbChangesetCampaign, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*dbChangesetCampaign
	for rows.Next() {
		var t dbChangesetCampaign
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Name, &t.Description); err != nil {
			return nil, err
		}
		results = append(results, &t)
	}
	return results, nil
}

// Count counts all changeset campaigns that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the campaigns.
func (dbChangesetCampaigns) Count(ctx context.Context, opt dbChangesetCampaignsListOptions) (int, error) {
	if mocks.campaigns.Count != nil {
		return mocks.campaigns.Count(opt)
	}

	q := sqlf.Sprintf("SELECT COUNT(*) FROM campaigns WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Delete deletes a changeset campaign given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to delete the changeset campaign.
func (s dbChangesetCampaigns) DeleteByID(ctx context.Context, id int64) error {
	if mocks.campaigns.DeleteByID != nil {
		return mocks.campaigns.DeleteByID(id)
	}
	return s.delete(ctx, sqlf.Sprintf("id=%d", id))
}

func (dbChangesetCampaigns) delete(ctx context.Context, cond *sqlf.Query) error {
	conds := []*sqlf.Query{cond, sqlf.Sprintf("TRUE")}
	q := sqlf.Sprintf("DELETE FROM campaigns WHERE (%s)", sqlf.Join(conds, ") AND ("))

	res, err := dbconn.Global.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errChangesetCampaignNotFound
	}
	return nil
}

// mockChangesetCampaigns mocks the changeset campaigns-related DB operations.
type mockChangesetCampaigns struct {
	Create     func(*dbChangesetCampaign) (*dbChangesetCampaign, error)
	Update     func(int64, dbChangesetCampaignUpdate) (*dbChangesetCampaign, error)
	GetByID    func(int64) (*dbChangesetCampaign, error)
	List       func(dbChangesetCampaignsListOptions) ([]*dbChangesetCampaign, error)
	Count      func(dbChangesetCampaignsListOptions) (int, error)
	DeleteByID func(int64) error
}
