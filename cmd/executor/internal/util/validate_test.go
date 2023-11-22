package util_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestValidateGitVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		exitStatus  int
		stdout      string
		expectedErr error
	}{
		{
			name:       "Version is minimum",
			exitStatus: 0,
			stdout:     "2.26",
		},
		{
			name:        "Version is below minimum",
			exitStatus:  0,
			stdout:      "1.1",
			expectedErr: errors.New("git version is too old, install at least git 2.26, current version: 1.1"),
		},
		{
			name:        "Failed to parse version",
			exitStatus:  0,
			stdout:      "",
			expectedErr: errors.New("failed to semver parse git version: "),
		},
		{
			name:        "Failed to get version",
			exitStatus:  1,
			stdout:      "failed to get version",
			expectedErr: errors.New("getting git version: 'git version': failed to get version: exit status 1"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := new(fakeCmdRunner)
			runner.On("CombinedOutput", mock.Anything, mock.Anything, mock.Anything).
				Return(test.exitStatus, fmt.Sprintf(test.stdout))

			err := util.ValidateGitVersion(context.Background(), runner)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateSrcCLIVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		latestVersion  string
		currentVersion string
		expectedErr    error
		isSrcPatchErr  bool
	}{
		{
			name:           "Matches",
			latestVersion:  "1.2.3",
			currentVersion: "1.2.3",
		},
		{
			name:           "Current patch behind",
			latestVersion:  "1.2.3",
			currentVersion: "1.2.2",
			expectedErr:    errors.New("consider upgrading actual=1.2.2, latest=1.2.3: installed src-cli is not the latest version"),
			isSrcPatchErr:  true,
		},
		{
			name:           "Latest patch behind",
			latestVersion:  "1.2.2",
			currentVersion: "1.2.3",
		},
		{
			name:           "Current minor behind",
			latestVersion:  "1.2.3",
			currentVersion: "1.1.0",
			expectedErr:    errors.New("installed src-cli is not the recommended version, consider switching actual=1.1.0, recommended=1.2.3"),
			isSrcPatchErr:  false,
		},
		{
			name:           "Latest minor behind",
			latestVersion:  "1.1.0",
			currentVersion: "1.2.0",
			expectedErr:    errors.New("installed src-cli is not the recommended version, consider switching actual=1.2.0, recommended=1.1.0"),
			isSrcPatchErr:  false,
		},
		{
			name:           "Current major behind",
			latestVersion:  "2.0.0",
			currentVersion: "1.0.0",
			expectedErr:    errors.New("installed src-cli is not the recommended version, consider switching actual=1.0.0, recommended=2.0.0"),
			isSrcPatchErr:  false,
		},
		{
			name:           "Latest major behind",
			latestVersion:  "1.0.0",
			currentVersion: "2.0.0",
			expectedErr:    errors.New("installed src-cli is not the recommended version, consider switching actual=2.0.0, recommended=1.0.0"),
			isSrcPatchErr:  false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err := json.NewEncoder(w).Encode(struct {
					Version string `json:"version"`
				}{test.latestVersion})
				require.NoError(t, err)
			}))
			defer server.Close()

			client, err := apiclient.NewBaseClient(logtest.Scoped(t), apiclient.BaseClientOptions{
				EndpointOptions: apiclient.EndpointOptions{
					URL: server.URL,
				},
			})
			require.NoError(t, err)

			runner := new(fakeCmdRunner)
			runner.On("CombinedOutput", mock.Anything, "src", []string{"version", "-client-only"}).
				Return(0, fmt.Sprintf("Current version: %s", test.currentVersion))

			err = util.ValidateSrcCLIVersion(context.Background(), runner, client)
			if test.expectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, errors.Is(err, util.ErrSrcPatchBehind), test.isSrcPatchErr)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateDockerTools(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mockFunc    func(runner *fakeCmdRunner)
		expectedErr error
	}{
		{
			name: "Docker is valid",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "docker").
					Return("", nil)
				runner.On("LookPath", "git").
					Return("", nil)
				runner.On("LookPath", "src").
					Return("", nil)
			},
		},
		{
			name: "Docker missing",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "docker").
					Return("", exec.ErrNotFound)
				runner.On("LookPath", "git").
					Return("", nil)
				runner.On("LookPath", "src").
					Return("", nil)
			},
			expectedErr: errors.New("docker not found in PATH, is it installed?\nCheck out https://docs.docker.com/get-docker/ on how to install."),
		},
		{
			name: "Docker error",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "docker").
					Return("", errors.New("failed to find docker"))
			},
			expectedErr: errors.New("failed to find docker"),
		},
		{
			name: "Git missing",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "docker").
					Return("", nil)
				runner.On("LookPath", "git").
					Return("", exec.ErrNotFound)
				runner.On("LookPath", "src").
					Return("", nil)
			},
			expectedErr: errors.New("git not found in PATH, is it installed?\nUse your package manager, or build from source."),
		},
		{
			name: "Git error",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "docker").
					Return("", nil)
				runner.On("LookPath", "git").
					Return("", errors.New("failed to find git"))
			},
			expectedErr: errors.New("failed to find git"),
		},
		{
			name: "Src missing",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "docker").
					Return("", nil)
				runner.On("LookPath", "git").
					Return("", nil)
				runner.On("LookPath", "src").
					Return("", exec.ErrNotFound)
			},
			expectedErr: errors.New("src not found in PATH, is it installed?\nRun executor install src-cli, or refer to https://github.com/sourcegraph/src-cli to install src-cli yourself."),
		},
		{
			name: "Src error",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "docker").
					Return("", nil)
				runner.On("LookPath", "git").
					Return("", nil)
				runner.On("LookPath", "src").
					Return("", errors.New("failed to find src"))
			},
			expectedErr: errors.New("failed to find src"),
		},
		{
			name: "Missing all",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "docker").
					Return("", exec.ErrNotFound)
				runner.On("LookPath", "git").
					Return("", exec.ErrNotFound)
				runner.On("LookPath", "src").
					Return("", exec.ErrNotFound)
			},
			expectedErr: errors.New("3 errors occurred:\n\t* docker not found in PATH, is it installed?\nCheck out https://docs.docker.com/get-docker/ on how to install.\n\t* git not found in PATH, is it installed?\nUse your package manager, or build from source.\n\t* src not found in PATH, is it installed?\nRun executor install src-cli, or refer to https://github.com/sourcegraph/src-cli to install src-cli yourself."),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := new(fakeCmdRunner)
			if test.mockFunc != nil {
				test.mockFunc(runner)
			}

			err := util.ValidateDockerTools(runner)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateFirecrackerTools(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mockFunc    func(runner *fakeCmdRunner)
		expectedErr error
	}{
		{
			name: "Firecracker is valid",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "dmsetup").
					Return("", nil)
				runner.On("LookPath", "losetup").
					Return("", nil)
				runner.On("LookPath", "mkfs.ext4").
					Return("", nil)
				runner.On("LookPath", "strings").
					Return("", nil)
			},
		},
		{
			name: "Dmsetup missing",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "dmsetup").
					Return("", exec.ErrNotFound)
				runner.On("LookPath", "losetup").
					Return("", nil)
				runner.On("LookPath", "mkfs.ext4").
					Return("", nil)
				runner.On("LookPath", "strings").
					Return("", nil)
			},
			expectedErr: errors.New("dmsetup not found in PATH, is it installed?"),
		},
		{
			name: "Dmsetup error",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "dmsetup").
					Return("", errors.New("failed to find"))
			},
			expectedErr: errors.New("failed to find"),
		},
		{
			name: "Losetup missing",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "dmsetup").
					Return("", nil)
				runner.On("LookPath", "losetup").
					Return("", exec.ErrNotFound)
				runner.On("LookPath", "mkfs.ext4").
					Return("", nil)
				runner.On("LookPath", "strings").
					Return("", nil)
			},
			expectedErr: errors.New("losetup not found in PATH, is it installed?"),
		},
		{
			name: "Losetup error",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "dmsetup").
					Return("", nil)
				runner.On("LookPath", "losetup").
					Return("", errors.New("failed to find"))
			},
			expectedErr: errors.New("failed to find"),
		},
		{
			name: "Mkfs missing",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "dmsetup").
					Return("", nil)
				runner.On("LookPath", "losetup").
					Return("", nil)
				runner.On("LookPath", "mkfs.ext4").
					Return("", exec.ErrNotFound)
				runner.On("LookPath", "strings").
					Return("", nil)
			},
			expectedErr: errors.New("mkfs.ext4 not found in PATH, is it installed?"),
		},
		{
			name: "Mkfs error",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "dmsetup").
					Return("", nil)
				runner.On("LookPath", "losetup").
					Return("", nil)
				runner.On("LookPath", "mkfs.ext4").
					Return("", errors.New("failed to find"))
			},
			expectedErr: errors.New("failed to find"),
		},
		{
			name: "Strings missing",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "dmsetup").
					Return("", nil)
				runner.On("LookPath", "losetup").
					Return("", nil)
				runner.On("LookPath", "mkfs.ext4").
					Return("", nil)
				runner.On("LookPath", "strings").
					Return("", exec.ErrNotFound)
			},
			expectedErr: errors.New("strings not found in PATH, is it installed?"),
		},
		{
			name: "Strings error",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "dmsetup").
					Return("", nil)
				runner.On("LookPath", "losetup").
					Return("", nil)
				runner.On("LookPath", "mkfs.ext4").
					Return("", nil)
				runner.On("LookPath", "strings").
					Return("", errors.New("failed to find"))
			},
			expectedErr: errors.New("failed to find"),
		},
		{
			name: "All missing",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "dmsetup").
					Return("", exec.ErrNotFound)
				runner.On("LookPath", "losetup").
					Return("", exec.ErrNotFound)
				runner.On("LookPath", "mkfs.ext4").
					Return("", exec.ErrNotFound)
				runner.On("LookPath", "strings").
					Return("", exec.ErrNotFound)
			},
			expectedErr: errors.New("4 errors occurred:\n\t* dmsetup not found in PATH, is it installed?\n\t* losetup not found in PATH, is it installed?\n\t* mkfs.ext4 not found in PATH, is it installed?\n\t* strings not found in PATH, is it installed?"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := new(fakeCmdRunner)
			if test.mockFunc != nil {
				test.mockFunc(runner)
			}

			err := util.ValidateFirecrackerTools(runner)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateIgniteInstalled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mockFunc    func(runner *fakeCmdRunner)
		expectedErr error
	}{
		{
			name: "Ignite valid",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "ignite").
					Return("", nil)
				runner.On("CombinedOutput", mock.Anything, mock.Anything, mock.Anything).
					Return(0, "0.10.5")
			},
		},
		{
			name: "Unsupported ignite version",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "ignite").
					Return("", nil)
				runner.On("CombinedOutput", mock.Anything, mock.Anything, mock.Anything).
					Return(0, "1.2.3")
			},
			expectedErr: errors.New("using unsupported ignite version, if things don't work alright, consider switching to the supported version. have=1.2.3, want=0.10.5"),
		},
		{
			name: "Missing ignite",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "ignite").
					Return("", exec.ErrNotFound)
			},
			expectedErr: errors.New("Ignite not found in PATH. Is it installed correctly?\n\nTry running \"executor install ignite\", or:\n  $ curl -sfLo ignite https://github.com/sourcegraph/ignite/releases/download/v0.10.5/ignite-amd64\n  $ chmod +x ignite\n  $ mv ignite /usr/local/bin"),
		},
		{
			name: "Failed to parse ignite version",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "ignite").
					Return("", nil)
				runner.On("CombinedOutput", mock.Anything, mock.Anything, mock.Anything).
					Return(0, "")
			},
			expectedErr: errors.New("failed to parse ignite version: Invalid Semantic Version"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := new(fakeCmdRunner)
			if test.mockFunc != nil {
				test.mockFunc(runner)
			}

			err := util.ValidateIgniteInstalled(context.Background(), runner)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TODO: visit this later. It uses os.Stat on a constant path. Maybe mock os.Stat or use a temp dir??
//func TestValidateCNIInstalled(t *testing.T) {
//	tests := []struct {
//		name string
//	}{
//	}
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//		})
//	}
//}
