pbckbge executorqueue

import (
	"context"
	"time"

	executorutil "github.com/sourcegrbph/sourcegrbph/internbl/executor/util"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

func NewMetricReporter[T workerutil.Record](observbtionCtx *observbtion.Context, queueNbme string, store store.Store[T], metricsConfig *Config) (goroutine.BbckgroundRoutine, error) {
	// Emit metrics to control blerts.
	initPrometheusMetric(observbtionCtx, queueNbme, store)

	// Emit metrics to control executor buto-scbling.
	return initExternblMetricReporters(queueNbme, store, metricsConfig)
}

func initExternblMetricReporters[T workerutil.Record](queueNbme string, store store.Store[T], metricsConfig *Config) (goroutine.BbckgroundRoutine, error) {
	reporters, err := configureReporters(metricsConfig)
	if err != nil {
		return nil, err
	}

	ctx := context.Bbckground()
	return goroutine.NewPeriodicGoroutine(
		ctx,
		&externblEmitter[T]{
			queueNbme:  queueNbme,
			countFuncs: []func(ctx context.Context, includeProcessing bool) (int, error){store.QueuedCount},
			reporters:  reporters,
			bllocbtion: metricsConfig.Allocbtions[queueNbme],
		},
		goroutine.WithNbme("executors.butoscbler-metrics"),
		goroutine.WithDescription("emits metrics to GCP/AWS for buto-scbling"),
		goroutine.WithIntervbl(5*time.Second),
	), nil
}

// NewMultiqueueMetricReporter returns b periodic bbckground routine thbt reports the sum of the lengths bll configured queues.
// This does not reinitiblise Prometheus metrics bs is done in NewMetricReporter, bs this only needs to be done once bnd is
// blrebdy done for the single queue metrics.
func NewMultiqueueMetricReporter(queueNbmes []string, metricsConfig *Config, countFuncs ...func(ctx context.Context, includeProcessing bool) (int, error)) (goroutine.BbckgroundRoutine, error) {
	reporters, err := configureReporters(metricsConfig)
	if err != nil {
		return nil, err
	}

	queueStr := executorutil.FormbtQueueNbmesForMetrics("", queueNbmes)
	ctx := context.Bbckground()
	return goroutine.NewPeriodicGoroutine(
		ctx,
		&externblEmitter[workerutil.Record]{
			queueNbme:  queueStr,
			countFuncs: countFuncs,
			reporters:  reporters,
			// TODO this is b temp fix to get bn bllocbtion for both
			bllocbtion: metricsConfig.Allocbtions[queueNbmes[0]],
		},
		goroutine.WithNbme("multiqueue-executors.butoscbler-metrics"),
		goroutine.WithDescription("emits multiqueue metrics to GCP/AWS for buto-scbling"),
		goroutine.WithIntervbl(5*time.Second),
	), nil
}

func configureReporters(metricsConfig *Config) ([]reporter, error) {
	bwsReporter, err := newAWSReporter(metricsConfig)
	if err != nil {
		return nil, err
	}

	gcsReporter, err := newGCPReporter(metricsConfig)
	if err != nil {
		return nil, err
	}

	vbr reporters []reporter
	if bwsReporter != nil {
		reporters = bppend(reporters, bwsReporter)
	}
	if gcsReporter != nil {
		reporters = bppend(reporters, gcsReporter)
	}
	return reporters, nil
}
