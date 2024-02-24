package worker

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/apiclient/files"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/apiclient/queue"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/janitor"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/metrics"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/runner"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/runtime"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/workspace"
	"github.com/sourcegraph/sourcegraph/internal/executor/types"
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

	// QueueName is the name of the queue to process work from. Having this configurable
	// allows us to have multiple worker pools with different resource requirements and
	// horizontal scaling factors while still uniformly processing events. Only one of
	// QueueName and QueueNames can be set.
	QueueName string

	// QueueNames is the list of queue names to process work from. When multiple queues
	// are configured the frontend will dequeue a job from one of those queues and return
	// it to the executor to process. Only one of QueueNames and QueueName can be set.
	QueueNames []string

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

	RunnerOptions runner.Options

	// NodeExporterEndpoint is the URL of the local node_exporter endpoint, without
	// the /metrics path.
	NodeExporterEndpoint string

	// DockerRegistryNodeExporterEndpoint is the URL of the intermediary caching docker registry,
	// for scraping and forwarding metrics.
	DockerRegistryNodeExporterEndpoint string

	// EnableJobAuditLogging enables logging of the job payload to the executor logs.
	// Note that this mode might contain secret information and logs very verbosely.
	EnableJobAuditLogging bool
}

// NewWorker creates a worker that polls a remote job queue API for work.
func NewWorker(observationCtx *observation.Context, nameSet *janitor.NameSet, options Options) (goroutine.WaitableBackgroundRoutine, error) {
	observationCtx = observation.ContextWithLogger(observationCtx.Logger.Scoped("worker"), observationCtx)

	gatherer := metrics.MakeExecutorMetricsGatherer(log.Scoped("executor-worker.metrics-gatherer"), prometheus.DefaultGatherer, options.NodeExporterEndpoint, options.DockerRegistryNodeExporterEndpoint)
	queueClient, err := queue.New(observationCtx, options.QueueOptions, gatherer)
	if err != nil {
		return nil, errors.Wrap(err, "building queue worker client")
	}

	if !connectToFrontend(observationCtx.Logger, queueClient, options) {
		os.Exit(1)
	}

	filesClient, err := files.New(observationCtx, options.FilesOptions)
	if err != nil {
		return nil, errors.Wrap(err, "building files store")
	}

	commandOps := command.NewOperations(observationCtx)
	cloneOptions := workspace.CloneOptions{
		ExecutorName:   options.WorkerOptions.WorkerHostname,
		EndpointURL:    options.QueueOptions.BaseClientOptions.EndpointOptions.URL,
		GitServicePath: options.GitServicePath,
		ExecutorToken:  options.QueueOptions.BaseClientOptions.EndpointOptions.Token,
	}

	cmdRunner := &util.RealCmdRunner{}
	cmd := &command.RealCommand{
		CmdRunner: cmdRunner,
		Logger:    log.Scoped("executor-worker.command"),
	}

	// Configure the supported runtimes
	jobRuntime, err := runtime.New(observationCtx.Logger, commandOps, filesClient, cloneOptions, options.RunnerOptions, cmdRunner, cmd)
	if err != nil {
		return nil, err
	}

	h := &handler{
		nameSet:      nameSet,
		cmdRunner:    cmdRunner,
		cmd:          cmd,
		logStore:     queueClient,
		filesStore:   filesClient,
		options:      options,
		cloneOptions: cloneOptions,
		operations:   commandOps,
		jobRuntime:   jobRuntime,
	}

	return workerutil.NewWorker[types.Job](context.Background(), queueClient, h, options.WorkerOptions), nil
}

// connectToFrontend will ping the configured Sourcegraph instance until it receives a 200 response.
// For the first minute, "connection refused" errors will not be emitted. This is to stop log spam
// in dev environments where the executor may start up before the frontend. This method returns true
// after a ping is successful and returns false if a user signal is received.
func connectToFrontend(logger log.Logger, queueClient *queue.Client, options Options) bool {
	start := time.Now()
	logger = logger.With(log.String("url", options.QueueOptions.BaseClientOptions.EndpointOptions.URL))
	logger.Debug("Connecting to Sourcegraph instance")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	for {
		err := queueClient.Ping(context.Background())
		if err == nil {
			logger.Debug("Connected to Sourcegraph instance")
			return true
		}

		var e *os.SyscallError
		if errors.As(err, &e) && e.Syscall == "connect" && time.Since(start) < time.Minute {
			// Hide initial connection logs due to services starting up in an nondeterminstic order.
			// Logs occurring one minute after startup or later are not filtered, nor are non-expected
			// connection errors during app startup.
		} else {
			logger.Error("Failed to connect to Sourcegraph instance", log.Error(err))
		}

		select {
		case <-ticker.C:
		case sig := <-signals:
			logger.Error("Signal received while connecting to Sourcegraph", log.String("signal", sig.String()))
			return false
		}
	}
}
