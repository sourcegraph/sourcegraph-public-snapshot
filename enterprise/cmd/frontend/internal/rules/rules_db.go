package rules

import (
	"context"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// dbRule describes a rule for a discussion thread.
type dbRule struct {
	ID          int64
	Container   ruleContainer
	Name        string // the name (case-preserving)
	Description *string
	Definition  string
}

// errRuleNotFound occurs when a database operation expects a specific rule to exist but it does not
// exist.
var errRuleNotFound = errors.New("rule not found")

type dbRules struct{}

const selectColumns = `id, container_campaign_id, container_thread_id, name, description, definition, created_at, updated_at`

// Create creates a rule. The rule argument's (Rule).ID field is ignored.
func (dbRules) Create(ctx context.Context, rule *dbRule) (*dbRule, error) {
	if mocks.rules.Create != nil {
		return mocks.rules.Create(rule)
	}

	args := []interface{}{
		rule.Container.Campaign,
		rule.Container.Thread,
		rule.Name,
		rule.Description,
		rule.Definition,
	}
	query := sqlf.Sprintf(
		`INSERT INTO rules(`+selectColumns+`) VALUES(DEFAULT`+strings.Repeat(", %v", len(args))+`, DEFAULT,  DEFAULT) RETURNING `+selectColumns,
		args...,
	)
	return dbRules{}.scanRow(dbconn.Global.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...))
}

type dbRuleUpdate struct {
	Name        *string
	Description *string
	Definition  *string
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
	if update.Definition != nil {
		setFields = append(setFields, sqlf.Sprintf("definition=%s", *update.Definition))
	}

	if len(setFields) == 0 {
		return nil, nil
	}

	results, err := s.query(ctx, sqlf.Sprintf(`UPDATE rules SET %v WHERE id=%s RETURNING `+selectColumns, sqlf.Join(setFields, ", "), id))
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
	Container ruleContainer
	*db.LimitOffset
}

func (o dbRulesListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.Query != "" {
		conds = append(conds, sqlf.Sprintf("name ILIKE %s", "%"+o.Query+"%"))
	}
	addCondition := func(id int64, column string) {
		if id != 0 {
			conds = append(conds, sqlf.Sprintf(column+"=%d", id))
		}
	}
	addCondition(o.Container.Campaign, "container_campaign_id")
	addCondition(o.Container.Thread, "container_thread_id")
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
SELECT `+selectColumns+` FROM rules
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
		t, err := dbRules{}.scanRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, nil
}

func (dbRules) scanRow(row interface {
	Scan(dest ...interface{}) error
}) (*dbRule, error) {
	var t dbRule
	if err := row.Scan(
		&t.ID,
		&t.Container.Campaign,
		&t.Container.Thread,
		&t.Name,
		&t.Description,
		&t.Definition,
		&t.CreatedAt,
		&t.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &t, nil
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
