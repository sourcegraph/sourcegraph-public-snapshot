package cliutil

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
)

type Runner interface {
	Run(ctx context.Context, options runner.Options) error
	Validate(ctx context.Context, schemaNames ...string) error
	Store(ctx context.Context, schemaName string) (Store, error)
}

type Store interface {
	WithMigrationLog(ctx context.Context, definition definition.Definition, up bool, f func() error) error
}

type RunnerFactory func(ctx context.Context, schemaNames []string) (Runner, error)

type runnerShim struct {
	*runner.Runner
}

func NewShim(runner *runner.Runner) Runner {
	return &runnerShim{Runner: runner}
}

func (r *runnerShim) Store(ctx context.Context, schemaName string) (Store, error) {
	return r.Runner.Store(ctx, schemaName)
}
