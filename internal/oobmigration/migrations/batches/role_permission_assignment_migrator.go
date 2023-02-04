package batches

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type rolePermissionAssignmentMigrator struct {
	store *basestore.Store
}

func NewRolePermissionAssignmentMigrator(store *basestore.Store) *rolePermissionAssignmentMigrator {
	return &rolePermissionAssignmentMigrator{
		store: store,
	}
}

var _ oobmigration.Migrator = &rolePermissionAssignmentMigrator{}

func (m *rolePermissionAssignmentMigrator) ID() int                 { return 21 }
func (m *rolePermissionAssignmentMigrator) Interval() time.Duration { return time.Second * 10 }

const rolePermissionAssignmentMigratorProgressQuery = `
WITH system_roles AS (
	SELECT id FROM roles WHERE system
)
SELECT
	CASE p.count WHEN 0 THEN 1 ELSE
		-- we multiply the amount of permissions by the amount of system roles
		-- because we expect all permissions to be assigned to all system roles
		CAST(rp.count AS FLOAT) / CAST((p.count * r.count) AS FLOAT)
	END
FROM
	(SELECT COUNT(1) FROM permissions) p,
	(SELECT COUNT(1) FROM system_roles) r,
	(SELECT COUNT(1) FROM role_permissions WHERE role_id IN (SELECT id FROM system_roles)) rp
`

// Progress returns the percentage (ranged [0, 1]) of system roles that have been assigned permissions.
func (m *rolePermissionAssignmentMigrator) Progress(ctx context.Context, _ bool) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(rolePermissionAssignmentMigratorProgressQuery)))
	return progress, err
}

const rolePermissionMigratorUpQuery = `
WITH user_system_role AS MATERIALIZED (
    SELECT id FROM roles WHERE name = %s
),
site_admin_system_role AS MATERIALIZED (
    SELECT id FROM roles WHERE name = %s
)
INSERT INTO role_permissions (permission_id, role_id)
	SELECT id, (SELECT id FROM user_system_role) FROM permissions
		UNION ALL
	SELECT id, (SELECT id FROM site_admin_system_role) FROM permissions
ON CONFLICT DO NOTHING
`

func (m *rolePermissionAssignmentMigrator) Up(ctx context.Context) (err error) {
	return m.store.Exec(ctx, sqlf.Sprintf(rolePermissionMigratorUpQuery, string(types.UserSystemRole), string(types.SiteAdministratorSystemRole)))
}

func (m *rolePermissionAssignmentMigrator) Down(ctx context.Context) error {
	// non-destructive
	return nil
}
