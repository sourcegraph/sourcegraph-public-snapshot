package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/grafana-tools/sdk"
)

const overviewDashboardUID = "overview"

// newGrafanaRunCmd instantiates a new command to run grafana.
func newGrafanaRunCmd() *exec.Cmd {
	cmd := exec.Command("/grafana-run.sh")
	cmd.Env = os.Environ() // propagate env to grafana
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd
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

func waitForGrafana(ctx context.Context, grafana *sdk.Client) error {
	ping := func(ctx context.Context) error {
		resp, err := grafana.GetHealth(ctx)
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
