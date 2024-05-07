package scim

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/scim"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Init sets SCIMHandler to a real handler.
func Init(ctx context.Context, observationCtx *observation.Context, db database.DB, _ codeintel.Services, _ conftypes.UnifiedWatchable, s *enterprise.Services) error {
	s.SCIMHandler = scim.NewHandler(ctx, db, observationCtx)

	return nil
}
