package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	iauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

// NewAuthzStore returns an OSS db.AuthzStore set with enterprise implementation.
func NewAuthzStore() db.AuthzStore {
	return &authzStore{}
}

type authzStore struct {
	once  sync.Once
	store *iauthz.Store
}

// The global DB is not initialized when NewAuthzStore is called, so we need to create the store at runtime.
func (s *authzStore) init() {
	s.once.Do(func() {
		s.store = iauthz.NewStore(dbconn.Global, time.Now)
	})
}

// 🚨 SECURITY: It is the caller's responsibility to ensure the supplied email is verified.
func (s *authzStore) GrantPendingPermissions(ctx context.Context, args *db.GrantPendingPermissionsArgs) error {
	s.init()

	// Note: we purposely don't check cfg.PermissionsUserMapping.Enabled here because admin could disable the
	// feature by mistake while a user has valid pending permissions.
	var bindID string
	cfg := globals.PermissionsUserMapping()
	switch cfg.BindID {
	case "email":
		bindID = args.VerifiedEmail
	case "username":
		bindID = args.Username
	default:
		return fmt.Errorf("unrecognized user mapping bind ID type %q", cfg.BindID)
	}

	if bindID == "" {
		return nil
	}
	return s.store.GrantPendingPermissions(ctx, args.UserID, &iauthz.UserPendingPermissions{
		BindID: bindID,
		Perm:   args.Perm,
		Type:   args.Type,
	})
}

func (s *authzStore) AuthorizedRepos(ctx context.Context, args *db.AuthorizedReposArgs) ([]*types.Repo, error) {
	if len(args.Repos) == 0 {
		return args.Repos, nil
	}

	s.init()

	p := &iauthz.UserPermissions{
		UserID:   args.UserID,
		Perm:     args.Perm,
		Type:     args.Type,
		Provider: args.Provider,
	}
	if err := s.store.LoadUserPermissions(ctx, p); err != nil {
		if err == iauthz.ErrNotFound {
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
