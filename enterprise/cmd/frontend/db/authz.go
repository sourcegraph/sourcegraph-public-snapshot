package db

import (
	"context"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
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

func (s *authzStore) GrantPendingPermissions(ctx context.Context, args *db.GrantPendingPermissionsArgs) error {
	s.init()
	return s.store.GrantPendingPermissions(ctx, args.UserID, &iauthz.UserPendingPermissions{
		BindID: args.BindID,
		Perm:   args.Perm,
		Type:   args.Type,
	})
}

func (s *authzStore) AuthorizedRepos(ctx context.Context, args *db.AuthorizedReposArgs) ([]*types.Repo, error) {
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
	filtered := args.Repos[:0]
	for _, r := range perms {
		filtered = append(filtered, r.Repo) // In-place filtering
	}
	return filtered, nil
}
