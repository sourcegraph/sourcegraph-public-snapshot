module github.com/sourcegraph/sourcegraph/observability

go 1.14

require (
	github.com/gosimple/slug v1.9.0 // indirect
	github.com/grafana-tools/sdk v0.0.0-20200326200416-f0431e44c1c3
	gopkg.in/yaml.v2 v2.2.8
)

// https://github.com/grafana-tools/sdk/pull/80
replace github.com/grafana-tools/sdk => github.com/slimsag/sdk v0.0.0-20200402190125-fc52c0aed0b7
