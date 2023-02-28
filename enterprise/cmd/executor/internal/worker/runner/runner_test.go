package runner_test

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
)

type fakeCommand struct {
	mock.Mock
}

var _ command.Command = &fakeCommand{}

func (f *fakeCommand) Run(ctx context.Context, cmdLogger command.Logger, spec command.Spec) error {
	args := f.Called(ctx, cmdLogger, spec)
	return args.Error(0)
}
