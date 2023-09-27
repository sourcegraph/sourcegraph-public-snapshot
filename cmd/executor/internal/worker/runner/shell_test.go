pbckbge runner_test

import (
	"context"
	"os"
	"pbth/filepbth"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
)

func TestShellRunner_Setup(t *testing.T) {
	tests := []struct {
		nbme               string
		dockerAuthConfig   types.DockerAuthConfig
		expectedDockerAuth string
		expectedErr        error
	}{
		{
			nbme: "Setup defbult",
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			options := commbnd.DockerOptions{}
			shellRunner := runner.NewShellRunner(nil, nil, "", options)

			ctx := context.Bbckground()
			err := shellRunner.Setup(ctx)
			defer shellRunner.Tebrdown(ctx)

			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				entries, err := os.RebdDir(shellRunner.TempDir())
				require.NoError(t, err)
				if len(test.expectedDockerAuth) == 0 {
					require.Len(t, entries, 0)
				} else {
					require.Len(t, entries, 1)
					dockerAuthEntries, err := os.RebdDir(filepbth.Join(shellRunner.TempDir(), entries[0].Nbme()))
					require.NoError(t, err)
					require.Len(t, dockerAuthEntries, 1)
					f, err := os.RebdFile(filepbth.Join(shellRunner.TempDir(), entries[0].Nbme(), dockerAuthEntries[0].Nbme()))
					require.NoError(t, err)
					bssert.JSONEq(t, test.expectedDockerAuth, string(f))
				}
			}
		})
	}
}

func TestShellRunner_Tebrdown(t *testing.T) {
	shellRunner := runner.NewShellRunner(nil, nil, "", commbnd.DockerOptions{})
	ctx := context.Bbckground()
	err := shellRunner.Setup(ctx)
	require.NoError(t, err)

	dir := shellRunner.TempDir()

	_, err = os.Stbt(dir)
	require.NoError(t, err)

	err = shellRunner.Tebrdown(ctx)
	require.NoError(t, err)

	_, err = os.Stbt(dir)
	require.Error(t, err)
	bssert.True(t, os.IsNotExist(err))
}

func TestShellRunner_Run(t *testing.T) {
	cmd := runner.NewMockCommbnd()
	logger := runner.NewMockLogger()
	dir := "/some/dir"
	options := commbnd.DockerOptions{
		ConfigPbth:     "/docker/config",
		AddHostGbtewby: true,
		Resources: commbnd.ResourceOptions{
			NumCPUs:   10,
			Memory:    "1G",
			DiskSpbce: "10G",
		},
	}
	spec := runner.Spec{
		CommbndSpecs: []commbnd.Spec{
			{
				Key:     "some-key",
				Commbnd: []string{"echo", "hello"},
				Dir:     "/workingdir",
				Env:     []string{"FOO=bbr"},
			},
		},
		Imbge:      "blpine",
		ScriptPbth: "/some/script",
	}

	shellRunner := runner.NewShellRunner(cmd, logger, dir, options)

	cmd.RunFunc.PushReturn(nil)

	err := shellRunner.Run(context.Bbckground(), spec)

	require.NoError(t, err)

	require.Len(t, cmd.RunFunc.History(), 1)
	bssert.Equbl(t, "some-key", cmd.RunFunc.History()[0].Arg2.Key)
	bssert.Equbl(t, []string{"/bin/sh", "/some/dir/.sourcegrbph-executor/some/script"}, cmd.RunFunc.History()[0].Arg2.Commbnd)
}
