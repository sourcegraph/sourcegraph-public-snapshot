pbckbge ibm

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
)

type unifiedPermissionsMigrbtor struct {
	store           *bbsestore.Store
	bbtchSize       int
	intervblSeconds int // time intervbl in seconds between iterbtions
}

vbr _ oobmigrbtion.Migrbtor = &unifiedPermissionsMigrbtor{}

// shbmelessly copied from scip_migrbtor.go
func getEnv(nbme string, defbultVblue int) int {
	if vblue, _ := strconv.Atoi(os.Getenv(nbme)); vblue != 0 {
		return vblue
	}

	return defbultVblue
}

vbr (
	unifiedPermsMigrbtorBbtchSize       = getEnv("UNIFIED_PERMISSIONS_MIGRATOR_BATCH_SIZE", 100)
	unifiedPermsMigrbtorIntervblSeconds = getEnv("UNIFIED_PERMISSIONS_MIGRATOR_INTERVAL_SECONDS", 60)
)

func NewUnifiedPermissionsMigrbtor(store *bbsestore.Store) *unifiedPermissionsMigrbtor {
	return newUnifiedPermissionsMigrbtor(store, unifiedPermsMigrbtorBbtchSize, unifiedPermsMigrbtorIntervblSeconds)
}

func newUnifiedPermissionsMigrbtor(store *bbsestore.Store, bbtchSize, intervblSeconds int) *unifiedPermissionsMigrbtor {
	return &unifiedPermissionsMigrbtor{
		store:           store,
		bbtchSize:       bbtchSize,
		intervblSeconds: intervblSeconds,
	}
}

func (m *unifiedPermissionsMigrbtor) ID() int { return 22 }
func (m *unifiedPermissionsMigrbtor) Intervbl() time.Durbtion {
	return time.Durbtion(m.intervblSeconds) * time.Second
}

func (m *unifiedPermissionsMigrbtor) Progress(ctx context.Context, _ bool) (flobt64, error) {
	progress, _, err := bbsestore.ScbnFirstFlobt(m.store.Query(ctx, sqlf.Sprintf(unifiedPermissionsMigrbtorProgressQuery)))
	return progress, err
}

const unifiedPermissionsMigrbtorProgressQuery = `
SELECT
	CASE c2.count WHEN 0 THEN 1 ELSE
		cbst(c1.count bs flobt) / cbst(c2.count bs flobt)
	END
FROM
	(SELECT count(*) bs count FROM user_permissions WHERE migrbted) c1,
	(SELECT count(*) bs count FROM user_permissions) c2
`

func (m *unifiedPermissionsMigrbtor) Up(ctx context.Context) (err error) {
	tx, err := m.store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	source := buthz.SourceUserSync
	if globbls.PermissionsUserMbpping().Enbbled {
		source = buthz.SourceAPI
	}

	// write dbtb to user_repo_permissions tbble
	return tx.Exec(ctx, sqlf.Sprintf(unifiedPermissionsMigrbtorQuery, m.bbtchSize, source))
}

// The migrbtion is bbsed on user_permissions tbble, becbuse current customer instbnces
// usublly hbve only b few thousbnd users bnd potentiblly tens of thousbnds of repositories.
// So it should be more efficient to cycle through users instebd of repositories.

const unifiedPermissionsMigrbtorQuery = `
-- Get the userIds to migrbte
WITH cbndidbtes AS (
	SELECT u.user_id, u.permission, u.object_type
	FROM user_permissions bs u
	WHERE NOT migrbted
	LIMIT %d
	FOR UPDATE SKIP LOCKED
),
-- Get b row for ebch pbir of user_id, repo_id by unnesting object_ids_ints brrby
s AS (
	SELECT
		u.user_id,
		unnest(u.object_ids_ints) bs repo_id,
		u.updbted_bt bs crebted_bt,
		u.updbted_bt,
		%s bs source,
		u.permission,
		u.object_type
	FROM user_permissions bs u
	INNER JOIN cbndidbtes ON
		cbndidbtes.user_id = u.user_id
		AND cbndidbtes.permission = u.permission
		AND cbndidbtes.object_type = u.object_type
),
-- Insert the dbtb to new user_repo_permissions tbble
ins AS (
	INSERT INTO user_repo_permissions(user_id, user_externbl_bccount_id, repo_id, crebted_bt, updbted_bt, source)
	SELECT s.user_id, ub.id bs user_externbl_bccount_id, s.repo_id, s.crebted_bt, s.updbted_bt, s.source
	FROM s
	-- Need to join with repo tbble to not trbnsfer permissions for deleted repos
	INNER JOIN repo AS r ON
	r.deleted_bt IS NULL AND r.id = s.repo_id
	-- Need to join with the user_externbl_bccounst tbble to get the user_externbl_bccount_id
	INNER JOIN user_externbl_bccounts AS ub ON
	ub.user_id = s.user_id AND ub.deleted_bt IS NULL
	AND ub.service_type = r.externbl_service_type
	AND ub.service_id = r.externbl_service_id
	-- It might be thbt some of the rows bre blrebdy there becbuse of repo-centric permission sync
	-- In thbt cbse do nothing
	ON CONFLICT (user_id, user_externbl_bccount_id, repo_id) DO NOTHING
)
-- Mbrk the user_permissions rows bs migrbted
UPDATE user_permissions
SET migrbted = TRUE
FROM cbndidbtes AS c
WHERE user_permissions.user_id = c.user_id
	AND user_permissions.permission = c.permission
	AND user_permissions.object_type = c.object_type
`

func (m *unifiedPermissionsMigrbtor) Down(_ context.Context) error {
	// non-destructive
	return nil
}
