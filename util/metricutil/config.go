package metricutil

import (
	"log"
	"time"

	"golang.org/x/net/context"

	"gopkg.in/inconshreveable/log15.v2"

	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

// Flags defines settings for metrics.
type Flags struct {
	ForwardURL string `long:"metrics.forward" value-name:"URL" description:"Sourcegraph metric sink to forward metrics to (empty to disable)" default:"https://sourcegraph.com"`

	StoreURL string `long:"metrics.store" value-name:"URL" description:"Elasticsearch server to store metrics in (if set)" env:"SG_ELASTICSEARCH_URL"`

	DeprecatedGraphUplinkPeriod time.Duration `long:"graphuplink" hidden:"yes"`
}

// config is the currently active metrics config (as set by CLI flags).
var config Flags

// EnableForwarding is true if this server should forward usage
// metrics and stats to another server for collection.
//
// It depends on the CLI flags being set, so it only returns the
// correct value when called from an invocation of `src serve`.
func EnableForwarding() bool {
	return config.ForwardURL != ""
}

func init() {
	cli.PostInit = append(cli.PostInit, func() {
		_, err := cli.Serve.AddGroup("Metrics forwarding", "Metrics forwarding", &config)
		if err != nil {
			log.Fatal(err)
		}
	})

	cli.ServeInit = append(cli.ServeInit, func() {
		if fed.Config.DeprecatedIsRoot {
			log15.Warn("The --fed.is-root option is DEPRECATED. If you were using it to disable sending metrics to an external server, switch to using --metrics.forward='' instead. Otherwise, you can remove it, as it is a no-op.")
			config.ForwardURL = ""
		}

		if config.ForwardURL != "" && config.StoreURL != "" {
			log.Fatal("At most one of the --metrics.forward and --metrics.store and CLI flags may be specified.")
		}

		if config.DeprecatedGraphUplinkPeriod != 0 {
			log15.Warn("The --graphuplink flag is DEPRECATED. The standard interval is now 10 minutes and is not configurable.")
		}
	})
}

// Start starts the event logger and event storer using the CLI
// configuration.
func Start(ctx context.Context, channelCapacity, workerBufferSize int, flushInterval time.Duration) {
	// Listen for events and flush them to Elasticsearch.
	startEventStorer(ctx)

	// Listen for events and periodically push them upstream.
	startEventLogger(ctx, channelCapacity, workerBufferSize, flushInterval)
}
