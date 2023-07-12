package migrations

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type StoreFactory interface {
	Store(ctx context.Context, schemaName string) (*basestore.Store, error)
}
