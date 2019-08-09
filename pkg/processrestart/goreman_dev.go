package processrestart

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	pkgerrors "github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
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
		"query-runner",
		"repo-updater",
		"searcher",
		"symbols",
		"github-proxy",
		"management-console",
		"syntect_server",
		"zoekt-indexserver",
		"zoekt-webserver",
		// frontend is restarted separately last
	}

	runCommand := func(command string, processes ...string) error {
		group, ctx := errgroup.WithContext(context.Background())
		for _, proc := range processes {
			proc := proc
			group.Go(func() error {
				args := append(strings.Fields(goreman), "run", command, proc)
				cmd := exec.CommandContext(ctx, args[0], args[1:]...)
				cmd.Stdout = os.Stderr
				cmd.Stderr = os.Stderr
				if runErr := cmd.Run(); runErr != nil {
					if err := ctx.Err(); err != nil {
						return err
					}
					return pkgerrors.Wrap(runErr, "process "+proc)
				}
				return nil
			})
		}
		if err := group.Wait(); err != nil {
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_859(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
