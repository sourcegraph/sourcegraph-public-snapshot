package conf

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

// parseConfigData parses the provided config string into the given cfg struct
// pointer.
func parseConfigData(data string, cfg interface{}) error {
	if data != "" {
		data, err := jsonc.Parse(data)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, cfg); err != nil {
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
		ServiceConnections: data.ServiceConnections,
	}
	if err := parseConfigData(data.Site, &cfg.SiteConfiguration); err != nil {
		return nil, err
	}
	return cfg, nil
}

// ConfigWriteResult defines the actions that should be performed after a config
// property is changed in order for the user to see the change take effect.
type ConfigWriteResult struct {
	ClientReloadRequired  bool
	ServerRestartRequired bool
}

type configPropertyResultSchema map[string]ConfigWriteResult

// configPropertiesRequiringAction describes the list of config properties that
// require action to be taken after being changed.
//
// Experimental features are special in that they are denoted individually via
// e.g. "experimentalFeatures::myFeatureFlag".
var configPropertiesRequiringAction = configPropertyResultSchema{
	"auth.accessTokens":                {ServerRestartRequired: true},
	"auth.providers":                   {ServerRestartRequired: true},
	"auth.sessionExpiry":               {ServerRestartRequired: true},
	"auth.userOrgMap":                  {ServerRestartRequired: true},
	"disablePublicRepoRedirects":       {ServerRestartRequired: true},
	"experimentalFeatures::automation": {ClientReloadRequired: true},
	"extensions":                       {ServerRestartRequired: true},
	"externalURL":                      {ServerRestartRequired: true},
	"git.cloneURLToRepositoryName":     {ServerRestartRequired: true},
	"lightstepAccessToken":             {ServerRestartRequired: true},
	"lightstepProject":                 {ServerRestartRequired: true},
	"searchScopes":                     {ServerRestartRequired: true},
	"update.channel":                   {ServerRestartRequired: true},
	"useJaeger":                        {ServerRestartRequired: true},
}

func calculateConfigChangeResult(before, after *Unified, schema configPropertyResultSchema) ConfigWriteResult {
	result := ConfigWriteResult{}

	// Check every option that changed to determine whether or not any flags
	// should be set.
	for option := range diff(before, after) {
		if optionResult, ok := schema[option]; ok {
			if optionResult.ClientReloadRequired {
				result.ClientReloadRequired = true
			}
			if optionResult.ServerRestartRequired {
				result.ServerRestartRequired = true
			}
		}
	}

	return result
}

// CalculateConfigChangeResult determines the actions that need to be taken to
// apply the changes between the two configurations.
func CalculateConfigChangeResult(before, after *Unified) ConfigWriteResult {
	return calculateConfigChangeResult(before, after, configPropertiesRequiringAction)
}
