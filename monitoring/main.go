//go:generate go build -o /tmp/monitoring-generator
//go:generate /tmp/monitoring-generator
package main

import "github.com/sourcegraph/sourcegraph/monitoring/monitoring"

func main() {
	monitoring.Generate([]*monitoring.Container{
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
	})
}
