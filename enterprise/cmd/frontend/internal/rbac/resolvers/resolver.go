package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
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

func (r *Resolver) roleByID(ctx context.Context, id graphql.ID) (graphqlbackend.RoleResolver, error) {
	roleID, err := unmarshalRoleID(id)
	if err != nil {
		return nil, err
	}

	if roleID == 0 {
		return nil, ErrIDIsZero{}
	}

	role, err := r.db.Roles().GetByID(ctx, database.GetRoleOpts{
		ID: roleID,
	})
	if err != nil {
		return nil, err
	}
	return &roleResolver{role: role}, nil
}

func (r *Resolver) permissionByID(ctx context.Context, id graphql.ID) (graphqlbackend.PermissionResolver, error) {
	permissionID, err := unmarshalPermissionID(id)
	if err != nil {
		return nil, err
	}

	if permissionID == 0 {
		return nil, ErrIDIsZero{}
	}

	permission, err := r.db.Permissions().GetByID(ctx, database.GetPermissionOpts{
		ID: permissionID,
	})
	if err != nil {
		return nil, err
	}
	return &permissionResolver{permission: permission}, nil
}

func (r *Resolver) Role(ctx context.Context, args *gql.RoleArgs) (gql.RoleResolver, error) {
	roleID, err := unmarshalRoleID(args.ID)
	if err != nil {
		return nil, err
	}

	role, err := r.db.Roles().GetByID(ctx, database.GetRoleOpts{
		ID: roleID,
	})
	if err != nil {
		return nil, err
	}
	return &roleResolver{role: role}, nil
}

func (r *Resolver) Roles(ctx context.Context, args *gql.ListRoleArgs) (gql.RoleConnectionResolver, error) {
	var opts = database.RolesListOptions{}

	if args.User != nil {
		userID, err := gql.UnmarshalUserID(*args.User)
		if err != nil {
			return nil, err
		}

		opts.UserID = userID
	}

	return &roleConnectionResolver{
		db:   r.db,
		opts: opts,
	}, nil
}

func (r *Resolver) Permissions(ctx context.Context, args *gql.ListPermissionArgs) (gql.PermissionConnectionResolver, error) {
	var opts = database.PermissionListOpts{}

	if args.Role != nil {
		roleID, err := unmarshalRoleID(*args.Role)
		if err != nil {
			return nil, err
		}

		opts.RoleID = roleID
	}

	if args.User != nil {
		userID, err := gql.UnmarshalUserID(*args.User)
		if err != nil {
			return nil, err
		}

		opts.UserID = userID
	}

	return &permissionConnectionResolver{
		db:   r.db,
		opts: opts,
	}, nil
}
