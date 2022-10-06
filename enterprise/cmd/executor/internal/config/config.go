package config

import (
	"runtime"
	"strconv"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/google/uuid"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Config struct {
	env.BaseConfig

	FrontendURL                   string
	FrontendAuthorizationToken    string
	QueueName                     string
	QueuePollInterval             time.Duration
	MaximumNumJobs                int
	FirecrackerImage              string
	FirecrackerKernelImage        string
	FirecrackerSandboxImage       string
	VMStartupScriptPath           string
	VMPrefix                      string
	KeepWorkspaces                bool
	DockerHostMountPath           string
	UseFirecracker                bool
	JobNumCPUs                    int
	JobMemory                     string
	FirecrackerDiskSpace          string
	FirecrackerBandwidthIngress   int
	FirecrackerBandwidthEgress    int
	MaximumRuntimePerJob          time.Duration
	CleanupTaskInterval           time.Duration
	NumTotalJobs                  int
	MaxActiveTime                 time.Duration
	NodeExporterURL               string
	DockerRegistryNodeExporterURL string
	WorkerHostname                string
	DockerRegistryMirrorURL       string
}

func (c *Config) Load() {
	c.FrontendURL = c.Get("EXECUTOR_FRONTEND_URL", "", "The external URL of the sourcegraph instance.")
	c.FrontendAuthorizationToken = c.Get("EXECUTOR_FRONTEND_PASSWORD", "", "The authorization token supplied to the frontend.")
	c.QueueName = c.Get("EXECUTOR_QUEUE_NAME", "", "The name of the queue to listen to.")
	c.QueuePollInterval = c.GetInterval("EXECUTOR_QUEUE_POLL_INTERVAL", "1s", "Interval between dequeue requests.")
	c.MaximumNumJobs = c.GetInt("EXECUTOR_MAXIMUM_NUM_JOBS", "1", "Number of virtual machines or containers that can be running at once.")
	c.UseFirecracker = c.GetBool("EXECUTOR_USE_FIRECRACKER", strconv.FormatBool(runtime.GOOS == "linux"), "Whether to isolate commands in virtual machines. Requires ignite and firecracker. Linux hosts only.")
	c.FirecrackerImage = c.Get("EXECUTOR_FIRECRACKER_IMAGE", DefaultFirecrackerImage, "The base image to use for virtual machines.")
	c.FirecrackerKernelImage = c.Get("EXECUTOR_FIRECRACKER_KERNEL_IMAGE", DefaultFirecrackerKernelImage, "The base image containing the kernel binary to use for virtual machines.")
	c.FirecrackerSandboxImage = c.Get("EXECUTOR_FIRECRACKER_SANDBOX_IMAGE", DefaultFirecrackerSandboxImage, "The OCI image for the ignite VM sandbox.")
	c.VMStartupScriptPath = c.GetOptional("EXECUTOR_VM_STARTUP_SCRIPT_PATH", "A path to a file on the host that is loaded into a fresh virtual machine and executed on startup.")
	c.VMPrefix = c.Get("EXECUTOR_VM_PREFIX", "executor", "A name prefix for virtual machines controlled by this instance.")
	c.KeepWorkspaces = c.GetBool("EXECUTOR_KEEP_WORKSPACES", "false", "Whether to skip deletion of workspaces after a job completes (or fails). Note that when Firecracker is enabled that the workspace is initially copied into the VM, so modifications will not be observed.")
	c.DockerHostMountPath = c.GetOptional("EXECUTOR_DOCKER_HOST_MOUNT_PATH", "The target workspace as it resides on the Docker host (used to enable Docker-in-Docker).")
	c.JobNumCPUs = c.GetInt(env.ChooseFallbackVariableName("EXECUTOR_JOB_NUM_CPUS", "EXECUTOR_FIRECRACKER_NUM_CPUS"), "4", "How many CPUs to allocate to each virtual machine or container. A value of zero sets no resource bound (in Docker, but not VMs).")
	c.JobMemory = c.Get(env.ChooseFallbackVariableName("EXECUTOR_JOB_MEMORY", "EXECUTOR_FIRECRACKER_MEMORY"), "12G", "How much memory to allocate to each virtual machine or container. A value of zero sets no resource bound (in Docker, but not VMs).")
	c.FirecrackerDiskSpace = c.Get("EXECUTOR_FIRECRACKER_DISK_SPACE", "20G", "How much disk space to allocate to each virtual machine.")
	c.FirecrackerBandwidthIngress = c.GetInt("EXECUTOR_FIRECRACKER_BANDWIDTH_INGRESS", "524288000", "How much bandwidth to allow for ingress packets to the VM in bytes/s.")
	c.FirecrackerBandwidthEgress = c.GetInt("EXECUTOR_FIRECRACKER_BANDWIDTH_EGRESS", "524288000", "How much bandwidth to allow for egress packets to the VM in bytes/s.")
	c.MaximumRuntimePerJob = c.GetInterval("EXECUTOR_MAXIMUM_RUNTIME_PER_JOB", "30m", "The maximum wall time that can be spent on a single job.")
	c.CleanupTaskInterval = c.GetInterval("EXECUTOR_CLEANUP_TASK_INTERVAL", "1m", "The frequency with which to run periodic cleanup tasks.")
	c.NumTotalJobs = c.GetInt("EXECUTOR_NUM_TOTAL_JOBS", "0", "The maximum number of jobs that will be dequeued by the worker.")
	c.NodeExporterURL = c.GetOptional("NODE_EXPORTER_URL", "The URL of the node_exporter instance, without the /metrics path.")
	c.DockerRegistryNodeExporterURL = c.GetOptional("DOCKER_REGISTRY_NODE_EXPORTER_URL", "The URL of the Docker Registry instance's node_exporter, without the /metrics path.")
	c.MaxActiveTime = c.GetInterval("EXECUTOR_MAX_ACTIVE_TIME", "0", "The maximum time that can be spent by the worker dequeueing records to be handled.")
	c.DockerRegistryMirrorURL = c.GetOptional("EXECUTOR_DOCKER_REGISTRY_MIRROR_URL", "The address of a docker registry mirror to use in firecracker VMs.")

	hn := hostname.Get()
	// Be unique but also descriptive.
	c.WorkerHostname = hn + "-" + uuid.New().String()
}

func (c *Config) Validate() error {
	if c.QueueName != "" && c.QueueName != "batches" && c.QueueName != "codeintel" {
		c.AddError(errors.New("EXECUTOR_QUEUE_NAME must be set to 'batches' or 'codeintel'"))
	}

	if c.UseFirecracker {
		// Validate that firecracker can work on this host.
		if runtime.GOOS != "linux" {
			c.AddError(errors.New("EXECUTOR_USE_FIRECRACKER is only supported on linux hosts."))
		}
		if runtime.GOARCH != "amd64" {
			c.AddError(errors.New("EXECUTOR_USE_FIRECRACKER is only supported on amd64 hosts."))
		}

		// Required by Firecracker: The vCPU number can only be 1 or an even number when hyperthreading is enabled.
		if c.JobNumCPUs != 1 && c.JobNumCPUs%2 != 0 {
			c.AddError(errors.New("EXECUTOR_JOB_NUM_CPUS must be 1 or an even number"))
		}

		// Make sure disk space is a valid datasize string.
		_, err := datasize.ParseString(c.FirecrackerDiskSpace)
		if err != nil {
			c.AddError(errors.Wrapf(err, "invalid disk size provided for EXECUTOR_FIRECRACKER_DISK_SPACE: %q", c.FirecrackerDiskSpace))
		}
	}

	return c.BaseConfig.Validate()
}
