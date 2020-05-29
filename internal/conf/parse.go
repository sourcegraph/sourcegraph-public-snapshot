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

// PostConfigWriteActions defines actions that should be taken after a config
// property is changed in order for the user to see the change take effect.
type PostConfigWriteActions struct {
	FrontendReloadRequired bool
	ServerRestartRequired  bool
}

type configPropertyActionSchema map[string]PostConfigWriteActions

// configPropertiesRequiringAction describes the list of config properties that
// require action to be taken after being changed.
//
// Experimental features are special in that they are denoted individually via
// e.g. "experimentalFeatures::myFeatureFlag".
var configPropertiesRequiringAction = configPropertyActionSchema{
	"auth.accessTokens":                {ServerRestartRequired: true},
	"auth.providers":                   {ServerRestartRequired: true},
	"auth.sessionExpiry":               {ServerRestartRequired: true},
	"auth.userOrgMap":                  {ServerRestartRequired: true},
	"disablePublicRepoRedirects":       {ServerRestartRequired: true},
	"experimentalFeatures::automation": {FrontendReloadRequired: true},
	"extensions":                       {ServerRestartRequired: true},
	"externalURL":                      {ServerRestartRequired: true},
	"git.cloneURLToRepositoryName":     {ServerRestartRequired: true},
	"lightstepAccessToken":             {ServerRestartRequired: true},
	"lightstepProject":                 {ServerRestartRequired: true},
	"searchScopes":                     {ServerRestartRequired: true},
	"update.channel":                   {ServerRestartRequired: true},
	"useJaeger":                        {ServerRestartRequired: true},
}

func needActionToApply(before, after *Unified, schema configPropertyActionSchema) PostConfigWriteActions {
	actions := PostConfigWriteActions{}

	// Check every option that changed to determine whether or not action should
	// be taken.
	for option := range diff(before, after) {
		if action, ok := schema[option]; ok {
			if action.FrontendReloadRequired {
				actions.FrontendReloadRequired = true
			}
			if action.ServerRestartRequired {
				actions.ServerRestartRequired = true
			}
		}
	}

	return actions
}

// NeedActionToApply determines if action needs to be taken to apply the changes
// between the two configurations.
func NeedActionToApply(before, after *Unified) PostConfigWriteActions {
	return needActionToApply(before, after, configPropertiesRequiringAction)
}
