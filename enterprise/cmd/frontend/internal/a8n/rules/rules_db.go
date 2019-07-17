package rules

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// dbRule describes a rule for a discussion thread.
type dbRule struct {
	ID          int64
	ProjectID   int64  // the project that defines the rule
	Name        string // the name (case-preserving)
	Description *string
	Settings       string // the hex settings code (omitting the '#' prefix)
}

// errRuleNotFound occurs when a database operation expects a specific rule to exist but it does
// not exist.
var errRuleNotFound = errors.New("rule not found")

type dbRules struct{}

// Create creates a rule. The rule argument's (Rule).ID field is ignored. The database ID of the
// new rule is returned.
func (dbRules) Create(ctx context.Context, rule *dbRule) (*dbRule, error) {
	if mocks.rules.Create != nil {
		return mocks.rules.Create(rule)
	}

	var id int64
	if err := dbconn.Global.QueryRowContext(ctx,
		`INSERT INTO rules(project_id, name, description, settings) VALUES($1, $2, $3, $4) RETURNING id`,
		rule.ProjectID, rule.Name, rule.Description, rule.Settings,
	).Scan(&id); err != nil {
		return nil, err
	}
	created := *rule
	created.ID = id
	return &created, nil
}

type dbRuleUpdate struct {
	Name        *string
	Description *string
	Settings       *string
}

// Update updates a rule given its ID.
func (s dbRules) Update(ctx context.Context, id int64, update dbRuleUpdate) (*dbRule, error) {
	if mocks.rules.Update != nil {
		return mocks.rules.Update(id, update)
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
	if update.Settings != nil {
		setFields = append(setFields, sqlf.Sprintf("settings=%s", *update.Settings))
	}

	if len(setFields) == 0 {
		return nil, nil
	}

	results, err := s.query(ctx, sqlf.Sprintf(`UPDATE rules SET %v WHERE id=%s RETURNING id, project_id, name, description, settings`, sqlf.Join(setFields, ", "), id))
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errRuleNotFound
	}
	return results[0], nil
}

// GetByID retrieves the rule (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this rule.
func (s dbRules) GetByID(ctx context.Context, id int64) (*dbRule, error) {
	if mocks.rules.GetByID != nil {
		return mocks.rules.GetByID(id)
	}

	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errRuleNotFound
	}
	return results[0], nil
}

// dbRulesListOptions contains options for listing rules.
type dbRulesListOptions struct {
	Query     string // only list rules matching this query (case-insensitively)
	ProjectID int64  // only list rules defined in this project
	*db.LimitOffset
}

func (o dbRulesListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.Query != "" {
		conds = append(conds, sqlf.Sprintf("name LIKE %s", "%"+o.Query+"%"))
	}
	if o.ProjectID != 0 {
		conds = append(conds, sqlf.Sprintf("project_id=%d", o.ProjectID))
	}
	return conds
}

// List lists all rules that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s dbRules) List(ctx context.Context, opt dbRulesListOptions) ([]*dbRule, error) {
	if mocks.rules.List != nil {
		return mocks.rules.List(opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s dbRules) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*dbRule, error) {
	q := sqlf.Sprintf(`
SELECT id, project_id, name, description, settings FROM rules
WHERE (%s)
ORDER BY name ASC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)
	return s.query(ctx, q)
}

func (dbRules) query(ctx context.Context, query *sqlf.Query) ([]*dbRule, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*dbRule
	for rows.Next() {
		var t dbRule
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Name, &t.Description, &t.Settings); err != nil {
			return nil, err
		}
		results = append(results, &t)
	}
	return results, nil
}

// Count counts all rules that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the rules.
func (dbRules) Count(ctx context.Context, opt dbRulesListOptions) (int, error) {
	if mocks.rules.Count != nil {
		return mocks.rules.Count(opt)
	}

	q := sqlf.Sprintf("SELECT COUNT(*) FROM rules WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Delete deletes a rule given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to delete the rule.
func (s dbRules) DeleteByID(ctx context.Context, id int64) error {
	if mocks.rules.DeleteByID != nil {
		return mocks.rules.DeleteByID(id)
	}
	return s.delete(ctx, sqlf.Sprintf("id=%d", id))
}

func (dbRules) delete(ctx context.Context, cond *sqlf.Query) error {
	conds := []*sqlf.Query{cond, sqlf.Sprintf("TRUE")}
	q := sqlf.Sprintf("DELETE FROM rules WHERE (%s)", sqlf.Join(conds, ") AND ("))

	res, err := dbconn.Global.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errRuleNotFound
	}
	return nil
}

// mockRules mocks the rules-related DB operations.
type mockRules struct {
	Create     func(*dbRule) (*dbRule, error)
	Update     func(int64, dbRuleUpdate) (*dbRule, error)
	GetByID    func(int64) (*dbRule, error)
	List       func(dbRulesListOptions) ([]*dbRule, error)
	Count      func(dbRulesListOptions) (int, error)
	DeleteByID func(int64) error
}
