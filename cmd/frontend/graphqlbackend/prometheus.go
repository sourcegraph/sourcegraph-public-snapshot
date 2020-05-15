package graphqlbackend

import (
	"context"

	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

var prometheusURL = env.Get("PROMETHEUS_SERVER_URL", "http://prometheus:9090", "prometheus server URL")

//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend -i prometheusQuerier -o prometheus_mock.go

// prometheusQuerier provides a shim around prometheus.API
type prometheusQuerier interface {
	// QueryRange performs a query for the given range.
	QueryRange(ctx context.Context, query string, r prometheus.Range) (model.Value, prometheus.Warnings, error)
}
