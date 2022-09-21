//go:generate go build -o /tmp/monitoring-generator
//go:generate /tmp/monitoring-generator
package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/monitoring/definitions"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func optsFromEnv(logger log.Logger) monitoring.GenerateOptions {
	return monitoring.GenerateOptions{
		DisablePrune: getEnvBool("NO_PRUNE", false),
		Reload:       getEnvBool("RELOAD", false),

		GrafanaDir:    getEnvStr("GRAFANA_DIR", "../docker-images/grafana/config/provisioning/dashboards/sourcegraph/"),
		PrometheusDir: getEnvStr("PROMETHEUS_DIR", "../docker-images/prometheus/config/"),
		DocsDir:       getEnvStr("DOCS_DIR", "../doc/admin/observability/"),

		InjectLabelMatchers: func() []*labels.Matcher {
			matcherEntries := strings.Split(getEnvStr("INJECT_LABEL_MATCHERS", ""), ",")
			if len(matcherEntries) == 0 {
				return nil
			}
			var matchers []*labels.Matcher
			for _, entry := range matcherEntries {
				if len(entry) == 0 {
					continue
				}
				parts := strings.Split(entry, "=")
				if len(parts) != 2 {
					logger.Error("discarding invalid INJECT_LABEL_MATCHERS entry",
						log.String("entry", entry))
					continue
				}

				label := parts[0]
				value := parts[1]
				matcher, err := labels.NewMatcher(labels.MatchEqual, label, value)
				if err != nil {
					logger.Error("discarding invalid INJECT_LABEL_MATCHERS entry",
						log.String("entry", entry),
						log.Error(err))
					continue
				}
				matchers = append(matchers, matcher)
			}
			return matchers
		}(),
	}
}

func main() {
	// Configure logger
	if _, set := os.LookupEnv(log.EnvDevelopment); !set {
		os.Setenv(log.EnvDevelopment, "true")
	}
	if _, set := os.LookupEnv(log.EnvLogFormat); !set {
		os.Setenv(log.EnvLogFormat, "console")
	}

	liblog := log.Init(log.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	})
	defer liblog.Sync()

	logger := log.Scoped("monitoring-generator", "generates monitoring dashboards")

	// Runs the monitoring generator. Ensure that any dashboards created or removed are
	// updated in the arguments here as required.
	if err := monitoring.Generate(logger, optsFromEnv(logger.Scoped("opts", "options builder")),
		definitions.Frontend(),
		definitions.GitServer(),
		definitions.GitHubProxy(),
		definitions.Postgres(),
		definitions.PreciseCodeIntelWorker(),
		definitions.Redis(),
		definitions.Worker(),
		definitions.RepoUpdater(),
		definitions.Searcher(),
		definitions.Symbols(),
		definitions.SyntectServer(),
		definitions.Zoekt(),
		definitions.Prometheus(),
		definitions.Executor(),
		definitions.Containers(),
		definitions.CodeIntelAutoIndexing(),
		definitions.CodeIntelUploads(),
		definitions.CodeIntelPolicies(),
		definitions.Telemetry(),
	); err != nil {
		// Dump error as plain output for readability.
		println(err.Error())

		logger.Fatal("Error encountered")
	}
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
