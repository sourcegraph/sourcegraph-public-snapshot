package conf

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf/confdefaults"
	"github.com/sourcegraph/sourcegraph/pkg/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/pkg/legacyconf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	deployType := DeployType()
	if !IsValidDeployType(deployType) {
		log.Fatalf("The 'DEPLOY_TYPE' environment variable is invalid. Expected one of: %q, %q, %q. Got: %q", DeployCluster, DeployDocker, DeployDev, deployType)
	}

	defaultConfig := defaultConfigForDeployment()

	// If a legacy configuration file is available (specified via
	// SOURCEGRAPH_CONFIG_FILE), use it as the default for the critical and
	// site configs.
	//
	// This relies on the fact that the old v2.13.6 site config schema has
	// most fields align directly with the v3.0+ critical and site config
	// schemas.
	//
	// This code can be removed in the next significant version after 3.0 (NOT
	// preview), after which critical/site config schemas no longer need to
	// align generally.
	//
	// TODO(slimsag): Remove after 3.0 (NOT preview).
	{
		legacyConf := jsonc.Normalize(legacyconf.Raw())

		var criticalDecoded schema.CriticalConfiguration
		_ = json.Unmarshal(legacyConf, &criticalDecoded)

		// Backwards compatability for deprecated environment variables
		// that we previously considered deprecated but are actually
		// widespread in use in user's deployments and/or are suggested for
		// use in our public documentation (i.e., even though these were
		// long deprecated, our docs were not up to date).
		criticalBackcompatVars := map[string]func(value string){
			"LIGHTSTEP_PROJECT":      func(v string) { criticalDecoded.LightstepProject = v },
			"LIGHTSTEP_ACCESS_TOKEN": func(v string) { criticalDecoded.LightstepAccessToken = v },
		}
		for envVar, setter := range criticalBackcompatVars {
			val := os.Getenv(envVar)
			if val != "" {
				setter(val)
			}
		}

		critical, err := json.MarshalIndent(criticalDecoded, "", "  ")
		if string(critical) != "{}" && err == nil {
			defaultConfig.Critical = string(critical)
		}

		var siteDecoded schema.SiteConfiguration
		_ = json.Unmarshal(legacyConf, &siteDecoded)
		site, err := json.MarshalIndent(siteDecoded, "", "  ")
		if string(site) != "{}" && err == nil {
			defaultConfig.Site = string(site)
		}
	}

	confdefaults.Default = defaultConfig
}

func defaultConfigForDeployment() conftypes.RawUnified {
	deployType := DeployType()
	switch {
	case IsDev(deployType):
		return confdefaults.DevAndTesting
	case IsDeployTypeDockerContainer(deployType):
		return confdefaults.DockerContainer
	case IsDeployTypeCluster(deployType):
		return confdefaults.Cluster
	default:
		panic("deploy type did not register default configuration")
	}
}

func AWSCodeCommitConfigs(ctx context.Context) ([]*schema.AWSCodeCommitConnection, error) {
	var config []*schema.AWSCodeCommitConnection
	if err := api.InternalClient.ExternalServiceConfigs(ctx, "AWSCODECOMMIT", &config); err != nil {
		return nil, err
	}
	return config, nil
}

func BitbucketServerConfigs(ctx context.Context) ([]*schema.BitbucketServerConnection, error) {
	var config []*schema.BitbucketServerConnection
	if err := api.InternalClient.ExternalServiceConfigs(ctx, "BITBUCKETSERVER", &config); err != nil {
		return nil, err
	}
	return config, nil
}

func GitHubConfigs(ctx context.Context) ([]*schema.GitHubConnection, error) {
	var config []*schema.GitHubConnection
	if err := api.InternalClient.ExternalServiceConfigs(ctx, "GITHUB", &config); err != nil {
		return nil, err
	}
	return config, nil
}

func GitLabConfigs(ctx context.Context) ([]*schema.GitLabConnection, error) {
	var config []*schema.GitLabConnection
	if err := api.InternalClient.ExternalServiceConfigs(ctx, "GITLAB", &config); err != nil {
		return nil, err
	}
	return config, nil
}

func GitoliteConfigs(ctx context.Context) ([]*schema.GitoliteConnection, error) {
	var config []*schema.GitoliteConnection
	if err := api.InternalClient.ExternalServiceConfigs(ctx, "GITOLITE", &config); err != nil {
		return nil, err
	}
	return config, nil
}

func PhabricatorConfigs(ctx context.Context) ([]*schema.PhabricatorConnection, error) {
	var config []*schema.PhabricatorConnection
	if err := api.InternalClient.ExternalServiceConfigs(ctx, "PHABRICATOR", &config); err != nil {
		return nil, err
	}
	return config, nil
}

type AccessTokAllow string

const (
	AccessTokensNone  AccessTokAllow = "none"
	AccessTokensAll   AccessTokAllow = "all-users-create"
	AccessTokensAdmin AccessTokAllow = "site-admin-create"
)

// AccessTokensAllow returns whether access tokens are enabled, disabled, or restricted to creation by admin users.
func AccessTokensAllow() AccessTokAllow {
	cfg := Get().AuthAccessTokens
	if cfg == nil {
		return AccessTokensAll
	}
	switch cfg.Allow {
	case "":
		return AccessTokensAll
	case string(AccessTokensAll):
		return AccessTokensAll
	case string(AccessTokensNone):
		return AccessTokensNone
	case string(AccessTokensAdmin):
		return AccessTokensAdmin
	default:
		return AccessTokensNone
	}
}

// EmailVerificationRequired returns whether users must verify an email address before they
// can perform most actions on this site.
//
// It's false for sites that do not have an email sending API key set up.
func EmailVerificationRequired() bool {
	return Get().EmailSmtp != nil
}

// CanSendEmail returns whether the site can send emails (e.g., to reset a password or
// invite a user to an org).
//
// It's false for sites that do not have an email sending API key set up.
func CanSendEmail() bool {
	return Get().EmailSmtp != nil
}

// CanReadEmail tells if an IMAP server is configured and reading email is possible.
func CanReadEmail() bool {
	return Get().EmailImap != nil
}

const (
	DeployCluster = "cluster"
	DeployDocker  = "docker-container"
	DeployDev     = "dev"
)

// DeployType tells the deployment type.
func DeployType() string {
	if e := os.Getenv("DEPLOY_TYPE"); e != "" {
		return e
	}
	// Default to Cluster so that every Cluster deployment doesn't need to be
	// configured with DEPLOY_TYPE.
	return DeployCluster
}

// IsDeployTypeCluster tells if the given deployment type is a cluster (and
// non-dev, non-single Docker image).
func IsDeployTypeCluster(deployType string) bool {
	if deployType == "k8s" {
		// backwards compatibility for older deployments
		return true
	}
	return deployType == DeployCluster
}

// IsDeployTypeDockerContainer tells if the given deployment type is Docker sourcegraph/server
// single-container (non-Kubernetes, non-cluster, non-dev).
func IsDeployTypeDockerContainer(deployType string) bool {
	return deployType == DeployDocker
}

// IsDev tells if the given deployment type is "dev".
func IsDev(deployType string) bool {
	return deployType == DeployDev
}

// IsValidDeployType returns true iff the given deployType is a Kubernetes deployment, Docker deployment, or a
// local development environmnent.
func IsValidDeployType(deployType string) bool {
	return IsDeployTypeCluster(deployType) || IsDeployTypeDockerContainer(deployType) || IsDev(deployType)
}

// UpdateChannel tells the update channel. Default is "release".
func UpdateChannel() string {
	channel := Get().Critical.UpdateChannel
	if channel == "" {
		return "release"
	}
	return channel
}

// SearchIndexEnabled returns true if sourcegraph should index all
// repositories for text search. If the configuration is unset, it returns
// false for the docker server image (due to resource usage) but true
// elsewhere. Additionally it also checks for the outdated environment
// variable INDEXED_SEARCH.
func SearchIndexEnabled() bool {
	if v := Get().SearchIndexEnabled; v != nil {
		return *v
	}
	if v := os.Getenv("INDEXED_SEARCH"); v != "" {
		enabled, _ := strconv.ParseBool(v)
		return enabled
	}
	return DeployType() != DeployDocker
}

// SrcGitServers represents the SRC_GIT_SERVERS environment variable.
//
// Non-frontend callers should go through api.InternalClient.GitServerAddrs() instead.
var SrcGitServers = readSrcGitServers()

func readSrcGitServers() []string {
	v := env.Get("SRC_GIT_SERVERS", "", "addresses of the remote gitservers")
	if v == "" {
		// Detect 'go test' and setup default addresses in that case.
		p, err := os.Executable()
		if err == nil && filepath.Ext(p) == ".test" {
			v = "gitserver:3178"
		}
	}
	return strings.Fields(v)
}

func UsingExternalURL() bool {
	url := Get().Critical.ExternalURL
	return !(url == "" || strings.HasPrefix(url, "http://localhost") || strings.HasPrefix(url, "https://localhost") || strings.HasPrefix(url, "http://127.0.0.1") || strings.HasPrefix(url, "https://127.0.0.1")) // CI:LOCALHOST_OK
}

func IsExternalURLSecure() bool {
	return strings.HasPrefix(Get().Critical.ExternalURL, "https:")
}

func IsBuiltinSignupAllowed() bool {
	provs := Get().Critical.AuthProviders
	for _, prov := range provs {
		if prov.Builtin != nil {
			return prov.Builtin.AllowSignup
		}
	}
	return false
}

func Branding() schema.Branding {
	brandSettings := Get().Branding
	if brandSettings == nil {
		return schema.Branding{
			Logo:    "",
			Favicon: "",
		}
	}
	return *brandSettings
}
