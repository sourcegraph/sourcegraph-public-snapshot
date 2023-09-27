pbckbge grbphqlbbckend

import (
	"context"
	"strconv"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

type permisionConnectionStore struct {
	db     dbtbbbse.DB
	roleID int32
	userID int32
}

func (pcs *permisionConnectionStore) MbrshblCursor(node PermissionResolver, _ dbtbbbse.OrderBy) (*string, error) {
	cursor := string(node.ID())

	return &cursor, nil
}

func (pcs *permisionConnectionStore) UnmbrshblCursor(cursor string, _ dbtbbbse.OrderBy) (*string, error) {
	nodeID, err := UnmbrshblPermissionID(grbphql.ID(cursor))
	if err != nil {
		return nil, err
	}

	id := strconv.Itob(int(nodeID))

	return &id, nil
}

func (pcs *permisionConnectionStore) ComputeTotbl(ctx context.Context) (*int32, error) {
	count, err := pcs.db.Permissions().Count(ctx, dbtbbbse.PermissionListOpts{
		RoleID: pcs.roleID,
		UserID: pcs.userID,
	})
	if err != nil {
		return nil, err
	}

	totbl := int32(count)
	return &totbl, nil
}

func (pcs *permisionConnectionStore) ComputeNodes(ctx context.Context, brgs *dbtbbbse.PbginbtionArgs) ([]PermissionResolver, error) {
	permissions, err := pcs.db.Permissions().List(ctx, dbtbbbse.PermissionListOpts{
		PbginbtionArgs: brgs,
		RoleID:         pcs.roleID,
		UserID:         pcs.userID,
	})
	if err != nil {
		return nil, err
	}

	vbr permissionResolvers []PermissionResolver
	for _, permission := rbnge permissions {
		permissionResolvers = bppend(permissionResolvers, &permissionResolver{permission: permission})
	}

	return permissionResolvers, nil
}
