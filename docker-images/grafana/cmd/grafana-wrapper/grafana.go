package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/inconshreveable/log15"
)

// newGrafanaRunCmd instantiates a new command to run grafana.
// note that exec.Cmd can not be reused - instead, after calling cmd.Start()
// maintain a reference to cmd.Process to control it.
func newGrafanaRunCmd() *exec.Cmd {
	cmd := exec.Command("/run.sh")
	cmd.Env = os.Environ() // propagate env to grafana
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd
}

type grafanaController struct {
	mux     sync.Mutex
	process *os.Process
	log     log15.Logger
}

func newGrafanaController(log log15.Logger) *grafanaController {
	return &grafanaController{
		log: log.New("logger", "grafana-controller"),
	}
}

func (c *grafanaController) Restart() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	// if grafana is running, try to stop it. log and swallow errors in while doing so,
	// since regardless of what happens we should still attempt to restart grafana.
	if c.process != nil {
		c.log.Debug("stopping grafana")

		// signal process to be killed and if able to do so, wait for process to stop
		if err := c.process.Kill(); err != nil {
			c.log.Error("error occurred while killing grafana", "error", err)
		} else if _, err := c.process.Wait(); err != nil {
			c.log.Error("error occurred while waiting for grafana to stop", "error", err)
		}

		// reset process
		if err := c.process.Release(); err != nil {
			c.log.Warn("error occurred while releasing process", "error", err)
		}
		c.process = nil
	} else {
		c.log.Debug("grafana is not running")
	}

	// spin up grafana
	c.log.Debug("starting grafana")
	cmd := newGrafanaRunCmd()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start grafana: %w", err)
	}
	c.process = cmd.Process
	return nil
}
