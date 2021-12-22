package runner

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
)

type Store interface {
	Version(ctx context.Context) (int, bool, bool, error)
	Lock(ctx context.Context) (bool, func(err error) error, error)
	TryLock(ctx context.Context) (bool, func(err error) error, error)
	Up(ctx context.Context, migration definition.Definition) error
	Down(ctx context.Context, migration definition.Definition) error
}
