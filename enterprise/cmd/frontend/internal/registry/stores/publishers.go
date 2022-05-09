package stores

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

// Publisher is a publisher of extensions to the registry.
type Publisher struct {
	UserID, OrgID int32 // exactly 1 is nonzero

	// NonCanonicalName is the publisher's username (for a user) or name (for an org) as of
	// when the query executed. Do not persist this.
	NonCanonicalName string
}

// IsZero reports whether p is the zero value.
func (p Publisher) IsZero() bool { return p == Publisher{} }

// PublisherNotFoundError occurs when a registry extension publisher is not found.
type PublisherNotFoundError struct {
	args []any
}

// NotFound implements errcode.NotFounder.
func (err PublisherNotFoundError) NotFound() bool { return true }

func (err PublisherNotFoundError) Error() string {
	return fmt.Sprintf("registry extension publisher not found: %v (there is no user or organization with the specified name)", err.args)
}

// PublishersListOptions contains options for listing publishers of extensions in the
// registry.
type PublishersListOptions struct {
	*database.LimitOffset
}

func (o PublishersListOptions) sqlConditions() []*sqlf.Query {
	var conds []*sqlf.Query
	if len(conds) == 0 {
		conds = append(conds, sqlf.Sprintf("TRUE"))
	}
	return conds
}

func (s *extensionStore) ListPublishers(ctx context.Context, opt PublishersListOptions) ([]*Publisher, error) {
	return s.listPublishers(ctx, nil, opt.LimitOffset)
}

func (s *extensionStore) publishersSQLCTE() *sqlf.Query {
	return sqlf.Sprintf(`WITH publishers AS (
  (SELECT DISTINCT ON (publisher_user_id) publisher_user_id AS user_id, NULL AS org_id FROM registry_extensions WHERE publisher_user_id IS NOT NULL AND deleted_at IS NULL)
  UNION
  (SELECT DISTINCT ON (publisher_org_id) NULL AS user_id, publisher_org_id AS org_id FROM registry_extensions WHERE publisher_org_id IS NOT NULL AND deleted_at IS NULL)
) `)
}

func (s *extensionStore) listPublishers(ctx context.Context, conds []*sqlf.Query, limitOffset *database.LimitOffset) ([]*Publisher, error) {
	conds = append(conds, sqlf.Sprintf("TRUE"))
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/frontend/internal/registry/stores/publishers.go:listPublishers
%s
SELECT
	user_id,
	org_id,
	COALESCE(users.username, orgs.name) AS non_canonical_name
FROM publishers
LEFT JOIN users ON users.id=user_id
LEFT JOIN orgs ON orgs.id=org_id
WHERE (%s)
ORDER BY
	org_id ASC NULLS LAST,
	user_id ASC NULLS LAST
%s`,
		s.publishersSQLCTE(),
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var results []*Publisher
	for rows.Next() {
		var t Publisher
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

func (s *extensionStore) CountPublishers(ctx context.Context, opt PublishersListOptions) (int, error) {
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/frontend/internal/registry/stores/publishers.go:CountPublishers
%s
SELECT COUNT(*) FROM publishers WHERE (%s)`, s.publishersSQLCTE(), sqlf.Join(opt.sqlConditions(), ") AND ("))

	var count int
	if err := s.QueryRow(ctx, q).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *extensionStore) GetPublisher(ctx context.Context, name string) (*Publisher, error) {
	var userID, orgID sql.NullInt64
	var p Publisher
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/frontend/internal/registry/stores/publishers.go:GetPublisher
WITH publishers AS (
	(
		SELECT
			id AS user_id,
			NULL AS org_id,
			username AS non_canonical_name
		FROM
			users
		WHERE
			username = %s
			AND
			deleted_at IS NULL
	)
  	UNION
  	(
		SELECT
			NULL AS user_id,
			id AS org_id,
			name AS non_canonical_name
		FROM
			orgs
		WHERE
			name = %s
			AND
			deleted_at IS NULL
	)
)
SELECT
	user_id,
	org_id,
	non_canonical_name
FROM
	publishers
ORDER BY
	user_id NULLS LAST
LIMIT 1
`, name, name)

	err := s.QueryRow(ctx, q).Scan(&userID, &orgID, &p.NonCanonicalName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &PublisherNotFoundError{[]any{"name", name}}
		}
		return nil, err
	}

	p.UserID = int32(userID.Int64)
	p.OrgID = int32(orgID.Int64)

	return &p, nil
}
