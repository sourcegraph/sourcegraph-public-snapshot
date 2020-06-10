package v0

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite/store"
)

// Migrate v0: A no-op migration that serves as the base of all "legacy" (un-versioned)
// schema versions as a place to uniformly begin the migration process.
func Migrate(ctx context.Context, s *store.Store, serializer serialization.Serializer) error {
	return nil
}
