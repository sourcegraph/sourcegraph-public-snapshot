package worker_test

import (
	"context"
	"os/exec"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/mock"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
)

func TestHandler_PreDequeue(t *testing.T) {
	logger := logtest.Scoped(t)

	tests := []struct {
		name              string
		options           worker.Options
		expectedDequeue   bool
		expectedExtraArgs any
		expectedErr       error
	}{
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmdRunner := new(fakeCmdRunner)
		})
	}
}

type fakeCmdRunner struct {
	mock.Mock
}

var _ util.CmdRunner = &fakeCmdRunner{}

func (f *fakeCmdRunner) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	panic("not needed")
}

func (f *fakeCmdRunner) CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	calledArgs := f.Called(ctx, name, args)
	return calledArgs.Get(0).([]byte), calledArgs.Error(1)
}

func (f *fakeCmdRunner) LookPath(file string) (string, error) {
	panic("not needed")
}

type fakeCommand struct {
	mock.Mock
}

var _ command.Command = &fakeCommand{}

func (f *fakeCommand) Run(ctx context.Context, cmdLogger command.Logger, spec command.Spec) error {
	args := f.Called(ctx, cmdLogger, spec)
	return args.Error(0)
}
