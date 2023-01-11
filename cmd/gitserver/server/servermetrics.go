package server

import (
	"os/exec"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/mountinfo"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *Server) RegisterMetrics(observationCtx *observation.Context, db dbutil.DB) {
	// test the latency of exec, which may increase under certain memory
	// conditions
	echoDuration := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "src_gitserver_echo_duration_seconds",
		Help: "Duration of executing the echo command.",
	})
	prometheus.MustRegister(echoDuration)
	go func(server *Server) {
		for {
			time.Sleep(10 * time.Second)
			s := time.Now()
			if err := exec.Command("echo").Run(); err != nil {
				server.Logger.Warn("exec measurement failed", log.Error(err))
				continue
			}
			echoDuration.Set(time.Since(s).Seconds())
		}
	}(s)

	// report the size of the repos dir
	if s.ReposDir == "" {
		s.Logger.Error("ReposDir is not set, cannot export disk_space_available and gitserver_mount_info metric.")
		return
	}

	opts := mountinfo.CollectorOpts{Namespace: "gitserver"}
	m := mountinfo.NewCollector(s.Logger, opts, map[string]string{"reposDir": s.ReposDir})
	observationCtx.Registerer.MustRegister(m)

	metrics.MustRegisterDiskMonitor(s.ReposDir)

	// TODO(keegan) these are older names for the above disk metric. Keeping
	// them to prevent breaking dashboards. Can remove once no
	// alert/dashboards use them.
	c := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_gitserver_disk_space_available",
		Help: "Amount of free space disk space on the repos mount.",
	}, func() float64 {
		var stat syscall.Statfs_t
		_ = syscall.Statfs(s.ReposDir, &stat)
		return float64(stat.Bavail * uint64(stat.Bsize))
	})
	prometheus.MustRegister(c)

	c = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_gitserver_disk_space_total",
		Help: "Amount of total disk space in the repos directory.",
	}, func() float64 {
		var stat syscall.Statfs_t
		_ = syscall.Statfs(s.ReposDir, &stat)
		return float64(stat.Blocks * uint64(stat.Bsize))
	})
	prometheus.MustRegister(c)

	// Register uniform observability via internal/observation
	s.operations = newOperations(observationCtx)
}
