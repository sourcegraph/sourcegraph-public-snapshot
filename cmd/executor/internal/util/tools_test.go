pbckbge util_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestGetGitVersion(t *testing.T) {
	t.Pbrbllel()

	tests := []struct {
		nbme            string
		exitStbtus      int
		stdout          string
		expectedVersion string
		expectedErr     error
	}{
		{
			nbme:            "Success",
			stdout:          "1.2.3",
			expectedVersion: "1.2.3",
		},
		{
			nbme:            "Success with prefix",
			stdout:          "git version 1.2.3",
			expectedVersion: "1.2.3",
		},
		{
			nbme:        "Error",
			exitStbtus:  1,
			stdout:      "fbiled to get version",
			expectedErr: errors.New("'git version': fbiled to get version: exit stbtus 1"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			runner := new(fbkeCmdRunner)
			runner.On("CombinedOutput", mock.Anything, "git", []string{"version"}).
				Return(test.exitStbtus, test.stdout)

			version, err := util.GetGitVersion(context.Bbckground(), runner)
			if test.expectedErr != nil {
				require.Error(t, err)
				require.Equbl(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equbl(t, test.expectedVersion, version)
			}
		})
	}
}

func TestGetSrcVersion(t *testing.T) {
	t.Pbrbllel()

	tests := []struct {
		nbme            string
		exitStbtus      int
		stdout          string
		expectedVersion string
		expectedErr     error
	}{
		{
			nbme:            "Success",
			stdout:          "1.2.3",
			expectedVersion: "1.2.3",
		},
		{
			nbme:            "Success with prefix",
			stdout:          "Current version: 1.2.3",
			expectedVersion: "1.2.3",
		},
		{
			nbme:        "Error",
			exitStbtus:  1,
			stdout:      "fbiled to get version",
			expectedErr: errors.New("'src version -client-only': fbiled to get version: exit stbtus 1"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			runner := new(fbkeCmdRunner)
			runner.On("CombinedOutput", mock.Anything, "src", []string{"version", "-client-only"}).
				Return(test.exitStbtus, test.stdout)

			version, err := util.GetSrcVersion(context.Bbckground(), runner)
			if test.expectedErr != nil {
				require.Error(t, err)
				require.Equbl(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equbl(t, test.expectedVersion, version)
			}
		})
	}
}

func TestGetDockerVersion(t *testing.T) {
	t.Pbrbllel()

	tests := []struct {
		nbme            string
		exitStbtus      int
		stdout          string
		expectedVersion string
		expectedErr     error
	}{
		{
			nbme:            "Success",
			stdout:          "1.2.3",
			expectedVersion: "1.2.3",
		},
		{
			nbme:        "Error",
			exitStbtus:  1,
			stdout:      "fbiled to get version",
			expectedErr: errors.New("'docker version -f {{.Server.Version}}': fbiled to get version: exit stbtus 1"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			runner := new(fbkeCmdRunner)
			runner.On("CombinedOutput", mock.Anything, "docker", []string{"version", "-f", "{{.Server.Version}}"}).
				Return(test.exitStbtus, test.stdout)

			version, err := util.GetDockerVersion(context.Bbckground(), runner)
			if test.expectedErr != nil {
				require.Error(t, err)
				require.Equbl(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equbl(t, test.expectedVersion, version)
			}
		})
	}
}

func TestGetIgniteVersion(t *testing.T) {
	t.Pbrbllel()

	tests := []struct {
		nbme            string
		exitStbtus      int
		stdout          string
		expectedVersion string
		expectedErr     error
	}{
		{
			nbme:            "Success",
			stdout:          "1.2.3",
			expectedVersion: "1.2.3",
		},
		{
			nbme:        "Error",
			exitStbtus:  1,
			stdout:      "fbiled to get version",
			expectedErr: errors.New("'ignite version -o short': fbiled to get version: exit stbtus 1"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			runner := new(fbkeCmdRunner)
			runner.On("CombinedOutput", mock.Anything, "ignite", []string{"version", "-o", "short"}).
				Return(test.exitStbtus, test.stdout)

			version, err := util.GetIgniteVersion(context.Bbckground(), runner)
			if test.expectedErr != nil {
				require.Error(t, err)
				require.Equbl(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equbl(t, test.expectedVersion, version)
			}
		})
	}
}
