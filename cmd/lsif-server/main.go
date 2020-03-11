package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goreman"
)

var (
	servers = env.Get("LSIF_NUM_SERVERS", "1", "the number of server instances to run (defaults to one)")
	workers = env.Get("LSIF_NUM_WORKERS", "1", "the number of worker instances to run (defaults to one)")
)

func main() {
	procfile, err := makeProcfile()
	if err != nil {
		log.Fatalf(err.Error())
	}

	// This mirrors the behavior from cmd/start
	if err := goreman.Start([]byte(strings.Join(procfile, "\n")), goreman.Options{
		RPCAddr:        "127.0.0.1:5005",
		ProcDiedAction: goreman.Shutdown,
	}); err != nil {
		log.Fatalf(err.Error())
	}
}

func makeProcfile() ([]string, error) {
	numServers, err := strconv.ParseInt(servers, 10, 64)
	if err != nil || numServers < 0 || numServers > 1 {
		return nil, fmt.Errorf("invalid int %q for LSIF_NUM_SERVERS: %s", servers, err)
	}

	numWorkers, err := strconv.ParseInt(workers, 10, 64)
	if err != nil || numWorkers < 0 {
		return nil, fmt.Errorf("invalid int %q for LSIF_NUM_WORKERS: %s", workers, err)
	}

	procfile := []string{}

	for i := int64(0); i < numServers; i++ {
		procfile = append(procfile, fmt.Sprintf(`lsif-server-%d: node /lsif/out/server/server.js`, i))
	}

	for i := int64(0); i < numWorkers; i++ {
		procfile = append(procfile, fmt.Sprintf(`lsif-worker-%d: env WORKER_METRICS_PORT=%d node /lsif/out/worker/worker.js`, i, 3187+i))
	}

	return procfile, nil
}
