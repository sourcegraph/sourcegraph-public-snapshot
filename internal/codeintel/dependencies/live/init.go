// This package contains the version of the dependencies service initialization that should
// be used in all non-test uses. We need to have a separate package here because gitserver
// requires the dependencies service (abstractly), but our gitService dependency requires
// the internal/vcs/git. This breaks that circular dependency.
package live

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// GetService creates or returns an already-initialized dependencies service. If the service is
// new, it will use the given database handle and syncer instance. If the given syncer is nil,
// then ErrorSyncer will be used instead.
func GetService(db database.DB) *dependencies.Service {
	return dependencies.GetService(db)
}

// TestService creates a fresh dependencies service with the given database handle and syncer
// instance. If the given syncer is nil, then ErrorSyncer will be used instead.
func TestService(db database.DB) *dependencies.Service {
	return dependencies.TestService(db)
}
