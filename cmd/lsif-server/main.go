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
	dumpManagers   = env.Get("LSIF_NUM_DUMP_MANAGERS", "1", "the number of dump manager instances to run (defaults to one)")
	dumpProcessors = env.Get("LSIF_NUM_DUMP_PROCESSORS", "1", "the number of dump processor instances to run (defaults to one)")

	// Set in docker image
	prometheusStorageDir       = os.Getenv("PROMETHEUS_STORAGE_DIR")
	prometheusConfigurationDir = os.Getenv("PROMETHEUS_CONFIGURATION_DIR")
)

const (
	apiPort           = 3186
	dumpManagerPort   = 3187
	dumpProcessorPort = 3188
)

func main() {
	numAPIs, err := strconv.ParseInt(apis, 10, 64)
	if err != nil || numAPIs < 0 || numAPIs > 1 {
		log.Fatalf("Invalid int %q for LSIF_NUM_APIS: %s", apis, err)
	}

	numDumpManagers, err := strconv.ParseInt(dumpManagers, 10, 64)
	if err != nil || numDumpManagers < 0 || numDumpManagers > 1 {
		log.Fatalf("Invalid int %q for LSIF_NUM_DUMP_MANAGERS: %s", dumpManagers, err)
	}

	numDumpProcessors, err := strconv.ParseInt(dumpProcessors, 10, 64)
	if err != nil || numDumpProcessors < 0 {
		log.Fatalf("Invalid int %q for LSIF_NUM_DUMP_PROCESSORS: %s", numDumpProcessors, err)
	}

	if err := ioutil.WriteFile(
		filepath.Join(prometheusConfigurationDir, "targets.yml"),
		[]byte(makePrometheusTargets(numAPIs, numDumpManagers, numDumpProcessors)),
		0644,
	); err != nil {
		log.Fatalf("Writing prometheus config: %v", err)
	}

	// This mirrors the behavior from cmd/start
	if err := goreman.Start([]byte(makeProcfile(numAPIs, numDumpManagers, numDumpProcessors)), goreman.Options{
		RPCAddr:        "127.0.0.1:5005",
		ProcDiedAction: goreman.Shutdown,
	}); err != nil {
		log.Fatalf("Starting goreman: %v", err)
	}
}

func makeProcfile(numAPIs, numDumpManagers, numDumpProcessors int64) string {
	procfile := []string{}
	addProcess := func(name, command string) {
		procfile = append(procfile, fmt.Sprintf("%s: %s", name, command))
	}

	if numAPIs > 0 {
		addProcess("lsif-api", "node /lsif/out/api/api.js")
	}

	if numDumpManagers > 0 {
		addProcess("lsif-dump-manager", "node /lsif/out/dump-manager/dump-manager.js")
	}

	for i := 0; i < int(numDumpProcessors); i++ {
		addProcess(
			fmt.Sprintf("lsif-dump-processor-%d", i),
			fmt.Sprintf("env METRICS_PORT=%d node /lsif/out/dump-processor/dump-processor.js", dumpProcessorPort+i),
		)
	}

	addProcess("prometheus", fmt.Sprintf("prometheus '--storage.tsdb.path=%s' '--config.file=%s/prometheus.yml'",
		prometheusStorageDir,
		prometheusConfigurationDir,
	))

	return strings.Join(procfile, "\n") + "\n"
}

func makePrometheusTargets(numAPIs, numDumpManagers, numDumpProcessors int64) string {
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
		addTarget("lsif-api", apiPort)
	}

	if numDumpManagers > 0 {
		addTarget("lsif-dump-manager", dumpManagerPort)
	}

	for i := 0; i < int(numDumpProcessors); i++ {
		addTarget("lsif-dump-processor", dumpProcessorPort+i)
	}

	return strings.Join(content, "\n") + "\n"
}
