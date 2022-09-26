package worker

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient/files"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient/queue"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/metrics"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Options struct {
	// VMPrefix is a unique string used to namespace virtual machines controlled by
	// this executor instance. Different values for executors running on the same host
	// (as in dev) will allow the janitors not to see each other's jobs as orphans.
	VMPrefix string

	// KeepWorkspaces prevents deletion of a workspace after a job completes. Setting
	// this value to true will continually use more and more disk, so it should only
	// be used as a debugging mechanism.
	KeepWorkspaces bool

	// QueueName is the name of the queue to process work from. Having this configurable
	// allows us to have multiple worker pools with different resource requirements and
	// horizontal scaling factors while still uniformly processing events.
	QueueName string

	// GitServicePath is the path to the internal git service API proxy in the frontend.
	// This path should contain the endpoints info/refs and git-upload-pack.
	GitServicePath string

	// RedactedValues is a map from strings to replace to their replacement in the command
	// output before sending it to the underlying job store. This should contain all worker
	// environment variables, as well as secret values passed along with the dequeued job
	// payload, which may be sensitive (e.g. shared API tokens, URLs with credentials).
	RedactedValues map[string]string

	// WorkerOptions configures the worker behavior.
	WorkerOptions workerutil.WorkerOptions

	// QueueOptions configures the client that interacts with the queue API.
	QueueOptions queue.Options

	// FilesOptions configures the client that interacts with the files API.
	FilesOptions apiclient.BaseClientOptions

	// FirecrackerOptions configures the behavior of Firecracker virtual machine creation.
	FirecrackerOptions command.FirecrackerOptions

	// ResourceOptions configures the resource limits of docker container and Firecracker
	// virtual machines running on the executor.
	ResourceOptions command.ResourceOptions

	// NodeExporterEndpoint is the URL of the local node_exporter endpoint, without
	// the /metrics path.
	NodeExporterEndpoint string

	// DockerRegsitryEndpoint is the URL of the intermediary caching docker registry,
	// for scraping and forwarding metrics.
	DockerRegistryNodeExporterEndpoint string
}

// NewWorker creates a worker that polls a remote job queue API for work. The returned
// routine contains both a worker that periodically polls for new work to perform, as well
// as a heartbeat routine that will periodically hit the remote API with the work that is
// currently being performed, which is necessary so the job queue API doesn't hand out jobs
// it thinks may have been dropped.
func NewWorker(nameSet *janitor.NameSet, options Options, observationContext *observation.Context) (goroutine.WaitableBackgroundRoutine, error) {
	gatherer := metrics.MakeExecutorMetricsGatherer(log.Scoped("executor-worker.metrics-gatherer", ""), prometheus.DefaultGatherer, options.NodeExporterEndpoint, options.DockerRegistryNodeExporterEndpoint)
	queueStore, err := queue.New(options.QueueOptions, gatherer, observationContext)
	if err != nil {
		return nil, errors.Wrap(err, "building queue store")
	}
	filesStore, err := files.New(options.FilesOptions, observationContext)
	if err != nil {
		return nil, errors.Wrap(err, "building files store")
	}
	shim := &store.QueueShim{Name: options.QueueName, Store: queueStore}

	if !connectToFrontend(queueStore, options) {
		os.Exit(1)
	}

	h := &handler{
		nameSet:       nameSet,
		store:         shim,
		filesStore:    filesStore,
		options:       options,
		operations:    command.NewOperations(observationContext),
		runnerFactory: command.NewRunner,
	}

	ctx := context.Background()

	return workerutil.NewWorker(ctx, shim, h, options.WorkerOptions), nil
}

// connectToFrontend will ping the configured Sourcegraph instance until it receives a 200 response.
// For the first minute, "connection refused" errors will not be emitted. This is to stop log spam
// in dev environments where the executor may start up before the frontend. This method returns true
// after a ping is successful and returns false if a user signal is received.
func connectToFrontend(queueStore *queue.Client, options Options) bool {
	start := time.Now()
	log15.Info("Connecting to Sourcegraph instance", "url", options.QueueOptions.BaseClientOptions.EndpointOptions.URL)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	for {
		err := queueStore.Ping(context.Background(), options.QueueName, nil)
		if err == nil {
			log15.Info("Connected to Sourcegraph instance")
			return true
		}

		var e *os.SyscallError
		if errors.As(err, &e) && e.Syscall == "connect" && time.Since(start) < time.Minute {
			// Hide initial connection logs due to services starting up in an nondeterminstic order.
			// Logs occurring one minute after startup or later are not filtered, nor are non-expected
			// connection errors during app startup.
		} else {
			log15.Error("Failed to connect to Sourcegraph instance", "error", err)
		}

		select {
		case <-ticker.C:
		case <-signals:
			log15.Error("Signal received while connecting to Sourcegraph")
			return false
		}
	}
}
