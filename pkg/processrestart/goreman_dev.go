package processrestart

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	pkgerrors "github.com/pkg/errors"
)

// usingGoremanDev is whether we are running goreman in dev/launch.sh
var usingGoremanDev = os.Getenv("GOREMAN") != ""

// restartGoremanDev restarts the processes when running goreman in dev/launch.sh. It takes care to
// avoid a race condition where some services have started up with the new config and some are still
// running with the old config.
func restartGoremanDev() error {
	goreman := os.Getenv("GOREMAN")
	if goreman == "" {
		return errors.New("unable to reload site")
	}

	// Should be kept in sync with Procfile.
	allProcessesExceptFrontend := []string{
		"gitserver",
		"indexer",
		"query-runner",
		"repo-updater",
		"searcher",
		"symbols",
		"github-proxy",
		"lsp-proxy",
		"xlang-go",
		"syntect_server",
		"zoekt-indexserver",
		"zoekt-webserver",
		// frontend is restarted separately last
	}

	runCommand := func(command string, processes ...string) error {
		var (
			wg  sync.WaitGroup
			mu  sync.Mutex
			err error
		)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		for _, proc := range processes {
			wg.Add(1)
			go func(proc string) {
				defer wg.Done()
				args := append(strings.Fields(goreman), "run", command, proc)
				cmd := exec.CommandContext(ctx, args[0], args[1:]...)
				cmd.Stdout = os.Stderr
				cmd.Stderr = os.Stderr
				if runErr := cmd.Run(); err != nil {
					if ctx.Err() != nil {
						return
					}
					mu.Lock()
					defer mu.Unlock()
					err = pkgerrors.Wrapf(err, "process %s: %s", proc, runErr)
					cancel()
				}
			}(proc)
		}
		wg.Wait()
		if err != nil {
			return fmt.Errorf("failed to run %q command on all processes: %s", command, err)
		}
		return nil
	}

	if err := runCommand("stop", allProcessesExceptFrontend...); err != nil {
		return err
	}

	// Make the frontend process unreachable from the other processes because they'll try to read
	// config/data from frontend (us), and until frontend restarts, it has the pre-restart config.
	close(WillRestart)
	time.Sleep(100 * time.Millisecond)

	// Start all other processes. If they need to communicate with frontend (us), they must be
	// designed to wait for a short period until frontend is reachable again (after frontend is
	// restarted right below).
	if err := runCommand("start", allProcessesExceptFrontend...); err != nil {
		return err
	}

	return runCommand("restart", "frontend")
}
