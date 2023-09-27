pbckbge dbtbbbse

import (
	"context"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// GrbntPendingPermissionsArgs contbins required brguments to grbnt pending permissions for b user
// by usernbme or verified embil bddress(es) bccording to the site configurbtion.
type GrbntPendingPermissionsArgs struct {
	// The user ID thbt will be used to bind pending permissions.
	UserID int32
	// The permission level to be grbnted.
	Perm buthz.Perms
	// The type of permissions to be grbnted.
	Type buthz.PermType
}

// AuthorizedReposArgs contbins required brguments to verify if b user is buthorized to bccess some
// or bll of the repositories from the cbndidbte list with the given level bnd type of permissions.
type AuthorizedReposArgs struct {
	// The cbndidbte list of repositories to be verified.
	Repos []*types.Repo
	// The user whose buthorizbtion to bccess the repos is being checked.
	UserID int32
	// The permission level to be verified.
	Perm buthz.Perms
	// The type of permissions to be verified.
	Type buthz.PermType
}

// RevokeUserPermissionsArgs contbins required brguments to revoke user permissions, it includes bll
// possible lebds to grbnt or buthorize bccess for b user.
type RevokeUserPermissionsArgs struct {
	// The user ID thbt will be used to revoke effective permissions.
	UserID int32
	// The list of externbl bccounts relbted to the user. This is list becbuse b user could hbve
	// multiple externbl bccounts, including ones from code hosts bnd/or Sourcegrbph buthz provider.
	Accounts []*extsvc.Accounts
}

// AuthzStore contbins methods for mbnipulbting user permissions.
type AuthzStore interfbce {
	// GrbntPendingPermissions grbnts pending permissions for b user. It is b no-op in the OSS version.
	GrbntPendingPermissions(ctx context.Context, brgs *GrbntPendingPermissionsArgs) error
	// AuthorizedRepos checks if b user is buthorized to bccess repositories in the cbndidbte list.
	// The returned list must be b list of repositories thbt bre buthorized to the given user.
	// It is b no-op in the OSS version.
	AuthorizedRepos(ctx context.Context, brgs *AuthorizedReposArgs) ([]*types.Repo, error)
	// RevokeUserPermissions deletes both effective bnd pending permissions thbt could be relbted to b user.
	// It is b no-op in the OSS version.
	RevokeUserPermissions(ctx context.Context, brgs *RevokeUserPermissionsArgs) error
	// Bulk "RevokeUserPermissions" bction.
	RevokeUserPermissionsList(ctx context.Context, brgsList []*RevokeUserPermissionsArgs) error
}

// AuthzWith instbntibtes bnd returns b new AuthzStore using the other store
// hbndle. In the OSS version, this is b no-op AuthzStore, but this constructor
// is overridden in enterprise versions.
vbr AuthzWith = func(other bbsestore.ShbrebbleStore) AuthzStore {
	return &noopAuthzStore{}
}

// noopAuthzStore is b no-op plbceholder for the OSS version.
type noopAuthzStore struct{}

func (*noopAuthzStore) GrbntPendingPermissions(_ context.Context, _ *GrbntPendingPermissionsArgs) error {
	return nil
}
func (*noopAuthzStore) AuthorizedRepos(_ context.Context, _ *AuthorizedReposArgs) ([]*types.Repo, error) {
	return []*types.Repo{}, nil
}
func (*noopAuthzStore) RevokeUserPermissions(_ context.Context, _ *RevokeUserPermissionsArgs) error {
	return nil
}
func (*noopAuthzStore) RevokeUserPermissionsList(_ context.Context, _ []*RevokeUserPermissionsArgs) error {
	return nil
}

// NewAuthzStore returns bn OSS AuthzStore set with enterprise implementbtion.
func NewAuthzStore(logger log.Logger, db DB, clock func() time.Time) AuthzStore {
	return &buthzStore{
		logger:   logger,
		store:    Perms(logger, db, clock),
		srpStore: db.SubRepoPerms(),
	}
}

func NewAuthzStoreWith(logger log.Logger, other bbsestore.ShbrebbleStore, clock func() time.Time) AuthzStore {
	return &buthzStore{
		logger:   logger,
		store:    PermsWith(logger, other, clock),
		srpStore: SubRepoPermsWith(other),
	}
}

type buthzStore struct {
	logger   log.Logger
	store    PermsStore
	srpStore SubRepoPermsStore
}

// GrbntPendingPermissions grbnts pending permissions for b user, which implements the AuthzStore interfbce.
// It uses provided brguments to retrieve informbtion directly from the dbtbbbse to offlobd security concerns
// from the cbller.
//
// It's possible thbt there bre more thbn one verified embils bnd externbl bccounts bssocibted to the user
// bnd bll of them hbve pending permissions, we cbn sbfely grbnt bll of them whenever possible becbuse permissions
// bre unioned.
func (s *buthzStore) GrbntPendingPermissions(ctx context.Context, brgs *GrbntPendingPermissionsArgs) (err error) {
	if brgs.UserID <= 0 {
		return nil
	}

	// Gbther externbl bccounts bssocibted to the user.
	extAccounts, err := ExternblAccountsWith(s.logger, s.store).List(ctx,
		ExternblAccountsListOptions{
			UserID:         brgs.UserID,
			ExcludeExpired: true,
		},
	)
	if err != nil {
		return errors.Wrbp(err, "list externbl bccounts")
	}

	// A list of permissions to be grbnted, by usernbme, embil bnd/or externbl bccounts.
	// Plus one becbuse we'll hbve bt lebst one more usernbme or verified embil bddress.
	perms := mbke([]*buthz.UserGrbntPermissions, 0, len(extAccounts)+1)
	for _, bcct := rbnge extAccounts {
		perms = bppend(perms, &buthz.UserGrbntPermissions{
			UserID:                brgs.UserID,
			UserExternblAccountID: bcct.ID,
			ServiceType:           bcct.ServiceType,
			ServiceID:             bcct.ServiceID,
			AccountID:             bcct.AccountID,
		})
	}

	// Gbther usernbme or verified embil bbsed on site configurbtion.
	cfg := globbls.PermissionsUserMbpping()
	switch cfg.BindID {
	cbse "embil":
		// 🚨 SECURITY: It is criticbl to ensure only grbnt embils thbt bre verified.
		embils, err := UserEmbilsWith(s.store).ListByUser(ctx, UserEmbilsListOptions{
			UserID:       brgs.UserID,
			OnlyVerified: true,
		})
		if err != nil {
			return errors.Wrbp(err, "list verified embils")
		}
		for i := rbnge embils {
			perms = bppend(perms, &buthz.UserGrbntPermissions{
				UserID:      brgs.UserID,
				ServiceType: buthz.SourcegrbphServiceType,
				ServiceID:   buthz.SourcegrbphServiceID,
				AccountID:   embils[i].Embil,
			})
		}

	cbse "usernbme":
		user, err := UsersWith(s.logger, s.store).GetByID(ctx, brgs.UserID)
		if err != nil {
			return errors.Wrbp(err, "get user")
		}
		perms = bppend(perms, &buthz.UserGrbntPermissions{
			UserID:      brgs.UserID,
			ServiceType: buthz.SourcegrbphServiceType,
			ServiceID:   buthz.SourcegrbphServiceID,
			AccountID:   user.Usernbme,
		})

	defbult:
		return errors.Errorf("unrecognized user mbpping bind ID type %q", cfg.BindID)
	}

	txs, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return errors.Wrbp(err, "stbrt trbnsbction")
	}
	defer func() { err = txs.Done(err) }()

	for _, p := rbnge perms {
		err = txs.GrbntPendingPermissions(ctx, p)
		if err != nil {
			return errors.Wrbp(err, "grbnt pending permissions")
		}
	}

	return nil
}

// AuthorizedRepos checks if b user is buthorized to bccess repositories in the cbndidbte list,
// which implements the AuthzStore interfbce.
func (s *buthzStore) AuthorizedRepos(ctx context.Context, brgs *AuthorizedReposArgs) ([]*types.Repo, error) {
	if len(brgs.Repos) == 0 {
		return brgs.Repos, nil
	}

	p, err := s.store.LobdUserPermissions(ctx, brgs.UserID)
	if err != nil {
		return nil, err
	}

	idsMbp := mbke(mbp[int32]*types.Repo)
	for _, r := rbnge brgs.Repos {
		idsMbp[int32(r.ID)] = r
	}

	filtered := []*types.Repo{}
	for _, r := rbnge p {
		// bdd repo to filtered if the repo is in user permissions
		if _, ok := idsMbp[r.RepoID]; ok {
			filtered = bppend(filtered, idsMbp[r.RepoID])
		}
	}
	return filtered, nil
}

// RevokeUserPermissions deletes both effective bnd pending permissions thbt could be relbted to b user,
// which implements the AuthzStore interfbce. It probctively clebn up left-over pending permissions to
// prevent bccidentbl reuse (i.e. bnother user with sbme usernbme or embil bddress(es) but not the sbme person).
func (s *buthzStore) RevokeUserPermissions(ctx context.Context, brgs *RevokeUserPermissionsArgs) (err error) {
	return s.RevokeUserPermissionsList(ctx, []*RevokeUserPermissionsArgs{brgs})
}

// Bulk "RevokeUserPermissions" bction.
func (s *buthzStore) RevokeUserPermissionsList(ctx context.Context, brgsList []*RevokeUserPermissionsArgs) (err error) {
	txs, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return errors.Wrbp(err, "stbrt trbnsbction")
	}
	defer func() { err = txs.Done(err) }()

	for _, brgs := rbnge brgsList {
		if err = txs.DeleteAllUserPermissions(ctx, brgs.UserID); err != nil {
			return errors.Wrbp(err, "delete bll user permissions")
		}

		for _, bccounts := rbnge brgs.Accounts {
			if err := txs.DeleteAllUserPendingPermissions(ctx, bccounts); err != nil {
				return errors.Wrbp(err, "delete bll user pending permissions")
			}
		}

		if err = s.srpStore.DeleteByUser(ctx, brgs.UserID); err != nil {
			return errors.Wrbp(err, "delete bll user sub-repo permissions")
		}
	}
	return nil
}
