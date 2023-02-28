package runner_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/runner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	errors "github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestFirecrackerRunner_Setup(t *testing.T) {
	tests := []struct {
		name             string
		workspaceDevice  string
		vmName           string
		options          runner.FirecrackerOptions
		dockerAuthConfig types.DockerAuthConfig
		mockFunc         func(t *testing.T, cmd *fakeCommand, ops *command.Operations)
		expectedEntries  map[string]string
		expectedErr      error
	}{
		{
			name:            "Setup default",
			workspaceDevice: "/dev/sda",
			vmName:          "test",
			mockFunc: func(t *testing.T, cmd *fakeCommand, ops *command.Operations) {
				// Start command
				cmd.
					On("Run", mock.Anything, mock.Anything, mock.MatchedBy(matchCmd("setup.firecracker.start"))).
					Once().
					// Since the environment variables has a temp dir, we can't check for equality like normal.
					// Using Run() will let us check the arguments in a custom way.
					Run(func(args mock.Arguments) {
						actualSpec := args.Get(2).(command.Spec)
						assert.Equal(t, "setup.firecracker.start", actualSpec.Key)
						assert.Equal(
							t,
							[]string{
								"ignite",
								"run",
								"--runtime",
								"docker",
								"--network-plugin",
								"cni",
								"--cpus",
								"0",
								"--memory",
								"",
								"--size",
								"",
								"--volumes",
								"/dev/sda:/work",
								"--ssh",
								"--name",
								"test",
								"--kernel-image",
								"",
								"--kernel-args",
								"console=ttyS0 reboot=k panic=1 pci=off ip=dhcp random.trust_cpu=on i8042.noaux i8042.nomux i8042.nopnp i8042.dumbkbd",
								"--sandbox-image",
								"",
								"",
							},
							actualSpec.Command,
						)
						assert.Empty(t, actualSpec.Dir)
						require.Len(t, actualSpec.Env, 1)
						assert.True(t, strings.HasPrefix(actualSpec.Env[0], "CNI_CONF_DIR="))
						assert.Equal(t, ops.SetupFirecrackerStart, actualSpec.Operation)
					}).
					Return(nil)

				// Teardown
				cmd.
					On("Run", mock.Anything, mock.Anything, mock.MatchedBy(matchCmd("teardown.firecracker.remove"))).
					Return(nil)
			},
			expectedEntries: map[string]string{
				"cni": defaultCNIConfig,
			},
		},
		{
			name:            "Failed to start firecracker",
			workspaceDevice: "/dev/sda",
			vmName:          "test",
			mockFunc: func(t *testing.T, cmd *fakeCommand, ops *command.Operations) {
				// Start command
				cmd.
					On("Run", mock.Anything, mock.Anything, mock.MatchedBy(matchCmd("setup.firecracker.start"))).
					Once().
					Return(errors.New("failed"))

				// Teardown
				cmd.
					On("Run", mock.Anything, mock.Anything, mock.MatchedBy(matchCmd("teardown.firecracker.remove"))).
					Return(nil)
			},
			expectedErr: errors.New("failed to start firecracker vm: failed"),
		},
		{
			name:            "Docker registry mirrors",
			workspaceDevice: "/dev/sda",
			vmName:          "test",
			options: runner.FirecrackerOptions{
				DockerRegistryMirrorURLs: []string{"https://mirror1", "https://mirror2"},
			},
			mockFunc: func(t *testing.T, cmd *fakeCommand, ops *command.Operations) {
				// Start command
				cmd.
					On("Run", mock.Anything, mock.Anything, mock.MatchedBy(matchCmd("setup.firecracker.start"))).
					Once().
					// Since the environment variables has a temp dir, we can't check for equality like normal.
					// Using Run() will let us check the arguments in a custom way.
					Run(func(args mock.Arguments) {
						actualSpec := args.Get(2).(command.Spec)
						// Having docker mirrors add "--copy-files" to the command. The location to copy is the temp
						// directory. So we need to do extra work.
						for i, val := range actualSpec.Command {
							if val == "--copy-files" {
								assert.True(t, strings.HasSuffix(actualSpec.Command[i+1], "/docker-daemon.json:/etc/docker/daemon.json"))
								break
							}
						}
					}).
					Return(nil)

				// Teardown
				cmd.
					On("Run", mock.Anything, mock.Anything, mock.MatchedBy(matchCmd("teardown.firecracker.remove"))).
					Return(nil)
			},
			expectedEntries: map[string]string{
				"cni":                defaultCNIConfig,
				"docker-daemon.json": `{"registry-mirrors":["https://mirror1","https://mirror2"]}`,
			},
		},
		{
			name:            "Docker auth config",
			workspaceDevice: "/dev/sda",
			vmName:          "test",
			dockerAuthConfig: types.DockerAuthConfig{
				Auths: map[string]types.DockerAuthConfigAuth{
					"index.docker.io": {
						Auth: []byte("foobar"),
					},
				},
			},
			mockFunc: func(t *testing.T, cmd *fakeCommand, ops *command.Operations) {
				// Start command
				cmd.
					On("Run", mock.Anything, mock.Anything, mock.MatchedBy(matchCmd("setup.firecracker.start"))).
					Once().
					// Since the environment variables has a temp dir, we can't check for equality like normal.
					// Using Run() will let us check the arguments in a custom way.
					Run(func(args mock.Arguments) {
						actualSpec := args.Get(2).(command.Spec)
						// Having docker mirrors add "--copy-files" to the command. The location to copy is the temp
						// directory. So we need to do extra work.
						for i, val := range actualSpec.Command {
							if val == "--copy-files" {
								assert.True(t, strings.HasSuffix(actualSpec.Command[i+1], "/etc/docker/cli"))
								break
							}
						}
					}).
					Return(nil)

				// Teardown
				cmd.
					On("Run", mock.Anything, mock.Anything, mock.MatchedBy(matchCmd("teardown.firecracker.remove"))).
					Return(nil)
			},
			expectedEntries: map[string]string{
				"cni":        defaultCNIConfig,
				"dockerAuth": `{"auths":{"index.docker.io":{"auth":"Zm9vYmFy"}}}`,
			},
		},
		{
			name:            "Startup script",
			workspaceDevice: "/dev/sda",
			vmName:          "test",
			options: runner.FirecrackerOptions{
				VMStartupScriptPath: "/tmp/startup.sh",
			},
			mockFunc: func(t *testing.T, cmd *fakeCommand, ops *command.Operations) {
				// Start command
				cmd.
					On("Run", mock.Anything, mock.Anything, mock.MatchedBy(matchCmd("setup.firecracker.start"))).
					Once().
					Run(func(args mock.Arguments) {
						actualSpec := args.Get(2).(command.Spec)
						assert.Equal(
							t,
							[]string{
								"ignite",
								"run",
								"--runtime",
								"docker",
								"--network-plugin",
								"cni",
								"--cpus",
								"0",
								"--memory",
								"",
								"--size",
								"",
								"--copy-files",
								"/tmp/startup.sh:/tmp/startup.sh",
								"--volumes",
								"/dev/sda:/work",
								"--ssh",
								"--name",
								"test",
								"--kernel-image",
								"",
								"--kernel-args",
								"console=ttyS0 reboot=k panic=1 pci=off ip=dhcp random.trust_cpu=on i8042.noaux i8042.nomux i8042.nopnp i8042.dumbkbd",
								"--sandbox-image",
								"",
								"",
							},
							actualSpec.Command,
						)
					}).
					Return(nil)

				// startup script
				cmd.
					On("Run", mock.Anything, mock.Anything, command.Spec{
						Key: "setup.startup-script",
						Command: []string{
							"ignite",
							"exec",
							"test",
							"--",
							"/tmp/startup.sh",
						},
						Operation: ops.SetupStartupScript,
					}).
					Once().
					Return(nil)

				// Teardown
				cmd.
					On("Run", mock.Anything, mock.Anything, mock.MatchedBy(matchCmd("teardown.firecracker.remove"))).
					Return(nil)
			},
			expectedEntries: map[string]string{
				"cni": defaultCNIConfig,
			},
		},
		{
			name:            "Failed to run startup script",
			workspaceDevice: "/dev/sda",
			vmName:          "test",
			options: runner.FirecrackerOptions{
				VMStartupScriptPath: "/tmp/startup.sh",
			},
			mockFunc: func(t *testing.T, cmd *fakeCommand, ops *command.Operations) {
				cmd.
					On("Run", mock.Anything, mock.Anything, mock.MatchedBy(matchCmd("setup.firecracker.start"))).
					Once().
					Return(nil)
				cmd.
					On("Run", mock.Anything, mock.Anything, mock.MatchedBy(matchCmd("setup.startup-script"))).
					Once().
					Return(errors.New("failed"))
				cmd.
					On("Run", mock.Anything, mock.Anything, mock.MatchedBy(matchCmd("teardown.firecracker.remove"))).
					Return(nil)
			},
			expectedErr: errors.New("failed to run startup script: failed"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := new(fakeCommand)
			logger := command.NewMockLogger()
			operations := command.NewOperations(&observation.TestContext)
			firecrackerRunner := runner.NewFirecrackerRunner(
				cmd,
				logger,
				test.workspaceDevice,
				test.vmName,
				test.options,
				test.dockerAuthConfig,
				operations,
			)

			if test.mockFunc != nil {
				test.mockFunc(t, cmd, operations)
			}

			ctx := context.Background()
			err := firecrackerRunner.Setup(ctx)
			t.Cleanup(func() {
				firecrackerRunner.Teardown(ctx)
				// the Teardown is messing with the mock invocation count. Move AssertExpectationsForObjects here to
				// capture the teardown.
				mock.AssertExpectationsForObjects(t, cmd)
			})

			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				fmt.Println(firecrackerRunner.TempDir())
				entries, err := os.ReadDir(firecrackerRunner.TempDir())
				require.NoError(t, err)
				require.Len(t, entries, len(test.expectedEntries))
				for k, expectedVal := range test.expectedEntries {
					if k == "cni" {
						cniEntries, err := os.ReadDir(filepath.Join(firecrackerRunner.TempDir(), k))
						require.NoError(t, err)
						require.Len(t, cniEntries, 1)
						f, err := os.ReadFile(filepath.Join(firecrackerRunner.TempDir(), k, cniEntries[0].Name()))
						require.NoError(t, err)
						assert.JSONEq(t, expectedVal, string(f))
					} else if k == "docker-daemon.json" {
						f, err := os.ReadFile(filepath.Join(firecrackerRunner.TempDir(), k))
						require.NoError(t, err)
						require.JSONEq(t, expectedVal, string(f))
					} else if k == "dockerAuth" {
						var name string
						for _, entry := range entries {
							if strings.HasPrefix(entry.Name(), "docker_auth") {
								name = entry.Name()
								break
							}
						}
						require.NotEmpty(t, name)
						dockerAuthEntries, err := os.ReadDir(filepath.Join(firecrackerRunner.TempDir(), name))
						require.NoError(t, err)
						require.Len(t, dockerAuthEntries, 1)
						f, err := os.ReadFile(filepath.Join(firecrackerRunner.TempDir(), name, dockerAuthEntries[0].Name()))
						require.NoError(t, err)
						assert.JSONEq(t, expectedVal, string(f))
					}
				}
			}
		})
	}
}

const defaultCNIConfig = `
{
  "cniVersion": "0.4.0",
  "name": "ignite-cni-bridge",
  "plugins": [
    {
  	  "type": "bridge",
  	  "bridge": "ignite0",
  	  "isGateway": true,
  	  "isDefaultGateway": true,
  	  "promiscMode": false,
  	  "ipMasq": true,
  	  "ipam": {
  	    "type": "host-local",
  	    "subnet": "10.61.0.0/16"
  	  }
    },
    {
  	  "type": "portmap",
  	  "capabilities": {
  	    "portMappings": true
  	  }
    },
    {
  	  "type": "firewall"
    },
    {
  	  "type": "isolation"
    },
    {
  	  "name": "slowdown",
  	  "type": "bandwidth",
  	  "ingressRate": 0,
  	  "ingressBurst": 0,
  	  "egressRate": 0,
  	  "egressBurst": 0
    }
  ]
}
`

func TestFirecrackerRunner_Teardown(t *testing.T) {
	cmd := new(fakeCommand)
	logger := command.NewMockLogger()
	operations := command.NewOperations(&observation.TestContext)
	firecrackerRunner := runner.NewFirecrackerRunner(cmd, logger, "/dev", "test", runner.FirecrackerOptions{}, types.DockerAuthConfig{}, operations)

	cmd.
		On("Run", mock.Anything, mock.Anything, mock.MatchedBy(matchCmd("setup.firecracker.start"))).
		Once().
		Return(nil)

	ctx := context.Background()
	err := firecrackerRunner.Setup(ctx)
	require.NoError(t, err)

	dir := firecrackerRunner.TempDir()

	_, err = os.Stat(dir)
	require.NoError(t, err)

	cmd.
		On("Run", mock.Anything, mock.Anything, command.Spec{
			Key:       "teardown.firecracker.remove",
			Command:   []string{"ignite", "rm", "-f", "test"},
			Operation: operations.TeardownFirecrackerRemove,
		}).
		Once().
		Return(nil)

	err = firecrackerRunner.Teardown(ctx)
	require.NoError(t, err)

	_, err = os.Stat(dir)
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func matchCmd(key string) func(spec command.Spec) bool {
	return func(spec command.Spec) bool {
		return spec.Key == key
	}
}

func TestFirecrackerRunner_Run(t *testing.T) {
	cmd := new(fakeCommand)
	logger := command.NewMockLogger()
	operations := command.NewOperations(&observation.TestContext)
	options := runner.FirecrackerOptions{
		DockerOptions: command.DockerOptions{
			ConfigPath:     "/docker/config",
			AddHostGateway: true,
			Resources: command.ResourceOptions{
				NumCPUs:   10,
				Memory:    "1G",
				DiskSpace: "10G",
			},
		},
	}
	spec := runner.Spec{
		CommandSpec: command.Spec{
			Key:     "some-key",
			Command: []string{"echo", "hello"},
			Dir:     "/workingdir",
			Env:     []string{"FOO=bar"},
		},
		Image:      "alpine",
		ScriptPath: "/some/script",
	}

	firecrackerRunner := runner.NewFirecrackerRunner(cmd, logger, "/dev", "test", options, types.DockerAuthConfig{}, operations)

	expectedCommandSpec := command.Spec{
		Key: "some-key",
		Command: []string{
			"ignite",
			"exec",
			"test",
			"--",
			"sh",
			"-c",
			"docker --config /docker/config run --rm --add-host=host.docker.internal:host-gateway --cpus 10 --memory 1G -v /work:/data -w /data/workingdir -e FOO=bar --entrypoint /bin/sh alpine /data/.sourcegraph-executor/some/script",
		},
	}
	cmd.
		On("Run", mock.Anything, logger, expectedCommandSpec).
		Return(nil)

	err := firecrackerRunner.Run(context.Background(), spec)
	require.NoError(t, err)
}
