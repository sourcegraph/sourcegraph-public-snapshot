package metrics

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type metricsSyncPoint struct {
	notify *sync.Cond
	result chan metricsResult
}

func newMetricsSyncPoint() metricsSyncPoint {
	return metricsSyncPoint{
		notify: sync.NewCond(&sync.Mutex{}),
		result: make(chan metricsResult, 1),
	}
}

type metricsResult struct {
	metrics map[string]*dto.MetricFamily
	err     error
}

// MakeExecutorMetricsGatherer uses the given prometheus gatherer to collect all current
// metrics, and optionally also gathers metrics from node exporter and the docker
// registry mirror, if configured.
func MakeExecutorMetricsGatherer(
	logger log.Logger,
	gatherer prometheus.Gatherer,
	// nodeExporterEndpoint is the URL of the local node_exporter endpoint, without
	// the /metrics path. Disabled, when an empty string.
	nodeExporterEndpoint string,
	// dockerRegsitryEndpoint is the URL of the intermediary caching docker registry,
	// for scraping and forwarding metrics. Disabled, when an empty string.
	dockerRegistryNodeExporterEndpoint string,
) prometheus.GathererFunc {
	nodeMetrics := newMetricsSyncPoint()
	registryMetrics := newMetricsSyncPoint()

	go backgroundCollectNodeExporterMetrics(nodeExporterEndpoint, nodeMetrics)
	go backgroundCollectNodeExporterMetrics(dockerRegistryNodeExporterEndpoint, registryMetrics)

	return func() (mfs []*dto.MetricFamily, err error) {
		// notify to start a scrape
		nodeMetrics.notify.Signal()
		registryMetrics.notify.Signal()

		mfs, err = gatherer.Gather()
		if err != nil {
			return nil, errors.Wrap(err, "getting default gatherer")
		}

		if nodeExporterEndpoint != "" {
			result := <-registryMetrics.result
			if result.err != nil {
				logger.Warn("failed to get metrics for node exporter", log.Error(result.err))
			}
			for key, mf := range result.metrics {
				if strings.HasPrefix(key, "go_") || strings.HasPrefix(key, "promhttp_") || strings.HasPrefix(key, "process_") {
					continue
				}

				mfs = append(mfs, mf)
			}
		}

		if dockerRegistryNodeExporterEndpoint != "" {
			result := <-registryMetrics.result
			if result.err != nil {
				logger.Warn("failed to get metrics for docker registry", log.Error(result.err))
			}
			for key, mf := range result.metrics {
				if strings.HasPrefix(key, "go_") || strings.HasPrefix(key, "promhttp_") || strings.HasPrefix(key, "process_") {
					continue
				}

				// should only be 1 registry, so we give it a set instance value
				metricLabelInstance := "sg_instance"
				instanceName := "docker-regsitry"
				for _, m := range mf.Metric {
					m.Label = append(m.Label, &dto.LabelPair{Name: &metricLabelInstance, Value: &instanceName})
				}

				mfs = append(mfs, mf)
			}
		}

		return mfs, nil
	}
}

// On notify, scrapes the specified endpoint for prometheus metrics and sends them down the
// associated channel. If the endpoint is "", then the channel is closed so that subsequent
// reads return an empty value instead of blocking indefinitely.
func backgroundCollectNodeExporterMetrics(endpoint string, syncPoint metricsSyncPoint) {
	if endpoint == "" {
		close(syncPoint.result)
		return
	}

	collect := func() (map[string]*dto.MetricFamily, error) {
		resp, err := (&http.Client{
			Timeout: 2 * time.Second,
		}).Get(endpoint + "/metrics")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var parser expfmt.TextParser
		mfMap, err := parser.TextToMetricFamilies(bytes.NewReader(b))
		return mfMap, errors.Wrapf(err, "parsing node_exporter metrics, response: %s", string(b))
	}

	for {
		syncPoint.notify.L.Lock()
		syncPoint.notify.Wait()
		mfMap, err := collect()
		if err != nil {
			syncPoint.result <- metricsResult{err: err}
		} else {
			syncPoint.result <- metricsResult{metrics: mfMap}
		}
		syncPoint.notify.L.Unlock()
	}
}
