package runtime_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/runner"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/runtime"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/workspace"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		runnerOpts     runner.Options
		mockFunc       func(cmdRunner *runtime.MockCmdRunner)
		expectedName   runtime.Name
		expectedErr    error
		assertMockFunc func(t *testing.T, cmdRunner *runtime.MockCmdRunner)
	}{
		{
			name: "Docker",
			mockFunc: func(cmdRunner *runtime.MockCmdRunner) {
				cmdRunner.LookPathFunc.SetDefaultReturn("", nil)
			},
			expectedName: runtime.NameDocker,
			assertMockFunc: func(t *testing.T, cmdRunner *runtime.MockCmdRunner) {
				require.Len(t, cmdRunner.LookPathFunc.History(), 3)
				assert.Equal(t, "docker", cmdRunner.LookPathFunc.History()[0].Arg0)
				assert.Equal(t, "git", cmdRunner.LookPathFunc.History()[1].Arg0)
				assert.Equal(t, "src", cmdRunner.LookPathFunc.History()[2].Arg0)
			},
		},
		{
			name: "Firecracker",
			runnerOpts: runner.Options{
				FirecrackerOptions: runner.FirecrackerOptions{
					Enabled: true,
				},
			},
			mockFunc: func(cmdRunner *runtime.MockCmdRunner) {
				// ValidateFirecrackerTools + ValidateIgniteInstalled
				cmdRunner.LookPathFunc.SetDefaultReturn("", nil)
				// ValidateIgniteInstalled (GetIgniteVersion)
				cmdRunner.CombinedOutputFunc.SetDefaultReturn([]byte("v0.10.5"), nil)
				// ValidateCNIInstalled
				cmdRunner.StatFunc.PushReturn(&fileutil.FileInfo{Mode_: os.ModeDir}, nil)
				cmdRunner.StatFunc.PushReturn(&fileutil.FileInfo{Mode_: 0}, nil)
				cmdRunner.StatFunc.PushReturn(&fileutil.FileInfo{Mode_: 0}, nil)
				cmdRunner.StatFunc.PushReturn(&fileutil.FileInfo{Mode_: 0}, nil)
				cmdRunner.StatFunc.PushReturn(&fileutil.FileInfo{Mode_: 0}, nil)
				cmdRunner.StatFunc.PushReturn(&fileutil.FileInfo{Mode_: 0}, nil)
				cmdRunner.StatFunc.PushReturn(&fileutil.FileInfo{Mode_: 0}, nil)
				cmdRunner.StatFunc.PushReturn(&fileutil.FileInfo{Mode_: 0}, nil)
			},
			expectedName: runtime.NameFirecracker,
			assertMockFunc: func(t *testing.T, cmdRunner *runtime.MockCmdRunner) {
				require.Len(t, cmdRunner.LookPathFunc.History(), 5)
				assert.Equal(t, "dmsetup", cmdRunner.LookPathFunc.History()[0].Arg0)
				assert.Equal(t, "losetup", cmdRunner.LookPathFunc.History()[1].Arg0)
				assert.Equal(t, "mkfs.ext4", cmdRunner.LookPathFunc.History()[2].Arg0)
				assert.Equal(t, "strings", cmdRunner.LookPathFunc.History()[3].Arg0)
				assert.Equal(t, "ignite", cmdRunner.LookPathFunc.History()[4].Arg0)

				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 1)
				assert.Equal(t, "ignite", cmdRunner.CombinedOutputFunc.History()[0].Arg1)
				assert.Equal(t, []string{"version", "-o", "short"}, cmdRunner.CombinedOutputFunc.History()[0].Arg2)

				require.Len(t, cmdRunner.StatFunc.History(), 8)
				assert.Equal(t, "/opt/cni/bin", cmdRunner.StatFunc.History()[0].Arg0)
				assert.Equal(t, "/opt/cni/bin/bandwidth", cmdRunner.StatFunc.History()[1].Arg0)
				assert.Equal(t, "/opt/cni/bin/bridge", cmdRunner.StatFunc.History()[2].Arg0)
				assert.Equal(t, "/opt/cni/bin/firewall", cmdRunner.StatFunc.History()[3].Arg0)
				assert.Equal(t, "/opt/cni/bin/host-local", cmdRunner.StatFunc.History()[4].Arg0)
				assert.Equal(t, "/opt/cni/bin/isolation", cmdRunner.StatFunc.History()[5].Arg0)
				assert.Equal(t, "/opt/cni/bin/loopback", cmdRunner.StatFunc.History()[6].Arg0)
				assert.Equal(t, "/opt/cni/bin/portmap", cmdRunner.StatFunc.History()[7].Arg0)
			},
		},
		{
			name: "Missing Firecracker tools",
			runnerOpts: runner.Options{
				FirecrackerOptions: runner.FirecrackerOptions{
					Enabled: true,
				},
			},
			mockFunc: func(cmdRunner *runtime.MockCmdRunner) {
				cmdRunner.LookPathFunc.SetDefaultReturn("", exec.ErrNotFound)
			},
			expectedName: runtime.NameFirecracker,
			assertMockFunc: func(t *testing.T, cmdRunner *runtime.MockCmdRunner) {
				require.Len(t, cmdRunner.LookPathFunc.History(), 4)
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 0)
				require.Len(t, cmdRunner.StatFunc.History(), 0)
			},
			expectedErr: errors.New("4 errors occurred:\n\t* dmsetup not found in PATH, is it installed?\n\t* losetup not found in PATH, is it installed?\n\t* mkfs.ext4 not found in PATH, is it installed?\n\t* strings not found in PATH, is it installed?"),
		},
		{
			name: "Ignite not installed",
			runnerOpts: runner.Options{
				FirecrackerOptions: runner.FirecrackerOptions{
					Enabled: true,
				},
			},
			mockFunc: func(cmdRunner *runtime.MockCmdRunner) {
				// ValidateFirecrackerTools + ValidateIgniteInstalled
				cmdRunner.LookPathFunc.PushReturn("", nil)
				cmdRunner.LookPathFunc.PushReturn("", nil)
				cmdRunner.LookPathFunc.PushReturn("", nil)
				cmdRunner.LookPathFunc.PushReturn("", nil)
				cmdRunner.LookPathFunc.PushReturn("", exec.ErrNotFound)
			},
			expectedName: runtime.NameFirecracker,
			assertMockFunc: func(t *testing.T, cmdRunner *runtime.MockCmdRunner) {
				require.Len(t, cmdRunner.LookPathFunc.History(), 5)
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 0)
				require.Len(t, cmdRunner.StatFunc.History(), 0)
			},
			expectedErr: errors.New("Ignite not found in PATH. Is it installed correctly?\n\nTry running \"executor install ignite\", or:\n  $ curl -sfLo ignite https://github.com/sourcegraph/ignite/releases/download/v0.10.5/ignite-amd64\n  $ chmod +x ignite\n  $ mv ignite /usr/local/bin"),
		},
		{
			name: "Wrong ignite version",
			runnerOpts: runner.Options{
				FirecrackerOptions: runner.FirecrackerOptions{
					Enabled: true,
				},
			},
			mockFunc: func(cmdRunner *runtime.MockCmdRunner) {
				// ValidateFirecrackerTools + ValidateIgniteInstalled
				cmdRunner.LookPathFunc.SetDefaultReturn("", nil)
				// ValidateIgniteInstalled (GetIgniteVersion)
				cmdRunner.CombinedOutputFunc.SetDefaultReturn([]byte("v0.1.0"), nil)
			},
			expectedName: runtime.NameFirecracker,
			assertMockFunc: func(t *testing.T, cmdRunner *runtime.MockCmdRunner) {
				require.Len(t, cmdRunner.LookPathFunc.History(), 5)
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 1)
				require.Len(t, cmdRunner.StatFunc.History(), 0)
			},
			expectedErr: errors.New("using unsupported ignite version, if things don't work alright, consider switching to the supported version. have=0.1.0, want=0.10.5"),
		},
		{
			name: "CNI not installed",
			runnerOpts: runner.Options{
				FirecrackerOptions: runner.FirecrackerOptions{
					Enabled: true,
				},
			},
			mockFunc: func(cmdRunner *runtime.MockCmdRunner) {
				// ValidateFirecrackerTools + ValidateIgniteInstalled
				cmdRunner.LookPathFunc.SetDefaultReturn("", nil)
				// ValidateIgniteInstalled (GetIgniteVersion)
				cmdRunner.CombinedOutputFunc.SetDefaultReturn([]byte("v0.10.5"), nil)
				// ValidateCNIInstalled
				cmdRunner.StatFunc.PushReturn(nil, os.ErrNotExist)
			},
			expectedName: runtime.NameFirecracker,
			assertMockFunc: func(t *testing.T, cmdRunner *runtime.MockCmdRunner) {
				require.Len(t, cmdRunner.LookPathFunc.History(), 5)
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 1)
				require.Len(t, cmdRunner.StatFunc.History(), 1)
			},
			expectedErr: errors.New("2 errors occurred:\n\t* Cannot find directory /opt/cni/bin. Are the CNI plugins for firecracker installed correctly?\n\t* Cannot find CNI plugins [bandwidth bridge firewall host-local isolation loopback portmap], are the CNI plugins for firecracker installed correctly?\nTo install the CNI plugins used by ignite run \"executor install cni\" or the following:\n  $ mkdir -p /opt/cni/bin\n  $ curl -sSL https://github.com/containernetworking/plugins/releases/download/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz | tar -xz -C /opt/cni/bin\n  $ curl -sSL https://github.com/AkihiroSuda/cni-isolation/releases/download/v0.0.4/cni-isolation-amd64.tgz | tar -xz -C /opt/cni/bin"),
		},
		{
			name: "No Runtime",
			mockFunc: func(cmdRunner *runtime.MockCmdRunner) {
				cmdRunner.LookPathFunc.PushReturn("", exec.ErrNotFound)
			},
			expectedErr: runtime.ErrNoRuntime,
			assertMockFunc: func(t *testing.T, cmdRunner *runtime.MockCmdRunner) {
				require.Len(t, cmdRunner.LookPathFunc.History(), 3)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmdRunner := runtime.NewMockCmdRunner()
			if test.mockFunc != nil {
				test.mockFunc(cmdRunner)
			}
			logger := logtest.Scoped(t)
			// Most of the arguments can be nil/empty since we are not doing anything with them
			r, err := runtime.New(
				logger,
				nil,
				nil,
				workspace.CloneOptions{},
				test.runnerOpts,
				cmdRunner,
				nil,
			)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.Nil(t, r)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, r)
				assert.Equal(t, test.expectedName, r.Name())
			}

			if test.assertMockFunc != nil {
				test.assertMockFunc(t, cmdRunner)
			}
		})
	}
}

func TestNew_Kubernetes(t *testing.T) {
	tempFile, err := os.CreateTemp("", "kubeconfig")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	content := `
apiVersion: v1
clusters:
- cluster:
    server: https://localhost:8080
  name: foo-cluster
contexts:
- context:
    cluster: foo-cluster
    user: foo-user
    namespace: bar
  name: foo-context
current-context: foo-context
kind: Config
`
	err = os.WriteFile(tempFile.Name(), []byte(content), 0644)
	require.NoError(t, err)

	r, err := runtime.New(
		logtest.Scoped(t),
		nil,
		nil,
		workspace.CloneOptions{},
		runner.Options{
			KubernetesOptions: runner.KubernetesOptions{
				Enabled:          true,
				ConfigPath:       tempFile.Name(),
				ContainerOptions: command.KubernetesContainerOptions{},
			},
		},
		runtime.NewMockCmdRunner(),
		nil,
	)
	require.NoError(t, err)

	assert.Equal(t, runtime.NameKubernetes, r.Name())
}

func TestCommandKey(t *testing.T) {
	tests := []struct {
		name        string
		runtimeName runtime.Name
		key         string
		index       int
		expectedKey string
	}{
		{
			name:        "Docker",
			runtimeName: runtime.NameDocker,
			key:         "step.1.pre",
			index:       0,
			expectedKey: "step.docker.step.1.pre",
		},
		{
			name:        "Docker with index",
			runtimeName: runtime.NameDocker,
			key:         "",
			index:       1,
			expectedKey: "step.docker.1",
		},
		{
			name:        "Firecracker",
			runtimeName: runtime.NameFirecracker,
			key:         "step.1.pre",
			index:       0,
			expectedKey: "step.docker.step.1.pre",
		},
		{
			name:        "Firecracker with index",
			runtimeName: runtime.NameFirecracker,
			key:         "",
			index:       1,
			expectedKey: "step.docker.1",
		},
		{
			name:        "Kubernetes",
			runtimeName: runtime.NameKubernetes,
			key:         "step.1.pre",
			index:       0,
			expectedKey: "step.kubernetes.step.1.pre",
		},
		{
			name:        "Kubernetes with index",
			runtimeName: runtime.NameKubernetes,
			key:         "",
			index:       1,
			expectedKey: "step.kubernetes.1",
		},
		{
			name:        "Shell",
			runtimeName: runtime.NameShell,
			key:         "step.1.pre",
			index:       0,
			expectedKey: "step.docker.step.1.pre",
		},
		{
			name:        "Shell with index",
			runtimeName: runtime.NameShell,
			key:         "",
			index:       1,
			expectedKey: "step.docker.1",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := runtime.CommandKey(test.runtimeName, test.key, test.index)
			assert.Equal(t, test.expectedKey, key)
		})
	}
}
