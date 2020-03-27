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
	apis           = env.Get("LSIF_NUM_APIS", "1", "the number of API instances to run (defaults to one)")
	bundleManagers = env.Get("LSIF_NUM_BUNDLE_MANAGERS", "1", "the number of bundle manager instances to run (defaults to one)")
	workers        = env.Get("LSIF_NUM_WORKERS", "1", "the number of worker instances to run (defaults to one)")

	// Set in docker image
	prometheusStorageDir       = os.Getenv("PROMETHEUS_STORAGE_DIR")
	prometheusConfigurationDir = os.Getenv("PROMETHEUS_CONFIGURATION_DIR")
)

const (
	apiPort           = 3186
	bundleManagerPort = 3187
	workerPort        = 3188
)

func main() {
	numAPIs, err := strconv.ParseInt(apis, 10, 64)
	if err != nil || numAPIs < 0 || numAPIs > 1 {
		log.Fatalf("Invalid int %q for LSIF_NUM_APIS: %s", apis, err)
	}

	numBundleManagers, err := strconv.ParseInt(bundleManagers, 10, 64)
	if err != nil || numBundleManagers < 0 || numBundleManagers > 1 {
		log.Fatalf("Invalid int %q for LSIF_NUM_BUNDLE_MANAGERS: %s", bundleManagers, err)
	}

	numWorkers, err := strconv.ParseInt(workers, 10, 64)
	if err != nil || numWorkers < 0 {
		log.Fatalf("Invalid int %q for LSIF_NUM_WORKERS: %s", workers, err)
	}

	if err := ioutil.WriteFile(
		filepath.Join(prometheusConfigurationDir, "targets.yml"),
		[]byte(makePrometheusTargets(numAPIs, numBundleManagers, numWorkers)),
		0644,
	); err != nil {
		log.Fatalf("Writing prometheus config: %v", err)
	}

	// This mirrors the behavior from cmd/start
	if err := goreman.Start([]byte(makeProcfile(numAPIs, numBundleManagers, numWorkers)), goreman.Options{
		RPCAddr:        "127.0.0.1:5005",
		ProcDiedAction: goreman.Shutdown,
	}); err != nil {
		log.Fatalf("Starting goreman: %v", err)
	}
}

func makeProcfile(numAPIs, numBundleManagers, numWorkers int64) string {
	procfile := []string{}
	addProcess := func(name, command string) {
		procfile = append(procfile, fmt.Sprintf("%s: %s", name, command))
	}

	if numAPIs > 0 {
		addProcess("lsif-api-server", "node /lsif/out/api-server/api.js")
	}

	if numBundleManagers > 0 {
		addProcess("lsif-bundle-manager", "node /lsif/out/bundle-manager/manager.js")
	}

	for i := 0; i < int(numWorkers); i++ {
		addProcess(
			fmt.Sprintf("lsif-worker-%d", i),
			fmt.Sprintf("env METRICS_PORT=%d node /lsif/out/worker/worker.js", workerPort+i),
		)
	}

	addProcess("prometheus", fmt.Sprintf("prometheus '--storage.tsdb.path=%s' '--config.file=%s/prometheus.yml'",
		prometheusStorageDir,
		prometheusConfigurationDir,
	))

	return strings.Join(procfile, "\n") + "\n"
}

func makePrometheusTargets(numAPIs, numBundleManagers, numWorkers int64) string {
	content := []string{"---"}
	addTarget := func(job string, port int) {
		content = append(content,
			"- labels:",
			fmt.Sprintf("    job: %s", job),
			"  targets:",
			fmt.Sprintf("    - 127.0.0.1:%d", port),
		)
	}

	if numAPIs > 0 {
		addTarget("lsif-api-server", apiPort)
	}

	if numBundleManagers > 0 {
		addTarget("lsif-bundle-manager", bundleManagerPort)
	}

	for i := 0; i < int(numWorkers); i++ {
		addTarget("lsif-worker", workerPort+i)
	}

	return strings.Join(content, "\n") + "\n"
}
