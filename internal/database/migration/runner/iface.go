package runner

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
)

type Store interface {
	Transact(ctx context.Context) (Store, error)
	Done(err error) error

	Version(ctx context.Context) (int, bool, bool, error)
	Versions(ctx context.Context) (appliedVersions, pendingVersions, failedVersions []int, _ error)
	Lock(ctx context.Context) (bool, func(err error) error, error)
	TryLock(ctx context.Context) (bool, func(err error) error, error)
	Up(ctx context.Context, migration definition.Definition) error
	Down(ctx context.Context, migration definition.Definition) error
	WithMigrationLog(ctx context.Context, definition definition.Definition, up bool, f func() error) error
}
