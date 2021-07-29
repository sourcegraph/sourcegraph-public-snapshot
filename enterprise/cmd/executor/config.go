package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	apiworker "github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type Config struct {
	env.BaseConfig

	FrontendURL          string
	FrontendUsername     string
	FrontendPassword     string
	QueueName            string
	QueuePollInterval    time.Duration
	MaximumNumJobs       int
	FirecrackerImage     string
	VMPrefix             string
	UseFirecracker       bool
	FirecrackerNumCPUs   int
	FirecrackerMemory    string
	FirecrackerDiskSpace string
	ImageArchivesPath    string
	DisableHealthServer  bool
	HealthServerPort     int
	MaximumRuntimePerJob time.Duration
	CleanupTaskInterval  time.Duration
}

func (c *Config) Load() {
	c.FrontendURL = c.Get("EXECUTOR_FRONTEND_URL", "", "The external URL of the sourcegraph instance.")
	c.FrontendUsername = c.Get("EXECUTOR_FRONTEND_USERNAME", "", "The username supplied to the frontend.")
	c.FrontendPassword = c.Get("EXECUTOR_FRONTEND_PASSWORD", "", "The password supplied to the frontend.")
	c.QueueName = c.Get("EXECUTOR_QUEUE_NAME", "", "The name of the queue to listen to.")
	c.QueuePollInterval = c.GetInterval("EXECUTOR_QUEUE_POLL_INTERVAL", "1s", "Interval between dequeue requests.")
	c.MaximumNumJobs = c.GetInt("EXECUTOR_MAXIMUM_NUM_JOBS", "1", "Number of virtual machines or containers that can be running at once.")
	c.UseFirecracker = c.GetBool("EXECUTOR_USE_FIRECRACKER", "true", "Whether to isolate commands in virtual machines.")
	c.FirecrackerImage = c.Get("EXECUTOR_FIRECRACKER_IMAGE", "sourcegraph/ignite-ubuntu:insiders", "The base image to use for virtual machines.")
	c.VMPrefix = c.Get("EXECUTOR_VM_PREFIX", "executor", "A name prefix for virtual machines controlled by this instance.")
	c.FirecrackerNumCPUs = c.GetInt("EXECUTOR_FIRECRACKER_NUM_CPUS", "4", "How many CPUs to allocate to each virtual machine or container.")
	c.FirecrackerMemory = c.Get("EXECUTOR_FIRECRACKER_MEMORY", "12G", "How much memory to allocate to each virtual machine or container.")
	c.FirecrackerDiskSpace = c.Get("EXECUTOR_FIRECRACKER_DISK_SPACE", "20G", "How much disk space to allocate to each virtual machine or container.")
	c.ImageArchivesPath = c.Get("EXECUTOR_IMAGE_ARCHIVE_PATH", "", "Where to store tar archives of docker images shared by virtual machines.")
	c.DisableHealthServer = c.GetBool("EXECUTOR_DISABLE_HEALTHSERVER", "false", "Whether or not to disable the health server.")
	c.HealthServerPort = c.GetInt("EXECUTOR_HEALTH_SERVER_PORT", "3192", "The port to listen on for the health server.")
	c.MaximumRuntimePerJob = c.GetInterval("EXECUTOR_MAXIMUM_RUNTIME_PER_JOB", "30m", "The maximum wall time that can be spent on a single job.")
	c.CleanupTaskInterval = c.GetInterval("EXECUTOR_CLEANUP_TASK_INTERVAL", "1m", "The frequency with which to run periodic cleanup tasks.")
}

func (c *Config) Validate() error {
	if c.FirecrackerNumCPUs != 1 && c.FirecrackerNumCPUs%2 != 0 {
		// Required by Firecracker: The vCPU number is invalid! The vCPU number can only be 1 or an even number when hyperthreading is enabled
		c.AddError(fmt.Errorf("EXECUTOR_FIRECRACKER_NUM_CPUS must be 1 or an even number"))
	}

	return c.BaseConfig.Validate()
}

func (c *Config) APIWorkerOptions(transport http.RoundTripper) apiworker.Options {
	return apiworker.Options{
		VMPrefix:             c.VMPrefix,
		QueueName:            c.QueueName,
		WorkerOptions:        c.WorkerOptions(),
		FirecrackerOptions:   c.FirecrackerOptions(),
		ResourceOptions:      c.ResourceOptions(),
		MaximumRuntimePerJob: c.MaximumRuntimePerJob,
		GitServicePath:       "/.executors/git",
		ClientOptions:        c.ClientOptions(transport),
		RedactedValues: map[string]string{
			// 🚨 SECURITY: Catch uses of the shared frontend token used to clone
			// git repositories that make it into commands or stdout/stderr streams.
			c.FrontendPassword: "PASSWORD_REMOVED",
		},
	}
}

func (c *Config) WorkerOptions() workerutil.WorkerOptions {
	return workerutil.WorkerOptions{
		Name:              fmt.Sprintf("executor_%s_worker", c.QueueName),
		NumHandlers:       c.MaximumNumJobs,
		Interval:          c.QueuePollInterval,
		HeartbeatInterval: 1 * time.Second,
		Metrics:           makeWorkerMetrics(c.QueueName),
	}
}

func (c *Config) FirecrackerOptions() command.FirecrackerOptions {
	return command.FirecrackerOptions{
		Enabled:           c.UseFirecracker,
		Image:             c.FirecrackerImage,
		ImageArchivesPath: c.ImageArchivesPath,
	}
}

func (c *Config) ResourceOptions() command.ResourceOptions {
	return command.ResourceOptions{
		NumCPUs:   c.FirecrackerNumCPUs,
		Memory:    c.FirecrackerMemory,
		DiskSpace: c.FirecrackerDiskSpace,
	}
}

func (c *Config) ClientOptions(transport http.RoundTripper) apiclient.Options {
	hn := hostname.Get()

	return apiclient.Options{
		// Be unique but also descriptive.
		ExecutorName:      hn + "-" + uuid.New().String(),
		ExecutorHostname:  hn,
		PathPrefix:        "/.executors/queue",
		EndpointOptions:   c.EndpointOptions(),
		BaseClientOptions: c.BaseClientOptions(transport),
	}
}

func (c *Config) BaseClientOptions(transport http.RoundTripper) apiclient.BaseClientOptions {
	return apiclient.BaseClientOptions{
		Transport: transport,
	}
}

func (c *Config) EndpointOptions() apiclient.EndpointOptions {
	return apiclient.EndpointOptions{
		URL:      c.FrontendURL,
		Username: c.FrontendUsername,
		Password: c.FrontendPassword,
	}
}
