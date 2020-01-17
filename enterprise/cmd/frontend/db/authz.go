package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	iauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
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

	cfg := conf.Get().SiteConfiguration
	// Note: we purposely don't check cfg.PermissionsUserMapping.Enabled here because admin could disable the
	// feature by mistake while a user has valid pending permissions.
	if cfg.PermissionsUserMapping == nil {
		return nil
	}

	var bindID string
	switch cfg.PermissionsUserMapping.BindID {
	case "email":
		bindID = args.VerifiedEmail
	case "username":
		bindID = args.Username
	default:
		return fmt.Errorf("unrecognized user mapping bind ID type %q", cfg.PermissionsUserMapping.BindID)
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
