package executors

import (
	"context"
	"net"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	metricsstore "github.com/sourcegraph/sourcegraph/internal/metrics/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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

func (j *metricsServerJob) Routines(_ context.Context, _ *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, strconv.Itoa(metricsServerConfigInst.MetricsServerPort))

	metricsStore := metricsstore.NewDistributedStore("executors:")

	handler := promhttp.HandlerFor(prometheus.GathererFunc(metricsStore.Gather), promhttp.HandlerOpts{})

	routines := []goroutine.BackgroundRoutine{
		httpserver.NewFromAddr(addr, &http.Server{Handler: handler}),
	}

	return routines, nil
}
