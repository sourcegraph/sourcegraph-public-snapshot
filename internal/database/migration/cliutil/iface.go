package cliutil

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
)

type Runner interface {
	Run(ctx context.Context, options runner.Options) error
	Validate(ctx context.Context, schemaNames ...string) error
}

type RunnerFactory func(ctx context.Context, schemaNames []string) (Runner, error)
