package cliutil

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type Runner interface {
	Run(ctx context.Context, options runner.Options) error
	Validate(ctx context.Context, schemaNames ...string) error
	Store(ctx context.Context, schemaName string) (Store, error)
}

type Store interface {
	WithMigrationLog(ctx context.Context, definition definition.Definition, up bool, f func() error) error
	Describe(ctx context.Context) (map[string]schemas.SchemaDescription, error)
	Versions(ctx context.Context) (appliedVersions, pendingVersions, failedVersions []int, _ error)
	RunDDLStatements(ctx context.Context, statements []string) error
}

// OutputFactory allows providing global output that might not be instantiated at compile time.
type OutputFactory func() *output.Output

type RunnerFactory func(schemaNames []string) (*runner.Runner, error)
