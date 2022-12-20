package database

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewAuthzStore returns an OSS database.AuthzStore set with enterprise implementation.
func NewAuthzStore(logger log.Logger, db database.DB, clock func() time.Time) database.AuthzStore {
	return &authzStore{
		logger:   logger,
		store:    Perms(logger, db, clock),
		srpStore: database.SubRepoPermsWith(basestore.NewWithHandle(db.Handle())),
	}
}

func NewAuthzStoreWith(logger log.Logger, other basestore.ShareableStore, clock func() time.Time) database.AuthzStore {
	return &authzStore{
		logger:   logger,
		store:    PermsWith(logger, other, clock),
		srpStore: database.SubRepoPermsWith(other),
	}
}

type authzStore struct {
	logger   log.Logger
	store    PermsStore
	srpStore database.SubRepoPermsStore
}

// GrantPendingPermissions grants pending permissions for a user, which implements the database.AuthzStore interface.
// It uses provided arguments to retrieve information directly from the database to offload security concerns
// from the caller.
//
// It's possible that there are more than one verified emails and external accounts associated to the user
// and all of them have pending permissions, we can safely grant all of them whenever possible because permissions
// are unioned.
func (s *authzStore) GrantPendingPermissions(ctx context.Context, args *database.GrantPendingPermissionsArgs) (err error) {
	if args.UserID <= 0 {
		return nil
	}

	// Gather external accounts associated to the user.
	extAccounts, err := database.ExternalAccountsWith(s.logger, s.store).List(ctx,
		database.ExternalAccountsListOptions{
			UserID:         args.UserID,
			ExcludeExpired: true,
		},
	)
	if err != nil {
		return errors.Wrap(err, "list external accounts")
	}

	// A list of permissions to be granted, by username, email and/or external accounts.
	// Plus one because we'll have at least one more username or verified email address.
	perms := make([]*authz.UserPendingPermissions, 0, len(extAccounts)+1)
	for _, acct := range extAccounts {
		perms = append(perms, &authz.UserPendingPermissions{
			ServiceType: acct.ServiceType,
			ServiceID:   acct.ServiceID,
			BindID:      acct.AccountID,
			Perm:        args.Perm,
			Type:        args.Type,
		})
	}

	// Gather username or verified email based on site configuration.
	cfg := globals.PermissionsUserMapping()
	switch cfg.BindID {
	case "email":
		// 🚨 SECURITY: It is critical to ensure only grant emails that are verified.
		emails, err := database.UserEmailsWith(s.store).ListByUser(ctx, database.UserEmailsListOptions{
			UserID:       args.UserID,
			OnlyVerified: true,
		})
		if err != nil {
			return errors.Wrap(err, "list verified emails")
		}
		for i := range emails {
			perms = append(perms, &authz.UserPendingPermissions{
				ServiceType: authz.SourcegraphServiceType,
				ServiceID:   authz.SourcegraphServiceID,
				BindID:      emails[i].Email,
				Perm:        args.Perm,
				Type:        args.Type,
			})
		}

	case "username":
		user, err := database.UsersWith(s.logger, s.store).GetByID(ctx, args.UserID)
		if err != nil {
			return errors.Wrap(err, "get user")
		}
		perms = append(perms, &authz.UserPendingPermissions{
			ServiceType: authz.SourcegraphServiceType,
			ServiceID:   authz.SourcegraphServiceID,
			BindID:      user.Username,
			Perm:        args.Perm,
			Type:        args.Type,
		})

	default:
		return errors.Errorf("unrecognized user mapping bind ID type %q", cfg.BindID)
	}

	txs, err := s.store.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "start transaction")
	}
	defer func() { err = txs.Done(err) }()

	for _, p := range perms {
		err = txs.GrantPendingPermissions(ctx, args.UserID, p)
		if err != nil {
			return errors.Wrap(err, "grant pending permissions")
		}
	}

	return nil
}

// AuthorizedRepos checks if a user is authorized to access repositories in the candidate list,
// which implements the database.AuthzStore interface.
func (s *authzStore) AuthorizedRepos(ctx context.Context, args *database.AuthorizedReposArgs) ([]*types.Repo, error) {
	if len(args.Repos) == 0 {
		return args.Repos, nil
	}

	p := &authz.UserPermissions{
		UserID: args.UserID,
		Perm:   args.Perm,
		Type:   args.Type,
	}
	if err := s.store.LoadUserPermissions(ctx, p); err != nil {
		if err == authz.ErrPermsNotFound {
			return []*types.Repo{}, nil
		}
		return nil, err
	}

	perms := p.AuthorizedRepos(args.Repos)
	filtered := make([]*types.Repo, len(perms))
	for i, r := range perms {
		filtered[i] = r.Repo
	}
	return filtered, nil
}

// RevokeUserPermissions deletes both effective and pending permissions that could be related to a user,
// which implements the database.AuthzStore interface. It proactively clean up left-over pending permissions to
// prevent accidental reuse (i.e. another user with same username or email address(es) but not the same person).
func (s *authzStore) RevokeUserPermissions(ctx context.Context, args *database.RevokeUserPermissionsArgs) (err error) {
	return s.RevokeUserPermissionsList(ctx, []*database.RevokeUserPermissionsArgs{args})
}

// Bulk "RevokeUserPermissions" action.
func (s *authzStore) RevokeUserPermissionsList(ctx context.Context, argsList []*database.RevokeUserPermissionsArgs) (err error) {
	txs, err := s.store.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "start transaction")
	}
	defer func() { err = txs.Done(err) }()

	for _, args := range argsList {
		if err = txs.DeleteAllUserPermissions(ctx, args.UserID); err != nil {
			return errors.Wrap(err, "delete all user permissions")
		}

		for _, accounts := range args.Accounts {
			if err := txs.DeleteAllUserPendingPermissions(ctx, accounts); err != nil {
				return errors.Wrap(err, "delete all user pending permissions")
			}
		}

		if err = s.srpStore.DeleteByUser(ctx, args.UserID); err != nil {
			return errors.Wrap(err, "delete all user sub-repo permissions")
		}
	}
	return nil
}
