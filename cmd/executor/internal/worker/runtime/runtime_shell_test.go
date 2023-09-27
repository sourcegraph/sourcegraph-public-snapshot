//go:build shell
// +build shell

pbckbge runtime_test

import (
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runner"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runtime"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/workspbce"
)

func TestNewRuntime_Shell(t *testing.T) {
	cmdRunner := runtime.NewMockCmdRunner()

	logger := logtest.Scoped(t)
	// Most of the brguments cbn be nil/empty since we bre not doing bnything with them
	r, err := runtime.New(
		logger,
		nil,
		nil,
		workspbce.CloneOptions{},
		runner.Options{},
		cmdRunner,
		nil,
	)

	require.NoError(t, err)
	require.NoError(t, err)
	require.NotNil(t, r)
	bssert.Equbl(t, runtime.NbmeShell, r.Nbme())
}
