package conf

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/conf/confdefaults"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	deployType := deploy.Type()
	if !deploy.IsValidDeployType(deployType) {
		log.Fatalf("The 'DEPLOY_TYPE' environment variable is invalid. Expected one of: %q, %q, %q, %q, %q, %q. Got: %q", deploy.Kubernetes, deploy.DockerCompose, deploy.PureDocker, deploy.SingleDocker, deploy.Dev, deploy.Helm, deployType)
	}

	confdefaults.Default = defaultConfigForDeployment()
}

func defaultConfigForDeployment() conftypes.RawUnified {
	deployType := deploy.Type()
	switch {
	case deploy.IsDev(deployType):
		return confdefaults.DevAndTesting
	case deploy.IsDeployTypeSingleDockerContainer(deployType):
		return confdefaults.DockerContainer
	case deploy.IsDeployTypeKubernetes(deployType), deploy.IsDeployTypeDockerCompose(deployType), deploy.IsDeployTypePureDocker(deployType):
		return confdefaults.KubernetesOrDockerComposeOrPureDocker
	default:
		panic("deploy type did not register default configuration")
	}
}

func BitbucketServerConfigs(ctx context.Context) ([]*schema.BitbucketServerConnection, error) {
	var config []*schema.BitbucketServerConnection
	if err := internalapi.Client.ExternalServiceConfigs(ctx, extsvc.KindBitbucketServer, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func GitHubConfigs(ctx context.Context) ([]*schema.GitHubConnection, error) {
	var config []*schema.GitHubConnection
	if err := internalapi.Client.ExternalServiceConfigs(ctx, extsvc.KindGitHub, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func GitLabConfigs(ctx context.Context) ([]*schema.GitLabConnection, error) {
	var config []*schema.GitLabConnection
	if err := internalapi.Client.ExternalServiceConfigs(ctx, extsvc.KindGitLab, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func GitoliteConfigs(ctx context.Context) ([]*schema.GitoliteConnection, error) {
	var config []*schema.GitoliteConnection
	if err := internalapi.Client.ExternalServiceConfigs(ctx, extsvc.KindGitolite, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func PhabricatorConfigs(ctx context.Context) ([]*schema.PhabricatorConnection, error) {
	var config []*schema.PhabricatorConnection
	if err := internalapi.Client.ExternalServiceConfigs(ctx, extsvc.KindPhabricator, &config); err != nil {
		return nil, err
	}
	return config, nil
}

type AccessTokenAllow string

const (
	AccessTokensNone  AccessTokenAllow = "none"
	AccessTokensAll   AccessTokenAllow = "all-users-create"
	AccessTokensAdmin AccessTokenAllow = "site-admin-create"
)

// AccessTokensAllow returns whether access tokens are enabled, disabled, or
// restricted creation to only site admins.
func AccessTokensAllow() AccessTokenAllow {
	cfg := Get().AuthAccessTokens
	if cfg == nil || cfg.Allow == "" {
		return AccessTokensAll
	}
	v := AccessTokenAllow(cfg.Allow)
	switch v {
	case AccessTokensAll, AccessTokensAdmin:
		return v
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

// UpdateChannel tells the update channel. Default is "release".
func UpdateChannel() string {
	channel := Get().UpdateChannel
	if channel == "" {
		return "release"
	}
	return channel
}

// SearchIndexEnabled returns true if sourcegraph should index all
// repositories for text search.
func SearchIndexEnabled() bool {
	if v := Get().SearchIndexEnabled; v != nil {
		return *v
	}
	return true // always on by default in all deployment types, see confdefaults.go
}

func BatchChangesEnabled() bool {
	if enabled := Get().BatchChangesEnabled; enabled != nil {
		return *enabled
	}
	return true
}

func BatchChangesRestrictedToAdmins() bool {
	if restricted := Get().BatchChangesRestrictToAdmins; restricted != nil {
		return *restricted
	}
	return false
}

func ExecutorsEnabled() bool {
	return Get().ExecutorsAccessToken != ""
}

func CodeIntelAutoIndexingEnabled() bool {
	if enabled := Get().CodeIntelAutoIndexingEnabled; enabled != nil {
		return *enabled
	}
	return false
}

func CodeIntelAutoIndexingAllowGlobalPolicies() bool {
	if enabled := Get().CodeIntelAutoIndexingAllowGlobalPolicies; enabled != nil {
		return *enabled
	}
	return false
}

func CodeIntelAutoIndexingPolicyRepositoryMatchLimit() int {
	val := Get().CodeIntelAutoIndexingPolicyRepositoryMatchLimit
	if val == nil || *val < -1 {
		return -1
	}

	return *val
}

func CodeInsightsGQLApiEnabled() bool {
	enabled, _ := strconv.ParseBool(os.Getenv("ENABLE_CODE_INSIGHTS_SETTINGS_STORAGE"))
	return !enabled
}

func ProductResearchPageEnabled() bool {
	if enabled := Get().ProductResearchPageEnabled; enabled != nil {
		return *enabled
	}
	return true
}

func ExternalURL() string {
	return Get().ExternalURL
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

// SearchSymbolsParallelism returns 20, or the site config
// "debug.search.symbolsParallelism" value if configured.
func SearchSymbolsParallelism() int {
	val := Get().DebugSearchSymbolsParallelism
	if val == 0 {
		return 20
	}
	return val
}

func BitbucketServerPluginPerm() bool {
	val := ExperimentalFeatures().BitbucketServerFastPerm
	return val == "enabled"
}

func EventLoggingEnabled() bool {
	val := ExperimentalFeatures().EventLogging
	if val == "" {
		return true
	}
	return val == "enabled"
}

func APIDocsSearchIndexingEnabled() bool {
	val := ExperimentalFeatures().ApidocsSearchIndexing
	if val == "" {
		return false // off by default until API docs search indexing stabilizes, see https://github.com/sourcegraph/sourcegraph/issues/26292
	}
	return val == "enabled"
}

func StructuralSearchEnabled() bool {
	val := ExperimentalFeatures().StructuralSearch
	if val == "" {
		return true
	}
	return val == "enabled"
}

func DependeciesSearchEnabled() bool {
	val := ExperimentalFeatures().DependenciesSearch
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

func Tracer() string {
	ot := Get().ObservabilityTracing
	if ot == nil {
		return ""
	}
	return ot.Type
}

// AuthMinPasswordLength returns the value of minimum password length requirement.
// If not set, it returns the default value 12.
func AuthMinPasswordLength() int {
	val := Get().AuthMinPasswordLength
	if val <= 0 {
		return 12
	}
	return val
}

// By default, password reset links are valid for 4 hours.
const defaultPasswordLinkExpiry = 14400

// AuthPasswordResetLinkExpiry returns the time (in seconds) indicating how long password
// reset links are considered valid. If not set, it returns the default value.
func AuthPasswordResetLinkExpiry() int {
	val := Get().AuthPasswordResetLinkExpiry
	if val <= 0 {
		return defaultPasswordLinkExpiry
	}
	return val
}

type ExternalServiceMode int

const (
	ExternalServiceModeDisabled ExternalServiceMode = 0
	ExternalServiceModePublic   ExternalServiceMode = 1
	ExternalServiceModeAll      ExternalServiceMode = 2
)

func (e ExternalServiceMode) String() string {
	switch e {
	case ExternalServiceModeDisabled:
		return "disabled"
	case ExternalServiceModePublic:
		return "public"
	case ExternalServiceModeAll:
		return "all"
	default:
		return "unknown"
	}
}

// ExternalServiceUserMode returns the site level mode describing if users are
// allowed to add external services for public and private repositories. It does
// NOT take into account permissions granted to the current user.
func ExternalServiceUserMode() ExternalServiceMode {
	switch Get().ExternalServiceUserMode {
	case "public":
		return ExternalServiceModePublic
	case "all":
		return ExternalServiceModeAll
	default:
		return ExternalServiceModeDisabled
	}
}

const defaultGitLongCommandTimeout = time.Hour

// GitLongCommandTimeout returns the maximum amount of time in seconds that a
// long Git command (e.g. clone or remote update) is allowed to execute. If not
// set, it returns the default value.
//
// In general, Git commands that are expected to take a long time should be
// executed in the background in a non-blocking fashion.
func GitLongCommandTimeout() time.Duration {
	val := Get().GitLongCommandTimeout
	if val < 1 {
		return defaultGitLongCommandTimeout
	}
	return time.Duration(val) * time.Second
}

// GitMaxCodehostRequestsPerSecond returns maximum number of remote code host
// git operations to be run per second per gitserver. If not set, it returns the
// default value -1.
func GitMaxCodehostRequestsPerSecond() int {
	val := Get().GitMaxCodehostRequestsPerSecond
	if val == nil || *val < -1 {
		return -1
	}
	return *val
}

func GitMaxConcurrentClones() int {
	v := Get().GitMaxConcurrentClones
	if v <= 0 {
		return 5
	}
	return v
}

func UserReposMaxPerUser() int {
	v := Get().UserReposMaxPerUser
	if v == 0 {
		return 2000
	}
	return v
}

func UserReposMaxPerSite() int {
	v := Get().UserReposMaxPerSite
	if v == 0 {
		return 200000
	}
	return v
}
