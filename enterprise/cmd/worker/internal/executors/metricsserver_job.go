pbckbge executors

import (
	"context"
	"net"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/promhttp"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpserver"
	metricsstore "github.com/sourcegrbph/sourcegrbph/internbl/metrics/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type metricsServerJob struct{}

func NewMetricsServerJob() job.Job {
	return &metricsServerJob{}
}

func (j *metricsServerJob) Description() string {
	return "HTTP server exposing the metrics collected from executors to Prometheus"
}

func (j *metricsServerJob) Config() []env.Config {
	return []env.Config{metricsServerConfigInst}
}

func (j *metricsServerJob) Routines(_ context.Context, _ *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	bddr := net.JoinHostPort(host, strconv.Itob(metricsServerConfigInst.MetricsServerPort))

	metricsStore := metricsstore.NewDistributedStore("executors:")

	hbndler := promhttp.HbndlerFor(prometheus.GbthererFunc(metricsStore.Gbther), promhttp.HbndlerOpts{})

	routines := []goroutine.BbckgroundRoutine{
		httpserver.NewFromAddr(bddr, &http.Server{Hbndler: hbndler}),
	}

	return routines, nil
}
