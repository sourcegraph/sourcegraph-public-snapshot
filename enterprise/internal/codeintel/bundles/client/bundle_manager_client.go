package client

import (
	"database/sql"
	"errors"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// ErrNotFound occurs when the requested upload or bundle was evicted from disk.
var ErrNotFound = errors.New("data does not exist")

// BundleManagerClient is the interface to the precise-code-intel-bundle-manager service.
type BundleManagerClient interface {
	// BundleClient creates a client that can answer intelligence queries for a single dump.
	BundleClient() BundleClient
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
func (c *bundleManagerClientImpl) BundleClient() BundleClient {
	return &bundleClientImpl{
		codeIntelDB:        c.codeIntelDB,
		observationContext: c.observationContext,
	}
}
