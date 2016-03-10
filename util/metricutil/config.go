package metricutil

import (
	"log"
	"time"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

// Flags defines settings for metrics.
type Flags struct {
	ForwardURL string `long:"metrics.forward" value-name:"URL" description:"Sourcegraph metric sink to forward metrics to (empty to disable)" default:"https://sourcegraph.com" env:"SRC_METRICS_FORWARD"`

	StoreURL string `long:"metrics.store" value-name:"URL" description:"Elasticsearch server to store metrics in (if set)" env:"SG_ELASTICSEARCH_URL"`

	DisableAllCollection bool `long:"metrics.disable" description:"Disable all metrics, stats and events data collection from this server" env:"SRC_METRICS_DISABLE"`
}

// config is the currently active metrics config (as set by CLI flags).
var config Flags

// DisableMetricsCollection is true if this server should not collect
// usage metrics, stats or events data.
//
// It depends on the CLI flags being set, so it only returns the
// correct value when called from an invocation of `src serve`.
func DisableMetricsCollection() bool {
	return config.DisableAllCollection
}

// EnableForwarding is true if this server should forward metrics
// and stats to another server.
//
// It depends on the CLI flags being set, so it only returns the
// correct value when called from an invocation of `src serve`.
func EnableForwarding() bool {
	return !config.DisableAllCollection && config.ForwardURL != ""
}

func init() {
	cli.PostInit = append(cli.PostInit, func() {
		_, err := cli.Serve.AddGroup("Metrics forwarding", "Metrics forwarding", &config)
		if err != nil {
			log.Fatal(err)
		}
	})

	cli.ServeInit = append(cli.ServeInit, func() {
		if config.ForwardURL != "" && config.StoreURL != "" {
			log.Fatal("At most one of the --metrics.forward and --metrics.store and CLI flags may be specified.")
		}
	})
}

// Start starts the event logger and event storer using the CLI
// configuration.
func Start(ctx context.Context, channelCapacity, workerBufferSize int, flushInterval time.Duration) {
	if DisableMetricsCollection() {
		return
	}

	if config.StoreURL != "" {
		// Listen for events and flush them to Elasticsearch.
		startEventStorer(ctx)
	}

	// Listen for events and periodically push them upstream.
	startEventLogger(ctx, channelCapacity, workerBufferSize, flushInterval)
}
