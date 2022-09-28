package run

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
	apiworker "github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func apiWorkerOptions(c *config.Config, telemetryOptions apiclient.TelemetryOptions) apiworker.Options {
	return apiworker.Options{
		VMPrefix:           c.VMPrefix,
		KeepWorkspaces:     c.KeepWorkspaces,
		QueueName:          c.QueueName,
		WorkerOptions:      workerOptions(c),
		FirecrackerOptions: firecrackerOptions(c),
		ResourceOptions:    resourceOptions(c),
		GitServicePath:     "/.executors/git",
		ClientOptions:      clientOptions(c, telemetryOptions),
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
		CancelInterval:       c.QueuePollInterval,
		Metrics:              makeWorkerMetrics(c.QueueName),
		NumTotalJobs:         c.NumTotalJobs,
		MaxActiveTime:        c.MaxActiveTime,
		WorkerHostname:       c.WorkerHostname,
		MaximumRuntimePerJob: c.MaximumRuntimePerJob,
	}
}

func firecrackerOptions(c *config.Config) command.FirecrackerOptions {
	return command.FirecrackerOptions{
		Enabled:                 c.UseFirecracker,
		Image:                   c.FirecrackerImage,
		KernelImage:             c.FirecrackerKernelImage,
		SandboxImage:            c.FirecrackerSandboxImage,
		VMStartupScriptPath:     c.VMStartupScriptPath,
		DockerRegistryMirrorURL: c.DockerRegistryMirrorURL,
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

func clientOptions(c *config.Config, telemetryOptions apiclient.TelemetryOptions) apiclient.Options {
	return apiclient.Options{
		ExecutorName:      c.WorkerHostname,
		PathPrefix:        "/.executors/queue",
		EndpointOptions:   endpointOptions(c),
		BaseClientOptions: baseClientOptions(c),
		TelemetryOptions:  telemetryOptions,
	}
}

func baseClientOptions(c *config.Config) apiclient.BaseClientOptions {
	return apiclient.BaseClientOptions{}
}

func endpointOptions(c *config.Config) apiclient.EndpointOptions {
	return apiclient.EndpointOptions{
		URL:   c.FrontendURL,
		Token: c.FrontendAuthorizationToken,
	}
}

func makeWorkerMetrics(queueName string) workerutil.WorkerMetrics {
	observationContext := &observation.Context{
		Logger:     log.Scoped("executor_processor", "executor worker processor"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	return workerutil.NewMetrics(observationContext, "executor_processor",
		// derived from historic data, ideally we will use spare high-res histograms once they're a reality
		// 										 30s 1m	 2.5m 5m   7.5m 10m  15m  20m	30m	  45m	1hr
		workerutil.WithDurationBuckets([]float64{30, 60, 150, 300, 450, 600, 900, 1200, 1800, 2700, 3600}),
		workerutil.WithLabels(map[string]string{
			"queue": queueName,
		}),
	)
}
