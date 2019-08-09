package rules

import (
	"context"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/nnz"
)

// DBRule describes a rule for a discussion thread.
type DBRule struct {
	ID          int64
	Container   RuleContainer
	Name        string // the name (case-preserving)
	Description *string
	Definition  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// errRuleNotFound occurs when a database operation expects a specific rule to exist but it does not
// exist.
var errRuleNotFound = errors.New("rule not found")

type DBRules struct{}

const selectColumns = `id, container_campaign_id, container_thread_id, name, description, definition, created_at, updated_at`

// Create creates a rule. The rule argument's (Rule).ID field is ignored.
func (DBRules) Create(ctx context.Context, rule *DBRule) (*DBRule, error) {
	if Mocks.Rules.Create != nil {
		return Mocks.Rules.Create(rule)
	}

	args := []interface{}{
		nnz.Int64(rule.Container.Campaign),
		nnz.Int64(rule.Container.Thread),
		rule.Name,
		rule.Description,
		rule.Definition,
	}
	query := sqlf.Sprintf(
		`INSERT INTO rules(`+selectColumns+`) VALUES(DEFAULT`+strings.Repeat(", %v", len(args))+`, DEFAULT,  DEFAULT) RETURNING `+selectColumns,
		args...,
	)
	return DBRules{}.scanRow(dbconn.Global.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...))
}

type dbRuleUpdate struct {
	Name        *string
	Description *string
	Definition  *string
}

// Update updates a rule given its ID.
func (s DBRules) Update(ctx context.Context, id int64, update dbRuleUpdate) (*DBRule, error) {
	if Mocks.Rules.Update != nil {
		return Mocks.Rules.Update(id, update)
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
func (s DBRules) GetByID(ctx context.Context, id int64) (*DBRule, error) {
	if Mocks.Rules.GetByID != nil {
		return Mocks.Rules.GetByID(id)
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

// DBRulesListOptions contains options for listing rules.
type DBRulesListOptions struct {
	Query     string // only list rules matching this query (case-insensitively)
	Container RuleContainer
	*db.LimitOffset
}

func (o DBRulesListOptions) sqlConditions() []*sqlf.Query {
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
func (s DBRules) List(ctx context.Context, opt DBRulesListOptions) ([]*DBRule, error) {
	if Mocks.Rules.List != nil {
		return Mocks.Rules.List(opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s DBRules) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*DBRule, error) {
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

func (DBRules) query(ctx context.Context, query *sqlf.Query) ([]*DBRule, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*DBRule
	for rows.Next() {
		t, err := DBRules{}.scanRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, nil
}

func (DBRules) scanRow(row interface {
	Scan(dest ...interface{}) error
}) (*DBRule, error) {
	var t DBRule
	if err := row.Scan(
		&t.ID,
		(*nnz.Int64)(&t.Container.Campaign),
		(*nnz.Int64)(&t.Container.Thread),
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
func (DBRules) Count(ctx context.Context, opt DBRulesListOptions) (int, error) {
	if Mocks.Rules.Count != nil {
		return Mocks.Rules.Count(opt)
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
func (s DBRules) DeleteByID(ctx context.Context, id int64) error {
	if Mocks.Rules.DeleteByID != nil {
		return Mocks.Rules.DeleteByID(id)
	}
	return s.delete(ctx, sqlf.Sprintf("id=%d", id))
}

func (DBRules) delete(ctx context.Context, cond *sqlf.Query) error {
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

// mockRules Mocks the rules-related DB operations.
type mockRules struct {
	Create     func(*DBRule) (*DBRule, error)
	Update     func(int64, dbRuleUpdate) (*DBRule, error)
	GetByID    func(int64) (*DBRule, error)
	List       func(DBRulesListOptions) ([]*DBRule, error)
	Count      func(DBRulesListOptions) (int, error)
	DeleteByID func(int64) error
}
