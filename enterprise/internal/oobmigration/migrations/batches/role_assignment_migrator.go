package batches

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type roleAssignerMigrator struct {
	store     *basestore.Store
	batchSize int
}

func NewRoleAssignmentMigrator(store *basestore.Store, batchSize int) *roleAssignerMigrator {
	return &roleAssignerMigrator{
		store:     store,
		batchSize: batchSize,
	}
}

var _ oobmigration.Migrator = &roleAssignerMigrator{}

func (m *roleAssignerMigrator) ID() int                 { return 19 }
func (m *roleAssignerMigrator) Interval() time.Duration { return time.Second * 10 }

// Progress returns the percentage (ranged [0, 1]) of users wuthout a role.
func (m *roleAssignerMigrator) Progress(ctx context.Context, _ bool) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(roleAssignmentMigratorProgressQuery)))
	return progress, err
}

const roleAssignmentMigratorProgressQuery = `
SELECT 
	CASE u1.regular_count WHEN 0 THEN 1 ELSE
		CAST(ur1.count AS FLOAT) / CAST((u1.regular_count + u1.siteadmin_count) AS FLOAT)
	END
FROM 
	(SELECT COUNT(1) AS regular_count, COUNT(1) FILTER (WHERE site_admin) AS siteadmin_count from users u) u1,
	(SELECT COUNT(1) AS count FROM user_roles) ur1
`

func (m *roleAssignerMigrator) Up(ctx context.Context) (err error) {
	return m.store.Exec(ctx, sqlf.Sprintf(userRolesMigratorUpQuery, m.batchSize))
}

func (m *roleAssignerMigrator) Down(ctx context.Context) error {
	// non-destructive
	return nil
}

const userRolesMigratorUpQuery = `
WITH default_role AS MATERIALIZED (
    SELECT id FROM roles WHERE name = 'DEFAULT'
),
site_admin_role AS MATERIALIZED (
    SELECT id FROM roles WHERE name = 'SITE_ADMINISTRATOR'
),
users_without_roles AS MATERIALIZED (
	SELECT 
		id, site_admin 
	FROM users u 
	WHERE 
		u.deleted_at IS NULL AND u.id NOT IN (SELECT user_id from user_roles)
	LIMIT %s
	FOR UPDATE SKIP LOCKED
)
INSERT INTO user_roles 
	SELECT id, (SELECT id FROM default_role) FROM users_without_roles
		UNION ALL
	SELECT id, (SELECT id FROM site_admin_role) FROM users_without_roles uwr WHERE uwr.site_admin
ON CONFLICT DO NOTHING
`
