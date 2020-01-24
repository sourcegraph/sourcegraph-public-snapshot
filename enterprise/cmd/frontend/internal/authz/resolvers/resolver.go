package resolvers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	iauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

type Resolver struct {
	store *iauthz.Store
}

var _ graphqlbackend.AuthzResolver = &Resolver{}

func NewResolver() graphqlbackend.AuthzResolver {
	return &Resolver{
		store: iauthz.NewStore(dbconn.Global, time.Now),
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

	p := &iauthz.RepoPermissions{
		RepoID:   int32(repoID),
		Perm:     authz.Read, // Note: We currently only support read for repository permissions.
		UserIDs:  roaring.NewBitmap(),
		Provider: authz.ProviderSourcegraph,
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

	// Note: We're not wrapping these two operations in a transaction because PostgreSQL 9.6 (the minimal version
	// we support) does not support nested transactions. Besides, these two operations will acquire row-level locks
	// over 4 tables, which could greatly increase chances of causing deadlocks with other methods. Practically,
	// the result of SetRepoPermissions is much more important because it takes effect immediately. If the call of
	// the SetRepoPendingPermissions method failed, a retry from client won't hurt.
	if err = r.store.SetRepoPermissions(ctx, p); err != nil {
		return nil, err
	} else if err = r.store.SetRepoPendingPermissions(ctx, pendingBindIDs, p); err != nil {
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
		p := &iauthz.UserPermissions{
			UserID:   user.ID,
			Perm:     authz.Read, // Note: We currently only support read for repository permissions.
			Type:     authz.PermRepos,
			Provider: authz.ProviderSourcegraph,
		}
		err = r.store.LoadUserPermissions(ctx, p)
		ids = p.IDs
	} else {
		p := &iauthz.UserPendingPermissions{
			BindID: bindID,
			Perm:   authz.Read, // Note: We currently only support read for repository permissions.
			Type:   authz.PermRepos,
		}
		err = r.store.LoadUserPendingPermissions(ctx, p)
		ids = p.IDs
	}
	if err != nil && err != iauthz.ErrNotFound {
		return nil, err
	}
	// If no row is found, we return an empty list to the consumer.
	if err == iauthz.ErrNotFound {
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

	return r.store.ListPendingUsers(ctx)
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

	p := &iauthz.RepoPermissions{
		RepoID:   int32(repoID),
		Perm:     authz.Read, // Note: We currently only support read for repository permissions.
		Provider: authz.ProviderSourcegraph,
	}
	err = r.store.LoadRepoPermissions(ctx, p)
	if err != nil && err != iauthz.ErrNotFound {
		return nil, err
	}
	// If no row is found, we return an empty list to the consumer.
	if err == iauthz.ErrNotFound {
		p.UserIDs = roaring.NewBitmap()
	}

	return &userConnectionResolver{
		ids:   p.UserIDs,
		first: args.First,
		after: args.After,
	}, nil
}
