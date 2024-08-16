package config

import (
	"embed"
)

var (
	//go:embed otel/*
	//go:embed postgres/*
	//go:embed prometheus/default.yml.gotmpl
	//go:embed grafana/default.yml.gotmpl
	fs embed.FS

	PgsqlConfig                     []byte
	PrometheusDefaultConfigTemplate []byte
	GrafanaDefaultConfigTemplate    []byte
	CodeIntelConfig                 []byte
	CodeInsightsConfig              []byte
	OtelAgentConfig                 []byte
	OtelCollectorConfigTemplate     []byte
)

func init() {
	CodeIntelConfig = mustReadFile("postgres/codeintel.conf")
	CodeInsightsConfig = mustReadFile("postgres/codeinsights.conf")
	PgsqlConfig = mustReadFile("postgres/pgsql.conf")
	PrometheusDefaultConfigTemplate = mustReadFile("prometheus/default.yml.gotmpl")
	GrafanaDefaultConfigTemplate = mustReadFile("grafana/default.yml.gotmpl")
	OtelAgentConfig = mustReadFile("otel/agent.yaml")
	OtelCollectorConfigTemplate = mustReadFile("otel/collector.yaml.gotmpl")
}

func mustReadFile(name string) []byte {
	b, err := fs.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return b
}
