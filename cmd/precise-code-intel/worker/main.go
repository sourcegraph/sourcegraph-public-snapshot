package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goreman"
)

var (
	workers = env.Get("NUM_WORKERS", "2", "the number of worker instances to run (defaults to one)")

	// Set in docker image
	prometheusStorageDir       = os.Getenv("PROMETHEUS_STORAGE_DIR")
	prometheusConfigurationDir = os.Getenv("PROMETHEUS_CONFIGURATION_DIR")
)

const workerPort = 3188

func main() {
	numWorkers, err := strconv.ParseInt(workers, 10, 64)
	if err != nil || numWorkers < 0 {
		log.Fatalf("Invalid int %q for NUM_WORKERS: %s", workers, err)
	}

	if err := ioutil.WriteFile(
		filepath.Join(prometheusConfigurationDir, "targets.yml"),
		[]byte(makePrometheusTargets(numWorkers)),
		0644,
	); err != nil {
		log.Fatalf("Writing prometheus config: %v", err)
	}

	// This mirrors the behavior from cmd/start
	if err := goreman.Start([]byte(makeProcfile(numWorkers)), goreman.Options{
		RPCAddr:        "127.0.0.1:5005",
		ProcDiedAction: goreman.Shutdown,
	}); err != nil {
		log.Fatalf("Starting goreman: %v", err)
	}
}

func makeProcfile(numWorkers int64) string {
	procfile := []string{}
	addProcess := func(name, command string) {
		procfile = append(procfile, fmt.Sprintf("%s: %s", name, command))
	}

	for i := 0; i < int(numWorkers); i++ {
		addProcess(
			fmt.Sprintf("worker-%d", i),
			fmt.Sprintf("env METRICS_PORT=%d node /precise-code-intel/out/worker/worker.js", workerPort+i),
		)
	}

	addProcess("prometheus", fmt.Sprintf("prometheus '--storage.tsdb.path=%s' '--config.file=%s/prometheus.yml'",
		prometheusStorageDir,
		prometheusConfigurationDir,
	))

	return strings.Join(procfile, "\n") + "\n"
}

func makePrometheusTargets(numWorkers int64) string {
	content := []string{"---"}
	addTarget := func(job string, port int) {
		content = append(content,
			"- labels:",
			fmt.Sprintf("    job: %s", job),
			"  targets:",
			fmt.Sprintf("    - 127.0.0.1:%d", port),
		)
	}

	for i := 0; i < int(numWorkers); i++ {
		addTarget("worker", workerPort+i)
	}

	return strings.Join(content, "\n") + "\n"
}
