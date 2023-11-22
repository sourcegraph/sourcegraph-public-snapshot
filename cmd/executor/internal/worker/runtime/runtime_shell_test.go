//go:build shell
// +build shell

package runtime_test

import (
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/runner"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/runtime"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/workspace"
)

func TestNewRuntime_Shell(t *testing.T) {
	cmdRunner := runtime.NewMockCmdRunner()

	logger := logtest.Scoped(t)
	// Most of the arguments can be nil/empty since we are not doing anything with them
	r, err := runtime.New(
		logger,
		nil,
		nil,
		workspace.CloneOptions{},
		runner.Options{},
		cmdRunner,
		nil,
	)

	require.NoError(t, err)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, runtime.NameShell, r.Name())
}
