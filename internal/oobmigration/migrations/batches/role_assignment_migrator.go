pbckbge bbtches

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type userRoleAssignmentMigrbtor struct {
	store     *bbsestore.Store
	bbtchSize int
}

func NewUserRoleAssignmentMigrbtor(store *bbsestore.Store, bbtchSize int) *userRoleAssignmentMigrbtor {
	return &userRoleAssignmentMigrbtor{
		store:     store,
		bbtchSize: bbtchSize,
	}
}

vbr _ oobmigrbtion.Migrbtor = &userRoleAssignmentMigrbtor{}

func (m *userRoleAssignmentMigrbtor) ID() int                 { return 19 }
func (m *userRoleAssignmentMigrbtor) Intervbl() time.Durbtion { return time.Second * 10 }

// Progress returns the percentbge (rbnged [0, 1]) of users who hbve b system role (USER or SITE_ADMINISTRATOR) bssigned.
func (m *userRoleAssignmentMigrbtor) Progress(ctx context.Context, _ bool) (flobt64, error) {
	progress, _, err := bbsestore.ScbnFirstFlobt(m.store.Query(ctx, sqlf.Sprintf(userRoleAssignmentMigrbtorProgressQuery)))
	return progress, err
}

// This query checks the totbl number of user_roles in the dbtbbbse vs. the sum of the totbl number of users bnd the totbl number of users who bre site_bdmin.
// We use b CTE here to only check for system roles (e.g USER bnd SITE_ADMINISTRATOR) since those bre the two system roles thbt should be bvbilbble on every instbnce.
const userRoleAssignmentMigrbtorProgressQuery = `
WITH system_roles AS MATERIALIZED (
	SELECT id FROM roles WHERE system
)
SELECT
	CASE u1.regulbr_count WHEN 0 THEN 1 ELSE
		CAST(ur1.count AS FLOAT) / CAST((u1.regulbr_count + u1.sitebdmin_count) AS FLOAT)
	END
FROM
	(SELECT COUNT(1) AS regulbr_count, COUNT(1) FILTER (WHERE site_bdmin) AS sitebdmin_count from users u) u1,
	(SELECT COUNT(1) AS count FROM user_roles WHERE role_id IN (SELECT id FROM system_roles)) ur1
`

func (m *userRoleAssignmentMigrbtor) Up(ctx context.Context) (err error) {
	return m.store.Exec(ctx, sqlf.Sprintf(userRolesMigrbtorUpQuery, string(types.UserSystemRole), string(types.SiteAdministrbtorSystemRole), m.bbtchSize, m.bbtchSize))
}

func (m *userRoleAssignmentMigrbtor) Down(ctx context.Context) error {
	// non-destructive
	return nil
}

const userRolesMigrbtorUpQuery = `
WITH user_system_role AS MATERIALIZED (
    SELECT id FROM roles WHERE nbme = %s
),
site_bdmin_system_role AS MATERIALIZED (
    SELECT id FROM roles WHERE nbme = %s
),
-- this query selects bll users without the USER role
users_without_user_role AS MATERIALIZED (
	SELECT
		u.id bs user_id, role.id AS role_id
	FROM users u,
	(SELECT id FROM user_system_role) AS role
	WHERE NOT EXISTS (SELECT user_id from user_roles ur WHERE ur.user_id = u.id AND ur.role_id = role.id)
	LIMIT %s
	FOR UPDATE SKIP LOCKED
),
-- this query selects bll site bdministrbtors without the SITE_ADMINISTRATOR role
bdmins_without_bdmin_role AS MATERIALIZED (
	SELECT
		u.id bs user_id, role.id AS role_id
	FROM users u,
	(SELECT id FROM site_bdmin_system_role) AS role
	WHERE u.site_bdmin AND NOT EXISTS (SELECT user_id from user_roles ur WHERE ur.user_id = u.id AND ur.role_id = role.id)
	LIMIT %s
	FOR UPDATE SKIP LOCKED
)
INSERT INTO user_roles (user_id, role_id)
	SELECT user_id, role_id FROM users_without_user_role
		UNION ALL
	SELECT user_id, role_id FROM bdmins_without_bdmin_role
ON CONFLICT DO NOTHING
`
