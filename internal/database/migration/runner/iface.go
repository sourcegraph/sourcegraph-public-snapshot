package runner

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/shared"
)

type Store interface {
	Transact(ctx context.Context) (Store, error)
	Done(err error) error

	Versions(ctx context.Context) (appliedVersions, pendingVersions, failedVersions []int, _ error)
	RunDDLStatements(ctx context.Context, statements []string) error
	TryLock(ctx context.Context) (bool, func(err error) error, error)
	Up(ctx context.Context, migration definition.Definition) error
	Down(ctx context.Context, migration definition.Definition) error
	WithMigrationLog(ctx context.Context, definition definition.Definition, up bool, f func() error) error
	IndexStatus(ctx context.Context, tableName, indexName string) (shared.IndexStatus, bool, error)
	Describe(ctx context.Context) (map[string]schemas.SchemaDescription, error)
}
