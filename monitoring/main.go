//go:generate go build -o /tmp/monitoring-generator
//go:generate /tmp/monitoring-generator
package main

import (
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func optsFromEnv() monitoring.GenerateOptions {
	return monitoring.GenerateOptions{
		DisablePrune: getEnvBool("NO_PRUNE", false),
		LiveReload:   getEnvBool("RELOAD", false),

		GrafanaDir:    getEnvStr("GRAFANA_DIR", "../docker-images/grafana/config/provisioning/dashboards/sourcegraph/"),
		PrometheusDir: getEnvStr("PROMETHEUS_DIR", "../docker-images/prometheus/config/"),
		DocsDir:       getEnvStr("DOCS_DIR", "../doc/admin/observability/"),
	}
}

func main() {
	// Runs the monitoring generator. Ensure that any dashboards created or removed are
	// updated in the arguments here as required.
	monitoring.Generate(optsFromEnv(),
		definitions.Frontend(),
		definitions.GitServer(),
		definitions.GitHubProxy(),
		definitions.Postgres(),
		definitions.PreciseCodeIntelWorker(),
		definitions.QueryRunner(),
		definitions.RepoUpdater(),
		definitions.Searcher(),
		definitions.Symbols(),
		definitions.SyntectServer(),
		definitions.ZoektIndexServer(),
		definitions.ZoektWebServer(),
		definitions.Prometheus(),
		definitions.ExecutorAndExecutorQueue(),
	)
}

func getEnvStr(key, defaultValue string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			panic(err)
		}
		return b
	}
	return defaultValue
}
