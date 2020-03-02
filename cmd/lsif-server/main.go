package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/goreman"
)

func main() {
	procfile := []string{}

	serverOnly, _ := strconv.ParseBool(os.Getenv("SERVER_ONLY"))
	workerOnly, _ := strconv.ParseBool(os.Getenv("WORKER_ONLY"))

	if serverOnly && workerOnly {
		log.Fatal("Flags server_only and worker_only are mutually exclusive")
	}

	if !workerOnly {
		procfile = append(procfile, `lsif-server: node /lsif/out/server/server.js`)
	}

	if !serverOnly {
		procfile = append(procfile, `lsif-worker: node /lsif/out/worker/worker.js`)
	}

	// Shutdown if any process dies
	procDiedAction := goreman.Shutdown
	if ignore, _ := strconv.ParseBool(os.Getenv("IGNORE_PROCESS_DEATH")); ignore {
		// IGNORE_PROCESS_DEATH is an escape hatch that matches the same behavior
		// as in sourcegraph/server.
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
