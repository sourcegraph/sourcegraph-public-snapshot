package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/goreman"
)

func main() {
	procfile := []string{
		`lsif-server: node /lsif/out/server/server.js`,
		`lsif-worker: node /lsif/out/worker/worker.js`,
	}

	// Shutdown if any process dies
	procDiedAction := goreman.Shutdown
	if ignore, _ := strconv.ParseBool(os.Getenv("IGNORE_PROCESS_DEATH")); ignore {
		// IGNORE_PROCESS_DEATH is an escape hatch so that sourcegraph/lsif-server
		// keeps running in the case of a subprocess dieing on startup. An example
		// use case is connecting to postgres even though frontend is dieing due
		// to a bad migration.
		procDiedAction = goreman.Ignore
	}

	err := goreman.Start([]byte(strings.Join(procfile, "\n")), goreman.Options{
		RPCAddr:        "127.0.0.1:5005",
		ProcDiedAction: procDiedAction,
	})
	if err != nil {
		log.Fatal(err)
	}
}
