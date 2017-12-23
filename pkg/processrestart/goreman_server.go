package processrestart

import (
	"net/rpc"
	"os"
	"sync"

	pkgerrors "github.com/pkg/errors"
)

// usingGoremanServer is whether we are running goreman in cmd/server.
var usingGoremanServer = os.Getenv("GOREMAN_RPC_ADDR") != ""

// restartGoremanServer restarts the processes when running goreman in
// dev/start.sh
func restartGoremanServer() error {
	// Should be kept in sync with cmd/server's process list.
	toRestart := []string{
		"frontend",
		"gitserver",
		"indexer",
		"repo-updater",
		"searcher",
		"github-proxy",
		"syntect_server",
	}

	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		err error
	)
	for _, proc := range toRestart {
		wg.Add(1)
		go func(proc string) {
			defer wg.Done()
			if callErr := goremanServerRPCRestartProcess(proc); err != nil {
				mu.Lock()
				defer mu.Unlock()
				err = pkgerrors.Wrapf(err, "failed to restart %s: %s", proc, callErr)
			}
		}(proc)
	}
	wg.Wait()

	return err
}

func goremanServerRPCRestartProcess(proc string) error {
	client, err := rpc.Dial("tcp", os.Getenv("GOREMAN_RPC_ADDR"))
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Call("Goreman.Restart", proc, nil)
}
