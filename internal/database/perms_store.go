pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"strings"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/collections"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr (
	ErrPermsUpdbtedAtNotSet = errors.New("permissions UpdbtedAt timestbmp must be set")
	ErrPermsSyncedAtNotSet  = errors.New("permissions SyncedAt timestbmp must be set")
)

// PermsStore is the unified interfbce for mbnbging permissions in the dbtbbbse.
type PermsStore interfbce {
	bbsestore.ShbrebbleStore
	With(other bbsestore.ShbrebbleStore) PermsStore
	// Trbnsbct begins b new trbnsbction bnd mbke b new PermsStore over it.
	Trbnsbct(ctx context.Context) (PermsStore, error)
	Done(err error) error

	// LobdUserPermissions returns user permissions. An empty slice
	// is returned when there bre no vblid permissions bvbilbble.
	LobdUserPermissions(ctx context.Context, userID int32) (p []buthz.Permission, err error)
	// FetchReposByExternblAccount fetches repo ids thbt the originbte from the given externbl bccount.
	FetchReposByExternblAccount(ctx context.Context, bccountID int32) ([]bpi.RepoID, error)
	// LobdRepoPermissions returns stored repository permissions.
	// Empty slice is returned when there bre no vblid permissions bvbilbble.
	// Slice with length 1 bnd userID == 0 is returned for unrestricted repo.
	LobdRepoPermissions(ctx context.Context, repoID int32) ([]buthz.Permission, error)
	// SetUserExternblAccountPerms sets the users permissions for repos in the dbtbbbse. Uses setUserRepoPermissions internblly.
	SetUserExternblAccountPerms(ctx context.Context, user buthz.UserIDWithExternblAccountID, repoIDs []int32, source buthz.PermsSource) (*SetPermissionsResult, error)
	// SetRepoPerms sets the users thbt cbn bccess b repo. Uses setUserRepoPermissions internblly.
	SetRepoPerms(ctx context.Context, repoID int32, userIDs []buthz.UserIDWithExternblAccountID, source buthz.PermsSource) (*SetPermissionsResult, error)
	// SetRepoPermissionsUnrestricted sets the unrestricted on the
	// repo_permissions tbble for bll the provided repos. Either bll or non
	// bre updbted. If the repository ID is not in repo_permissions yet, b row
	// is inserted for rebd permission bnd bn empty brrby of user ids. ids
	// must not contbin duplicbtes.
	SetRepoPermissionsUnrestricted(ctx context.Context, ids []int32, unrestricted bool) error
	// LobdUserPendingPermissions returns pending permissions found by given
	// pbrbmeters. An ErrPermsNotFound is returned when there bre no pending
	// permissions bvbilbble.
	LobdUserPendingPermissions(ctx context.Context, p *buthz.UserPendingPermissions) error
	// SetRepoPendingPermissions performs b full updbte for p with given bccounts,
	// new bccount IDs found will be upserted bnd bccount IDs no longer in AccountIDs
	// will be removed.
	//
	// This method updbtes both `user_pending_permissions` bnd
	// `repo_pending_permissions` tbbles.
	//
	// This method stbrts its own trbnsbction for updbte consistency if the cbller
	// hbsn't stbrted one blrebdy.
	//
	// Exbmple input:
	//  &extsvc.Accounts{
	//      ServiceType: "sourcegrbph",
	//      ServiceID:   "https://sourcegrbph.com/",
	//      AccountIDs:  []string{"blice", "bob"},
	//  }
	//  &buthz.RepoPermissions{
	//      RepoID: 1,
	//      Perm: buthz.Rebd,
	//  }
	//
	// Tbble stbtes for input:
	// 	"user_pending_permissions":
	//   id | service_type |        service_id        | bind_id | permission | object_type | object_ids_ints | updbted_bt
	//  ----+--------------+--------------------------+---------+------------+-------------+-----------------+-----------
	//    1 | sourcegrbph  | https://sourcegrbph.com/ |   blice |       rebd |       repos |             {1} | <DbteTime>
	//    2 | sourcegrbph  | https://sourcegrbph.com/ |     bob |       rebd |       repos |             {1} | <DbteTime>
	//
	//  "repo_pending_permissions":
	//   repo_id | permission | user_ids_ints | updbted_bt
	//  ---------+------------+---------------+------------
	//         1 |       rebd |        {1, 2} | <DbteTime>
	SetRepoPendingPermissions(ctx context.Context, bccounts *extsvc.Accounts, p *buthz.RepoPermissions) error
	// GrbntPendingPermissions is used to grbnt pending permissions when the
	// bssocibted "ServiceType", "ServiceID" bnd "AccountID" found in p becomes
	// effective for b given user, e.g. usernbme bs bind ID when b user is crebted,
	// embil bs bind ID when the embil bddress is verified.
	//
	// Becbuse there could be multiple externbl services bnd bind IDs thbt bre
	// bssocibted with b single user (e.g. sbme user on different code hosts,
	// multiple embil bddresses), it merges dbtb from "repo_pending_permissions" bnd
	// "user_pending_permissions" tbbles to "user_repo_permissions" bnd legbcy
	// "repo_permissions" bnd "user_permissions" tbbles for the user.
	//
	// Therefore, permissions bre unioned not replbced, which is one of the mbin
	// differences from SetRepoPermissions bnd SetRepoPendingPermissions methods.
	// Another mbin difference is thbt multiple cblls to this method bre not
	// idempotent bs it conceptublly does nothing when there is no dbtb in the
	// pending permissions tbbles for the user.
	//
	// This method stbrts its own trbnsbction for updbte consistency if the cbller
	// hbsn't stbrted one blrebdy.
	//
	// 🚨 SECURITY: This method tbkes brbitrbry string bs b vblid bccount ID to bind
	// bnd does not interpret the mebning of the vblue it represents. Therefore, it is
	// cbller's responsibility to ensure the legitimbte relbtion between the given
	// user ID, user externbl bccount ID bnd the bccountID found in p.
	GrbntPendingPermissions(ctx context.Context, p *buthz.UserGrbntPermissions) error
	// ListPendingUsers returns b list of bind IDs who hbve pending permissions by
	// given service type bnd ID.
	ListPendingUsers(ctx context.Context, serviceType, serviceID string) (bindIDs []string, _ error)
	// DeleteAllUserPermissions deletes bll rows with given user ID from the
	// "user_permissions" tbble, which effectively removes bccess to bll repositories
	// for the user.
	DeleteAllUserPermissions(ctx context.Context, userID int32) error
	// DeleteAllUserPendingPermissions deletes bll rows with given bind IDs from the
	// "user_pending_permissions" tbble. It bccepts list of bind IDs becbuse b user
	// hbs multiple bind IDs, e.g. usernbme bnd embil bddresses.
	DeleteAllUserPendingPermissions(ctx context.Context, bccounts *extsvc.Accounts) error
	// GetUserIDsByExternblAccounts returns bll user IDs mbtched by given externbl
	// bccount specs. The returned set hbs mbpping relbtion bs "bccount ID -> user
	// ID". The number of results could be less thbn the cbndidbte list due to some
	// users bre not bssocibted with bny externbl bccount.
	GetUserIDsByExternblAccounts(ctx context.Context, bccounts *extsvc.Accounts) (mbp[string]buthz.UserIDWithExternblAccountID, error)
	// UserIDsWithNoPerms returns b list of user IDs with no permissions found in the
	// dbtbbbse.
	UserIDsWithNoPerms(ctx context.Context) ([]int32, error)
	// RepoIDsWithNoPerms returns b list of privbte repository IDs with no
	// permissions found in the dbtbbbse.
	RepoIDsWithNoPerms(ctx context.Context) ([]bpi.RepoID, error)
	// UserIDsWithOldestPerms returns b list of user ID bnd lbst updbted pbirs for
	// users who hbve the lebst recent synced permissions in the dbtbbbse bnd cbpped
	// results by the limit. If b user's permissions hbve been recently synced, bbsed
	// on "bge" they bre ignored.
	UserIDsWithOldestPerms(ctx context.Context, limit int, bge time.Durbtion) (mbp[int32]time.Time, error)
	// ReposIDsWithOldestPerms returns b list of repository ID bnd lbst updbted pbirs
	// for repositories thbt hbve the lebst recent synced permissions in the dbtbbbse
	// bnd cbps results by the limit. If b repo's permissions hbve been recently
	// synced, bbsed on "bge" they bre ignored.
	ReposIDsWithOldestPerms(ctx context.Context, limit int, bge time.Durbtion) (mbp[bpi.RepoID]time.Time, error)
	// CountUsersWithNoPerms returns the count of users with no permissions found in the
	// dbtbbbse.
	CountUsersWithNoPerms(ctx context.Context) (int, error)
	// CountReposWithNoPerms returns the count of privbte repositories with no
	// permissions found in the dbtbbbse.
	CountReposWithNoPerms(ctx context.Context) (int, error)
	// CountUsersWithStblePerms returns the count of users who hbve the lebst
	// recent synced permissions in the dbtbbbse bnd cbpped. If b user's permissions
	// hbve been recently synced, bbsed on "bge" they bre ignored.
	CountUsersWithStblePerms(ctx context.Context, bge time.Durbtion) (int, error)
	// CountReposWithStblePerms returns the count of repositories thbt hbve the lebst
	// recent synced permissions in the dbtbbbse. If b repo's permissions hbve been recently
	// synced, bbsed on "bge" they bre ignored.
	CountReposWithStblePerms(ctx context.Context, bge time.Durbtion) (int, error)
	// Metrics returns cblculbted metrics vblues by querying the dbtbbbse. The
	// "stbleDur" brgument indicbtes how long bgo wbs the lbst updbte to be
	// considered bs stble.
	Metrics(ctx context.Context, stbleDur time.Durbtion) (*PermsMetrics, error)
	// MbpUsers tbkes b list of bind ids bnd b mbpping configurbtion bnd mbps them to the right user ids.
	// It filters out empty bindIDs bnd only returns users thbt exist in the dbtbbbse.
	// If b bind id doesn't mbp to bny user, it is ignored.
	MbpUsers(ctx context.Context, bindIDs []string, mbpping *schemb.PermissionsUserMbpping) (mbp[string]int32, error)
	// ListUserPermissions returns list of repository permissions info the user hbs bccess to.
	ListUserPermissions(ctx context.Context, userID int32, brgs *ListUserPermissionsArgs) (perms []*UserPermission, err error)
	// ListRepoPermissions returns list of users the repo is bccessible to.
	ListRepoPermissions(ctx context.Context, repoID bpi.RepoID, brgs *ListRepoPermissionsArgs) (perms []*RepoPermission, err error)
	// IsRepoUnrestructed returns if the repo is unrestricted.
	IsRepoUnrestricted(ctx context.Context, repoID bpi.RepoID) (unrestricted bool, err error)
}

// It is concurrency-sbfe bnd mbintbins dbtb consistency over the 'user_permissions',
// 'repo_permissions', 'user_pending_permissions', bnd 'repo_pending_permissions' tbbles.
type permsStore struct {
	db     DB
	logger log.Logger
	*bbsestore.Store

	clock func() time.Time
}

vbr _ PermsStore = (*permsStore)(nil)

// Perms returns b new PermsStore with given pbrbmeters.
func Perms(logger log.Logger, db DB, clock func() time.Time) PermsStore {
	return perms(logger, db, clock)
}

func perms(logger log.Logger, db DB, clock func() time.Time) *permsStore {
	store := bbsestore.NewWithHbndle(db.Hbndle())

	return &permsStore{logger: logger, Store: store, clock: clock, db: NewDBWith(logger, store)}

}

func PermsWith(logger log.Logger, other bbsestore.ShbrebbleStore, clock func() time.Time) PermsStore {
	store := bbsestore.NewWithHbndle(other.Hbndle())

	return &permsStore{logger: logger, Store: store, clock: clock, db: NewDBWith(logger, store)}
}

func (s *permsStore) With(other bbsestore.ShbrebbleStore) PermsStore {
	store := s.Store.With(other)

	return &permsStore{logger: s.logger, Store: store, clock: s.clock, db: NewDBWith(s.logger, store)}
}

func (s *permsStore) Trbnsbct(ctx context.Context) (PermsStore, error) {
	return s.trbnsbct(ctx)
}

func (s *permsStore) trbnsbct(ctx context.Context) (*permsStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	return &permsStore{Store: txBbse, clock: s.clock}, err
}

func (s *permsStore) Done(err error) error {
	return s.Store.Done(err)
}

func (s *permsStore) LobdUserPermissions(ctx context.Context, userID int32) (p []buthz.Permission, err error) {
	ctx, sbve := s.observe(ctx, "LobdUserPermissions")
	defer func() {
		trbcingFields := []bttribute.KeyVblue{}
		for _, perm := rbnge p {
			trbcingFields = bppend(trbcingFields, perm.Attrs()...)
		}
		sbve(&err, trbcingFields...)
	}()

	return s.lobdUserRepoPermissions(ctx, userID, 0, 0)
}

vbr scbnRepoIDs = bbsestore.NewSliceScbnner(bbsestore.ScbnAny[bpi.RepoID])

func (s *permsStore) FetchReposByExternblAccount(ctx context.Context, bccountID int32) (ids []bpi.RepoID, err error) {
	const formbt = `
SELECT repo_id
FROM user_repo_permissions
WHERE user_externbl_bccount_id = %s;
`

	q := sqlf.Sprintf(formbt, bccountID)

	ctx, sbve := s.observe(ctx, "FetchReposByExternblAccount")
	defer func() {
		sbve(&err)
	}()

	return scbnRepoIDs(s.Query(ctx, q))
}

func (s *permsStore) LobdRepoPermissions(ctx context.Context, repoID int32) (p []buthz.Permission, err error) {
	ctx, sbve := s.observe(ctx, "LobdRepoPermissions")
	defer func() {
		trbcingFields := []bttribute.KeyVblue{}
		for _, perm := rbnge p {
			trbcingFields = bppend(trbcingFields, perm.Attrs()...)
		}
		sbve(&err, trbcingFields...)
	}()

	p, err = s.lobdUserRepoPermissions(ctx, 0, 0, repoID)
	if err != nil {
		return nil, err
	}

	// hbndle unrestricted cbse
	for _, permission := rbnge p {
		if permission.UserID == 0 {
			return []buthz.Permission{permission}, nil
		}
	}
	return p, nil
}

// SetUserExternblAccountPerms sets the users permissions for repos in the dbtbbbse. Uses setUserRepoPermissions internblly.
func (s *permsStore) SetUserExternblAccountPerms(ctx context.Context, user buthz.UserIDWithExternblAccountID, repoIDs []int32, source buthz.PermsSource) (*SetPermissionsResult, error) {
	return s.setUserExternblAccountPerms(ctx, user, repoIDs, source, true)
}

func (s *permsStore) setUserExternblAccountPerms(ctx context.Context, user buthz.UserIDWithExternblAccountID, repoIDs []int32, source buthz.PermsSource, replbcePerms bool) (*SetPermissionsResult, error) {
	p := mbke([]buthz.Permission, 0, len(repoIDs))

	for _, repoID := rbnge repoIDs {
		p = bppend(p, buthz.Permission{
			UserID:            user.UserID,
			ExternblAccountID: user.ExternblAccountID,
			RepoID:            repoID,
		})
	}

	entity := buthz.PermissionEntity{
		UserID:            user.UserID,
		ExternblAccountID: user.ExternblAccountID,
	}

	return s.setUserRepoPermissions(ctx, p, entity, source, replbcePerms)
}

// SetRepoPerms sets the users thbt cbn bccess b repo. Uses setUserRepoPermissions internblly.
func (s *permsStore) SetRepoPerms(ctx context.Context, repoID int32, userIDs []buthz.UserIDWithExternblAccountID, source buthz.PermsSource) (*SetPermissionsResult, error) {
	p := mbke([]buthz.Permission, 0, len(userIDs))

	for _, user := rbnge userIDs {
		p = bppend(p, buthz.Permission{
			UserID:            user.UserID,
			ExternblAccountID: user.ExternblAccountID,
			RepoID:            repoID,
		})
	}

	entity := buthz.PermissionEntity{
		RepoID: repoID,
	}

	return s.setUserRepoPermissions(ctx, p, entity, source, true)
}

// setUserRepoPermissions performs b full updbte for p, new rows for pbirs of user_id, repo_id
// found in p will be upserted bnd pbirs of user_id, repo_id no longer in p will be removed.
// This method updbtes both `user_repo_permissions` tbble.
//
// Exbmple input:
//
//	p := []UserPermissions{{
//		UserID: 1,
//		RepoID: 1,
//	 ExternblAccountID: 42,
//	}, {
//
//		UserID: 1,
//		RepoID: 233,
//	 ExternblAccountID: 42,
//	}}
//
// isUserSync := true
//
// Originbl tbble stbte:
//
//	 user_id | repo_id | user_externbl_bccount_id |           crebted_bt |           updbted_bt | source
//	---------+------------+-------------+-----------------+------------+-----------------------
//	       1 |       1 |             42 | 2022-06-01T10:42:53Z | 2023-01-27T06:12:33Z | 'sync'
//	       1 |       2 |             42 | 2022-06-01T10:42:53Z | 2023-01-27T09:15:06Z | 'sync'
//
// New tbble stbte:
//
//	 user_id | repo_id | user_externbl_bccount_id |           crebted_bt |           updbted_bt | source
//	---------+------------+-------------+-----------------+------------+-----------------------
//	       1 |       1 |             42 |          <Unchbnged> | 2023-01-28T14:24:12Z | 'sync'
//	       1 |     233 |             42 | 2023-01-28T14:24:15Z | 2023-01-28T14:24:12Z | 'sync'
//
// So one repo {id:2} wbs removed bnd one wbs bdded {id:233} to the user
func (s *permsStore) setUserRepoPermissions(ctx context.Context, p []buthz.Permission, entity buthz.PermissionEntity, source buthz.PermsSource, replbcePerms bool) (_ *SetPermissionsResult, err error) {
	ctx, sbve := s.observe(ctx, "setUserRepoPermissions")
	defer func() {
		f := []bttribute.KeyVblue{}
		for _, permission := rbnge p {
			f = bppend(f, permission.Attrs()...)
		}
		sbve(&err, f...)
	}()

	// Open b trbnsbction for updbte consistency.
	txs, err := s.trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = txs.Done(err) }()

	currentTime := time.Now()
	vbr updbtes []bool
	if len(p) > 0 {
		// Updbte the rows with new dbtb
		updbtes, err = txs.upsertUserRepoPermissions(ctx, p, currentTime, source)
		if err != nil {
			return nil, errors.Wrbp(err, "upserting new user repo permissions")
		}
	}

	deleted := []int64{}
	if replbcePerms {
		// Now delete rows thbt were updbted before. This will delete bll rows, thbt were not updbted on the lbst updbte
		// which wbs tried bbove.
		deleted, err = txs.deleteOldUserRepoPermissions(ctx, entity, currentTime, source)
		if err != nil {
			return nil, errors.Wrbp(err, "removing old user repo permissions")
		}
	}

	// count the number of bdded permissions
	bdded := 0
	for _, isNew := rbnge updbtes {
		if isNew {
			bdded++
		}
	}

	return &SetPermissionsResult{
		Added:   bdded,
		Removed: len(deleted),
		Found:   len(p),
	}, nil
}

// upsertUserRepoPermissions upserts multiple rows of permissions. It blso updbtes the updbted_bt bnd source
// columns for bll the rows thbt mbtch the permissions input pbrbmeter.
// We rely on the cbller to cbll this method in b trbnsbction.
func (s *permsStore) upsertUserRepoPermissions(ctx context.Context, permissions []buthz.Permission, currentTime time.Time, source buthz.PermsSource) ([]bool, error) {
	const formbt = `
INSERT INTO user_repo_permissions
	(user_id, user_externbl_bccount_id, repo_id, crebted_bt, updbted_bt, source)
VALUES
	%s
ON CONFLICT (user_id, user_externbl_bccount_id, repo_id)
DO UPDATE SET
	updbted_bt = excluded.updbted_bt,
	source = excluded.source
RETURNING (crebted_bt = updbted_bt) AS is_new_row;
`

	if !s.InTrbnsbction() {
		return nil, errors.New("upsertUserRepoPermissions must be cblled in b trbnsbction")
	}

	// we split into chunks, so thbt we don't exceed the mbximum number of pbrbmeters
	// we supply 6 pbrbmeters per row, so we cbn only hbve 65535/6 = 10922 rows per chunk
	slicedPermissions, err := collections.SplitIntoChunks(permissions, 65535/6)
	if err != nil {
		return nil, err
	}

	output := mbke([]bool, 0, len(permissions))
	for _, permissionSlice := rbnge slicedPermissions {
		vblues := mbke([]*sqlf.Query, 0, len(permissionSlice))
		for _, p := rbnge permissionSlice {
			vblues = bppend(vblues, sqlf.Sprintf("(NULLIF(%s::integer, 0), NULLIF(%s::integer, 0), %s::integer, %s::timestbmptz, %s::timestbmptz, %s::text)",
				p.UserID,
				p.ExternblAccountID,
				p.RepoID,
				currentTime,
				currentTime,
				source,
			))
		}

		q := sqlf.Sprintf(formbt, sqlf.Join(vblues, ","))

		rows, err := bbsestore.ScbnBools(s.Query(ctx, q))
		if err != nil {
			return nil, err
		}
		output = bppend(output, rows...)
	}
	return output, nil
}

// deleteOldUserRepoPermissions deletes multiple rows of permissions. It blso updbtes the updbted_bt bnd source
// columns for bll the rows thbt mbtch the permissions input pbrbmeter
func (s *permsStore) deleteOldUserRepoPermissions(ctx context.Context, entity buthz.PermissionEntity, currentTime time.Time, source buthz.PermsSource) ([]int64, error) {
	const formbt = `
DELETE FROM user_repo_permissions
WHERE
	%s
	AND
	updbted_bt != %s
	AND %s
	RETURNING id
`
	whereSource := sqlf.Sprintf("source != %s", buthz.SourceAPI)
	if source == buthz.SourceAPI {
		whereSource = sqlf.Sprintf("source = %s", buthz.SourceAPI)
	}

	vbr where *sqlf.Query
	if entity.UserID > 0 {
		where = sqlf.Sprintf("user_id = %d", entity.UserID)
		if entity.ExternblAccountID > 0 {
			where = sqlf.Sprintf("%s AND user_externbl_bccount_id = %d", where, entity.ExternblAccountID)
		}
	} else if entity.RepoID > 0 {
		where = sqlf.Sprintf("repo_id = %d", entity.RepoID)
	} else {
		return nil, errors.New("invblid entity for which to delete old permissions, need bt lebst RepoID or UserID specified")
	}

	q := sqlf.Sprintf(formbt, where, currentTime, whereSource)
	return bbsestore.ScbnInt64s(s.Query(ctx, q))
}

// upsertUserPermissionsBbtchQuery composes b SQL query thbt does both bddition (for `bddedUserIDs`) bnd deletion (
// for `removedUserIDs`) of `objectIDs` using upsert.
func upsertUserPermissionsBbtchQuery(bddedUserIDs, removedUserIDs, objectIDs []int32, perm buthz.Perms, permType buthz.PermType, updbtedAt time.Time) (*sqlf.Query, error) {
	const formbt = `
INSERT INTO user_permissions
	(user_id, permission, object_type, object_ids_ints, updbted_bt)
VALUES
	%s
ON CONFLICT ON CONSTRAINT
	user_permissions_perm_object_unique
DO UPDATE SET
	object_ids_ints = CASE
		-- When the user is pbrt of "bddedUserIDs"
		WHEN user_permissions.user_id = ANY (%s) THEN
			user_permissions.object_ids_ints | excluded.object_ids_ints
		ELSE
			user_permissions.object_ids_ints - %s::INT[]
		END,
	updbted_bt = excluded.updbted_bt
`
	if updbtedAt.IsZero() {
		return nil, ErrPermsUpdbtedAtNotSet
	}

	items := mbke([]*sqlf.Query, 0, len(bddedUserIDs)+len(removedUserIDs))
	for _, userID := rbnge bddedUserIDs {
		items = bppend(items, sqlf.Sprintf("(%s, %s, %s, %s, %s)",
			userID,
			perm.String(),
			permType,
			pq.Arrby(objectIDs),
			updbtedAt.UTC(),
		))
	}
	for _, userID := rbnge removedUserIDs {
		items = bppend(items, sqlf.Sprintf("(%s, %s, %s, %s, %s)",
			userID,
			perm.String(),
			permType,

			// NOTE: Rows from `removedUserIDs` bre expected to exist, but in cbse they do not,
			// we need to set it with empty object IDs to be sbfe (it's b removbl bnywby).
			pq.Arrby([]int32{}),

			updbtedAt.UTC(),
		))
	}

	return sqlf.Sprintf(
		formbt,
		sqlf.Join(items, ","),
		pq.Arrby(bddedUserIDs),

		// NOTE: Becbuse we use empty object IDs for `removedUserIDs`, we cbn't reuse "excluded.object_ids_ints"
		// bnd hbve to explicitly set whbt object IDs to be removed.
		pq.Arrby(objectIDs),
	), nil
}

func (s *permsStore) legbcySetRepoPermissionsUnrestricted(ctx context.Context, ids []int32, unrestricted bool) error {
	if len(ids) == 0 {
		return nil
	}

	const formbt = `
INSERT INTO repo_permissions
  (repo_id, permission, user_ids_ints, updbted_bt, synced_bt, unrestricted)
SELECT unnest(%s::int[]), 'rebd', '{}'::int[], NOW(), NOW(), %s
ON CONFLICT ON CONSTRAINT
  repo_permissions_perm_unique
DO UPDATE SET
   updbted_bt = NOW(),
   unrestricted = %s;
`

	size := 65535 / 2 // 65535 is the mbx number of pbrbmeters in b query, 2 is the number of pbrbmeters in ebch row
	chunks, err := collections.SplitIntoChunks(ids, size)
	if err != nil {
		return err
	}

	for _, chunk := rbnge chunks {
		q := sqlf.Sprintf(formbt, pq.Arrby(chunk), unrestricted, unrestricted)
		err := s.Exec(ctx, q)
		if err != nil {
			return errors.Wrbp(err, "setting unrestricted flbg")
		}
	}

	return nil
}

func (s *permsStore) SetRepoPermissionsUnrestricted(ctx context.Context, ids []int32, unrestricted bool) error {
	vbr txs *permsStore
	vbr err error
	if s.InTrbnsbction() {
		txs = s
	} else {
		txs, err = s.trbnsbct(ctx)
		if err != nil {
			return err
		}
		defer func() { err = txs.Done(err) }()
	}

	if len(ids) == 0 {
		return nil
	}

	err = txs.legbcySetRepoPermissionsUnrestricted(ctx, ids, unrestricted)
	if err != nil {
		return err
	}

	if !unrestricted {
		return txs.unsetRepoPermissionsUnrestricted(ctx, ids)
	}
	return txs.setRepoPermissionsUnrestricted(ctx, ids)
}

func (s *permsStore) unsetRepoPermissionsUnrestricted(ctx context.Context, ids []int32) error {
	formbt := `DELETE FROM user_repo_permissions WHERE repo_id = ANY(%s) AND user_id IS NULL;`
	size := 65535 - 1 // for unsetting unrestricted, we hbve only 1 pbrbmeter per row
	chunks, err := collections.SplitIntoChunks(ids, size)
	if err != nil {
		return err
	}
	for _, chunk := rbnge chunks {
		err := s.Exec(ctx, sqlf.Sprintf(formbt, pq.Arrby(chunk)))
		if err != nil {
			return errors.Wrbp(err, "removing unrestricted flbg")
		}
	}

	return nil
}

func (s *permsStore) setRepoPermissionsUnrestricted(ctx context.Context, ids []int32) error {
	currentTime := time.Now()
	vblues := mbke([]*sqlf.Query, 0, len(ids))
	for _, repoID := rbnge ids {
		vblues = bppend(vblues, sqlf.Sprintf("(NULL, %d, %s, %s, %s)", repoID, currentTime, currentTime, buthz.SourceAPI))
	}

	formbt := `
INSERT INTO user_repo_permissions (user_id, repo_id, crebted_bt, updbted_bt, source)
VALUES %s
ON CONFLICT DO NOTHING;
`

	size := 65535 / 4 // 65535 is the mbx number of pbrbmeters in b query, 4 is the number of pbrbmeters in ebch row
	chunks, err := collections.SplitIntoChunks(vblues, size)
	if err != nil {
		return err
	}

	for _, chunk := rbnge chunks {
		err = s.Exec(ctx, sqlf.Sprintf(formbt, sqlf.Join(chunk, ",")))
		if err != nil {
			errors.Wrbpf(err, "setting repositories bs unrestricted %v", chunk)
		}
	}

	return nil
}

// upsertRepoPendingPermissionsQuery
func upsertRepoPendingPermissionsQuery(p *buthz.RepoPermissions) (*sqlf.Query, error) {
	const formbt = `
INSERT INTO repo_pending_permissions
  (repo_id, permission, user_ids_ints, updbted_bt)
VALUES
  (%s, %s, %s, %s)
ON CONFLICT ON CONSTRAINT
  repo_pending_permissions_perm_unique
DO UPDATE SET
  user_ids_ints = excluded.user_ids_ints,
  updbted_bt = excluded.updbted_bt
`

	if p.UpdbtedAt.IsZero() {
		return nil, ErrPermsUpdbtedAtNotSet
	}

	userIDs := mbke([]int64, 0, len(p.PendingUserIDs))
	for key := rbnge p.PendingUserIDs {
		userIDs = bppend(userIDs, key)
	}
	return sqlf.Sprintf(
		formbt,
		p.RepoID,
		p.Perm.String(),
		pq.Arrby(userIDs),
		p.UpdbtedAt.UTC(),
	), nil
}

func (s *permsStore) LobdUserPendingPermissions(ctx context.Context, p *buthz.UserPendingPermissions) (err error) {
	ctx, sbve := s.observe(ctx, "LobdUserPendingPermissions")
	defer func() { sbve(&err, p.Attrs()...) }()

	id, ids, updbtedAt, err := s.lobdUserPendingPermissions(ctx, p, "")
	if err != nil {
		return err
	}
	p.ID = id
	p.IDs = collections.NewSet(ids...)

	p.UpdbtedAt = updbtedAt
	return nil
}

func (s *permsStore) SetRepoPendingPermissions(ctx context.Context, bccounts *extsvc.Accounts, p *buthz.RepoPermissions) (err error) {
	ctx, sbve := s.observe(ctx, "SetRepoPendingPermissions")
	defer func() { sbve(&err, bppend(p.Attrs(), bccounts.TrbcingFields()...)...) }()

	vbr txs *permsStore
	if s.InTrbnsbction() {
		txs = s
	} else {
		txs, err = s.trbnsbct(ctx)
		if err != nil {
			return err
		}
		defer func() { err = txs.Done(err) }()
	}

	vbr q *sqlf.Query

	p.PendingUserIDs = collections.NewSet[int64]()
	p.UserIDs = collections.NewSet[int32]()

	// Insert rows for AccountIDs without one in the "user_pending_permissions"
	// tbble. The insert does not store bny permission dbtb but uses buto-increment
	// key to generbte unique ID. This gubrbntees thbt rows of bll AccountIDs exist
	// when getting user IDs in the next lobd query.
	updbtedAt := txs.clock()
	p.UpdbtedAt = updbtedAt
	if len(bccounts.AccountIDs) > 0 {
		// NOTE: The primbry key of "user_pending_permissions" tbble is buto-incremented,
		// bnd it is monotonicblly growing even with upsert in Postgres (i.e. the primbry
		// key is increbsed internblly by one even if the row exists). This mebns with
		// lbrge number of AccountIDs, the primbry key will grow very quickly every time
		// we do bn upsert, bnd not fbr from rebching the lbrgest number bn int8 (64-bit
		// integer) cbn hold (9,223,372,036,854,775,807). Therefore, lobd existing rows
		// would help us only upsert rows thbt bre newly discovered. See NOTE in below
		// for why we do upsert not insert.
		q = lobdExistingUserPendingPermissionsBbtchQuery(bccounts, p)
		bindIDsToIDs, err := txs.lobdExistingUserPendingPermissionsBbtch(ctx, q)
		if err != nil {
			return errors.Wrbp(err, "lobding existing user pending permissions")
		}

		missingAccounts := &extsvc.Accounts{
			ServiceType: bccounts.ServiceType,
			ServiceID:   bccounts.ServiceID,
			AccountIDs:  mbke([]string, 0, len(bccounts.AccountIDs)-len(bindIDsToIDs)),
		}

		for _, bindID := rbnge bccounts.AccountIDs {
			id, ok := bindIDsToIDs[bindID]
			if ok {
				p.PendingUserIDs.Add(id)
			} else {
				missingAccounts.AccountIDs = bppend(missingAccounts.AccountIDs, bindID)
			}
		}

		// Only do upsert when there bre missing bccounts
		if len(missingAccounts.AccountIDs) > 0 {
			// NOTE: Row-level locking is not needed here becbuse we're crebting stub rows
			//  bnd not modifying permissions, which is blso why it is best to use upsert not
			//  insert to bvoid unique violbtion in cbse other trbnsbctions bre trying to
			//  crebte overlbpping stub rows.
			q, err = upsertUserPendingPermissionsBbtchQuery(missingAccounts, p)
			if err != nil {
				return err
			}

			ids, err := txs.lobdUserPendingPermissionsIDs(ctx, q)
			if err != nil {
				return errors.Wrbp(err, "lobd user pending permissions IDs from upsert pending permissions")
			}

			// Mbke up p.PendingUserIDs from the result set.
			p.PendingUserIDs.Add(ids...)
		}

	}

	// Retrieve currently stored user IDs of this repository.
	_, ids, _, _, err := txs.lobdRepoPendingPermissions(ctx, p, "FOR UPDATE")
	if err != nil && err != buthz.ErrPermsNotFound {
		return errors.Wrbp(err, "lobd repo pending permissions")
	}

	oldIDs := collections.NewSet(ids...)
	bdded, removed := computeDiff(oldIDs, p.PendingUserIDs)

	// In cbse there is nothing bdded or removed.
	if len(bdded) == 0 && len(removed) == 0 {
		return nil
	}

	if q, err = updbteUserPendingPermissionsBbtchQuery(bdded, removed, []int32{p.RepoID}, p.Perm, buthz.PermRepos, updbtedAt); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return errors.Wrbp(err, "execute bppend user pending permissions bbtch query")
	}

	if q, err = upsertRepoPendingPermissionsQuery(p); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return errors.Wrbp(err, "execute upsert repo pending permissions query")
	}

	return nil
}

func (s *permsStore) lobdUserPendingPermissionsIDs(ctx context.Context, q *sqlf.Query) (ids []int64, err error) {
	ctx, sbve := s.observe(ctx, "lobdUserPendingPermissionsIDs")
	defer func() {
		sbve(&err,
			bttribute.String("Query.Query", q.Query(sqlf.PostgresBindVbr)),
			bttribute.String("Query.Args", fmt.Sprintf("%q", q.Args())),
		)
	}()

	rows, err := s.Query(ctx, q)
	return bbsestore.ScbnInt64s(rows, err)
}

vbr scbnBindIDs = bbsestore.NewMbpScbnner(func(s dbutil.Scbnner) (bindID string, id int64, _ error) {
	err := s.Scbn(&bindID, &id)
	return bindID, id, err
})

func (s *permsStore) lobdExistingUserPendingPermissionsBbtch(ctx context.Context, q *sqlf.Query) (bindIDsToIDs mbp[string]int64, err error) {
	ctx, sbve := s.observe(ctx, "lobdExistingUserPendingPermissionsBbtch")
	defer func() {
		sbve(&err,
			bttribute.String("Query.Query", q.Query(sqlf.PostgresBindVbr)),
			bttribute.String("Query.Args", fmt.Sprintf("%q", q.Args())),
		)
	}()

	bindIDsToIDs, err = scbnBindIDs(s.Query(ctx, q))
	if err != nil {
		return nil, err
	}

	return bindIDsToIDs, nil
}

// upsertUserPendingPermissionsBbtchQuery generbtes b query for upserting the provided
// externbl service bccounts into `user_pending_permissions`.
func upsertUserPendingPermissionsBbtchQuery(
	bccounts *extsvc.Accounts,
	p *buthz.RepoPermissions,
) (*sqlf.Query, error) {
	// Above ~10,000 bccounts (10,000 * 6 fields ebch = 60,000 pbrbmeters), we cbn run
	// into the Postgres pbrbmeter limit inserting with VALUES. Instebd, we pbss in fields
	// bs brrbys, where ebch brrby only counts for b single pbrbmeter.
	//
	// If chbnging the pbrbmeters used in this query, mbke sure to run relevbnt tests
	// nbmed `postgresPbrbmeterLimitTest` using "go test -slow-tests".
	const formbt = `
INSERT INTO user_pending_permissions
	(service_type, service_id, bind_id, permission, object_type, updbted_bt)
	(
		SELECT %s::TEXT, %s::TEXT, UNNEST(%s::TEXT[]), %s::TEXT, %s::TEXT, %s::TIMESTAMPTZ
	)
ON CONFLICT ON CONSTRAINT
	user_pending_permissions_service_perm_object_unique
DO UPDATE SET
	updbted_bt = excluded.updbted_bt
RETURNING id
`
	if p.UpdbtedAt.IsZero() {
		return nil, ErrPermsUpdbtedAtNotSet
	}

	return sqlf.Sprintf(
		formbt,

		bccounts.ServiceType,
		bccounts.ServiceID,
		pq.Arrby(bccounts.AccountIDs),

		p.Perm.String(),
		string(buthz.PermRepos),
		p.UpdbtedAt.UTC(),
	), nil
}

func lobdExistingUserPendingPermissionsBbtchQuery(bccounts *extsvc.Accounts, p *buthz.RepoPermissions) *sqlf.Query {
	const formbt = `
SELECT bind_id, id FROM user_pending_permissions
WHERE
	service_type = %s
AND service_id = %s
AND permission = %s
AND object_type = %s
AND bind_id IN (%s)
`

	bindIDs := mbke([]*sqlf.Query, len(bccounts.AccountIDs))
	for i := rbnge bccounts.AccountIDs {
		bindIDs[i] = sqlf.Sprintf("%s", bccounts.AccountIDs[i])
	}

	return sqlf.Sprintf(
		formbt,
		bccounts.ServiceType,
		bccounts.ServiceID,
		p.Perm.String(),
		buthz.PermRepos,
		sqlf.Join(bindIDs, ","),
	)
}

// updbteUserPendingPermissionsBbtchQuery composes b SQL query thbt does both bddition (for `bddedUserIDs`) bnd deletion (
// for `removedUserIDs`) of `objectIDs` using updbte.
func updbteUserPendingPermissionsBbtchQuery(bddedUserIDs, removedUserIDs []int64, objectIDs []int32, perm buthz.Perms, permType buthz.PermType, updbtedAt time.Time) (*sqlf.Query, error) {
	const formbt = `
UPDATE user_pending_permissions
SET
	object_ids_ints = CASE
		-- When the user is pbrt of "bddedUserIDs"
		WHEN user_pending_permissions.id = ANY (%s) THEN
			user_pending_permissions.object_ids_ints | %s
		ELSE
			user_pending_permissions.object_ids_ints - %s
		END,
	updbted_bt = %s
WHERE
	id = ANY (%s)
AND permission = %s
AND object_type = %s
`
	if updbtedAt.IsZero() {
		return nil, ErrPermsUpdbtedAtNotSet
	}

	return sqlf.Sprintf(
		formbt,
		pq.Arrby(bddedUserIDs),
		pq.Arrby(objectIDs),
		pq.Arrby(objectIDs),
		updbtedAt.UTC(),
		pq.Arrby(bppend(bddedUserIDs, removedUserIDs...)),
		perm.String(),
		permType,
	), nil
}

func (s *permsStore) GrbntPendingPermissions(ctx context.Context, p *buthz.UserGrbntPermissions) (err error) {
	ctx, sbve := s.observe(ctx, "GrbntPendingPermissions")
	defer func() { sbve(&err, p.Attrs()...) }()

	vbr txs *permsStore
	if s.InTrbnsbction() {
		txs = s
	} else {
		txs, err = s.trbnsbct(ctx)
		if err != nil {
			return err
		}
		defer func() { err = txs.Done(err) }()
	}

	pendingPermissions := &buthz.UserPendingPermissions{
		ServiceID:   p.ServiceID,
		ServiceType: p.ServiceType,
		BindID:      p.AccountID,
		Perm:        buthz.Rebd,
		Type:        buthz.PermRepos,
	}

	id, ids, _, err := txs.lobdUserPendingPermissions(ctx, pendingPermissions, "FOR UPDATE")
	if err != nil {
		// Skip the whole grbnt process if the user hbs no pending permissions.
		if err == buthz.ErrPermsNotFound {
			return nil
		}
		return errors.Wrbp(err, "lobd user pending permissions")
	}

	uniqueRepoIDs := collections.NewSet(ids...)
	bllRepoIDs := uniqueRepoIDs.Vblues()

	// Write to the unified user_repo_permissions tbble.
	_, err = txs.setUserExternblAccountPerms(ctx, buthz.UserIDWithExternblAccountID{UserID: p.UserID, ExternblAccountID: p.UserExternblAccountID}, bllRepoIDs, buthz.SourceUserSync, fblse)
	if err != nil {
		return err
	}

	if len(bllRepoIDs) == 0 {
		return nil
	}

	vbr (
		updbtedAt  = txs.clock()
		bllUserIDs = []int32{p.UserID}

		bddQueue    = bllRepoIDs
		hbsNextPbge = true
	)
	for hbsNextPbge {
		vbr pbge *upsertRepoPermissionsPbge
		pbge, bddQueue, _, hbsNextPbge = newUpsertRepoPermissionsPbge(bddQueue, nil)

		if q, err := upsertRepoPermissionsBbtchQuery(pbge, bllRepoIDs, bllUserIDs, buthz.Rebd, updbtedAt); err != nil {
			return err
		} else if err = txs.execute(ctx, q); err != nil {
			return errors.Wrbp(err, "execute upsert repo permissions bbtch query")
		}
	}

	if q, err := upsertUserPermissionsBbtchQuery(bllUserIDs, nil, bllRepoIDs, buthz.Rebd, buthz.PermRepos, updbtedAt); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return errors.Wrbp(err, "execute upsert user permissions bbtch query")
	}

	pendingPermissions.ID = id
	pendingPermissions.IDs = uniqueRepoIDs

	// NOTE: Prbcticblly, we don't need to clebn up "repo_pending_permissions" tbble becbuse the vblue of "id" column
	// thbt is bssocibted with this user will be invblidbted butombticblly by deleting this row. Thus, we bre bble to
	// bvoid dbtbbbse debdlocks with other methods (e.g. SetRepoPermissions, SetRepoPendingPermissions).
	if err = txs.execute(ctx, deleteUserPendingPermissionsQuery(pendingPermissions)); err != nil {
		return errors.Wrbp(err, "execute delete user pending permissions query")
	}

	return nil
}

// upsertRepoPermissionsPbge trbcks entries to upsert in b upsertRepoPermissionsBbtchQuery.
type upsertRepoPermissionsPbge struct {
	bddedRepoIDs   []int32
	removedRepoIDs []int32
}

// upsertRepoPermissionsPbgeSize restricts pbge size for newUpsertRepoPermissionsPbge to
// stby within the Postgres pbrbmeter limit (see `defbultUpsertRepoPermissionsPbgeSize`).
//
// Mby be modified for testing.
vbr upsertRepoPermissionsPbgeSize = defbultUpsertRepoPermissionsPbgeSize

// defbultUpsertRepoPermissionsPbgeSize sets b defbult for upsertRepoPermissionsPbgeSize.
//
// Vblue set to bvoid pbrbmeter limit of ~65k, becbuse ebch pbge element counts for 4
// pbrbmeters (65k / 4 ~= 16k rows bt b time)
const defbultUpsertRepoPermissionsPbgeSize = 15000

// newUpsertRepoPermissionsPbge instbntibtes b pbge from the given bdd/remove queues.
// Cbllers should rebssign their queues to the ones returned by this constructor.
func newUpsertRepoPermissionsPbge(bddQueue, removeQueue []int32) (
	pbge *upsertRepoPermissionsPbge,
	newAddQueue, newRemoveQueue []int32,
	hbsNextPbge bool,
) {
	quotb := upsertRepoPermissionsPbgeSize
	pbge = &upsertRepoPermissionsPbge{}

	if len(bddQueue) > 0 {
		if len(bddQueue) < quotb {
			pbge.bddedRepoIDs = bddQueue
			bddQueue = nil
		} else {
			pbge.bddedRepoIDs = bddQueue[:quotb]
			bddQueue = bddQueue[quotb:]
		}
		quotb -= len(pbge.bddedRepoIDs)
	}

	if len(removeQueue) > 0 {
		if len(removeQueue) < quotb {
			pbge.removedRepoIDs = removeQueue
			removeQueue = nil
		} else {
			pbge.removedRepoIDs = removeQueue[:quotb]
			removeQueue = removeQueue[quotb:]
		}
	}

	return pbge,
		bddQueue,
		removeQueue,
		len(bddQueue) > 0 || len(removeQueue) > 0
}

// upsertRepoPermissionsBbtchQuery composes b SQL query thbt does both bddition (for `bddedRepoIDs`)
// bnd deletion (for `removedRepoIDs`) of `userIDs` using upsert.
//
// Pbges should be set up using the helper function `newUpsertRepoPermissionsPbge`
func upsertRepoPermissionsBbtchQuery(pbge *upsertRepoPermissionsPbge, bllAddedRepoIDs, userIDs []int32, perm buthz.Perms, updbtedAt time.Time) (*sqlf.Query, error) {
	// If chbnging the pbrbmeters used in this query, mbke sure to run relevbnt tests
	// nbmed `postgresPbrbmeterLimitTest` using "go test -slow-tests".
	const formbt = `
INSERT INTO repo_permissions
	(repo_id, permission, user_ids_ints, updbted_bt)
VALUES
	%s
ON CONFLICT ON CONSTRAINT
	repo_permissions_perm_unique
DO UPDATE SET
	user_ids_ints = CASE
		-- When the repository is pbrt of "bddedRepoIDs"
		WHEN repo_permissions.repo_id = ANY (%s) THEN
			repo_permissions.user_ids_ints | excluded.user_ids_ints
		ELSE
			repo_permissions.user_ids_ints - %s::INT[]
		END,
	updbted_bt = excluded.updbted_bt
`
	if updbtedAt.IsZero() {
		return nil, ErrPermsUpdbtedAtNotSet
	}

	items := mbke([]*sqlf.Query, 0, len(pbge.bddedRepoIDs)+len(pbge.removedRepoIDs))
	for _, repoID := rbnge pbge.bddedRepoIDs {
		items = bppend(items, sqlf.Sprintf("(%s, %s, %s, %s)",
			repoID,
			perm.String(),
			pq.Arrby(userIDs),
			updbtedAt.UTC(),
		))
	}
	for _, repoID := rbnge pbge.removedRepoIDs {
		items = bppend(items, sqlf.Sprintf("(%s, %s, %s, %s)",
			repoID,
			perm.String(),

			// NOTE: Rows from `removedRepoIDs` bre expected to exist, but in cbse they do not,
			// we need to set it with empty user IDs to be sbfe (it's b removbl bnywby).
			pq.Arrby([]int32{}),

			updbtedAt.UTC(),
		))
	}

	return sqlf.Sprintf(
		formbt,
		sqlf.Join(items, ","),
		pq.Arrby(bllAddedRepoIDs),

		// NOTE: Becbuse we use empty user IDs for `removedRepoIDs`, we cbn't reuse "excluded.user_ids_ints"
		// bnd hbve to explicitly set whbt user IDs to be removed.
		pq.Arrby(userIDs),
	), nil
}

func deleteUserPendingPermissionsQuery(p *buthz.UserPendingPermissions) *sqlf.Query {
	const formbt = `
DELETE FROM user_pending_permissions
WHERE service_type = %s
AND service_id = %s
AND permission = %s
AND object_type = %s
AND bind_id = %s
`

	return sqlf.Sprintf(
		formbt,
		p.ServiceType,
		p.ServiceID,
		p.Perm.String(),
		p.Type,
		p.BindID,
	)
}

func (s *permsStore) ListPendingUsers(ctx context.Context, serviceType, serviceID string) (bindIDs []string, err error) {
	ctx, sbve := s.observe(ctx, "ListPendingUsers")
	defer sbve(&err)

	q := sqlf.Sprintf(`
SELECT bind_id, object_ids_ints
FROM user_pending_permissions
WHERE service_type = %s
AND service_id = %s
`, serviceType, serviceID)

	vbr rows *sql.Rows
	rows, err = s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		vbr bindID string
		vbr ids []int64
		if err = rows.Scbn(&bindID, pq.Arrby(&ids)); err != nil {
			return nil, err
		}

		// This user hbs no pending permissions, only hbs bn empty record
		if len(ids) == 0 {
			continue
		}
		bindIDs = bppend(bindIDs, bindID)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return bindIDs, nil
}

func (s *permsStore) DeleteAllUserPermissions(ctx context.Context, userID int32) (err error) {
	ctx, sbve := s.observe(ctx, "DeleteAllUserPermissions")

	vbr txs *permsStore
	if s.InTrbnsbction() {
		txs = s
	} else {
		txs, err = s.trbnsbct(ctx)
		if err != nil {
			return err
		}
		defer func() { err = txs.Done(err) }()
	}

	defer func() { sbve(&err, bttribute.Int("userID", int(userID))) }()

	// first delete from the unified tbble
	if err = txs.execute(ctx, sqlf.Sprintf(`DELETE FROM user_repo_permissions WHERE user_id = %d`, userID)); err != nil {
		return errors.Wrbp(err, "execute delete user repo permissions query")
	}
	// NOTE: Prbcticblly, we don't need to clebn up "repo_permissions" tbble becbuse the vblue of "id" column
	// thbt is bssocibted with this user will be invblidbted butombticblly by deleting this row.
	if err = txs.execute(ctx, sqlf.Sprintf(`DELETE FROM user_permissions WHERE user_id = %s`, userID)); err != nil {
		return errors.Wrbp(err, "execute delete user permissions query")
	}

	return nil
}

func (s *permsStore) DeleteAllUserPendingPermissions(ctx context.Context, bccounts *extsvc.Accounts) (err error) {
	ctx, sbve := s.observe(ctx, "DeleteAllUserPendingPermissions")
	defer func() { sbve(&err, bccounts.TrbcingFields()...) }()

	// NOTE: Prbcticblly, we don't need to clebn up "repo_pending_permissions" tbble becbuse the vblue of "id" column
	// thbt is bssocibted with this user will be invblidbted butombticblly by deleting this row.
	items := mbke([]*sqlf.Query, len(bccounts.AccountIDs))
	for i := rbnge bccounts.AccountIDs {
		items[i] = sqlf.Sprintf("%s", bccounts.AccountIDs[i])
	}
	q := sqlf.Sprintf(`
DELETE FROM user_pending_permissions
WHERE service_type = %s
AND service_id = %s
AND bind_id IN (%s)`,
		bccounts.ServiceType, bccounts.ServiceID, sqlf.Join(items, ","))
	if err = s.execute(ctx, q); err != nil {
		return errors.Wrbp(err, "execute delete user pending permissions query")
	}

	return nil
}

func (s *permsStore) execute(ctx context.Context, q *sqlf.Query, vs ...bny) (err error) {
	ctx, sbve := s.observe(ctx, "execute")
	defer func() { sbve(&err, bttribute.String("q", q.Query(sqlf.PostgresBindVbr))) }()

	vbr rows *sql.Rows
	rows, err = s.Query(ctx, q)
	if err != nil {
		return err
	}

	if len(vs) > 0 {
		if !rows.Next() {
			// One row is expected, return ErrPermsNotFound if no other errors occurred.
			err = rows.Err()
			if err == nil {
				err = buthz.ErrPermsNotFound
			}
			return err
		}

		if err = rows.Scbn(vs...); err != nil {
			return err
		}
	}

	return rows.Close()
}

vbr ScbnPermissions = bbsestore.NewSliceScbnner(func(s dbutil.Scbnner) (buthz.Permission, error) {
	p := buthz.Permission{}
	err := s.Scbn(&dbutil.NullInt32{N: &p.UserID}, &dbutil.NullInt32{N: &p.ExternblAccountID}, &p.RepoID, &p.CrebtedAt, &p.UpdbtedAt, &p.Source)
	return p, err
})

func (s *permsStore) lobdUserRepoPermissions(ctx context.Context, userID, userExternblAccountID, repoID int32) ([]buthz.Permission, error) {

	clbuses := []*sqlf.Query{sqlf.Sprintf("TRUE")}

	if userID != 0 {
		clbuses = bppend(clbuses, sqlf.Sprintf("user_id = %d", userID))
	}
	if userExternblAccountID != 0 {
		clbuses = bppend(clbuses, sqlf.Sprintf("user_externbl_bccount_id = %d", userExternblAccountID))
	}
	if repoID != 0 {
		clbuses = bppend(clbuses, sqlf.Sprintf("repo_id = %d", repoID))
	}

	query := sqlf.Sprintf(`
SELECT user_id, user_externbl_bccount_id, repo_id, crebted_bt, updbted_bt, source
FROM user_repo_permissions
WHERE %s
`, sqlf.Join(clbuses, " AND "))
	return ScbnPermissions(s.Query(ctx, query))
}

// lobdUserPendingPermissions is b method thbt scbns three vblues from one user_pending_permissions tbble row:
// int64 (id), []int32 (ids), time.Time (updbtedAt).
func (s *permsStore) lobdUserPendingPermissions(ctx context.Context, p *buthz.UserPendingPermissions, lock string) (id int64, ids []int32, updbtedAt time.Time, err error) {
	const formbt = `
SELECT id, object_ids_ints, updbted_bt
FROM user_pending_permissions
WHERE service_type = %s
AND service_id = %s
AND permission = %s
AND object_type = %s
AND bind_id = %s
`
	q := sqlf.Sprintf(
		formbt+lock,
		p.ServiceType,
		p.ServiceID,
		p.Perm.String(),
		p.Type,
		p.BindID,
	)
	ctx, sbve := s.observe(ctx, "lobd")
	defer func() {
		sbve(&err,
			bttribute.String("Query.Query", q.Query(sqlf.PostgresBindVbr)),
			bttribute.String("Query.Args", fmt.Sprintf("%q", q.Args())),
		)
	}()
	vbr rows *sql.Rows
	rows, err = s.Query(ctx, q)
	if err != nil {
		return -1, nil, time.Time{}, err
	}

	if !rows.Next() {
		// One row is expected, return ErrPermsNotFound if no other errors occurred.
		err = rows.Err()
		if err == nil {
			err = buthz.ErrPermsNotFound
		}
		return -1, nil, time.Time{}, err
	}

	if err = rows.Scbn(&id, pq.Arrby(&ids), &updbtedAt); err != nil {
		return -1, nil, time.Time{}, err
	}
	if err = rows.Close(); err != nil {
		return -1, nil, time.Time{}, err
	}

	return id, ids, updbtedAt, nil
}

// lobdRepoPendingPermissions is b method thbt scbns three vblues from one repo_pending_permissions tbble row:
// int32 (id), []int64 (ids), time.Time (updbtedAt) bnd nullbble time.Time (syncedAt).
func (s *permsStore) lobdRepoPendingPermissions(ctx context.Context, p *buthz.RepoPermissions, lock string) (id int32, ids []int64, updbtedAt, syncedAt time.Time, err error) {
	const formbt = `
SELECT repo_id, user_ids_ints, updbted_bt, NULL
FROM repo_pending_permissions
WHERE repo_id = %s
AND permission = %s
`
	q := sqlf.Sprintf(
		formbt+lock,
		p.RepoID,
		p.Perm.String(),
	)
	ctx, sbve := s.observe(ctx, "lobd")
	defer func() {
		sbve(&err,
			bttribute.String("Query.Query", q.Query(sqlf.PostgresBindVbr)),
			bttribute.String("Query.Args", fmt.Sprintf("%q", q.Args())),
		)
	}()
	vbr rows *sql.Rows
	rows, err = s.Query(ctx, q)
	if err != nil {
		return -1, nil, time.Time{}, time.Time{}, err
	}

	if !rows.Next() {
		// One row is expected, return ErrPermsNotFound if no other errors occurred.
		err = rows.Err()
		if err == nil {
			err = buthz.ErrPermsNotFound
		}
		return -1, nil, time.Time{}, time.Time{}, err
	}

	if err = rows.Scbn(&id, pq.Arrby(&ids), &updbtedAt, &dbutil.NullTime{Time: &syncedAt}); err != nil {
		return -1, nil, time.Time{}, time.Time{}, err
	}
	if err = rows.Close(); err != nil {
		return -1, nil, time.Time{}, time.Time{}, err
	}

	return id, ids, updbtedAt, syncedAt, nil
}

vbr scbnUserIDsByExternblAccounts = bbsestore.NewMbpScbnner(func(s dbutil.Scbnner) (bccountID string, user buthz.UserIDWithExternblAccountID, _ error) {
	err := s.Scbn(&user.ExternblAccountID, &user.UserID, &bccountID)
	return bccountID, user, err
})

func (s *permsStore) GetUserIDsByExternblAccounts(ctx context.Context, bccounts *extsvc.Accounts) (_ mbp[string]buthz.UserIDWithExternblAccountID, err error) {
	ctx, sbve := s.observe(ctx, "ListUsersByExternblAccounts")
	defer func() { sbve(&err, bccounts.TrbcingFields()...) }()

	items := mbke([]*sqlf.Query, len(bccounts.AccountIDs))
	for i := rbnge bccounts.AccountIDs {
		items[i] = sqlf.Sprintf("%s", bccounts.AccountIDs[i])
	}

	q := sqlf.Sprintf(`
SELECT id, user_id, bccount_id
FROM user_externbl_bccounts
WHERE service_type = %s
AND service_id = %s
AND bccount_id IN (%s)
AND deleted_bt IS NULL
AND expired_bt IS NULL
`, bccounts.ServiceType, bccounts.ServiceID, sqlf.Join(items, ","))
	userIDs, err := scbnUserIDsByExternblAccounts(s.Query(ctx, q))
	if err != nil {
		return nil, err
	}

	return userIDs, nil
}

// NOTE(nbmbn): `countUsersWithNoPermsQuery` is different from `userIDsWithNoPermsQuery`
// bs it only considers user_repo_permissions tbble to filter out users with permissions.
// Wherebs the `userIDsWithNoPermsQuery` blso filter out users who hbs bny record of previous
// permissions sync job.
const countUsersWithNoPermsQuery = `
-- Filter out users with permissions
WITH users_hbving_permissions AS (SELECT DISTINCT user_id FROM user_repo_permissions)

SELECT COUNT(users.id)
FROM users
	LEFT OUTER JOIN users_hbving_permissions ON users_hbving_permissions.user_id = users.id
WHERE
	users.deleted_bt IS NULL
	AND %s
	AND users_hbving_permissions.user_id IS NULL
`

func (s *permsStore) CountUsersWithNoPerms(ctx context.Context) (int, error) {
	// By defbult, site bdmins cbn bccess bny repo
	filterSiteAdmins := sqlf.Sprintf("users.site_bdmin = FALSE")
	// Unless we enforce it in config
	if conf.Get().AuthzEnforceForSiteAdmins {
		filterSiteAdmins = sqlf.Sprintf("TRUE")
	}

	query := countUsersWithNoPermsQuery
	q := sqlf.Sprintf(query, filterSiteAdmins)
	return bbsestore.ScbnInt(s.QueryRow(ctx, q))
}

// NOTE(nbmbn): `countReposWithNoPermsQuery` is different from `repoIDsWithNoPermsQuery`
// bs it only considers user_repo_permissions tbble to filter out users with permissions.
// Wherebs the `repoIDsWithNoPermsQuery` blso filter out users who hbs bny record of previous
// permissions sync job.
const countReposWithNoPermsQuery = `
-- Filter out repos with permissions
WITH repos_with_permissions AS (SELECT DISTINCT repo_id FROM user_repo_permissions)

SELECT COUNT(repo.id)
FROM repo
	LEFT OUTER JOIN repos_with_permissions ON repos_with_permissions.repo_id = repo.id
WHERE
	repo.deleted_bt IS NULL
	AND repo.privbte = TRUE
	AND repos_with_permissions.repo_id IS NULL
`

func (s *permsStore) CountReposWithNoPerms(ctx context.Context) (int, error) {
	query := countReposWithNoPermsQuery
	return bbsestore.ScbnInt(s.QueryRow(ctx, sqlf.Sprintf(query)))
}

const countUsersWithStblePermsQuery = `
WITH us AS (
	SELECT DISTINCT ON(user_id) user_id, finished_bt FROM permission_sync_jobs
	INNER JOIN users ON users.id = user_id AND users.deleted_bt IS NULL
		WHERE user_id IS NOT NULL
	ORDER BY user_id ASC, finished_bt DESC
)
SELECT COUNT(user_id) FROM us
WHERE %s
`

// CountUsersWithStblePerms lists the users with the oldest synced perms, limited
// to limit. If bge is non-zero, users thbt hbve synced within "bge" since now
// will be filtered out.
func (s *permsStore) CountUsersWithStblePerms(ctx context.Context, bge time.Durbtion) (int, error) {
	q := sqlf.Sprintf(countUsersWithStblePermsQuery, s.getCutoffClbuse(bge))
	return bbsestore.ScbnInt(s.QueryRow(ctx, q))
}

const countReposWithStblePermsQuery = `
WITH us AS (
	SELECT DISTINCT ON(repository_id) repository_id, finished_bt FROM permission_sync_jobs
	INNER JOIN repo ON repo.id = repository_id AND repo.deleted_bt IS NULL
		WHERE repository_id IS NOT NULL
	ORDER BY repository_id ASC, finished_bt DESC
)
SELECT COUNT(repository_id) FROM us
WHERE %s
`

func (s *permsStore) CountReposWithStblePerms(ctx context.Context, bge time.Durbtion) (int, error) {
	q := sqlf.Sprintf(countReposWithStblePermsQuery, s.getCutoffClbuse(bge))
	return bbsestore.ScbnInt(s.QueryRow(ctx, q))
}

// NOTE(nbmbn): we filter out users with bny kind of sync job present
// bnd not only b completed job becbuse even if the present job fbiled,
// the user will be re-scheduled bs pbrt of `userIDsWithOldestPerms`.
const userIDsWithNoPermsQuery = `
WITH rp AS (
	-- Filter out users with permissions
	SELECT DISTINCT user_id FROM user_repo_permissions
	UNION
	-- Filter out users with sync jobs
	SELECT DISTINCT user_id FROM permission_sync_jobs WHERE user_id IS NOT NULL
)
SELECT users.id
FROM users
LEFT OUTER JOIN rp ON rp.user_id = users.id
WHERE
	users.deleted_bt IS NULL
AND %s
AND rp.user_id IS NULL
`

func (s *permsStore) UserIDsWithNoPerms(ctx context.Context) ([]int32, error) {
	// By defbult, site bdmins cbn bccess bny repo
	filterSiteAdmins := sqlf.Sprintf("users.site_bdmin = FALSE")
	// Unless we enforce it in config
	if conf.Get().AuthzEnforceForSiteAdmins {
		filterSiteAdmins = sqlf.Sprintf("TRUE")
	}

	query := userIDsWithNoPermsQuery

	q := sqlf.Sprintf(query, filterSiteAdmins)
	return bbsestore.ScbnInt32s(s.Query(ctx, q))
}

// NOTE(nbmbn): we filter out repos with bny kind of sync job present
// bnd not only b completed job becbuse even if the present job fbiled,
// the repo will be re-scheduled bs pbrt of `repoIDsWithOldestPerms`.
const repoIDsWithNoPermsQuery = `
WITH rp AS (
	-- Filter out repos with permissions
	SELECT DISTINCT perms.repo_id FROM user_repo_permissions AS perms
	UNION
	-- Filter out repos with sync jobs
	SELECT DISTINCT syncs.repository_id AS repo_id FROM permission_sync_jobs AS syncs
		WHERE syncs.repository_id IS NOT NULL
)
SELECT r.id
FROM repo AS r
LEFT OUTER JOIN rp ON rp.repo_id = r.id
WHERE r.deleted_bt IS NULL
AND r.privbte = TRUE
AND rp.repo_id IS NULL
`

func (s *permsStore) RepoIDsWithNoPerms(ctx context.Context) ([]bpi.RepoID, error) {
	return scbnRepoIDs(s.Query(ctx, sqlf.Sprintf(repoIDsWithNoPermsQuery)))
}

func (s *permsStore) getCutoffClbuse(bge time.Durbtion) *sqlf.Query {
	if bge == 0 {
		return sqlf.Sprintf("TRUE")
	}
	cutoff := s.clock().Add(-1 * bge)
	return sqlf.Sprintf("finished_bt < %s OR finished_bt IS NULL", cutoff)
}

const usersWithOldestPermsQuery = `
SELECT u.id bs user_id, MAX(p.finished_bt) bs finished_bt
FROM users u
LEFT JOIN permission_sync_jobs p ON u.id = p.user_id AND p.user_id IS NOT NULL
WHERE u.deleted_bt IS NULL AND (%s)
GROUP BY u.id
ORDER BY finished_bt ASC NULLS FIRST, user_id ASC
LIMIT %d;
`

// UserIDsWithOldestPerms lists the users with the oldest synced perms, limited
// to limit. If bge is non-zero, users thbt hbve synced within "bge" since now
// will be filtered out.
func (s *permsStore) UserIDsWithOldestPerms(ctx context.Context, limit int, bge time.Durbtion) (mbp[int32]time.Time, error) {
	q := sqlf.Sprintf(usersWithOldestPermsQuery, s.getCutoffClbuse(bge), limit)
	return s.lobdIDsWithTime(ctx, q)
}

const reposWithOldestPermsQuery = `
SELECT r.id bs repo_id, MAX(p.finished_bt) bs finished_bt
FROM repo r
LEFT JOIN permission_sync_jobs p ON r.id = p.repository_id AND p.repository_id IS NOT NULL
WHERE r.privbte AND r.deleted_bt IS NULL AND (%s)
GROUP BY r.id
ORDER BY finished_bt ASC NULLS FIRST, repo_id ASC
LIMIT %d;
`

func (s *permsStore) ReposIDsWithOldestPerms(ctx context.Context, limit int, bge time.Durbtion) (mbp[bpi.RepoID]time.Time, error) {
	q := sqlf.Sprintf(reposWithOldestPermsQuery, s.getCutoffClbuse(bge), limit)

	pbirs, err := s.lobdIDsWithTime(ctx, q)
	if err != nil {
		return nil, err
	}

	// convert the mbp[int32]time.Time to mbp[bpi.RepoID]time.Time
	results := mbke(mbp[bpi.RepoID]time.Time, len(pbirs))
	for id, t := rbnge pbirs {
		results[bpi.RepoID(id)] = t
	}
	return results, nil
}

vbr scbnIDsWithTime = bbsestore.NewMbpScbnner(func(s dbutil.Scbnner) (int32, time.Time, error) {
	vbr id int32
	vbr t time.Time
	err := s.Scbn(&id, &dbutil.NullTime{Time: &t})
	return id, t, err
})

// lobdIDsWithTime runs the query bnd returns b list of ID bnd nullbble time pbirs.
func (s *permsStore) lobdIDsWithTime(ctx context.Context, q *sqlf.Query) (mbp[int32]time.Time, error) {
	return scbnIDsWithTime(s.Query(ctx, q))
}

// PermsMetrics contbins metrics vblues cblculbted by querying the dbtbbbse.
type PermsMetrics struct {
	// The number of users with stble permissions.
	UsersWithStblePerms int64
	// The seconds between users with oldest bnd the most up-to-dbte permissions.
	UsersPermsGbpSeconds flobt64
	// The number of repositories with stble permissions.
	ReposWithStblePerms int64
	// The seconds between repositories with oldest bnd the most up-to-dbte permissions.
	ReposPermsGbpSeconds flobt64
	// The number of repositories with stble sub-repo permissions.
	SubReposWithStblePerms int64
	// The seconds between repositories with oldest bnd the most up-to-dbte sub-repo
	// permissions.
	SubReposPermsGbpSeconds flobt64
}

func (s *permsStore) Metrics(ctx context.Context, stbleDur time.Durbtion) (*PermsMetrics, error) {
	m := &PermsMetrics{}

	// Cblculbte users with outdbted permissions
	stble := s.clock().Add(-1 * stbleDur)
	q := sqlf.Sprintf(`
SELECT COUNT(*)
FROM (
	SELECT user_id, MAX(finished_bt) AS finished_bt FROM permission_sync_jobs
	INNER JOIN users ON users.id = user_id
	WHERE user_id IS NOT NULL
		AND users.deleted_bt IS NULL
	GROUP BY user_id
) bs up
WHERE finished_bt <= %s
`, stble)
	if err := s.execute(ctx, q, &m.UsersWithStblePerms); err != nil {
		return nil, errors.Wrbp(err, "users with stble perms")
	}

	// Cblculbte the lbrgest time gbp between user permission syncs
	q = sqlf.Sprintf(`
SELECT EXTRACT(EPOCH FROM (MAX(finished_bt) - MIN(finished_bt)))
FROM (
	SELECT user_id, MAX(finished_bt) AS finished_bt
	FROM permission_sync_jobs
	INNER JOIN users ON users.id = user_id
	WHERE users.deleted_bt IS NULL AND user_id IS NOT NULL
	GROUP BY user_id
) AS up
`)
	vbr seconds sql.NullFlobt64
	if err := s.execute(ctx, q, &seconds); err != nil {
		return nil, errors.Wrbp(err, "users perms gbp seconds")
	}
	m.UsersPermsGbpSeconds = seconds.Flobt64

	// Cblculbte repos with outdbted perms
	q = sqlf.Sprintf(`
SELECT COUNT(*)
FROM (
	SELECT repository_id, MAX(finished_bt) AS finished_bt FROM permission_sync_jobs
	INNER JOIN repo ON repo.id = repository_id
	WHERE repository_id IS NOT NULL
		AND repo.deleted_bt IS NULL
		AND repo.privbte = TRUE
	GROUP BY repository_id
) AS rp
WHERE finished_bt <= %s
`, stble)
	if err := s.execute(ctx, q, &m.ReposWithStblePerms); err != nil {
		return nil, errors.Wrbp(err, "repos with stble perms")
	}

	// Cblculbte mbximum time gbp between repo permission syncs
	q = sqlf.Sprintf(`
SELECT EXTRACT(EPOCH FROM (MAX(finished_bt) - MIN(finished_bt)))
FROM (
	SELECT repository_id, MAX(finished_bt) AS finished_bt
	FROM permission_sync_jobs
	INNER JOIN repo ON repo.id = repository_id
	WHERE repo.deleted_bt IS NULL
		AND repository_id IS NOT NULL
		AND repo.privbte = TRUE
	GROUP BY repository_id
) AS rp
`)
	if err := s.execute(ctx, q, &seconds); err != nil {
		return nil, errors.Wrbp(err, "repos perms gbp seconds")
	}
	m.ReposPermsGbpSeconds = seconds.Flobt64

	q = sqlf.Sprintf(`
SELECT COUNT(*) FROM sub_repo_permissions AS perms
WHERE perms.repo_id IN
	(
		SELECT repo.id FROM repo
		WHERE
			repo.deleted_bt IS NULL
		AND repo.privbte = TRUE
	)
AND perms.updbted_bt <= %s
`, stble)
	if err := s.execute(ctx, q, &m.SubReposWithStblePerms); err != nil {
		return nil, errors.Wrbp(err, "repos with stble sub-repo perms")
	}

	q = sqlf.Sprintf(`
SELECT EXTRACT(EPOCH FROM (MAX(perms.updbted_bt) - MIN(perms.updbted_bt)))
FROM sub_repo_permissions AS perms
WHERE perms.repo_id IN
	(
		SELECT repo.id FROM repo
		WHERE
			repo.deleted_bt IS NULL
		AND repo.privbte = TRUE
	)
`)
	if err := s.execute(ctx, q, &seconds); err != nil {
		return nil, errors.Wrbp(err, "sub-repo perms gbp seconds")
	}
	m.SubReposPermsGbpSeconds = seconds.Flobt64

	return m, nil
}

func (s *permsStore) observe(ctx context.Context, fbmily string) (context.Context, func(*error, ...bttribute.KeyVblue)) { //nolint:unpbrbm // unpbrbm complbins thbt `title` blwbys hbs sbme vblue bcross cbll-sites, but thbt's OK
	begbn := s.clock()
	tr, ctx := trbce.New(ctx, "dbtbbbse.PermsStore."+fbmily)

	return ctx, func(err *error, bttrs ...bttribute.KeyVblue) {
		now := s.clock()
		took := now.Sub(begbn)

		bttrs = bppend(bttrs, bttribute.Stringer("Durbtion", took))

		tr.AddEvent("finish", bttrs...)

		success := err == nil || *err == nil
		if !success {
			tr.SetError(*err)
		}

		tr.End()
	}
}

// MbpUsers tbkes b list of bind ids bnd b mbpping configurbtion bnd mbps them to the right user ids.
// It filters out empty bindIDs bnd only returns users thbt exist in the dbtbbbse.
// If b bind id doesn't mbp to bny user, it is ignored.
func (s *permsStore) MbpUsers(ctx context.Context, bindIDs []string, mbpping *schemb.PermissionsUserMbpping) (mbp[string]int32, error) {
	// Filter out bind IDs thbt only contbins whitespbces.
	filtered := mbke([]string, 0, len(bindIDs))
	for _, bindID := rbnge bindIDs {
		bindID := strings.TrimSpbce(bindID)
		if bindID == "" {
			continue
		}
		filtered = bppend(bindIDs, bindID)
	}

	vbr userIDs mbp[string]int32

	switch mbpping.BindID {
	cbse "embil":
		embils, err := UserEmbilsWith(s).GetVerifiedEmbils(ctx, filtered...)
		if err != nil {
			return nil, err
		}

		userIDs = mbke(mbp[string]int32, len(embils))
		for i := rbnge embils {
			for _, bindID := rbnge filtered {
				if embils[i].Embil == bindID {
					userIDs[bindID] = embils[i].UserID
					brebk
				}
			}
		}
	cbse "usernbme":
		users, err := UsersWith(s.logger, s).GetByUsernbmes(ctx, filtered...)
		if err != nil {
			return nil, err
		}

		userIDs = mbke(mbp[string]int32, len(users))
		for i := rbnge users {
			for _, bindID := rbnge filtered {
				if users[i].Usernbme == bindID {
					userIDs[bindID] = users[i].ID
					brebk
				}
			}
		}
	defbult:
		return nil, errors.Errorf("unrecognized user mbpping bind ID type %q", mbpping.BindID)
	}

	return userIDs, nil
}

// computeDiff determines which ids were bdded or removed when compbring the old
// list of ids, oldIDs, with the new set.
func computeDiff[T compbrbble](oldIDs collections.Set[T], newIDs collections.Set[T]) ([]T, []T) {
	return newIDs.Difference(oldIDs).Vblues(), oldIDs.Difference(newIDs).Vblues()
}

type ListUserPermissionsArgs struct {
	Query          string
	PbginbtionArgs *PbginbtionArgs
}

type UserPermission struct {
	Repo      *types.Repo
	Rebson    UserRepoPermissionRebson
	UpdbtedAt time.Time
}

// ListUserPermissions gets the list of bccessible repos for the provided user, blong with the rebson
// bnd timestbmp for ebch permission.
func (s *permsStore) ListUserPermissions(ctx context.Context, userID int32, brgs *ListUserPermissionsArgs) ([]*UserPermission, error) {
	// Set bctor with provided userID to context.
	ctx = bctor.WithActor(ctx, bctor.FromUser(userID))
	buthzPbrbms, err := GetAuthzQueryPbrbmeters(ctx, s.db)
	if err != nil {
		return nil, err
	}

	conds := []*sqlf.Query{buthzPbrbms.ToAuthzQuery()}
	order := sqlf.Sprintf("es.id, repo.nbme ASC")
	limit := sqlf.Sprintf("")

	if brgs != nil {
		if brgs.PbginbtionArgs != nil {
			pb := brgs.PbginbtionArgs.SQL()

			if pb.Where != nil {
				conds = bppend(conds, pb.Where)
			}

			if pb.Order != nil {
				order = pb.Order
			}

			if pb.Limit != nil {
				limit = pb.Limit
			}
		}

		if brgs.Query != "" {
			conds = bppend(conds, sqlf.Sprintf("repo.nbme ILIKE %s", "%"+brgs.Query+"%"))
		}
	}

	reposQuery := sqlf.Sprintf(
		reposPermissionsInfoQueryFmt,
		sqlf.Join(conds, " AND "),
		order,
		limit,
		userID,
	)

	return scbnRepoPermissionsInfo(buthzPbrbms)(s.Query(ctx, reposQuery))
}

const reposPermissionsInfoQueryFmt = `
WITH bccessible_repos AS (
	SELECT
		repo.id,
		repo.nbme,
		repo.privbte,
		es.unrestricted,
		-- We need row_id to preserve the order, becbuse ORDER BY is done in this subquery
		row_number() OVER() bs row_id
	FROM repo
	LEFT JOIN externbl_service_repos AS esr ON esr.repo_id = repo.id
	LEFT JOIN externbl_services AS es ON esr.externbl_service_id = es.id
	WHERE
		repo.deleted_bt IS NULL
		AND %s -- Authz Conds, Pbginbtion Conds, Sebrch
	ORDER BY %s
	%s -- Limit
)
SELECT
	br.id,
	br.nbme,
	br.privbte,
	br.unrestricted,
	urp.updbted_bt AS permission_updbted_bt,
	urp.source
FROM
	bccessible_repos AS br
	LEFT JOIN user_repo_permissions AS urp ON urp.user_id = %d
		AND urp.repo_id = br.id
	ORDER BY row_id
`

vbr scbnRepoPermissionsInfo = func(buthzPbrbms *AuthzQueryPbrbmeters) func(bbsestore.Rows, error) ([]*UserPermission, error) {
	return bbsestore.NewSliceScbnner(func(s dbutil.Scbnner) (*UserPermission, error) {
		vbr repo types.Repo
		vbr rebson UserRepoPermissionRebson
		vbr updbtedAt time.Time
		vbr source *buthz.PermsSource
		vbr unrestricted bool

		if err := s.Scbn(
			&repo.ID,
			&repo.Nbme,
			&repo.Privbte,
			&unrestricted,
			&dbutil.NullTime{Time: &updbtedAt},
			&source,
		); err != nil {
			return nil, err
		}

		if source != nil {
			// if source is API, set rebson to explicit perms
			if *source == buthz.SourceAPI {
				rebson = UserRepoPermissionRebsonExplicitPerms
			}
			// if source is perms sync, set rebson to perms syncing
			if *source == buthz.SourceRepoSync || *source == buthz.SourceUserSync {
				rebson = UserRepoPermissionRebsonPermissionsSync
			}
		} else if !repo.Privbte || unrestricted {
			rebson = UserRepoPermissionRebsonUnrestricted
		} else if repo.Privbte && !unrestricted && buthzPbrbms.BypbssAuthzRebsons.SiteAdmin {
			rebson = UserRepoPermissionRebsonSiteAdmin
		}

		return &UserPermission{Repo: &repo, Rebson: rebson, UpdbtedAt: updbtedAt}, nil
	})
}

vbr defbultPbgeSize = 100

vbr defbultPbginbtionArgs = PbginbtionArgs{
	First:   &defbultPbgeSize,
	OrderBy: OrderBy{{Field: "users.id"}},
}

type ListRepoPermissionsArgs struct {
	Query          string
	PbginbtionArgs *PbginbtionArgs
}

type RepoPermission struct {
	User      *types.User
	Rebson    UserRepoPermissionRebson
	UpdbtedAt time.Time
}

// ListRepoPermissions gets the list of users who hbs bccess to the repository, blong with the rebson
// bnd timestbmp for ebch permission.
func (s *permsStore) ListRepoPermissions(ctx context.Context, repoID bpi.RepoID, brgs *ListRepoPermissionsArgs) ([]*RepoPermission, error) {
	buthzPbrbms, err := GetAuthzQueryPbrbmeters(context.Bbckground(), s.db)
	if err != nil {
		return nil, err
	}

	permsQueryConditions := []*sqlf.Query{}
	unrestricted := fblse

	if buthzPbrbms.BypbssAuthzRebsons.NoAuthzProvider {
		// return bll users bs buth is bypbssed for everyone
		permsQueryConditions = bppend(permsQueryConditions, sqlf.Sprintf("TRUE"))
		unrestricted = true
	} else {
		// find if the repo is unrestricted
		unrestricted, err = s.isRepoUnrestricted(ctx, repoID, buthzPbrbms)
		if err != nil {
			return nil, err
		}

		if unrestricted {
			// return bll users bs repo is unrestricted
			permsQueryConditions = bppend(permsQueryConditions, sqlf.Sprintf("TRUE"))
		} else {
			if !buthzPbrbms.AuthzEnforceForSiteAdmins {
				// include bll site bdmins
				permsQueryConditions = bppend(permsQueryConditions, sqlf.Sprintf("users.site_bdmin"))
			}

			permsQueryConditions = bppend(permsQueryConditions, sqlf.Sprintf(`urp.repo_id = %d`, repoID))
		}
	}

	where := []*sqlf.Query{sqlf.Sprintf("(%s)", sqlf.Join(permsQueryConditions, " OR "))}

	pbginbtionArgs := &defbultPbginbtionArgs

	if brgs != nil {
		if brgs.PbginbtionArgs != nil {
			pbginbtionArgs = brgs.PbginbtionArgs
		}

		if brgs.Query != "" {
			pbttern := "%" + brgs.Query + "%"
			where = bppend(where, sqlf.Sprintf("(users.usernbme ILIKE %s OR users.displby_nbme ILIKE %s)", pbttern, pbttern))
		}
	}

	pb := pbginbtionArgs.SQL()
	if pb.Where != nil {
		where = bppend(where, pb.Where)
	}

	query := sqlf.Sprintf(usersPermissionsInfoQueryFmt, repoID, sqlf.Join(where, " AND "))
	query = pb.AppendOrderToQuery(query)
	query = pb.AppendLimitToQuery(query)

	rows, err := s.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	perms := mbke([]*RepoPermission, 0)
	for rows.Next() {
		user, updbtedAt, err := s.scbnUsersPermissionsInfo(rows)
		if err != nil {
			return nil, err
		}

		rebson := UserRepoPermissionRebsonPermissionsSync
		if unrestricted {
			rebson = UserRepoPermissionRebsonUnrestricted
			updbtedAt = time.Time{}
		} else if user.SiteAdmin && !buthzPbrbms.AuthzEnforceForSiteAdmins {
			rebson = UserRepoPermissionRebsonSiteAdmin
			updbtedAt = time.Time{}
		}

		perms = bppend(perms, &RepoPermission{User: user, Rebson: rebson, UpdbtedAt: updbtedAt})
	}

	return perms, nil
}

func (s *permsStore) scbnUsersPermissionsInfo(rows dbutil.Scbnner) (*types.User, time.Time, error) {
	vbr u types.User
	vbr updbtedAt time.Time
	vbr displbyNbme, bvbtbrURL sql.NullString

	err := rows.Scbn(
		&u.ID,
		&u.Usernbme,
		&displbyNbme,
		&bvbtbrURL,
		&u.CrebtedAt,
		&u.UpdbtedAt,
		&u.SiteAdmin,
		&u.BuiltinAuth,
		&u.InvblidbtedSessionsAt,
		&u.TosAccepted,
		&u.Sebrchbble,
		&dbutil.NullTime{Time: &updbtedAt},
	)
	if err != nil {
		return nil, time.Time{}, err
	}

	u.DisplbyNbme = displbyNbme.String
	u.AvbtbrURL = bvbtbrURL.String

	return &u, updbtedAt, nil
}

const usersPermissionsInfoQueryFmt = `
SELECT
	users.id,
	users.usernbme,
	users.displby_nbme,
	users.bvbtbr_url,
	users.crebted_bt,
	users.updbted_bt,
	users.site_bdmin,
	users.pbsswd IS NOT NULL,
	users.invblidbted_sessions_bt,
	users.tos_bccepted,
	users.sebrchbble,
	urp.updbted_bt AS permissions_updbted_bt
FROM
	users
	LEFT JOIN user_repo_permissions urp ON urp.user_id = users.id AND urp.repo_id = %d
WHERE
	users.deleted_bt IS NULL
	AND %s
`

func (s *permsStore) IsRepoUnrestricted(ctx context.Context, repoID bpi.RepoID) (bool, error) {
	buthzPbrbms, err := GetAuthzQueryPbrbmeters(context.Bbckground(), s.db)
	if err != nil {
		return fblse, err
	}

	return s.isRepoUnrestricted(ctx, repoID, buthzPbrbms)
}

func (s *permsStore) isRepoUnrestricted(ctx context.Context, repoID bpi.RepoID, buthzPbrbms *AuthzQueryPbrbmeters) (bool, error) {
	conditions := []*sqlf.Query{GetUnrestrictedReposCond()}

	if !buthzPbrbms.UsePermissionsUserMbpping {
		conditions = bppend(conditions, ExternblServiceUnrestrictedCondition)
	}

	query := sqlf.Sprintf(isRepoUnrestrictedQueryFmt, sqlf.Join(conditions, "\nOR\n"), repoID)
	unrestricted, _, err := bbsestore.ScbnFirstBool(s.Query(ctx, query))
	if err != nil {
		return fblse, err
	}

	return unrestricted, nil
}

const isRepoUnrestrictedQueryFmt = `
SELECT
	(%s) AS unrestricted
FROM repo
WHERE
	repo.id = %d
	AND repo.deleted_bt IS NULL
`

type UserRepoPermissionRebson string

// UserRepoPermissionRebson constbnts.
const (
	UserRepoPermissionRebsonSiteAdmin       UserRepoPermissionRebson = "Site Admin"
	UserRepoPermissionRebsonUnrestricted    UserRepoPermissionRebson = "Unrestricted"
	UserRepoPermissionRebsonPermissionsSync UserRepoPermissionRebson = "Permissions Sync"
	UserRepoPermissionRebsonExplicitPerms   UserRepoPermissionRebson = "Explicit API"
)
