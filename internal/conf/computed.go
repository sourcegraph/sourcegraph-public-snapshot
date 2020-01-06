package conf

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/confdefaults"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	deployType := DeployType()
	if !IsValidDeployType(deployType) {
		log.Fatalf("The 'DEPLOY_TYPE' environment variable is invalid. Expected one of: %q, %q, %q. Got: %q", DeployCluster, DeployDocker, DeployDev, deployType)
	}

	confdefaults.Default = defaultConfigForDeployment()
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

// Deploy type constants. Any changes here should be reflected in the DeployType type declared in web/src/globals.d.ts:
// https://sourcegraph.com/search?q=r:github.com/sourcegraph/sourcegraph%24+%22type+DeployType%22
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
	channel := Get().UpdateChannel
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

func SymbolIndexEnabled() bool {
	enabled := SearchIndexEnabled()
	if v := Get().SearchIndexSymbolsEnabled; v != nil {
		enabled = enabled && *v
	}
	return enabled
}

func UsingExternalURL() bool {
	url := Get().ExternalURL
	return !(url == "" || strings.HasPrefix(url, "http://localhost") || strings.HasPrefix(url, "https://localhost") || strings.HasPrefix(url, "http://127.0.0.1") || strings.HasPrefix(url, "https://127.0.0.1")) // CI:LOCALHOST_OK
}

func IsExternalURLSecure() bool {
	return strings.HasPrefix(Get().ExternalURL, "https:")
}

func IsBuiltinSignupAllowed() bool {
	provs := Get().AuthProviders
	for _, prov := range provs {
		if prov.Builtin != nil {
			return prov.Builtin.AllowSignup
		}
	}
	return false
}

func Branding() *schema.Branding {
	branding := Get().Branding
	if branding != nil && branding.BrandName == "" {
		bcopy := *branding
		bcopy.BrandName = "Sourcegraph"
		branding = &bcopy
	}
	return branding
}

func BrandName() string {
	branding := Branding()
	if branding == nil || branding.BrandName == "" {
		return "Sourcegraph"
	}
	return branding.BrandName
}

// SearchSymbolsParallelism returns 20, or the site config
// "debug.search.symbolsParallelism" value if configured.
func SearchSymbolsParallelism() int {
	val := Get().DebugSearchSymbolsParallelism
	if val == 0 {
		return 20
	}
	return val
}

func BitbucketServerFastPerm() bool {
	val := Get().ExperimentalFeatures.BitbucketServerFastPerm
	if val == "" {
		return false
	}
	return val == "enabled"
}

func EventLoggingEnabled() bool {
	val := Get().ExperimentalFeatures.EventLogging
	if val == "" {
		return true
	}
	return val == "enabled"
}

func StructuralSearchEnabled() bool {
	val := Get().ExperimentalFeatures.StructuralSearch
	if val == "" {
		return true
	}
	return val == "enabled"
}

func ExperimentalFeatures() schema.ExperimentalFeatures {
	val := Get().ExperimentalFeatures
	if val == nil {
		return schema.ExperimentalFeatures{}
	}
	return *val
}
