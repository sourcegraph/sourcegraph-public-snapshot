package conf

import (
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

// parseConfigData parses the provided config string into the given cfg struct
// pointer.
func parseConfigData(data string, cfg any) error {
	if data != "" {
		if err := jsonc.Unmarshal(data, cfg); err != nil {
			return err
		}
	}

	if v, ok := cfg.(*schema.SiteConfiguration); ok {
		// For convenience, make sure this is not nil.
		if v.ExperimentalFeatures == nil {
			v.ExperimentalFeatures = &schema.ExperimentalFeatures{}
		}
	}
	return nil
}

// ParseConfig parses the raw configuration.
func ParseConfig(data conftypes.RawUnified) (*Unified, error) {
	cfg := &Unified{
		ServiceConnectionConfig: data.ServiceConnections,
	}
	if err := parseConfigData(data.Site, &cfg.SiteConfiguration); err != nil {
		return nil, err
	}
	return cfg, nil
}

// requireRestart describes the list of config properties that require
// restarting the Sourcegraph Server in order for the change to take effect.
//
// Experimental features are special in that they are denoted individually
// via e.g. "experimentalFeatures::myFeatureFlag".
var requireRestart = []string{
	"auth.providers",
	"insights.query.worker.concurrency",
	"insights.commit.indexer.interval",
	"permissions.syncUsersMaxConcurrency",
}

// needRestartToApply determines if a restart is needed to apply the changes
// between the two configurations.
func needRestartToApply(before, after *Unified) bool {
	// Check every option that changed to determine whether or not a server
	// restart is required.
	for option := range diff(before, after) {
		for _, requireRestartOption := range requireRestart {
			if option == requireRestartOption {
				return true
			}
		}
	}
	return false
}
