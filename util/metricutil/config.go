package metricutil

import (
	"log"

	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

// Flags defines settings for metrics.
type Flags struct {
	ForwardURL string `long:"metrics.forward" value-name:"URL" description:"Sourcegraph metric sink to forward metrics to (empty to disable)" default:"https://sourcegraph.com"`

	StoreURL string `long:"metrics.store" value-name:"URL" description:"Elasticsearch server to store metrics in (if set)" env:"SG_ELASTICSEARCH_URL"`
}

// config is the currently active metrics config (as set by CLI flags).
var config Flags

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

		if fed.Config.IsRoot && config.ForwardURL != "" {
			// Preserve existing behavior where setting --fed.is-root
			// disabled metrics forwarding.
			config.ForwardURL = ""
		}
	})
}
