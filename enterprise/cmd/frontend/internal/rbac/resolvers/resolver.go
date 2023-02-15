package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Resolver is the GraphQL resolver of all things related to batch changes.
type Resolver struct {
	logger log.Logger
	db     database.DB
}

func New(logger log.Logger, db database.DB) gql.RBACResolver {
	return &Resolver{logger: logger, db: db}
}

func (r *Resolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		roleIDKind: func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return r.roleByID(ctx, id)
		},
		permissionIDKind: func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return r.permissionByID(ctx, id)
		},
	}
}

func (r *Resolver) AssignPermissionsToRole(ctx context.Context, args gql.AssignPermissionToRoleArgs) (*gql.EmptyResponse, error) {
	// 🚨 SECURITY: Only site administrators can assign a permission to a role.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if len(args.Permissions) == 0 {
		return nil, errors.New("permissions is required")
	}

	roleID, err := unmarshalRoleID(args.Role)
	if err != nil {
		return nil, err
	}

	opts := database.BulkAssignPermissionsToRoleOpts{}
	opts.RoleID = roleID

	for _, p := range args.Permissions {
		pID, err := unmarshalPermissionID(p)
		if err != nil {
			return nil, err
		}
		opts.Permissions = append(opts.Permissions, pID)
	}

	if _, err = r.db.RolePermissions().BulkAssignPermissionsToRole(ctx, opts); err != nil {
		return nil, err
	}

	return &gql.EmptyResponse{}, nil
}
