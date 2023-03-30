package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestConfig_Load(t *testing.T) {
	cfg := &config.Config{}
	cfg.SetMockGetter(func(name, defaultValue, description string) string {
		switch name {
		case "EXECUTOR_QUEUE_POLL_INTERVAL":
			return "10s"
		case "EXECUTOR_MAXIMUM_NUM_JOBS":
			return "10"
		case "EXECUTOR_USE_FIRECRACKER":
			return "true"
		case "EXECUTOR_KEEP_WORKSPACES":
			return "true"
		case "EXECUTOR_JOB_NUM_CPUS":
			return "8"
		case "EXECUTOR_FIRECRACKER_BANDWIDTH_INGRESS":
			return "100"
		case "EXECUTOR_FIRECRACKER_BANDWIDTH_EGRESS":
			return "100"
		case "EXECUTOR_MAXIMUM_RUNTIME_PER_JOB":
			return "1m"
		case "EXECUTOR_CLEANUP_TASK_INTERVAL":
			return "10m"
		case "EXECUTOR_NUM_TOTAL_JOBS":
			return "10"
		case "EXECUTOR_MAX_ACTIVE_TIME":
			return "1h"
		case "EXECUTOR_KUBERNETES_CONFIG_PATH":
			return "/foo/bar"
		case "EXECUTOR_KUBERNETES_NODE_NAME":
			return "my-node"
		case "EXECUTOR_KUBERNETES_NODE_SELECTOR":
			return "app=my-app,zone=west"
		case "EXECUTOR_KUBERNETES_NODE_REQUIRED_AFFINITY_MATCH_EXPRESSIONS":
			return `[{"key": "foo", "operator": "In", "values": ["bar"]}]`
		case "EXECUTOR_KUBERNETES_NODE_REQUIRED_AFFINITY_MATCH_FIELDS":
			return `[{"key": "faz", "operator": "In", "values": ["baz"]}]`
		default:
			return name
		}
	})
	cfg.Load()

	assert.Equal(t, "EXECUTOR_FRONTEND_URL", cfg.FrontendURL)
	assert.Equal(t, "EXECUTOR_FRONTEND_PASSWORD", cfg.FrontendAuthorizationToken)
	assert.Equal(t, "EXECUTOR_QUEUE_NAME", cfg.QueueName)
	assert.Equal(t, 10*time.Second, cfg.QueuePollInterval)
	assert.Equal(t, 10, cfg.MaximumNumJobs)
	assert.True(t, cfg.UseFirecracker)
	assert.Equal(t, "EXECUTOR_FIRECRACKER_IMAGE", cfg.FirecrackerImage)
	assert.Equal(t, "EXECUTOR_FIRECRACKER_KERNEL_IMAGE", cfg.FirecrackerKernelImage)
	assert.Equal(t, "EXECUTOR_FIRECRACKER_SANDBOX_IMAGE", cfg.FirecrackerSandboxImage)
	assert.Equal(t, "EXECUTOR_VM_STARTUP_SCRIPT_PATH", cfg.VMStartupScriptPath)
	assert.Equal(t, "EXECUTOR_VM_PREFIX", cfg.VMPrefix)
	assert.True(t, cfg.KeepWorkspaces)
	assert.Equal(t, "EXECUTOR_DOCKER_HOST_MOUNT_PATH", cfg.DockerHostMountPath)
	assert.Equal(t, 8, cfg.JobNumCPUs)
	assert.Equal(t, "EXECUTOR_JOB_MEMORY", cfg.JobMemory)
	assert.Equal(t, "EXECUTOR_FIRECRACKER_DISK_SPACE", cfg.FirecrackerDiskSpace)
	assert.Equal(t, 100, cfg.FirecrackerBandwidthIngress)
	assert.Equal(t, 100, cfg.FirecrackerBandwidthEgress)
	assert.Equal(t, 1*time.Minute, cfg.MaximumRuntimePerJob)
	assert.Equal(t, 10*time.Minute, cfg.CleanupTaskInterval)
	assert.Equal(t, 10, cfg.NumTotalJobs)
	assert.Equal(t, "NODE_EXPORTER_URL", cfg.NodeExporterURL)
	assert.Equal(t, "DOCKER_REGISTRY_NODE_EXPORTER_URL", cfg.DockerRegistryNodeExporterURL)
	assert.Equal(t, time.Hour, cfg.MaxActiveTime)
	assert.Equal(t, "EXECUTOR_DOCKER_REGISTRY_MIRROR_URL", cfg.DockerRegistryMirrorURL)
	assert.Equal(t, "/foo/bar", cfg.KubernetesConfigPath)
	assert.Equal(t, "my-node", cfg.KubernetesNodeName)
	assert.Equal(t, "app=my-app,zone=west", cfg.KubernetesNodeSelector)
	assert.Equal(
		t,
		[]corev1.NodeSelectorRequirement{{Key: "foo", Operator: corev1.NodeSelectorOpIn, Values: []string{"bar"}}},
		cfg.KubernetesNodeRequiredAffinityMatchExpressions,
	)
	assert.Equal(
		t,
		[]corev1.NodeSelectorRequirement{{Key: "faz", Operator: corev1.NodeSelectorOpIn, Values: []string{"baz"}}},
		cfg.KubernetesNodeRequiredAffinityMatchFields,
	)
}

func TestConfig_Load_Defaults(t *testing.T) {
	cfg := &config.Config{}
	cfg.Load()

	assert.Empty(t, cfg.FrontendURL)
	assert.Empty(t, cfg.FrontendAuthorizationToken)
	assert.Empty(t, cfg.QueueName)
	assert.Equal(t, time.Second, cfg.QueuePollInterval)
	assert.Equal(t, 1, cfg.MaximumNumJobs)
	assert.False(t, cfg.UseFirecracker)
	assert.Equal(t, "sourcegraph/executor-vm:insiders", cfg.FirecrackerImage)
	assert.Equal(t, "sourcegraph/ignite-kernel:5.10.135-amd64", cfg.FirecrackerKernelImage)
	assert.Equal(t, "sourcegraph/ignite:v0.10.5", cfg.FirecrackerSandboxImage)
	assert.Empty(t, cfg.VMStartupScriptPath)
	assert.Equal(t, "executor", cfg.VMPrefix)
	assert.False(t, cfg.KeepWorkspaces)
	assert.Empty(t, cfg.DockerHostMountPath)
	assert.Equal(t, 4, cfg.JobNumCPUs)
	assert.Equal(t, "12G", cfg.JobMemory)
	assert.Equal(t, "20G", cfg.FirecrackerDiskSpace)
	assert.Equal(t, 524288000, cfg.FirecrackerBandwidthIngress)
	assert.Equal(t, 524288000, cfg.FirecrackerBandwidthEgress)
	assert.Equal(t, 30*time.Minute, cfg.MaximumRuntimePerJob)
	assert.Equal(t, 1*time.Minute, cfg.CleanupTaskInterval)
	assert.Zero(t, cfg.NumTotalJobs)
	assert.Empty(t, cfg.NodeExporterURL)
	assert.Empty(t, cfg.DockerRegistryNodeExporterURL)
	assert.Zero(t, cfg.MaxActiveTime)
	assert.Empty(t, cfg.DockerRegistryMirrorURL)
	assert.Empty(t, cfg.KubernetesConfigPath)
	assert.Empty(t, cfg.KubernetesNodeName)
	assert.Empty(t, cfg.KubernetesNodeSelector)
	assert.Nil(t, cfg.KubernetesNodeRequiredAffinityMatchExpressions)
	assert.Nil(t, cfg.KubernetesNodeRequiredAffinityMatchFields)
	assert.Equal(t, "default", cfg.KubernetesNamespace)
	assert.Equal(t, "sg-executor-pvc", cfg.KubernetesPersistenceVolumeName)
	assert.Empty(t, cfg.KubernetesResourceLimitCPU)
	assert.Equal(t, "12Gi", cfg.KubernetesResourceLimitMemory)
	assert.Empty(t, cfg.KubernetesResourceRequestCPU)
	assert.Equal(t, "12Gi", cfg.KubernetesResourceRequestMemory)
	assert.Equal(t, 600, cfg.KubernetesJobRetryBackoffLimit)
	assert.Equal(t, 100*time.Millisecond, cfg.KubernetesJobRetryBackoffDuration)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		getterFunc  env.GetterFunc
		expectedErr error
	}{
		{
			name: "Valid config",
			getterFunc: func(name string, defaultValue, description string) string {
				switch name {
				case "EXECUTOR_QUEUE_NAME":
					return "batches"
				case "EXECUTOR_FRONTEND_URL":
					return "http://some-url.com"
				case "EXECUTOR_FRONTEND_PASSWORD":
					return "some-password"
				default:
					return defaultValue
				}
			},
		},
		{
			name:        "Default config",
			expectedErr: errors.New("3 errors occurred:\n\t* invalid value \"\" for EXECUTOR_FRONTEND_URL: no value supplied\n\t* invalid value \"\" for EXECUTOR_FRONTEND_PASSWORD: no value supplied\n\t* invalid value \"\" for EXECUTOR_QUEUE_NAME: no value supplied"),
		},
		{
			name: "Invalid EXECUTOR_DOCKER_AUTH_CONFIG",
			getterFunc: func(name string, defaultValue, description string) string {
				switch name {
				case "EXECUTOR_QUEUE_NAME":
					return "batches"
				case "EXECUTOR_FRONTEND_URL":
					return "http://some-url.com"
				case "EXECUTOR_FRONTEND_PASSWORD":
					return "some-password"
				case "EXECUTOR_DOCKER_AUTH_CONFIG":
					return `{"foo": "bar"`
				default:
					return defaultValue
				}
			},
			expectedErr: errors.New("invalid EXECUTOR_DOCKER_AUTH_CONFIG, failed to parse: unexpected end of JSON input"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.SetMockGetter(test.getterFunc)
			cfg.Load()

			err := cfg.Validate()
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
package config

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestValidateConfig(t *testing.T) {
	t.Run("Frontend URL", func(t *testing.T) {
		tests := []struct {
			name        string
			frontendURL string
			expectedErr error
		}{
			{
				name:        "Valid URL",
				frontendURL: "https://sourcegraph.example.com",
				expectedErr: nil,
			},
			{
				name:        "Missing scheme",
				frontendURL: "sourcegraph.example.com",
				expectedErr: errors.New("EXECUTOR_FRONTEND_URL must be in the format scheme://host (and optionally :port)"),
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				conf := Config{
					FrontendURL:    test.frontendURL,
					QueueName:      "batches",
					UseFirecracker: false,
				}

				err := conf.Validate()
				if !errors.Is(err, test.expectedErr) {
					t.Errorf("Unexpected error returned: expected '%v', got '%v'", test.expectedErr, err)
				}
			})
		}
	})
}
