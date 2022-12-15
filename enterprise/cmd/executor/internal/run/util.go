package run

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient/queue"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
	apiworker "github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func newQueueTelemetryOptions(ctx context.Context, useFirecracker bool, logger log.Logger) queue.TelemetryOptions {
	t := queue.TelemetryOptions{
		OS:              runtime.GOOS,
		Architecture:    runtime.GOARCH,
		ExecutorVersion: version.Version(),
	}

	var err error

	t.GitVersion, err = getGitVersion(ctx)
	if err != nil {
		logger.Error("Failed to get git version", log.Error(err))
	}

	t.SrcCliVersion, err = getSrcVersion(ctx)
	if err != nil {
		logger.Error("Failed to get src-cli version", log.Error(err))
	}

	t.DockerVersion, err = getDockerVersion(ctx)
	if err != nil {
		logger.Error("Failed to get docker version", log.Error(err))
	}

	if useFirecracker {
		t.IgniteVersion, err = getIgniteVersion(ctx)
		if err != nil {
			logger.Error("Failed to get ignite version", log.Error(err))
		}
	}

	return t
}

func getGitVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "git version "), nil
}

func getSrcVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "src", "version", "-client-only")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "Current version: "), nil
}

func getDockerVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "version", "-f", "{{.Server.Version}}")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func getIgniteVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "ignite", "version", "-o", "short")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func apiWorkerOptions(c *config.Config, queueTelemetryOptions queue.TelemetryOptions) apiworker.Options {
	return apiworker.Options{
		VMPrefix:           c.VMPrefix,
		KeepWorkspaces:     c.KeepWorkspaces,
		QueueName:          c.QueueName,
		WorkerOptions:      workerOptions(c),
		FirecrackerOptions: firecrackerOptions(c),
		ResourceOptions:    resourceOptions(c),
		GitServicePath:     "/.executors/git",
		QueueOptions:       queueOptions(c, queueTelemetryOptions),
		FilesOptions:       filesOptions(c),
		RedactedValues: map[string]string{
			// 🚨 SECURITY: Catch uses of the shared frontend token used to clone
			// git repositories that make it into commands or stdout/stderr streams.
			c.FrontendAuthorizationToken: "SECRET_REMOVED",
		},

		NodeExporterEndpoint:               c.NodeExporterURL,
		DockerRegistryNodeExporterEndpoint: c.DockerRegistryNodeExporterURL,
	}
}

func workerOptions(c *config.Config) workerutil.WorkerOptions {
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

func firecrackerOptions(c *config.Config) command.FirecrackerOptions {
	dockerMirrors := []string{}
	if len(c.DockerRegistryMirrorURL) > 0 {
		dockerMirrors = strings.Split(c.DockerRegistryMirrorURL, ",")
	}
	return command.FirecrackerOptions{
		Enabled:                  c.UseFirecracker,
		Image:                    c.FirecrackerImage,
		KernelImage:              c.FirecrackerKernelImage,
		SandboxImage:             c.FirecrackerSandboxImage,
		VMStartupScriptPath:      c.VMStartupScriptPath,
		DockerRegistryMirrorURLs: dockerMirrors,
	}
}

func resourceOptions(c *config.Config) command.ResourceOptions {
	return command.ResourceOptions{
		NumCPUs:             c.JobNumCPUs,
		Memory:              c.JobMemory,
		DiskSpace:           c.FirecrackerDiskSpace,
		DockerHostMountPath: c.DockerHostMountPath,
		MaxIngressBandwidth: c.FirecrackerBandwidthIngress,
		MaxEgressBandwidth:  c.FirecrackerBandwidthEgress,
	}
}

func queueOptions(c *config.Config, telemetryOptions queue.TelemetryOptions) queue.Options {
	return queue.Options{
		ExecutorName:      c.WorkerHostname,
		BaseClientOptions: baseClientOptions(c, "/.executors/queue"),
		TelemetryOptions:  telemetryOptions,
		ResourceOptions: queue.ResourceOptions{
			NumCPUs:   c.JobNumCPUs,
			Memory:    c.JobMemory,
			DiskSpace: c.FirecrackerDiskSpace,
		},
	}
}

func filesOptions(c *config.Config) apiclient.BaseClientOptions {
	return apiclient.BaseClientOptions{
		EndpointOptions: endpointOptions(c, "/.executors/files"),
	}
}

func baseClientOptions(c *config.Config, pathPrefix string) apiclient.BaseClientOptions {
	return apiclient.BaseClientOptions{
		EndpointOptions: endpointOptions(c, pathPrefix),
	}
}

func endpointOptions(c *config.Config, pathPrefix string) apiclient.EndpointOptions {
	return apiclient.EndpointOptions{
		URL:        c.FrontendURL,
		PathPrefix: pathPrefix,
		Token:      c.FrontendAuthorizationToken,
	}
}

func makeWorkerMetrics(queueName string) workerutil.WorkerObservability {
	observationCtx := observation.NewContext(log.Scoped("executor_processor", "executor worker processor"))

	return workerutil.NewMetrics(observationCtx, "executor_processor", workerutil.WithSampler(func(job workerutil.Record) bool { return true }),
		// derived from historic data, ideally we will use spare high-res histograms once they're a reality
		// 										 30s 1m	 2.5m 5m   7.5m 10m  15m  20m	30m	  45m	1hr
		workerutil.WithDurationBuckets([]float64{30, 60, 150, 300, 450, 600, 900, 1200, 1800, 2700, 3600}),
		workerutil.WithLabels(map[string]string{
			"queue": queueName,
		}),
	)
}
