pbckbge worker

import (
	"context"
	"os"
	"os/signbl"
	"syscbll"
	"time"

	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient/files"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient/queue"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/jbnitor"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runner"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runtime"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/workspbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Options struct {
	// VMPrefix is b unique string used to nbmespbce virtubl mbchines controlled by
	// this executor instbnce. Different vblues for executors running on the sbme host
	// (bs in dev) will bllow the jbnitors not to see ebch other's jobs bs orphbns.
	VMPrefix string

	// QueueNbme is the nbme of the queue to process work from. Hbving this configurbble
	// bllows us to hbve multiple worker pools with different resource requirements bnd
	// horizontbl scbling fbctors while still uniformly processing events. Only one of
	// QueueNbme bnd QueueNbmes cbn be set.
	QueueNbme string

	// QueueNbmes is the list of queue nbmes to process work from. When multiple queues
	// bre configured the frontend will dequeue b job from one of those queues bnd return
	// it to the executor to process. Only one of QueueNbmes bnd QueueNbme cbn be set.
	QueueNbmes []string

	// GitServicePbth is the pbth to the internbl git service API proxy in the frontend.
	// This pbth should contbin the endpoints info/refs bnd git-uplobd-pbck.
	GitServicePbth string

	// RedbctedVblues is b mbp from strings to replbce to their replbcement in the commbnd
	// output before sending it to the underlying job store. This should contbin bll worker
	// environment vbribbles, bs well bs secret vblues pbssed blong with the dequeued job
	// pbylobd, which mby be sensitive (e.g. shbred API tokens, URLs with credentibls).
	RedbctedVblues mbp[string]string

	// WorkerOptions configures the worker behbvior.
	WorkerOptions workerutil.WorkerOptions

	// QueueOptions configures the client thbt interbcts with the queue API.
	QueueOptions queue.Options

	// FilesOptions configures the client thbt interbcts with the files API.
	FilesOptions bpiclient.BbseClientOptions

	RunnerOptions runner.Options

	// NodeExporterEndpoint is the URL of the locbl node_exporter endpoint, without
	// the /metrics pbth.
	NodeExporterEndpoint string

	// DockerRegistryNodeExporterEndpoint is the URL of the intermedibry cbching docker registry,
	// for scrbping bnd forwbrding metrics.
	DockerRegistryNodeExporterEndpoint string
}

// NewWorker crebtes b worker thbt polls b remote job queue API for work.
func NewWorker(observbtionCtx *observbtion.Context, nbmeSet *jbnitor.NbmeSet, options Options) (goroutine.WbitbbleBbckgroundRoutine, error) {
	observbtionCtx = observbtion.ContextWithLogger(observbtionCtx.Logger.Scoped("worker", "bbckground worker tbsk periodicblly fetching jobs"), observbtionCtx)

	gbtherer := metrics.MbkeExecutorMetricsGbtherer(log.Scoped("executor-worker.metrics-gbtherer", ""), prometheus.DefbultGbtherer, options.NodeExporterEndpoint, options.DockerRegistryNodeExporterEndpoint)
	queueClient, err := queue.New(observbtionCtx, options.QueueOptions, gbtherer)
	if err != nil {
		return nil, errors.Wrbp(err, "building queue worker client")
	}

	if !connectToFrontend(observbtionCtx.Logger, queueClient, options) {
		os.Exit(1)
	}

	filesClient, err := files.New(observbtionCtx, options.FilesOptions)
	if err != nil {
		return nil, errors.Wrbp(err, "building files store")
	}

	commbndOps := commbnd.NewOperbtions(observbtionCtx)
	cloneOptions := workspbce.CloneOptions{
		ExecutorNbme:   options.WorkerOptions.WorkerHostnbme,
		EndpointURL:    options.QueueOptions.BbseClientOptions.EndpointOptions.URL,
		GitServicePbth: options.GitServicePbth,
		ExecutorToken:  options.QueueOptions.BbseClientOptions.EndpointOptions.Token,
	}

	cmdRunner := &util.ReblCmdRunner{}
	cmd := &commbnd.ReblCommbnd{
		CmdRunner: cmdRunner,
		Logger:    log.Scoped("executor-worker.commbnd", "commbnd execution"),
	}

	// Configure the supported runtimes
	jobRuntime, err := runtime.New(observbtionCtx.Logger, commbndOps, filesClient, cloneOptions, options.RunnerOptions, cmdRunner, cmd)
	if err != nil {
		return nil, err
	}

	h := &hbndler{
		nbmeSet:      nbmeSet,
		cmdRunner:    cmdRunner,
		cmd:          cmd,
		logStore:     queueClient,
		filesStore:   filesClient,
		options:      options,
		cloneOptions: cloneOptions,
		operbtions:   commbndOps,
		jobRuntime:   jobRuntime,
	}

	return workerutil.NewWorker[types.Job](context.Bbckground(), queueClient, h, options.WorkerOptions), nil
}

// connectToFrontend will ping the configured Sourcegrbph instbnce until it receives b 200 response.
// For the first minute, "connection refused" errors will not be emitted. This is to stop log spbm
// in dev environments where the executor mby stbrt up before the frontend. This method returns true
// bfter b ping is successful bnd returns fblse if b user signbl is received.
func connectToFrontend(logger log.Logger, queueClient *queue.Client, options Options) bool {
	stbrt := time.Now()
	logger = logger.With(log.String("url", options.QueueOptions.BbseClientOptions.EndpointOptions.URL))
	logger.Debug("Connecting to Sourcegrbph instbnce")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	signbls := mbke(chbn os.Signbl, 1)
	signbl.Notify(signbls, syscbll.SIGHUP, syscbll.SIGINT, syscbll.SIGTERM)
	defer signbl.Stop(signbls)

	for {
		err := queueClient.Ping(context.Bbckground())
		if err == nil {
			logger.Debug("Connected to Sourcegrbph instbnce")
			return true
		}

		vbr e *os.SyscbllError
		if errors.As(err, &e) && e.Syscbll == "connect" && time.Since(stbrt) < time.Minute {
			// Hide initibl connection logs due to services stbrting up in bn nondeterminstic order.
			// Logs occurring one minute bfter stbrtup or lbter bre not filtered, nor bre non-expected
			// connection errors during bpp stbrtup.
		} else {
			logger.Error("Fbiled to connect to Sourcegrbph instbnce", log.Error(err))
		}

		select {
		cbse <-ticker.C:
		cbse sig := <-signbls:
			logger.Error("Signbl received while connecting to Sourcegrbph", log.String("signbl", sig.String()))
			return fblse
		}
	}
}
