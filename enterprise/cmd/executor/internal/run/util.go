package run

import (
	"context"
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient/queue"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/util"
	apiworker "github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/runner"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func newQueueTelemetryOptions(ctx context.Context, runner util.CmdRunner, useFirecracker bool, logger log.Logger) queue.TelemetryOptions {
	t := queue.TelemetryOptions{
		OS:              runtime.GOOS,
		Architecture:    runtime.GOARCH,
		ExecutorVersion: version.Version(),
	}

	var err error

	t.GitVersion, err = util.GetGitVersion(ctx, runner)
	if err != nil {
		logger.Error("Failed to get git version", log.Error(err))
	}

	// TODO: k8s handling??
	//t.SrcCliVersion, err = util.GetSrcVersion(ctx, runner)
	//if err != nil {
	//	logger.Error("Failed to get src-cli version", log.Error(err))
	//}

	//t.DockerVersion, err = util.GetDockerVersion(ctx, runner)
	//if err != nil {
	//	logger.Error("Failed to get docker version", log.Error(err))
	//}

	if useFirecracker {
		t.IgniteVersion, err = util.GetIgniteVersion(ctx, runner)
		if err != nil {
			logger.Error("Failed to get ignite version", log.Error(err))
		}
	}

	return t
}

func apiWorkerOptions(c *config.Config, queueTelemetryOptions queue.TelemetryOptions) apiworker.Options {
	return apiworker.Options{
		VMPrefix:      c.VMPrefix,
		QueueName:     c.QueueName,
		WorkerOptions: workerOptions(c),
		RunnerOptions: runner.Options{
			DockerOptions:      dockerOptions(c),
			FirecrackerOptions: firecrackerOptions(c),
		},
		GitServicePath: "/.executors/git",
		QueueOptions:   queueOptions(c, queueTelemetryOptions),
		FilesOptions:   filesOptions(c),
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

func dockerOptions(c *config.Config) command.DockerOptions {
	u, _ := url.Parse(c.FrontendURL)
	return command.DockerOptions{
		DockerAuthConfig: c.DockerAuthConfig,
		// If the configured Sourcegraph endpoint is host.docker.internal add a
		// host entry and route to it to the containers. This is used for LSIF
		// uploads and should not be required anymore once we support native uploads.
		AddHostGateway: u.Hostname() == "host.docker.internal",
		Resources:      resourceOptions(c),
	}
}

func firecrackerOptions(c *config.Config) runner.FirecrackerOptions {
	var dockerMirrors []string
	if len(c.DockerRegistryMirrorURL) > 0 {
		dockerMirrors = strings.Split(c.DockerRegistryMirrorURL, ",")
	}
	return runner.FirecrackerOptions{
		Enabled:                  c.UseFirecracker,
		Image:                    c.FirecrackerImage,
		KernelImage:              c.FirecrackerKernelImage,
		SandboxImage:             c.FirecrackerSandboxImage,
		VMStartupScriptPath:      c.VMStartupScriptPath,
		DockerRegistryMirrorURLs: dockerMirrors,
		DockerOptions:            dockerOptions(c),
		KeepWorkspaces:           c.KeepWorkspaces,
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
		QueueName:         c.QueueName,
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
		ExecutorName:    c.WorkerHostname,
		EndpointOptions: endpointOptions(c, "/.executors/files"),
	}
}

func baseClientOptions(c *config.Config, pathPrefix string) apiclient.BaseClientOptions {
	return apiclient.BaseClientOptions{
		ExecutorName:    c.WorkerHostname,
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
