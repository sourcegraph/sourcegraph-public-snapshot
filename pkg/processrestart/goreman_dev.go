package processrestart

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"

	pkgerrors "github.com/pkg/errors"
)

// usingGoremanDev is whether we are running goreman in dev/start.sh
var usingGoremanDev = os.Getenv("GOREMAN") != ""

// restartGoremanDev restarts the processes when running goreman in
// dev/start.sh
func restartGoremanDev() error {
	goreman := os.Getenv("GOREMAN")
	if goreman == "" {
		return errors.New("unable to reload site")
	}

	// Should be kept in sync with Procfile.
	toRestart := []string{
		"gitserver",
		"indexer",
		"repo-updater",
		"searcher",
		"github-proxy",
		"lsp-proxy",
		"xlang-go",
		"syntect_server",
		"zoekt-indexserver",
		"zoekt-webserver",
		"frontend",
	}

	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		err error
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, proc := range toRestart {
		wg.Add(1)
		go func(proc string) {
			defer wg.Done()
			args := append(strings.Fields(goreman), "run", "restart", proc)
			cmd := exec.CommandContext(ctx, args[0], args[1:]...)
			cmd.Stdout = os.Stderr
			cmd.Stderr = os.Stderr
			if runErr := cmd.Run(); err != nil {
				if ctx.Err() != nil {
					return
				}
				mu.Lock()
				defer mu.Unlock()
				err = pkgerrors.Wrapf(err, "failed to restart %s: %s", proc, runErr)
				cancel()
			}
		}(proc)
	}
	wg.Wait()

	return err
}
