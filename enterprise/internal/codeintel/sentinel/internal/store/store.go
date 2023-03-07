package store

import (
	"context"

	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store interface {
	VulnerabilityByID(ctx context.Context, id int) (_ shared.Vulnerability, _ bool, err error)
	GetVulnerabilitiesByIDs(ctx context.Context, ids ...int) (_ []shared.Vulnerability, err error)
	GetVulnerabilities(ctx context.Context, args shared.GetVulnerabilitiesArgs) (_ []shared.Vulnerability, _ int, err error)
	InsertVulnerabilities(ctx context.Context, vulnerabilities []shared.Vulnerability) (_ int, err error)

	VulnerabilityMatchByID(ctx context.Context, id int) (shared.VulnerabilityMatch, bool, error)
	GetVulnerabilityMatches(ctx context.Context, args shared.GetVulnerabilityMatchesArgs) ([]shared.VulnerabilityMatch, int, error)
	ScanMatches(ctx context.Context, batchSize int) (numReferencesScanned int, numVulnerabilityMatches int, _ error)
}

type store struct {
	db         *basestore.Store
	logger     logger.Logger
	operations *operations
}

// New returns a new sentinel store.
func New(observationCtx *observation.Context, db database.DB) Store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		logger:     logger.Scoped("sentinel.store", ""),
		operations: newOperations(observationCtx),
	}
}
