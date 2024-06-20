package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sourcegraph/log"
	"go.bobheadxi.dev/streamline/streamexec"
)

func runCmd(logger log.Logger, errs chan<- error, cmd *exec.Cmd) {
	commandLog := logger.With(log.Strings("cmd", append([]string{cmd.Path}, cmd.Args...)))
	commandLog.Info("running cmd")
	s, err := streamexec.Start(cmd, streamexec.Combined|streamexec.ErrWithStderr)
	if err != nil {
		commandLog.Error("command exited", log.Error(err))
		errs <- err
		return
	}
	if err := s.Stream(func(line string) {
		switch {
		case strings.Contains(line, "level=warn"):
			logger.Warn(line)
		case strings.Contains(line, "level=error"):
			logger.Error(line)
		default:
			logger.Info(line)
		}
	}); err != nil {
		commandLog.Error("command exited", log.Error(err))
		errs <- err
		return
	}
}

// NewPrometheusCmd instantiates a new command to run Prometheus.
// Parameter promArgs replicates "$@"
func NewPrometheusCmd(promArgs []string, promPort string) *exec.Cmd {
	promFlags := []string{
		fmt.Sprintf("--web.listen-address=0.0.0.0:%s", promPort),
	}
	cmd := exec.Command("/prometheus.sh", append(promFlags, promArgs...)...)
	cmd.Env = os.Environ()
	return cmd
}

func NewAlertmanagerCmd(configPath string) *exec.Cmd {
	cmd := exec.Command("/alertmanager.sh",
		fmt.Sprintf("--config.file=%s", configPath),
		fmt.Sprintf("--web.route-prefix=/%s", alertmanagerPathPrefix))
	// disable clustering unless otherwise configured - it is enabled by default, but
	// can cause alertmanager to fail to start up in some environments: https://github.com/sourcegraph/sourcegraph/issues/13079
	if alertmanagerEnableCluster != "true" {
		cmd.Args = append(cmd.Args, "--cluster.listen-address=")
	}
	cmd.Env = os.Environ()
	return cmd
}
