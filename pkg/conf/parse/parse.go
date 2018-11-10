package parse

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

// parseConfigData reads the provided config string, but NOT the environment
func parseConfigData(data string) (*schema.SiteConfiguration, error) {
	var tmpConfig schema.SiteConfiguration

	if data != "" {
		data, err := jsonc.Parse(data)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &tmpConfig); err != nil {
			return nil, err
		}
	}

	// For convenience, make sure this is not nil.
	if tmpConfig.ExperimentalFeatures == nil {
		tmpConfig.ExperimentalFeatures = &schema.ExperimentalFeatures{}
	}
	return &tmpConfig, nil
}

// ParseConfigEnvironment reads the provided string, then merges in additional
// data from the (deprecated) environment.
func ParseConfigEnvironment(data string) (*schema.SiteConfiguration, error) {
	tmpConfig, err := parseConfigData(data)
	if err != nil {
		return nil, err
	}

	// Env var config takes highest precedence but is deprecated.
	if v, envVarNames, err := configFromEnv(); err != nil {
		return nil, err
	} else if len(envVarNames) > 0 {
		if err := json.Unmarshal(v, tmpConfig); err != nil {
			return nil, err
		}
	}
	return tmpConfig, nil
}

// requireRestart describes the list of config properties that require
// restarting the Sourcegraph Server in order for the change to take effect.
//
// Experimental features are special in that they are denoted individually
// via e.g. "experimentalFeatures::myFeatureFlag".
var requireRestart = []string{
	"executeGradleOriginalRootPaths",
	"lightstepAccessToken",
	"lightstepProject",
	"auth.accessTokens",
	"privateArtifactRepoURL",
	"auth.userOrgMap",
	"auth.sessionExpiry",
	"noGoGetDomains",
	"auth.disableAccessTokens",
	"auth.providers",
	"externalURL",
	"tls.letsencrypt",
	"git.cloneURLToRepositoryName",
	"searchScopes",
	"extensions",
	"disableBrowserExtension",
	"tlsCert",
	"update.channel",
	"useJaeger",
	"privateArtifactRepoPassword",
	"disablePublicRepoRedirects",
	"privateArtifactRepoUsername",
	"blacklistGoGet",
	"privateArtifactRepoID",
	"tlsKey",
}

// NeedRestartToApply determines if a restart is needed to apply the changes
// between the two configurations.
func NeedRestartToApply(before, after *schema.SiteConfiguration) bool {
	diff := diff(before, after)

	// Check every option that changed to determine whether or not a server
	// restart is required.
	for option := range diff {
		for _, requireRestartOption := range requireRestart {
			if option == requireRestartOption {
				return true
			}
		}
	}
	return false
}
