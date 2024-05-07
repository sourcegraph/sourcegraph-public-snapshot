package internal

import (
	"os/exec"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"
)

func RegisterEchoMetric(logger log.Logger) {
	// This currently doesn't work on windows. Disabling.
	if runtime.GOOS == "windows" {
		return
	}
	// test the latency of exec, which may increase under certain memory
	// conditions
	echoDuration := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "src_gitserver_echo_duration_seconds",
		Help: "Duration of executing the echo command.",
	})
	prometheus.MustRegister(echoDuration)
	go func() {
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
