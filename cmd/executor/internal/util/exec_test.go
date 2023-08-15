package util_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/util"
)

type fakeCmdRunner struct {
	mock.Mock
}

var _ util.CmdRunner = &fakeCmdRunner{}

func (f *fakeCmdRunner) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	panic("not needed")
}

func (f *fakeCmdRunner) CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	cs := []string{"-test.run=TestExecCommandHelper", "--"}
	cs = append(cs, args...)
	calledArgs := f.Called(ctx, name, args)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{
		"GO_WANT_HELPER_PROCESS=1",
		fmt.Sprintf("EXIT_STATUS=%d", calledArgs.Int(0)),
		fmt.Sprintf("STDOUT=%s", calledArgs.String(1)),
	}
	out, err := cmd.CombinedOutput()
	return out, err
}

// TestExecCommandHelper a fake test that fakeExecCommand will run instead of calling the actual exec.CommandContext.
func TestExecCommandHelper(t *testing.T) {
	// Since this function must be big T test. We don't want to actually test anything. So if GO_WANT_HELPER_PROCESS
	// is not set, just exit right away.
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	_, err := fmt.Fprint(os.Stdout, os.Getenv("STDOUT"))
	require.NoError(t, err)

	i, err := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	require.NoError(t, err)

	os.Exit(i)
}

func (f *fakeCmdRunner) LookPath(file string) (string, error) {
	args := f.Called(file)
	return args.String(0), args.Error(1)
}

func (f *fakeCmdRunner) Stat(filename string) (os.FileInfo, error) {
	panic("not needed")
}
