//go:generate go build -o /tmp/monitoring-generator
//go:generate /tmp/monitoring-generator
package main

import "github.com/sourcegraph/sourcegraph/monitoring/monitoring"

func main() {
	// Runs the monitoring generator. Ensure that any dashboards created or removed are
	// updated in the arguments here as required.
	monitoring.Generate(
		Frontend(),
		GitServer(),
		GitHubProxy(),
		PreciseCodeIntelWorker(),
		QueryRunner(),
		RepoUpdater(),
		Searcher(),
		Symbols(),
		SyntectServer(),
		ZoektIndexServer(),
		ZoektWebServer(),
		Prometheus(),
		ExecutorAndExecutorQueue(),
	)
}
