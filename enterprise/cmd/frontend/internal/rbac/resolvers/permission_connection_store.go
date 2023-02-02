package resolvers

import (
	"context"
	"strconv"

	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type permisionConnectionStore struct {
	db     database.DB
	roleID int32
	userID int32
}

func (pcs *permisionConnectionStore) MarshalCursor(node gql.PermissionResolver, _ database.OrderBy) (*string, error) {
	cursor := string(node.ID())

	return &cursor, nil
}

func (pcs *permisionConnectionStore) UnmarshalCursor(cursor string, _ database.OrderBy) (*string, error) {
	nodeID, err := unmarshalPermissionID(graphql.ID(cursor))
	if err != nil {
		return nil, err
	}

	id := strconv.Itoa(int(nodeID))

	return &id, nil
}

func (pcs *permisionConnectionStore) ComputeTotal(ctx context.Context) (*int32, error) {
	count, err := pcs.db.Permissions().Count(ctx, database.PermissionListOpts{
		RoleID: pcs.roleID,
		UserID: pcs.userID,
	})
	if err != nil {
		return nil, err
	}

	total := int32(count)
	return &total, nil
}

func (pcs *permisionConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]gql.PermissionResolver, error) {
	permissions, err := pcs.db.Permissions().List(ctx, database.PermissionListOpts{
		PaginationArgs: args,
		RoleID:         pcs.roleID,
		UserID:         pcs.userID,
	})
	if err != nil {
		return nil, err
	}

	var permissionResolvers []gql.PermissionResolver
	for _, permission := range permissions {
		permissionResolvers = append(permissionResolvers, &permissionResolver{permission: permission})
	}

	return permissionResolvers, nil
}
