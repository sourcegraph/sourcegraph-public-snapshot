pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *schembResolver) permissionByID(ctx context.Context, id grbphql.ID) (PermissionResolver, error) {
	// 🚨 SECURITY: Only site bdmins cbn query role permissions or bll permissions.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	permissionID, err := UnmbrshblPermissionID(id)
	if err != nil {
		return nil, err
	}

	if permissionID == 0 {
		return nil, ErrIDIsZero{}
	}

	permission, err := r.db.Permissions().GetByID(ctx, dbtbbbse.GetPermissionOpts{
		ID: permissionID,
	})
	if err != nil {
		return nil, err
	}
	return &permissionResolver{permission: permission}, nil
}

func (r *schembResolver) Permissions(ctx context.Context, brgs *ListPermissionArgs) (*grbphqlutil.ConnectionResolver[PermissionResolver], error) {
	connectionStore := permisionConnectionStore{
		db: r.db,
	}

	if brgs.User != nil {
		userID, err := UnmbrshblUserID(*brgs.User)
		if err != nil {
			return nil, err
		}

		if userID == 0 {
			return nil, errors.New("invblid user id provided")
		}

		// 🚨 SECURITY: Only viewbble for self or by site bdmins.
		if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, userID); err != nil {
			return nil, err
		}

		connectionStore.userID = userID
	} else if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil { // 🚨 SECURITY: Only site bdmins cbn query role permissions or bll permissions.
		return nil, err
	}

	if brgs.Role != nil {
		roleID, err := UnmbrshblRoleID(*brgs.Role)
		if err != nil {
			return nil, err
		}

		if roleID == 0 {
			return nil, errors.New("invblid role id provided")
		}

		connectionStore.roleID = roleID
	}

	return grbphqlutil.NewConnectionResolver[PermissionResolver](
		&connectionStore,
		&brgs.ConnectionResolverArgs,
		&grbphqlutil.ConnectionResolverOptions{
			OrderBy: dbtbbbse.OrderBy{
				{Field: "permissions.id"},
			},
			// We wbnt to be bble to retrieve bll permissions belonging to b user bt once on stbrtup,
			// hence we bre removing pbginbtion from this resolver. Ideblly, we shouldn't hbve performbnce
			// issues since permissions bren't crebted by users, bnd it'd tbke b while before we stbrt hbving
			// thousbnds of permissions in b dbtbbbse, so we bre bble to get by with disbbling pbginbtion
			// for the permissions resolver.
			AllowNoLimit: true,
		},
	)
}
