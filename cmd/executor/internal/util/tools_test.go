package util_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGetGitVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		exitStatus      int
		stdout          string
		expectedVersion string
		expectedErr     error
	}{
		{
			name:            "Success",
			stdout:          "1.2.3",
			expectedVersion: "1.2.3",
		},
		{
			name:            "Success with prefix",
			stdout:          "git version 1.2.3",
			expectedVersion: "1.2.3",
		},
		{
			name:        "Error",
			exitStatus:  1,
			stdout:      "failed to get version",
			expectedErr: errors.New("'git version': failed to get version: exit status 1"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := new(fakeCmdRunner)
			runner.On("CombinedOutput", mock.Anything, "git", []string{"version"}).
				Return(test.exitStatus, test.stdout)

			version, err := util.GetGitVersion(context.Background(), runner)
			if test.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedVersion, version)
			}
		})
	}
}

func TestGetSrcVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		exitStatus      int
		stdout          string
		expectedVersion string
		expectedErr     error
	}{
		{
			name:            "Success",
			stdout:          "1.2.3",
			expectedVersion: "1.2.3",
		},
		{
			name:            "Success with prefix",
			stdout:          "Current version: 1.2.3",
			expectedVersion: "1.2.3",
		},
		{
			name:        "Error",
			exitStatus:  1,
			stdout:      "failed to get version",
			expectedErr: errors.New("'src version -client-only': failed to get version: exit status 1"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := new(fakeCmdRunner)
			runner.On("CombinedOutput", mock.Anything, "src", []string{"version", "-client-only"}).
				Return(test.exitStatus, test.stdout)

			version, err := util.GetSrcVersion(context.Background(), runner)
			if test.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedVersion, version)
			}
		})
	}
}

func TestGetDockerVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		exitStatus      int
		stdout          string
		expectedVersion string
		expectedErr     error
	}{
		{
			name:            "Success",
			stdout:          "1.2.3",
			expectedVersion: "1.2.3",
		},
		{
			name:        "Error",
			exitStatus:  1,
			stdout:      "failed to get version",
			expectedErr: errors.New("'docker version -f {{.Server.Version}}': failed to get version: exit status 1"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := new(fakeCmdRunner)
			runner.On("CombinedOutput", mock.Anything, "docker", []string{"version", "-f", "{{.Server.Version}}"}).
				Return(test.exitStatus, test.stdout)

			version, err := util.GetDockerVersion(context.Background(), runner)
			if test.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedVersion, version)
			}
		})
	}
}

func TestGetIgniteVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		exitStatus      int
		stdout          string
		expectedVersion string
		expectedErr     error
	}{
		{
			name:            "Success",
			stdout:          "1.2.3",
			expectedVersion: "1.2.3",
		},
		{
			name:        "Error",
			exitStatus:  1,
			stdout:      "failed to get version",
			expectedErr: errors.New("'ignite version -o short': failed to get version: exit status 1"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := new(fakeCmdRunner)
			runner.On("CombinedOutput", mock.Anything, "ignite", []string{"version", "-o", "short"}).
				Return(test.exitStatus, test.stdout)

			version, err := util.GetIgniteVersion(context.Background(), runner)
			if test.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedVersion, version)
			}
		})
	}
}
