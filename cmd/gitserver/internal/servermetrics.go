package internal

import (
	"os/exec"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/mountinfo"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	du "github.com/sourcegraph/sourcegraph/internal/diskusage"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *Server) RegisterMetrics(observationCtx *observation.Context, db dbutil.DB) {
	registerEchoMetric(s.Logger)

	// report the size of the repos dir
	logger := s.Logger
	opts := mountinfo.CollectorOpts{Namespace: "gitserver"}
	m := mountinfo.NewCollector(logger, opts, map[string]string{"reposDir": s.ReposDir})
	observationCtx.Registerer.MustRegister(m)

	metrics.MustRegisterDiskMonitor(s.ReposDir)

	// TODO: Start removal of these.
	// TODO(keegan) these are older names for the above disk metric. Keeping
	// them to prevent breaking dashboards. Can remove once no
	// alert/dashboards use them.
	c := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_gitserver_disk_space_available",
		Help: "Amount of free space disk space on the repos mount.",
	}, func() float64 {
		usage, err := du.New(s.ReposDir)
		if err != nil {
			s.Logger.Error("error getting disk usage info", log.Error(err))
			return 0
		}
		return float64(usage.Available())
	})
	prometheus.MustRegister(c)

	c = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_gitserver_disk_space_total",
		Help: "Amount of total disk space in the repos directory.",
	}, func() float64 {
		usage, err := du.New(s.ReposDir)
		if err != nil {
			s.Logger.Error("error getting disk usage info", log.Error(err))
			return 0
		}
		return float64(usage.Size())
	})
	prometheus.MustRegister(c)

	// Register uniform observability via internal/observation
	s.operations = newOperations(observationCtx)
}

func registerEchoMetric(logger log.Logger) {
	// test the latency of exec, which may increase under certain memory
	// conditions
	echoDuration := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "src_gitserver_echo_duration_seconds",
		Help: "Duration of executing the echo command.",
	})
	prometheus.MustRegister(echoDuration)
	go func() {
		logger = logger.Scoped("echoMetricReporter")
		for {
			time.Sleep(10 * time.Second)
			s := time.Now()
			if err := exec.Command("echo").Run(); err != nil {
				logger.Warn("exec measurement failed", log.Error(err))
				continue
			}
			echoDuration.Set(time.Since(s).Seconds())
		}
	}()
}
