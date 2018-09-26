package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/dbconn"
)

// RegistryPublisher is a publisher of extensions to the registry.
type RegistryPublisher struct {
	UserID, OrgID int32 // exactly 1 is nonzero

	// NonCanonicalExtensionID is the publisher's username (for a user) or name (for an org) as of
	// when the query executed. Do not persist this.
	NonCanonicalName string
}

// IsZero reports whether p is the zero value.
func (p RegistryPublisher) IsZero() bool { return p == RegistryPublisher{} }

// RegistryExtensionPublisherNotFoundError occurs when a registry extension publisher is not found.
type RegistryExtensionPublisherNotFoundError struct {
	args []interface{}
}

// NotFound implements errcode.NotFounder.
func (err RegistryExtensionPublisherNotFoundError) NotFound() bool { return true }

func (err RegistryExtensionPublisherNotFoundError) Error() string {
	return fmt.Sprintf("registry extension publisher not found: %v", err.args)
}

// RegistryPublishersListOptions contains options for listing publishers of extensions in the
// registry.
type RegistryPublishersListOptions struct {
	*LimitOffset
}

func (o RegistryPublishersListOptions) sqlConditions() []*sqlf.Query {
	var conds []*sqlf.Query
	if len(conds) == 0 {
		conds = append(conds, sqlf.Sprintf("TRUE"))
	}
	return conds
}

// ListPublishers lists all publishers of extensions to the registry.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s *registryExtensions) ListPublishers(ctx context.Context, opt RegistryPublishersListOptions) ([]*RegistryPublisher, error) {
	return s.listPublishers(ctx, nil, opt.LimitOffset)
}

func (s *registryExtensions) publishersSQLCTE() *sqlf.Query {
	return sqlf.Sprintf(`WITH publishers AS (
  (SELECT DISTINCT ON (publisher_user_id) publisher_user_id AS user_id, NULL AS org_id FROM registry_extensions WHERE publisher_user_id IS NOT NULL AND deleted_at IS NULL)
  UNION
  (SELECT DISTINCT ON (publisher_org_id) NULL AS user_id, publisher_org_id AS org_id FROM registry_extensions WHERE publisher_org_id IS NOT NULL AND deleted_at IS NULL)
) `)
}

func (s *registryExtensions) listPublishers(ctx context.Context, conds []*sqlf.Query, limitOffset *LimitOffset) ([]*RegistryPublisher, error) {
	conds = append(conds, sqlf.Sprintf("TRUE"))
	q := sqlf.Sprintf(`%s
SELECT user_id, org_id, COALESCE(users.username, orgs.name) AS non_canonical_name FROM publishers
LEFT JOIN users ON users.id=user_id
LEFT JOIN orgs ON orgs.id=org_id
WHERE (%s)
ORDER BY org_id ASC NULLS LAST, user_id ASC NULLS LAST
%s`,
		s.publishersSQLCTE(),
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)

	rows, err := dbconn.Global.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*RegistryPublisher
	for rows.Next() {
		var t RegistryPublisher
		var userID, orgID sql.NullInt64
		if err := rows.Scan(&userID, &orgID, &t.NonCanonicalName); err != nil {
			return nil, err
		}
		t.UserID = int32(userID.Int64)
		t.OrgID = int32(orgID.Int64)
		results = append(results, &t)
	}
	return results, nil
}

// CountPublishers counts all registry publishers that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the results.
func (s *registryExtensions) CountPublishers(ctx context.Context, opt RegistryPublishersListOptions) (int, error) {
	q := sqlf.Sprintf(`%s SELECT COUNT(*) FROM publishers WHERE (%s)`, s.publishersSQLCTE(), sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// GePublisher gets the registry publisher with the given name.
func (s *registryExtensions) GetPublisher(ctx context.Context, name string) (*RegistryPublisher, error) {
	var userID, orgID sql.NullInt64
	var p RegistryPublisher
	q := sqlf.Sprintf(`
WITH publishers AS (
  (SELECT id AS user_id, NULL AS org_id, username AS non_canonical_name FROM users WHERE username=%s AND deleted_at IS NULL)
  UNION
  (SELECT NULL AS user_id, id AS org_id, name AS non_canonical_name FROM orgs WHERE name=%s AND deleted_at IS NULL)
)
SELECT user_id, org_id, non_canonical_name FROM publishers ORDER BY user_id NULLS LAST LIMIT 1
`, name, name)
	err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&userID, &orgID, &p.NonCanonicalName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &RegistryExtensionPublisherNotFoundError{[]interface{}{"name", name}}
		}
		return nil, err
	}
	p.UserID = int32(userID.Int64)
	p.OrgID = int32(orgID.Int64)
	return &p, nil
}
