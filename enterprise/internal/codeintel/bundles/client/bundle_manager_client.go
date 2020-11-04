package client

import (
	"context"
	"database/sql"
	"errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/postgres"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// ErrNotFound occurs when the requested upload or bundle was evicted from disk.
var ErrNotFound = errors.New("data does not exist")

// BundleManagerClient is the interface to the precise-code-intel-bundle-manager service.
type BundleManagerClient interface {
	// BundleClient creates a client that can answer intelligence queries for a single dump.
	BundleClient(bundleID int) BundleClient
}

type bundleManagerClientImpl struct {
	codeIntelDB        *sql.DB
	observationContext *observation.Context
}

var _ BundleManagerClient = &bundleManagerClientImpl{}

func New(
	codeIntelDB *sql.DB,
	observationContext *observation.Context,
) BundleManagerClient {
	return &bundleManagerClientImpl{
		codeIntelDB:        codeIntelDB,
		observationContext: observationContext,
	}
}

// BundleClient creates a client that can answer intelligence queries for a single dump.
func (c *bundleManagerClientImpl) BundleClient(bundleID int) BundleClient {
	return &bundleClientImpl{
		bundleID: bundleID,
		store:    persistence.NewObserved(postgres.NewStore(c.codeIntelDB, bundleID), c.observationContext),
		databaseOpener: func(ctx context.Context, filename string, store persistence.Store) (database.Database, error) {
			db, err := database.OpenDatabase(ctx, filename, store)
			if err != nil {
				return nil, err
			}

			return database.NewObserved(db, filename, c.observationContext), nil
		},
	}
}
