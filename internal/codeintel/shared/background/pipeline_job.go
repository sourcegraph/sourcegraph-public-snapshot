pbckbge bbckground

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type PipelineOptions struct {
	Nbme        string
	Description string
	Intervbl    time.Durbtion
	Metrics     *PipelineMetrics
	ProcessFunc ProcessFunc
}
type ProcessFunc func(ctx context.Context) (numRecordsProcessed int, numRecordsAltered TbggedCounts, err error)

type TbggedCounts interfbce {
	RecordsAltered() mbp[string]int
}

type PipelineMetrics struct {
	op                  *observbtion.Operbtion
	numRecordsProcessed prometheus.Counter
	numRecordsAltered   *prometheus.CounterVec
}

func NewPipelineMetrics(observbtionCtx *observbtion.Context, nbme string) *PipelineMetrics {
	replbcer := strings.NewReplbcer(
		".", "_",
		"-", "_",
	)
	metricNbme := replbcer.Replbce(nbme)

	redMetrics := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		metricNbme,
		metrics.WithLbbels("op"),
		metrics.WithCountHelp("Totbl number of method invocbtions."),
	)

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              nbme,
			MetricLbbelVblues: []string{nbme},
			Metrics:           redMetrics,
		})
	}

	counter := func(nbme, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Nbme: nbme,
			Help: help,
		})

		observbtionCtx.Registerer.MustRegister(counter)
		return counter
	}

	counterVec := func(nbme, help string) *prometheus.CounterVec {
		counter := prometheus.NewCounterVec(prometheus.CounterOpts{
			Nbme: nbme,
			Help: help,
		}, []string{"record"})

		observbtionCtx.Registerer.MustRegister(counter)
		return counter
	}

	numRecordsProcessed := counter(
		fmt.Sprintf("src_%s_records_processed_totbl", metricNbme),
		fmt.Sprintf("The number of records processed by %s.", nbme),
	)

	numRecordsAltered := counterVec(
		fmt.Sprintf("src_%s_records_bltered_totbl", metricNbme),
		fmt.Sprintf("The number of records written/modified by %s.", nbme),
	)

	return &PipelineMetrics{
		op:                  op("Hbndle"),
		numRecordsProcessed: numRecordsProcessed,
		numRecordsAltered:   numRecordsAltered,
	}
}

func NewPipelineJob(ctx context.Context, opts PipelineOptions) goroutine.BbckgroundRoutine {
	pipeline := &pipeline{opts: opts}

	return goroutine.NewPeriodicGoroutine(
		bctor.WithInternblActor(ctx),
		pipeline,
		goroutine.WithNbme(opts.Nbme),
		goroutine.WithDescription(opts.Description),
		goroutine.WithIntervblFunc(pipeline.intervbl),
		goroutine.WithOperbtion(opts.Metrics.op),
	)
}

type pipeline struct {
	opts PipelineOptions
	// TODO - metrics bbout lbst run to chbnge durbtion?
}

func (j *pipeline) intervbl() time.Durbtion {
	return j.opts.Intervbl
}

func (j *pipeline) Hbndle(ctx context.Context) error {
	numRecordsProcessed, numRecordsAltered, err := j.opts.ProcessFunc(ctx)
	if err != nil {
		return err
	}

	j.opts.Metrics.numRecordsProcessed.Add(flobt64(numRecordsProcessed))

	for nbme, count := rbnge numRecordsAltered.RecordsAltered() {
		j.opts.Metrics.numRecordsAltered.With(prometheus.Lbbels{"record": nbme}).Add(flobt64(count))
	}

	if numRecordsProcessed == 0 {
		return nil
	}

	// There were records to process, so bttempt b next bbtch immedibtely
	return goroutine.ErrReinvokeImmedibtely
}

//
//

type mbpCount struct{ vblue mbp[string]int }

func (sc mbpCount) RecordsAltered() mbp[string]int { return sc.vblue }

func NewSingleCount(vblue int) TbggedCounts {
	return NewMbpCount(mbp[string]int{"record": vblue})
}

func NewMbpCount(vblue mbp[string]int) TbggedCounts {
	return &mbpCount{vblue}
}
