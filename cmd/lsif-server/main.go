package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/goreman"
)

func main() {
	targets, ok := os.LookupEnv("LSIF_RUN_TARGET")
	if !ok {
		targets = "server,worker"
	}

	procfile := []string{}
	for _, target := range strings.Split(targets, ",") {
		switch target {
		case "server":
			procfile = append(procfile, `lsif-server: node /lsif/out/server/server.js`)
		case "worker":
			procfile = append(procfile, `lsif-worker: node /lsif/out/worker/worker.js`)
		default:
			log.Fatalf("Unknown value '%s' for LSIF_RUN_TARGET (expected 'server' or 'worker')", target)
		}
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
