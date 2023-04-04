package batches

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type userRoleAssignmentMigrator struct {
	store     *basestore.Store
	batchSize int
}

func NewUserRoleAssignmentMigrator(store *basestore.Store, batchSize int) *userRoleAssignmentMigrator {
	return &userRoleAssignmentMigrator{
		store:     store,
		batchSize: batchSize,
	}
}

var _ oobmigration.Migrator = &userRoleAssignmentMigrator{}

func (m *userRoleAssignmentMigrator) ID() int                 { return 19 }
func (m *userRoleAssignmentMigrator) Interval() time.Duration { return time.Second * 10 }

// Progress returns the percentage (ranged [0, 1]) of users who have a system role (USER or SITE_ADMINISTRATOR) assigned.
func (m *userRoleAssignmentMigrator) Progress(ctx context.Context, _ bool) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(userRoleAssignmentMigratorProgressQuery)))
	return progress, err
}

// This query checks the total number of user_roles in the database vs. the sum of the total number of users and the total number of users who are site_admin.
// We use a CTE here to only check for system roles (e.g USER and SITE_ADMINISTRATOR) since those are the two system roles that should be available on every instance.
const userRoleAssignmentMigratorProgressQuery = `
WITH system_roles AS MATERIALIZED (
	SELECT id FROM roles WHERE system
)
SELECT
	CASE u1.regular_count WHEN 0 THEN 1 ELSE
		CAST(ur1.count AS FLOAT) / CAST((u1.regular_count + u1.siteadmin_count) AS FLOAT)
	END
FROM
	(SELECT COUNT(1) AS regular_count, COUNT(1) FILTER (WHERE site_admin) AS siteadmin_count from users u) u1,
	(SELECT COUNT(1) AS count FROM user_roles WHERE role_id IN (SELECT id FROM system_roles)) ur1
`

func (m *userRoleAssignmentMigrator) Up(ctx context.Context) (err error) {
	return m.store.Exec(ctx, sqlf.Sprintf(userRolesMigratorUpQuery, string(types.UserSystemRole), string(types.SiteAdministratorSystemRole), m.batchSize, m.batchSize))
}

func (m *userRoleAssignmentMigrator) Down(ctx context.Context) error {
	// non-destructive
	return nil
}

const userRolesMigratorUpQuery = `
WITH user_system_role AS MATERIALIZED (
    SELECT id FROM roles WHERE name = %s
),
site_admin_system_role AS MATERIALIZED (
    SELECT id FROM roles WHERE name = %s
),
-- this query selects all users without the USER role
users_without_user_role AS MATERIALIZED (
	SELECT
		u.id as user_id, role.id AS role_id
	FROM users u,
	(SELECT id FROM user_system_role) AS role
	WHERE NOT EXISTS (SELECT user_id from user_roles ur WHERE ur.user_id = u.id AND ur.role_id = role.id)
	LIMIT %s
	FOR UPDATE SKIP LOCKED
),
-- this query selects all site administrators without the SITE_ADMINISTRATOR role
admins_without_admin_role AS MATERIALIZED (
	SELECT
		u.id as user_id, role.id AS role_id
	FROM users u,
	(SELECT id FROM site_admin_system_role) AS role
	WHERE u.site_admin AND NOT EXISTS (SELECT user_id from user_roles ur WHERE ur.user_id = u.id AND ur.role_id = role.id)
	LIMIT %s
	FOR UPDATE SKIP LOCKED
)
INSERT INTO user_roles (user_id, role_id)
	SELECT user_id, role_id FROM users_without_user_role
		UNION ALL
	SELECT user_id, role_id FROM admins_without_admin_role
ON CONFLICT DO NOTHING
`
