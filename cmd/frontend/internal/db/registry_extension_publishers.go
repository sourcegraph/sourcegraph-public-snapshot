package db

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
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
	return s.listPublishers(ctx, opt.LimitOffset)
}

func (s *registryExtensions) publishersSQLCTE() *sqlf.Query {
	return sqlf.Sprintf(`WITH publishers AS (
  (SELECT DISTINCT ON (publisher_user_id) publisher_user_id AS user_id, NULL AS org_id FROM registry_extensions WHERE publisher_user_id IS NOT NULL AND deleted_at IS NULL)
  UNION
  (SELECT DISTINCT ON (publisher_org_id) NULL AS user_id, publisher_org_id AS org_id FROM registry_extensions WHERE publisher_org_id IS NOT NULL AND deleted_at IS NULL)
) `)
}

func (s *registryExtensions) listPublishers(ctx context.Context, limitOffset *LimitOffset) ([]*RegistryPublisher, error) {
	q := sqlf.Sprintf(`%s
SELECT user_id, org_id, COALESCE(users.username, orgs.name) AS non_canonical_name FROM publishers
LEFT JOIN users ON users.id=user_id
LEFT JOIN orgs ON orgs.id=org_id
ORDER BY org_id ASC NULLS LAST, user_id ASC NULLS LAST
%s`,
		s.publishersSQLCTE(),
		limitOffset.SQL(),
	)

	rows, err := globalDB.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
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
	if err := globalDB.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
