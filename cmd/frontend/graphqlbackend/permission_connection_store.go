package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type permisionConnectionStore struct {
	db     database.DB
	roleID int32
	userID int32
}

func (pcs *permisionConnectionStore) MarshalCursor(node PermissionResolver, _ database.OrderBy) (*string, error) {
	cursor := string(node.ID())

	return &cursor, nil
}

func (pcs *permisionConnectionStore) UnmarshalCursor(cursor string, _ database.OrderBy) ([]any, error) {
	nodeID, err := UnmarshalPermissionID(graphql.ID(cursor))
	if err != nil {
		return nil, err
	}

	return []any{nodeID}, nil
}

func (pcs *permisionConnectionStore) ComputeTotal(ctx context.Context) (int32, error) {
	count, err := pcs.db.Permissions().Count(ctx, database.PermissionListOpts{
		RoleID: pcs.roleID,
		UserID: pcs.userID,
	})
	if err != nil {
		return 0, err
	}

	return int32(count), nil
}

func (pcs *permisionConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]PermissionResolver, error) {
	permissions, err := pcs.db.Permissions().List(ctx, database.PermissionListOpts{
		PaginationArgs: args,
		RoleID:         pcs.roleID,
		UserID:         pcs.userID,
	})
	if err != nil {
		return nil, err
	}

	var permissionResolvers []PermissionResolver
	for _, permission := range permissions {
		permissionResolvers = append(permissionResolvers, &permissionResolver{permission: permission})
	}

	return permissionResolvers, nil
}
