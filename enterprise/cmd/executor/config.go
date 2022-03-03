package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	apiworker "github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Config struct {
	env.BaseConfig

	FrontendURL                string
	FrontendAuthorizationToken string
	QueueName                  string
	QueuePollInterval          time.Duration
	MaximumNumJobs             int
	FirecrackerImage           string
	VMStartupScriptPath        string
	VMPrefix                   string
	UseFirecracker             bool
	FirecrackerNumCPUs         int
	FirecrackerMemory          string
	FirecrackerDiskSpace       string
	MaximumRuntimePerJob       time.Duration
	CleanupTaskInterval        time.Duration
	NumTotalJobs               int
	MaxActiveTime              time.Duration
	WorkerHostname             string
}

func (c *Config) Load() {
	c.FrontendURL = c.Get("EXECUTOR_FRONTEND_URL", "", "The external URL of the sourcegraph instance.")
	c.FrontendAuthorizationToken = c.Get("EXECUTOR_FRONTEND_PASSWORD", "", "The authorization token supplied to the frontend.")
	c.QueueName = c.Get("EXECUTOR_QUEUE_NAME", "", "The name of the queue to listen to.")
	c.QueuePollInterval = c.GetInterval("EXECUTOR_QUEUE_POLL_INTERVAL", "1s", "Interval between dequeue requests.")
	c.MaximumNumJobs = c.GetInt("EXECUTOR_MAXIMUM_NUM_JOBS", "1", "Number of virtual machines or containers that can be running at once.")
	c.UseFirecracker = c.GetBool("EXECUTOR_USE_FIRECRACKER", "true", "Whether to isolate commands in virtual machines.")
	c.FirecrackerImage = c.Get("EXECUTOR_FIRECRACKER_IMAGE", "sourcegraph/ignite-ubuntu:insiders", "The base image to use for virtual machines.")
	c.VMStartupScriptPath = c.GetOptional("EXECUTOR_VM_STARTUP_SCRIPT_PATH", "A path to a file on the host that is loaded into a fresh virtual machine and executed on startup.")
	c.VMPrefix = c.Get("EXECUTOR_VM_PREFIX", "executor", "A name prefix for virtual machines controlled by this instance.")
	c.FirecrackerNumCPUs = c.GetInt("EXECUTOR_FIRECRACKER_NUM_CPUS", "4", "How many CPUs to allocate to each virtual machine or container.")
	c.FirecrackerMemory = c.Get("EXECUTOR_FIRECRACKER_MEMORY", "12G", "How much memory to allocate to each virtual machine or container.")
	c.FirecrackerDiskSpace = c.Get("EXECUTOR_FIRECRACKER_DISK_SPACE", "20G", "How much disk space to allocate to each virtual machine or container.")
	c.MaximumRuntimePerJob = c.GetInterval("EXECUTOR_MAXIMUM_RUNTIME_PER_JOB", "30m", "The maximum wall time that can be spent on a single job.")
	c.CleanupTaskInterval = c.GetInterval("EXECUTOR_CLEANUP_TASK_INTERVAL", "1m", "The frequency with which to run periodic cleanup tasks.")
	c.NumTotalJobs = c.GetInt("EXECUTOR_NUM_TOTAL_JOBS", "0", "The maximum number of jobs that will be dequeued by the worker.")
	c.MaxActiveTime = c.GetInterval("EXECUTOR_MAX_ACTIVE_TIME", "0", "The maximum time that can be spent by the worker dequeueing records to be handled.")

	hn := hostname.Get()
	// Be unique but also descriptive.
	c.WorkerHostname = hn + "-" + uuid.New().String()
}

func (c *Config) Validate() error {
	if c.FirecrackerNumCPUs != 1 && c.FirecrackerNumCPUs%2 != 0 {
		// Required by Firecracker: The vCPU number is invalid! The vCPU number can only be 1 or an even number when hyperthreading is enabled
		c.AddError(errors.Newf("EXECUTOR_FIRECRACKER_NUM_CPUS must be 1 or an even number"))
	}

	return c.BaseConfig.Validate()
}

func (c *Config) APIWorkerOptions(telemetryOptions apiclient.TelemetryOptions) apiworker.Options {
	return apiworker.Options{
		VMPrefix:           c.VMPrefix,
		QueueName:          c.QueueName,
		WorkerOptions:      c.WorkerOptions(),
		FirecrackerOptions: c.FirecrackerOptions(),
		ResourceOptions:    c.ResourceOptions(),
		GitServicePath:     "/.executors/git",
		ClientOptions:      c.ClientOptions(telemetryOptions),
		RedactedValues: map[string]string{
			// 🚨 SECURITY: Catch uses of the shared frontend token used to clone
			// git repositories that make it into commands or stdout/stderr streams.
			c.FrontendAuthorizationToken: "SECRET_REMOVED",
		},
	}
}

func (c *Config) WorkerOptions() workerutil.WorkerOptions {
	return workerutil.WorkerOptions{
		Name:                 fmt.Sprintf("executor_%s_worker", c.QueueName),
		NumHandlers:          c.MaximumNumJobs,
		Interval:             c.QueuePollInterval,
		HeartbeatInterval:    5 * time.Second,
		Metrics:              makeWorkerMetrics(c.QueueName),
		NumTotalJobs:         c.NumTotalJobs,
		MaxActiveTime:        c.MaxActiveTime,
		WorkerHostname:       c.WorkerHostname,
		MaximumRuntimePerJob: c.MaximumRuntimePerJob,
	}
}

func (c *Config) FirecrackerOptions() command.FirecrackerOptions {
	return command.FirecrackerOptions{
		Enabled:             c.UseFirecracker,
		Image:               c.FirecrackerImage,
		VMStartupScriptPath: c.VMStartupScriptPath,
	}
}

func (c *Config) ResourceOptions() command.ResourceOptions {
	return command.ResourceOptions{
		NumCPUs:   c.FirecrackerNumCPUs,
		Memory:    c.FirecrackerMemory,
		DiskSpace: c.FirecrackerDiskSpace,
	}
}

func (c *Config) ClientOptions(telemetryOptions apiclient.TelemetryOptions) apiclient.Options {
	return apiclient.Options{
		ExecutorName:      c.WorkerHostname,
		PathPrefix:        "/.executors/queue",
		EndpointOptions:   c.EndpointOptions(),
		BaseClientOptions: c.BaseClientOptions(),
		TelemetryOptions:  telemetryOptions,
	}
}

func (c *Config) BaseClientOptions() apiclient.BaseClientOptions {
	return apiclient.BaseClientOptions{}
}

func (c *Config) EndpointOptions() apiclient.EndpointOptions {
	return apiclient.EndpointOptions{
		URL:   c.FrontendURL,
		Token: c.FrontendAuthorizationToken,
	}
}
