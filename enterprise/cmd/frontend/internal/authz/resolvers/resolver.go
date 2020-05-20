package resolvers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	edb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

type Resolver struct {
	store             *edb.PermsStore
	repoupdaterClient interface {
		SchedulePermsSync(ctx context.Context, args protocol.PermsSyncRequest) error
	}
}

func NewResolver(db dbutil.DB, clock func() time.Time) graphqlbackend.AuthzResolver {
	return &Resolver{
		store:             edb.NewPermsStore(db, clock),
		repoupdaterClient: repoupdater.DefaultClient,
	}
}

func (r *Resolver) SetRepositoryPermissionsForUsers(ctx context.Context, args *graphqlbackend.RepoPermsArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can mutate repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	repoID, err := graphqlbackend.UnmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}
	// Make sure the repo ID is valid.
	if _, err = db.Repos.Get(ctx, repoID); err != nil {
		return nil, err
	}

	// Filter out bind IDs that only contains whitespaces.
	bindIDs := args.BindIDs[:0]
	for i := range args.BindIDs {
		args.BindIDs[i] = strings.TrimSpace(args.BindIDs[i])
		if len(args.BindIDs[i]) == 0 {
			continue
		}
		bindIDs = append(bindIDs, args.BindIDs[i])
	}

	bindIDSet := make(map[string]struct{})
	for i := range bindIDs {
		bindIDSet[bindIDs[i]] = struct{}{}
	}

	p := &authz.RepoPermissions{
		RepoID:  int32(repoID),
		Perm:    authz.Read, // Note: We currently only support read for repository permissions.
		UserIDs: roaring.NewBitmap(),
	}
	cfg := globals.PermissionsUserMapping()
	switch cfg.BindID {
	case "email":
		emails, err := db.UserEmails.GetVerifiedEmails(ctx, bindIDs...)
		if err != nil {
			return nil, err
		}

		for i := range emails {
			p.UserIDs.Add(uint32(emails[i].UserID))
			delete(bindIDSet, emails[i].Email)
		}

	case "username":
		users, err := db.Users.GetByUsernames(ctx, bindIDs...)
		if err != nil {
			return nil, err
		}

		for i := range users {
			p.UserIDs.Add(uint32(users[i].ID))
			delete(bindIDSet, users[i].Username)
		}

	default:
		return nil, fmt.Errorf("unrecognized user mapping bind ID type %q", cfg.BindID)
	}

	pendingBindIDs := make([]string, 0, len(bindIDSet))
	for id := range bindIDSet {
		pendingBindIDs = append(pendingBindIDs, id)
	}

	txs, err := r.store.Transact(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "start transaction")
	}
	defer txs.Done(&err)

	accounts := &extsvc.Accounts{
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountIDs:  pendingBindIDs,
	}

	if err = txs.SetRepoPermissions(ctx, p); err != nil {
		return nil, errors.Wrap(err, "set repository permissions")
	} else if err = txs.SetRepoPendingPermissions(ctx, accounts, p); err != nil {
		return nil, errors.Wrap(err, "set repository pending permissions")
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) ScheduleRepositoryPermissionsSync(ctx context.Context, args *graphqlbackend.RepositoryIDArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can query repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	repoID, err := graphqlbackend.UnmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}

	err = r.repoupdaterClient.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
		RepoIDs: []api.RepoID{repoID},
	})
	if err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) ScheduleUserPermissionsSync(ctx context.Context, args *graphqlbackend.UserIDArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can query repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	userID, err := graphqlbackend.UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	err = r.repoupdaterClient.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
		UserIDs: []int32{userID},
	})
	if err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) AuthorizedUserRepositories(ctx context.Context, args *graphqlbackend.AuthorizedRepoArgs) (graphqlbackend.RepositoryConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins can query repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	var (
		err    error
		bindID string
		user   *types.User
	)
	if args.Email != nil {
		bindID = *args.Email
		// 🚨 SECURITY: It is critical to ensure the email is verified.
		user, err = db.Users.GetByVerifiedEmail(ctx, *args.Email)
	} else if args.Username != nil {
		bindID = *args.Username
		user, err = db.Users.GetByUsername(ctx, *args.Username)
	} else {
		return nil, errors.New("neither email nor username is given to identify a user")
	}
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}

	var ids *roaring.Bitmap
	if user != nil {
		p := &authz.UserPermissions{
			UserID: user.ID,
			Perm:   authz.Read, // Note: We currently only support read for repository permissions.
			Type:   authz.PermRepos,
		}
		err = r.store.LoadUserPermissions(ctx, p)
		ids = p.IDs
	} else {
		p := &authz.UserPendingPermissions{
			ServiceType: authz.SourcegraphServiceType,
			ServiceID:   authz.SourcegraphServiceID,
			BindID:      bindID,
			Perm:        authz.Read, // Note: We currently only support read for repository permissions.
			Type:        authz.PermRepos,
		}
		err = r.store.LoadUserPendingPermissions(ctx, p)
		ids = p.IDs
	}
	if err != nil && err != authz.ErrPermsNotFound {
		return nil, err
	}
	// If no row is found, we return an empty list to the consumer.
	if err == authz.ErrPermsNotFound {
		ids = roaring.NewBitmap()
	}

	return &repositoryConnectionResolver{
		ids:   ids,
		first: args.First,
		after: args.After,
	}, nil
}

func (r *Resolver) UsersWithPendingPermissions(ctx context.Context) ([]string, error) {
	// 🚨 SECURITY: Only site admins can query repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	return r.store.ListPendingUsers(ctx, authz.SourcegraphServiceType, authz.SourcegraphServiceID)
}

func (r *Resolver) AuthorizedUsers(ctx context.Context, args *graphqlbackend.RepoAuthorizedUserArgs) (graphqlbackend.UserConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins can query repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	repoID, err := graphqlbackend.UnmarshalRepositoryID(args.RepositoryID)
	if err != nil {
		return nil, err
	}
	// Make sure the repo ID is valid.
	if _, err = db.Repos.Get(ctx, repoID); err != nil {
		return nil, err
	}

	p := &authz.RepoPermissions{
		RepoID: int32(repoID),
		Perm:   authz.Read, // Note: We currently only support read for repository permissions.
	}
	err = r.store.LoadRepoPermissions(ctx, p)
	if err != nil && err != authz.ErrPermsNotFound {
		return nil, err
	}
	// If no row is found, we return an empty list to the consumer.
	if err == authz.ErrPermsNotFound {
		p.UserIDs = roaring.NewBitmap()
	}

	return &userConnectionResolver{
		ids:   p.UserIDs,
		first: args.First,
		after: args.After,
	}, nil
}

type permissionsInfoResolver struct {
	perms     authz.Perms
	syncedAt  time.Time
	updatedAt time.Time
}

func (r *permissionsInfoResolver) Permissions() []string {
	return strings.Split(strings.ToUpper(r.perms.String()), ",")
}

func (r *permissionsInfoResolver) SyncedAt() *graphqlbackend.DateTime {
	if r.syncedAt.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.syncedAt}
}

func (r *permissionsInfoResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.updatedAt}
}

func (r *Resolver) RepositoryPermissionsInfo(ctx context.Context, id graphql.ID) (graphqlbackend.PermissionsInfoResolver, error) {
	// 🚨 SECURITY: Only site admins can query repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	repoID, err := graphqlbackend.UnmarshalRepositoryID(id)
	if err != nil {
		return nil, err
	}
	// Make sure the repo ID is valid and not soft-deleted.
	if _, err = db.Repos.Get(ctx, repoID); err != nil {
		return nil, err
	}

	p := &authz.RepoPermissions{
		RepoID: int32(repoID),
		Perm:   authz.Read, // Note: We currently only support read for repository permissions.
	}
	err = r.store.LoadRepoPermissions(ctx, p)
	if err != nil && err != authz.ErrPermsNotFound {
		return nil, err
	}

	if err == authz.ErrPermsNotFound {
		return nil, nil // It is acceptable to have no permissions information, i.e. nullable.
	}

	return &permissionsInfoResolver{
		perms:     p.Perm,
		syncedAt:  p.SyncedAt,
		updatedAt: p.UpdatedAt,
	}, nil
}

func (r *Resolver) UserPermissionsInfo(ctx context.Context, id graphql.ID) (graphqlbackend.PermissionsInfoResolver, error) {
	// 🚨 SECURITY: Only site admins can query user permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	userID, err := graphqlbackend.UnmarshalUserID(id)
	if err != nil {
		return nil, err
	}
	// Make sure the user ID is valid and not soft-deleted.
	if _, err = db.Users.GetByID(ctx, userID); err != nil {
		return nil, err
	}

	p := &authz.UserPermissions{
		UserID: userID,
		Perm:   authz.Read, // Note: We currently only support read for repository permissions.
		Type:   authz.PermRepos,
	}
	err = r.store.LoadUserPermissions(ctx, p)
	if err != nil && err != authz.ErrPermsNotFound {
		return nil, err
	}

	if err == authz.ErrPermsNotFound {
		return nil, nil // It is acceptable to have no permissions information, i.e. nullable.
	}

	return &permissionsInfoResolver{
		perms:     p.Perm,
		syncedAt:  p.SyncedAt,
		updatedAt: p.UpdatedAt,
	}, nil
}
