package main

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient/queue"
)

func TestConfig_Load_Envs(t *testing.T) {
	var envs = map[string]string{
		"EXECUTOR_FRONTEND_URL":             "http://foo-bar",
		"EXECUTOR_FRONTEND_PASSWORD":        "my-secret-pass",
		"EXECUTOR_QUEUE_NAME":               "test-executor",
		"EXECUTOR_QUEUE_POLL_INTERVAL":      "10s",
		"EXECUTOR_MAXIMUM_NUM_JOBS":         "10",
		"EXECUTOR_USE_FIRECRACKER":          "false",
		"EXECUTOR_FIRECRACKER_IMAGE":        "test-image",
		"EXECUTOR_FIRECRACKER_KERNEL_IMAGE": "test-kernel-image",
		"EXECUTOR_VM_STARTUP_SCRIPT_PATH":   "foo.sh",
		"EXECUTOR_VM_PREFIX":                "test",
		"EXECUTOR_KEEP_WORKSPACES":          "true",
		"EXECUTOR_DOCKER_HOST_MOUNT_PATH":   "/workdir",
		"EXECUTOR_JOB_NUM_CPUS":             "8",
		"EXECUTOR_JOB_MEMORY":               "24G",
		"EXECUTOR_FIRECRACKER_DISK_SPACE":   "40G",
		"EXECUTOR_MAXIMUM_RUNTIME_PER_JOB":  "60m",
		"EXECUTOR_CLEANUP_TASK_INTERVAL":    "10m",
		"EXECUTOR_NUM_TOTAL_JOBS":           "1",
		"NODE_EXPORTER_URL":                 "http://faz-baz",
		"DOCKER_REGISTRY_NODE_EXPORTER_URL": "http://raz-daz",
		"EXECUTOR_MAX_ACTIVE_TIME":          "1m",
	}

	config := &Config{}
	config.SetMockGetter(mapGetter(envs))
	config.Load()

	assert.Equal(t, "http://foo-bar", config.FrontendURL)
	assert.Equal(t, "my-secret-pass", config.FrontendAuthorizationToken)
	assert.Equal(t, "test-executor", config.QueueName)
	assert.Equal(t, 10*time.Second, config.QueuePollInterval)
	assert.Equal(t, 10, config.MaximumNumJobs)
	assert.False(t, config.UseFirecracker)
	assert.Equal(t, "test-image", config.FirecrackerImage)
	assert.Equal(t, "test-kernel-image", config.FirecrackerKernelImage)
	assert.Equal(t, "foo.sh", config.VMStartupScriptPath)
	assert.Equal(t, "test", config.VMPrefix)
	assert.True(t, config.KeepWorkspaces)
	assert.Equal(t, "/workdir", config.DockerHostMountPath)
	assert.Equal(t, 8, config.JobNumCPUs)
	assert.Equal(t, "24G", config.JobMemory)
	assert.Equal(t, "40G", config.FirecrackerDiskSpace)
	assert.Equal(t, 60*time.Minute, config.MaximumRuntimePerJob)
	assert.Equal(t, 10*time.Minute, config.CleanupTaskInterval)
	assert.Equal(t, 1, config.NumTotalJobs)
	assert.Equal(t, "http://faz-baz", config.NodeExporterURL)
	assert.Equal(t, "http://raz-daz", config.DockerRegistryNodeExporterURL)
	assert.Equal(t, 1*time.Minute, config.MaxActiveTime)
	// Since worker hostname is random each time, just make sure it matches a pattern.
	assert.Regexp(t, ".*-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", config.WorkerHostname)
}

func TestConfig_Load_Defaults(t *testing.T) {
	config := &Config{}
	config.Load()

	assert.Empty(t, config.FrontendURL)
	assert.Empty(t, config.FrontendAuthorizationToken)
	assert.Empty(t, config.QueueName)
	assert.Equal(t, 1*time.Second, config.QueuePollInterval)
	assert.Equal(t, 1, config.MaximumNumJobs)
	assert.True(t, config.UseFirecracker)
	assert.Equal(t, "sourcegraph/executor-vm:latest", config.FirecrackerImage)
	assert.Equal(t, "sourcegraph/executor-vm:latest", config.FirecrackerImage)
	assert.Equal(t, "sourcegraph/ignite-kernel:5.10.135-amd64", config.FirecrackerKernelImage)
	assert.Empty(t, config.VMStartupScriptPath)
	assert.Equal(t, "executor", config.VMPrefix)
	assert.False(t, config.KeepWorkspaces)
	assert.Empty(t, config.DockerHostMountPath)
	assert.Equal(t, 4, config.JobNumCPUs)
	assert.Equal(t, "12G", config.JobMemory)
	assert.Equal(t, "20G", config.FirecrackerDiskSpace)
	assert.Equal(t, 30*time.Minute, config.MaximumRuntimePerJob)
	assert.Equal(t, 1*time.Minute, config.CleanupTaskInterval)
	assert.Zero(t, config.NumTotalJobs)
	assert.Empty(t, config.NodeExporterURL)
	assert.Empty(t, config.DockerRegistryNodeExporterURL)
	assert.Zero(t, config.MaxActiveTime)
	// Since worker hostname is random each time, just make sure it matches a pattern.
	assert.Regexp(t, ".*-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", config.WorkerHostname)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		envs        map[string]string
		expectedErr error
	}{
		{
			name: "Valid batches config",
			envs: map[string]string{
				"EXECUTOR_FRONTEND_URL":      "http://foo-bar",
				"EXECUTOR_FRONTEND_PASSWORD": "my-secret-pass",
				"EXECUTOR_QUEUE_NAME":        "batches",
			},
		},
		{
			name: "Valid codeintel config",
			envs: map[string]string{
				"EXECUTOR_FRONTEND_URL":      "http://foo-bar",
				"EXECUTOR_FRONTEND_PASSWORD": "my-secret-pass",
				"EXECUTOR_QUEUE_NAME":        "codeintel",
			},
		},
		{
			name: "Invalid queue name",
			envs: map[string]string{
				"EXECUTOR_FRONTEND_URL":      "http://foo-bar",
				"EXECUTOR_FRONTEND_PASSWORD": "my-secret-pass",
				"EXECUTOR_QUEUE_NAME":        "my-queue",
			},
			expectedErr: errors.New("EXECUTOR_QUEUE_NAME must be set to 'batches' or 'codeintel'"),
		},
		{
			name: "Invalid number of CPUs",
			envs: map[string]string{
				"EXECUTOR_FRONTEND_URL":      "http://foo-bar",
				"EXECUTOR_FRONTEND_PASSWORD": "my-secret-pass",
				"EXECUTOR_QUEUE_NAME":        "batches",
				"EXECUTOR_JOB_NUM_CPUS":      "3",
			},
			expectedErr: errors.New("EXECUTOR_JOB_NUM_CPUS must be 1 or an even number"),
		},
		{
			name: "Missing frontend URL",
			envs: map[string]string{
				"EXECUTOR_FRONTEND_PASSWORD": "my-secret-pass",
				"EXECUTOR_QUEUE_NAME":        "batches",
			},
			expectedErr: errors.New("invalid value \"\" for EXECUTOR_FRONTEND_URL: no value supplied"),
		},
		{
			name: "Missing frontend password",
			envs: map[string]string{
				"EXECUTOR_FRONTEND_URL": "http://foo-bar",
				"EXECUTOR_QUEUE_NAME":   "batches",
			},
			expectedErr: errors.New("invalid value \"\" for EXECUTOR_FRONTEND_PASSWORD: no value supplied"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := &Config{}
			config.SetMockGetter(mapGetter(test.envs))
			config.Load()

			err := config.Validate()
			if test.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

var baseEnvs = map[string]string{
	"EXECUTOR_FRONTEND_URL":      "http://foo-bar",
	"EXECUTOR_FRONTEND_PASSWORD": "my-secret-pass",
	"EXECUTOR_QUEUE_NAME":        "batches",
}

func TestConfig_APIWorkerOptions(t *testing.T) {
	config := &Config{}
	config.SetMockGetter(mapGetter(baseEnvs))
	config.Load()

	options := config.APIWorkerOptions(queue.TelemetryOptions{})

	// base options
	assert.Equal(t, "executor", options.VMPrefix)
	assert.False(t, options.KeepWorkspaces)
	assert.Equal(t, "batches", options.QueueName)
	assert.NotNil(t, options.ResourceOptions)
	assert.Equal(t, "/.executors/git", options.GitServicePath)
	assert.NotNil(t, options.QueueOptions)
	assert.NotNil(t, options.FilesOptions)
	assert.Equal(t, map[string]string{"my-secret-pass": "SECRET_REMOVED"}, options.RedactedValues)
	assert.Empty(t, options.NodeExporterEndpoint)
	assert.Empty(t, options.DockerRegistryNodeExporterEndpoint)

	// worker options
	assert.Equal(t, "executor_batches_worker", options.WorkerOptions.Name)
	assert.Equal(t, 1, options.WorkerOptions.NumHandlers)
	assert.Equal(t, 1*time.Second, options.WorkerOptions.Interval)
	assert.Equal(t, 5*time.Second, options.WorkerOptions.HeartbeatInterval)
	assert.NotNil(t, options.WorkerOptions.Metrics)
	assert.Zero(t, options.WorkerOptions.NumTotalJobs)
	assert.Zero(t, options.WorkerOptions.MaxActiveTime)
	assert.Regexp(t, ".*-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", options.WorkerOptions.WorkerHostname)
	assert.Equal(t, 30*time.Minute, options.WorkerOptions.MaximumRuntimePerJob)

	// firecracker options
	assert.True(t, options.FirecrackerOptions.Enabled)
	assert.Equal(t, "sourcegraph/executor-vm:latest", options.FirecrackerOptions.Image)
	assert.Equal(t, "sourcegraph/ignite-kernel:5.10.135-amd64", options.FirecrackerOptions.KernelImage)
	assert.Empty(t, options.FirecrackerOptions.VMStartupScriptPath)

	// resource options
	assert.Equal(t, 4, options.ResourceOptions.NumCPUs)
	assert.Equal(t, "12G", options.ResourceOptions.Memory)
	assert.Equal(t, "20G", options.ResourceOptions.DiskSpace)
	assert.Empty(t, options.ResourceOptions.DockerHostMountPath)

	// queue options
	assert.Regexp(t, ".*-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", options.QueueOptions.ExecutorName)
	assert.Equal(t, "http://foo-bar", options.QueueOptions.BaseClientOptions.EndpointOptions.URL)
	assert.Equal(t, "/.executors/queue", options.QueueOptions.BaseClientOptions.EndpointOptions.PathPrefix)
	assert.Equal(t, "my-secret-pass", options.QueueOptions.BaseClientOptions.EndpointOptions.Token)
	assert.NotNil(t, options.QueueOptions.TelemetryOptions)

	// files options
	assert.Equal(t, "http://foo-bar", options.FilesOptions.EndpointOptions.URL)
	assert.Equal(t, "/.executors/files", options.FilesOptions.EndpointOptions.PathPrefix)
	assert.Equal(t, "my-secret-pass", options.FilesOptions.EndpointOptions.Token)
}

func mapGetter(env map[string]string) func(name, defaultValue, description string) string {
	return func(name, defaultValue, description string) string {
		if v, ok := env[name]; ok {
			return v
		}

		return defaultValue
	}
}
