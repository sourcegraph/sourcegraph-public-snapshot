package runner

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
)

type Store interface {
	// Required to build a locker from the same handle
	basestore.ShareableStore

	Version(ctx context.Context) (int, bool, bool, error)
	Up(ctx context.Context, migration definition.Definition) error
	Down(ctx context.Context, migration definition.Definition) error
}
