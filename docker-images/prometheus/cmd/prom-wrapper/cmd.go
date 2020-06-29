package main

import (
	"fmt"
	"os"
	"os/exec"
)

func runCmd(errs chan<- error, cmd *exec.Cmd) {
	errs <- cmd.Run()
}

// NewPrometheusCmd instantiates a new command to run Prometheus.
// Parameter promArgs replicates "$@"
func NewPrometheusCmd(promArgs []string, promPort string) *exec.Cmd {
	promFlags := []string{
		fmt.Sprintf("--web.listen-address=\"0.0.0.0:%s\"", promPort),
	}
	cmd := exec.Command("/prometheus.sh", append(promFlags, promArgs...)...)
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd
}

func NewAlertmanagerCmd(configPath string) *exec.Cmd {
	cmd := exec.Command("/bin/alertmanager",
		fmt.Sprintf("--config.file=%s", configPath),
		"--storage.path=/alertmanager")
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd
}
