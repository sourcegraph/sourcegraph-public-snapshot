package resolvers

import (
	"context"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type permissionsInfoResolver struct {
	db           database.DB
	userID       int32
	repoID       api.RepoID
	perms        authz.Perms
	syncedAt     time.Time
	updatedAt    time.Time
	source       string
	unrestricted bool
}

func (r *permissionsInfoResolver) Permissions() []string {
	return strings.Split(strings.ToUpper(r.perms.String()), ",")
}

func (r *permissionsInfoResolver) SyncedAt() *gqlutil.DateTime {
	if r.syncedAt.IsZero() {
		return nil
	}
	return &gqlutil.DateTime{Time: r.syncedAt}
}

func (r *permissionsInfoResolver) UpdatedAt() *gqlutil.DateTime {
	return gqlutil.FromTime(r.updatedAt)
}

func (r *permissionsInfoResolver) Source() *string {
	if r.source == "" {
		return nil
	}

	return &r.source
}

func (r *permissionsInfoResolver) Unrestricted(_ context.Context) bool {
	return r.unrestricted
}

var permissionsInfoRepositoryConnectionMaxPageSize = 100

var permissionsInfoRepositoryConnectionOptions = &gqlutil.ConnectionResolverOptions{
	OrderBy:     database.OrderBy{{Field: "repo.id"}},
	Ascending:   true,
	MaxPageSize: permissionsInfoRepositoryConnectionMaxPageSize,
}

func (r *permissionsInfoResolver) Repositories(_ context.Context, args graphqlbackend.PermissionsInfoRepositoriesArgs) (*gqlutil.ConnectionResolver[graphqlbackend.PermissionsInfoRepositoryResolver], error) {
	if r.userID == 0 {
		return nil, nil
	}

	query := ""
	if args.Query != nil {
		query = *args.Query
	}

	connectionStore := &permissionsInfoRepositoriesStore{
		userID: r.userID,
		db:     r.db,
		query:  query,
	}

	return gqlutil.NewConnectionResolver[graphqlbackend.PermissionsInfoRepositoryResolver](connectionStore, &args.ConnectionResolverArgs, permissionsInfoRepositoryConnectionOptions)
}

type permissionsInfoRepositoriesStore struct {
	userID int32
	db     database.DB
	query  string
}

func (s *permissionsInfoRepositoriesStore) MarshalCursor(node graphqlbackend.PermissionsInfoRepositoryResolver, _ database.OrderBy) (*string, error) {
	cursor := string(node.ID())

	return &cursor, nil
}

func (s *permissionsInfoRepositoriesStore) UnmarshalCursor(cursor string, _ database.OrderBy) ([]any, error) {
	repoID, err := graphqlbackend.UnmarshalRepositoryID(graphql.ID(cursor))
	if err != nil {
		return nil, err
	}

	return []any{int32(repoID)}, nil
}

func (s *permissionsInfoRepositoriesStore) ComputeTotal(ctx context.Context) (int32, error) {
	count, err := s.db.Repos().Count(actor.WithActor(ctx, actor.FromUser(s.userID)), database.ReposListOptions{Query: s.query})
	if err != nil {
		return 0, err
	}

	return int32(count), nil
}

func (s *permissionsInfoRepositoriesStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]graphqlbackend.PermissionsInfoRepositoryResolver, error) {
	permissions, err := s.db.Perms().ListUserPermissions(ctx, s.userID, &database.ListUserPermissionsArgs{Query: s.query, PaginationArgs: args})
	if err != nil {
		return nil, err
	}

	var permissionResolvers []graphqlbackend.PermissionsInfoRepositoryResolver
	for _, perm := range permissions {
		permissionResolvers = append(permissionResolvers, permissionsInfoRepositoryResolver{perm: perm, db: s.db})
	}

	return permissionResolvers, nil
}

type permissionsInfoRepositoryResolver struct {
	db   database.DB
	perm *database.UserPermission
}

func (r permissionsInfoRepositoryResolver) ID() graphql.ID {
	return graphqlbackend.MarshalRepositoryID(r.perm.Repo.ID)
}

func (r permissionsInfoRepositoryResolver) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	repo, err := r.db.Repos().Get(ctx, r.perm.Repo.ID)
	// If the errcode is NotFound, we return nil, nil, as we know that the repo should exist at this point.
	// So this should mean that this user simply cannot see the repository.
	if err != nil && errcode.IsNotFound(err) {
		return nil, nil
	}
	return graphqlbackend.NewRepositoryResolver(r.db, gitserver.NewClient("graphql.authz.permissions"), repo), err
}

func (r permissionsInfoRepositoryResolver) Reason() string {
	return string(r.perm.Reason)
}

func (r permissionsInfoRepositoryResolver) UpdatedAt() *gqlutil.DateTime {
	return gqlutil.FromTime(r.perm.UpdatedAt)
}

var permissionsInfoUserConnectionMaxPageSize = 100

var permissionsInfoUserConnectionOptions = &gqlutil.ConnectionResolverOptions{
	OrderBy:     database.OrderBy{{Field: "users.username"}},
	Ascending:   true,
	MaxPageSize: permissionsInfoUserConnectionMaxPageSize,
}

func (r *permissionsInfoResolver) Users(ctx context.Context, args graphqlbackend.PermissionsInfoUsersArgs) (*gqlutil.ConnectionResolver[graphqlbackend.PermissionsInfoUserResolver], error) {
	if r.repoID == 0 {
		return nil, nil
	}

	query := ""
	if args.Query != nil {
		query = *args.Query
	}

	connectionStore := &permissionsInfoUsersStore{
		ctx:    ctx,
		repoID: r.repoID,
		db:     r.db,
		query:  query,
	}

	return gqlutil.NewConnectionResolver[graphqlbackend.PermissionsInfoUserResolver](connectionStore, &args.ConnectionResolverArgs, permissionsInfoUserConnectionOptions)
}

type permissionsInfoUsersStore struct {
	ctx    context.Context
	repoID api.RepoID
	db     database.DB
	query  string
}

func (s *permissionsInfoUsersStore) MarshalCursor(node graphqlbackend.PermissionsInfoUserResolver, _ database.OrderBy) (*string, error) {
	cursor := node.User(s.ctx).Username()

	return &cursor, nil
}

func (s *permissionsInfoUsersStore) UnmarshalCursor(cursor string, _ database.OrderBy) ([]any, error) {
	return []any{cursor}, nil
}

// TODO(naman): implement total count
func (s *permissionsInfoUsersStore) ComputeTotal(ctx context.Context) (int32, error) {
	return 0, nil
}

func (s *permissionsInfoUsersStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]graphqlbackend.PermissionsInfoUserResolver, error) {
	permissions, err := s.db.Perms().ListRepoPermissions(ctx, s.repoID, &database.ListRepoPermissionsArgs{Query: s.query, PaginationArgs: args})
	if err != nil {
		return nil, err
	}

	permissionResolvers := make([]graphqlbackend.PermissionsInfoUserResolver, 0, len(permissions))
	for _, perm := range permissions {
		permissionResolvers = append(permissionResolvers, permissionsInfoUserResolver{perm: perm, db: s.db})
	}

	return permissionResolvers, nil
}

type permissionsInfoUserResolver struct {
	db   database.DB
	perm *database.RepoPermission
}

func (r permissionsInfoUserResolver) ID() graphql.ID {
	return graphqlbackend.MarshalUserID(r.perm.User.ID)
}

func (r permissionsInfoUserResolver) User(ctx context.Context) *graphqlbackend.UserResolver {
	return graphqlbackend.NewUserResolver(ctx, r.db, r.perm.User)
}

func (r permissionsInfoUserResolver) Reason() string {
	return string(r.perm.Reason)
}

func (r permissionsInfoUserResolver) UpdatedAt() *gqlutil.DateTime {
	return gqlutil.FromTime(r.perm.UpdatedAt)
}
