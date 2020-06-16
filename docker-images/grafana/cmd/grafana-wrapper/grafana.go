package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/grafana-tools/sdk"
	"github.com/inconshreveable/log15"
)

const overviewDashboardUID = "overview"

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

	// stop existing process if there is one
	if c.proc != nil {
		if err := c.proc.Kill(); err != nil {
			return err
		}
		_, err := c.proc.Wait()
		if err != nil {
			return err
		}
		if err := c.proc.Release(); err != nil {
			c.log.Warn("failed to release proccess", "error", err)
		}
		c.proc = nil
	}

	// spin up grafana and track process
	cmd := newGrafanaRunCmd()
	if err := cmd.Start(); err != nil {
		return err
	}
	c.proc = cmd.Process

	// capture errors from grafana starting
	go func() {
		if err := cmd.Wait(); err != nil {
			// TODO what do
			c.log.Crit("grafana exited unexpectedly", "error", err)
		}
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
