package server

import (
	"os/exec"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/inconshreveable/log15.v2"
)

func (s *Server) RegisterMetrics() {
	// test the latency of exec, which may increase under certain memory
	// conditions
	echoDuration := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "gitserver",
		Name:      "echo_duration_seconds",
		Help:      "Duration of executing the echo command.",
	})
	prometheus.MustRegister(echoDuration)
	go func() {
		for {
			time.Sleep(10 * time.Second)
			s := time.Now()
			if err := exec.Command("echo").Run(); err != nil {
				log15.Warn("exec measurement failed", "error", err)
				continue
			}
			echoDuration.Set(time.Since(s).Seconds())
		}
	}()

	// report the size of the repos dir
	if s.ReposDir == "" {
		log15.Error("ReposDir is not set, cannot export disk_space_available metric.")
		return
	}
	c := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "gitserver",
		Name:      "disk_space_available",
		Help:      "Amount of free space disk space on the repos mount.",
	}, func() float64 {
		var stat syscall.Statfs_t
		_ = syscall.Statfs(s.ReposDir, &stat)
		return float64(stat.Bavail * uint64(stat.Bsize))
	})
	prometheus.MustRegister(c)

	c = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "gitserver",
		Name:      "disk_space_total",
		Help:      "Amount of total disk space in the repos directory.",
	}, func() float64 {
		var stat syscall.Statfs_t
		_ = syscall.Statfs(s.ReposDir, &stat)
		return float64(stat.Blocks * uint64(stat.Bsize))
	})
	prometheus.MustRegister(c)
}
