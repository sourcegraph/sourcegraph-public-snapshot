package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/grafana-tools/sdk"
	"github.com/inconshreveable/log15"
	"gopkg.in/ini.v1"
)

const overviewDashboardUID = "overview"

var grafanaConfigPath = os.Getenv("GF_PATHS_CONFIG")

// newGrafanaRunCmd instantiates a new command to run grafana.
func newGrafanaRunCmd() *exec.Cmd {
	cmd := exec.Command("/run.sh")
	cmd.Env = os.Environ() // propagate env to grafana
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd
}

type grafanaController struct {
	log log15.Logger
	*sdk.Client

	mux  sync.Mutex
	proc *os.Process
}

func newGrafanaController(log log15.Logger, grafanaPort, grafanaCredentials string) *grafanaController {
	return &grafanaController{
		Client: sdk.NewClient(fmt.Sprintf("http://127.0.0.1:%s", grafanaPort), grafanaCredentials, http.DefaultClient),
		log:    log.New("logger", "grafana-ctrl"),
	}
}

func (c *grafanaController) RunServer() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	// spin up grafana and track process
	c.log.Debug("starting Grafana server")
	cmd := newGrafanaRunCmd()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Grafana: %w", err)
	}
	c.proc = cmd.Process

	// capture results from grafana process
	go func() {
		// cmd.Wait output:
		// * exits with status 0 => nil
		// * command fails to run or stopped => *ExitErr
		// * other IO error => error
		if err := cmd.Wait(); err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				exitCode := exitErr.ProcessState.ExitCode()
				// unfortunately grafana exits with code 1 on sigint
				if exitCode > 1 {
					c.log.Crit("grafana exited with unexpected code", "exitcode", exitCode)
					os.Exit(exitCode)
				}
				c.log.Info("grafana has stopped", "exitcode", exitCode)
				return
			}
			c.log.Warn("error waiting for grafana to stop", "error", err)
		}
		c.log.Debug("grafana has stopped")
	}()

	return nil
}

func (c *grafanaController) WaitForServer(ctx context.Context) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	ping := func(ctx context.Context) error {
		resp, err := c.Client.GetHealth(ctx)
		if err != nil {
			return err
		}
		if resp.Version == "" || resp.Commit == "" {
			return fmt.Errorf("ping: malformed health response: %+v", resp)
		}
		return nil
	}

	var lastErr error
	for {
		err := ping(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return fmt.Errorf("grafana not reachable: %s (last error: %v)", err, lastErr)
			}

			// Keep trying.
			lastErr = err
			time.Sleep(250 * time.Millisecond)
			continue
		}
		break
	}
	return nil
}

func (c *grafanaController) Stop() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.proc != nil {
		c.log.Debug("stopping current Grafana instance", "pid", c.proc.Pid)
		if err := c.proc.Signal(os.Interrupt); err != nil {
			return fmt.Errorf("failed to stop Grafana instance: %w", err)
		}
		_ = c.proc.Wait() // this can error for a variety of irrelvant reasons
		if err := c.proc.Release(); err != nil {
			c.log.Warn("failed to release process", "error", err)
		}
		c.proc = nil
	} else {
		c.log.Debug("already stopped")
	}
	return nil
}

// getOverviewDashboard parses the default home.json and returns it.
func getOverviewDashboard() (*sdk.Board, error) {
	data, err := ioutil.ReadFile("/usr/share/grafana/public/dashboards/home.json")
	if err != nil {
		return nil, err
	}
	var board sdk.Board
	if err := json.Unmarshal(data, &board); err != nil {
		return nil, err
	}
	board.ID = 0
	board.UID = overviewDashboardUID
	return &board, nil
}

func getGrafanaConfig(source interface{}) (*ini.File, error) {
	grafanaConfig, err := ini.Load(source)
	if err != nil {
		return nil, fmt.Errorf("failed to load Grafana config: %w", err)
	}
	// set pattern to automatically convert struct fields to Grafana's INI casing
	// see https://grafana.com/docs/grafana/latest/installation/configuration
	grafanaConfig.NameMapper = ini.TitleUnderscore
	return grafanaConfig, nil
}
