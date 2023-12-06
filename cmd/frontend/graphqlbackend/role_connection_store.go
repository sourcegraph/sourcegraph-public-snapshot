package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type roleConnectionStore struct {
	db     database.DB
	system bool
	userID int32
}

func (rcs *roleConnectionStore) MarshalCursor(node RoleResolver, _ database.OrderBy) (*string, error) {
	cursor := string(node.ID())

	return &cursor, nil
}

func (rcs *roleConnectionStore) UnmarshalCursor(cursor string, _ database.OrderBy) ([]any, error) {
	nodeID, err := UnmarshalRoleID(graphql.ID(cursor))
	if err != nil {
		return nil, err
	}

	return []any{nodeID}, nil
}

func (rcs *roleConnectionStore) ComputeTotal(ctx context.Context) (int32, error) {
	count, err := rcs.db.Roles().Count(ctx, database.RolesListOptions{
		UserID: rcs.userID,
	})
	if err != nil {
		return 0, err
	}

	return int32(count), nil
}

func (rcs *roleConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]RoleResolver, error) {
	roles, err := rcs.db.Roles().List(ctx, database.RolesListOptions{
		PaginationArgs: args,
		System:         rcs.system,
		UserID:         rcs.userID,
	})
	if err != nil {
		return nil, err
	}

	var roleResolvers []RoleResolver
	for _, role := range roles {
		roleResolvers = append(roleResolvers, &roleResolver{role: role, db: rcs.db})
	}

	return roleResolvers, nil
}
