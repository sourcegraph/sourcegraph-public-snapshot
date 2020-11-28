//go:generate go build -o /tmp/monitoring-generator
//go:generate /tmp/monitoring-generator
package main

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func main() {
	// Runs the monitoring generator. Ensure that any dashboards created or removed are
	// updated in the arguments here as required.
	monitoring.Generate(
		definitions.Frontend(),
		definitions.GitServer(),
		definitions.GitHubProxy(),
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
