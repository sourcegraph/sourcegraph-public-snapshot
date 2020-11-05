package client

import (
	"database/sql"
	"errors"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// ErrNotFound occurs when the requested upload or bundle was evicted from disk.
var ErrNotFound = errors.New("data does not exist")

func New(
	codeIntelDB *sql.DB,
	observationContext *observation.Context,
) BundleManagerClient {
	return &bundleManagerClientImpl{
		codeIntelDB:        codeIntelDB,
		observationContext: observationContext,
	}
}
