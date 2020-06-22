package shared

import (
	"os"

	"github.com/inconshreveable/log15"
)

const prometheusProcLine = `prometheus: prometheus --config.file=/sg_config_prometheus/prometheus.yml --web.enable-admin-api --storage.tsdb.path=/var/opt/sourcegraph/prometheus --web.console.libraries=/usr/share/prometheus/console_libraries --web.console.templates=/usr/share/prometheus/consoles >> /var/opt/sourcegraph/prometheus.log 2>&1`

const grafanaProcLine = `grafana: sh -c 'env PATH=/usr/share/grafana/bin:${PATH} GF_PATHS_HOME=/usr/share/grafana GF_PATHS_DATA=/var/opt/sourcegraph/grafana GF_PATHS_LOGS=/var/opt/sourcegraph/grafana/logs GF_PATHS_PLUGINS=/var/opt/sourcegraph/grafana/plugins /grafana-entry.sh >> /var/opt/sourcegraph/grafana.log 2>&1'`

const jaegerProcLine = `jaeger --memory.max-traces=20000 >> /var/opt/sourcegraph/jaeger.log 2>&1`

func maybeMonitoring() ([]string, error) {
	if os.Getenv("DISABLE_OBSERVABILITY") != "" {
		log15.Info("WARNING: Running with monitoring disabled")
		return []string{""}, nil
	}
	return []string{
		prometheusProcLine,
		grafanaProcLine,
		jaegerProcLine}, nil
}
