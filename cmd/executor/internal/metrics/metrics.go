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
	// dockerRegistryEndpoint is the URL of the intermediary caching docker registry,
	// for scraping and forwarding metrics. Disabled, when an empty string.
	dockerRegistryEndpoint string,
) prometheus.GathererFunc {
	nodeMetrics := newMetricsSyncPoint()
	registryMetrics := newMetricsSyncPoint()
	registryNodeMetrics := newMetricsSyncPoint()

	go backgroundCollectMetrics(nodeExporterEndpoint+"/metrics", nodeMetrics)
	go backgroundCollectMetrics(dockerRegistryEndpoint+"/proxy?module=registry", registryMetrics)
	go backgroundCollectMetrics(dockerRegistryEndpoint+"/proxy?module=node", registryNodeMetrics)

	return func() ([]*dto.MetricFamily, error) {
		// notify to start a scrape
		nodeMetrics.notify.Signal()
		registryMetrics.notify.Signal()
		registryNodeMetrics.notify.Signal()

		nodeMetricsGatherer := prometheus.GathererFunc(func() (mfs []*dto.MetricFamily, err error) {
			if nodeExporterEndpoint != "" {
				result := <-nodeMetrics.result
				if result.err != nil {
					logger.Warn("failed to get metrics for node exporter", log.Error(result.err))
				}
				for key, mf := range result.metrics {
					if filterMetric(key) {
						continue
					}

					mfs = append(mfs, mf)
				}
			}

			return mfs, err
		})

		registryMetricsGatherer := prometheus.GathererFunc(func() (mfs []*dto.MetricFamily, err error) {
			if dockerRegistryEndpoint != "" {
				result := <-registryMetrics.result
				if result.err != nil {
					logger.Warn("failed to get metrics for docker registry", log.Error(result.err))
				}
				for key, mf := range result.metrics {
					if filterMetric(key) {
						continue
					}

					// should only be 1 registry, so we give it a set instance value
					metricLabelInstance := "sg_instance"
					instanceName := "docker-registry"
					for _, m := range mf.Metric {
						m.Label = append(m.Label, &dto.LabelPair{Name: &metricLabelInstance, Value: &instanceName})
					}

					mfs = append(mfs, mf)
				}
			}

			return mfs, err
		})

		registryNodeMetricsGatherer := prometheus.GathererFunc(func() (mfs []*dto.MetricFamily, err error) {
			if dockerRegistryEndpoint != "" {
				result := <-registryNodeMetrics.result
				if result.err != nil {
					logger.Warn("failed to get metrics for docker registry", log.Error(result.err))
				}
				for key, mf := range result.metrics {
					if filterMetric(key) {
						continue
					}

					// should only be 1 registry, so we give it a set instance value
					metricLabelInstance := "sg_instance"
					instanceName := "docker-registry"
					for _, m := range mf.Metric {
						m.Label = append(m.Label, &dto.LabelPair{Name: &metricLabelInstance, Value: &instanceName})
					}

					mfs = append(mfs, mf)
				}
			}

			return mfs, err
		})

		gatherers := prometheus.Gatherers{
			gatherer,
			nodeMetricsGatherer,
			registryMetricsGatherer,
			registryNodeMetricsGatherer,
		}

		return gatherers.Gather()
	}
}

func filterMetric(key string) bool {
	return strings.HasPrefix(key, "go_") || strings.HasPrefix(key, "promhttp_") || strings.HasPrefix(key, "process_")
}

// On notify, scrapes the specified endpoint for prometheus metrics and sends them down the
// associated channel. If the endpoint is "", then the channel is closed so that subsequent
// reads return an empty value instead of blocking indefinitely.
func backgroundCollectMetrics(endpoint string, syncPoint metricsSyncPoint) {
	if endpoint == "" {
		close(syncPoint.result)
		return
	}

	collect := func() (map[string]*dto.MetricFamily, error) {
		resp, err := (&http.Client{
			Timeout: 2 * time.Second,
		}).Get(endpoint)
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
		return mfMap, errors.Wrapf(err, "parsing metrics, response: %s", string(b))
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
