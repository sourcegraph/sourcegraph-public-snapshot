package server

import (
	"context"
	"io"
	"os/exec"
	"strconv"
	"time"

	"github.com/mxk/go-flowrate/flowrate"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server/internal/accesslog"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/gitservice"
)

var gitServiceMaxEgressBytesPerSecond = func() int64 {
	logger := log.Scoped("gitServiceMaxEgressBytesPerSecond", "git service max egress bytes per second")
	bps, err := strconv.ParseInt(env.Get(
		"SRC_GIT_SERVICE_MAX_EGRESS_BYTES_PER_SECOND",
		"1000000000",
		"Git service egress rate limit in bytes per second (-1 = no limit, default = 1Gbps)"),
		10,
		64,
	)
	if err != nil {
		logger.Error("gitservice: failed parsing SRC_GIT_SERVICE_MAX_EGRESS_BYTES_PER_SECOND. defaulting to 1Gbps", log.Int64("bps", bps), log.Error(err))
		bps = 1000 * 1000 * 1000 // 1Gbps
	}
	return bps
}()

// flowrateWriter limits the write rate of w to 1 Gbps.
//
// We are cloning repositories from within the same network from another
// Sourcegraph service (zoekt-indexserver). This can end up being so fast that
// we harm our own network connectivity. In the case of zoekt-indexserver and
// gitserver running on the same host machine, we can even reach up to ~100
// Gbps and effectively DoS the Docker network, temporarily disrupting other
// containers running on the host.
//
// Google Compute Engine has a network bandwidth of about 1.64 Gbps
// between nodes, and AWS varies widely depending on instance type.
// We play it safe and default to 1 Gbps here (~119 MiB/s), which
// means we can fetch a 1 GiB archive in ~8.5 seconds.
func flowrateWriter(w io.Writer) io.Writer {
	if gitServiceMaxEgressBytesPerSecond > 0 {
		return flowrate.NewWriter(w, gitServiceMaxEgressBytesPerSecond)
	}
	return w
}

func (s *Server) gitServiceHandler() *gitservice.Handler {
	return &gitservice.Handler{
		Logger: s.Logger.Scoped("gitServiceHandler", "smart Git HTTP transfer protocol"),

		Dir: func(d string) string {
			return string(s.dir(api.RepoName(d)))
		},

		// Limit rate of stdout from git.
		CommandHook: func(cmd *exec.Cmd) {
			cmd.Stdout = flowrateWriter(cmd.Stdout)
		},

		Trace: func(ctx context.Context, svc, repo, protocol string) func(error) {
			start := time.Now()
			metricServiceRunning.WithLabelValues(svc).Inc()

			// Log which which actor is accessing the repo.
			accesslog.Record(ctx, repo, map[string]string{
				"svc":      svc,
				"protocol": protocol,
			})

			return func(err error) {
				errLabel := strconv.FormatBool(err != nil)
				metricServiceRunning.WithLabelValues(svc).Dec()
				metricServiceDuration.WithLabelValues(svc, errLabel).Observe(time.Since(start).Seconds())

				logger := s.Logger.With(
					log.String("svc", svc),
					log.String("repo", repo),
					log.String("protocol", protocol),
					log.Duration("duration", time.Since(start)),
				)

				if err != nil {
					logger.Error("gitservice.ServeHTTP", log.Error(err))
				} else if traceLogs {
					logger.Debug("TRACE gitserver git service")
				}
			}
		},
	}
}

var (
	metricServiceDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_gitserver_gitservice_duration_seconds",
		Help:    "A histogram of latencies for the git service (upload-pack for internal clones) endpoint.",
		Buckets: prometheus.ExponentialBuckets(.1, 4, 9),
		// [0.1 0.4 1.6 6.4 25.6 102.4 409.6 1638.4 6553.6]
	}, []string{"type", "error"})

	metricServiceRunning = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_gitserver_gitservice_running",
		Help: "A histogram of latencies for the git service (upload-pack for internal clones) endpoint.",
	}, []string{"type"})
)
