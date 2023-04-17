package conf

import (
	"context"
	"encoding/base64"
	"log"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/conf/confdefaults"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	srccli "github.com/sourcegraph/sourcegraph/internal/src-cli"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	deployType := deploy.Type()
	if !deploy.IsValidDeployType(deployType) {
		log.Fatalf("The 'DEPLOY_TYPE' environment variable is invalid. Expected one of: %q, %q, %q, %q, %q, %q, %q. Got: %q", deploy.Kubernetes, deploy.DockerCompose, deploy.PureDocker, deploy.SingleDocker, deploy.Dev, deploy.Helm, deploy.App, deployType)
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
	case deploy.IsDeployTypeApp(deployType):
		return confdefaults.App
	default:
		panic("deploy type did not register default configuration")
	}
}

func ExecutorsAccessToken() string {
	if deploy.IsApp() {
		return confdefaults.AppInMemoryExecutorPassword
	}
	return Get().ExecutorsAccessToken
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

func GitHubAppEnabled() bool {
	cfg, _ := GitHubAppConfig()
	return cfg.Configured()
}

type GitHubAppConfiguration struct {
	PrivateKey   []byte
	AppID        string
	Slug         string
	ClientID     string
	ClientSecret string
}

func (c GitHubAppConfiguration) Configured() bool {
	return c.AppID != "" && len(c.PrivateKey) != 0 && c.Slug != "" && c.ClientID != "" && c.ClientSecret != ""
}

func GitHubAppConfig() (config GitHubAppConfiguration, err error) {
	cfg := Get().GitHubApp
	if cfg == nil {
		return GitHubAppConfiguration{}, nil
	}

	privateKey, err := base64.StdEncoding.DecodeString(cfg.PrivateKey)
	if err != nil {
		return GitHubAppConfiguration{}, errors.Wrap(err, "decoding GitHub app private key failed")
	}
	return GitHubAppConfiguration{
		PrivateKey:   privateKey,
		AppID:        cfg.AppID,
		Slug:         cfg.Slug,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
	}, nil
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
	return CanSendEmail()
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

func ExecutorsFrontendURL() string {
	current := Get()
	if current.ExecutorsFrontendURL != "" {
		return current.ExecutorsFrontendURL
	}

	return current.ExternalURL
}

func ExecutorsSrcCLIImage() string {
	current := Get()
	if current.ExecutorsSrcCLIImage != "" {
		return current.ExecutorsSrcCLIImage
	}

	return "sourcegraph/src-cli"
}

func ExecutorsSrcCLIImageTag() string {
	current := Get()
	if current.ExecutorsSrcCLIImageTag != "" {
		return current.ExecutorsSrcCLIImageTag
	}

	return srccli.MinimumVersion
}

func ExecutorsBatcheshelperImage() string {
	current := Get()
	if current.ExecutorsBatcheshelperImage != "" {
		return current.ExecutorsBatcheshelperImage
	}

	return "sourcegraph/batcheshelper"
}

func ExecutorsBatcheshelperImageTag() string {
	current := Get()
	if current.ExecutorsBatcheshelperImageTag != "" {
		return current.ExecutorsBatcheshelperImageTag
	}

	if version.IsDev(version.Version()) {
		return "insiders"
	}

	return version.Version()
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

func CodeIntelRankingDocumentReferenceCountsEnabled() bool {
	if enabled := Get().CodeIntelRankingDocumentReferenceCountsEnabled; enabled != nil {
		return *enabled
	}
	return false
}

func CodeIntelRankingDocumentReferenceCountsGraphKey() string {
	if val := Get().CodeIntelRankingDocumentReferenceCountsGraphKey; val != "" {
		return val
	}
	return "dev"
}

func CodeIntelRankingDocumentReferenceCountsDerivativeGraphKeyPrefix() string {
	if val := Get().CodeIntelRankingDocumentReferenceCountsDerivativeGraphKeyPrefix; val != "" {
		return val
	}
	return ""
}

func CodeIntelRankingStaleResultAge() time.Duration {
	if val := Get().CodeIntelRankingStaleResultsAge; val > 0 {
		return time.Duration(val) * time.Hour
	}
	return 24 * time.Hour
}

func EmbeddingsEnabled() bool {
	embeddingsConfig := Get().Embeddings
	return embeddingsConfig != nil && embeddingsConfig.Enabled
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

// IsAccessRequestEnabled returns whether request access experimental feature is enabled or not.
func IsAccessRequestEnabled() bool {
	authAccessRequest := Get().AuthAccessRequest
	return authAccessRequest == nil || authAccessRequest.Enabled == nil || *authAccessRequest.Enabled
}

// AuthPrimaryLoginProvidersCount returns the number of primary login providers
// configured, or 3 (the default) if not explicitly configured.
// This is only used for the UI
func AuthPrimaryLoginProvidersCount() int {
	c := Get().AuthPrimaryLoginProvidersCount
	if c == 0 {
		return 3 // default to 3
	}
	return c
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

func EventLoggingEnabled() bool {
	val := ExperimentalFeatures().EventLogging
	if val == "" {
		return true
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

// SearchDocumentRanksWeight controls the impact of document ranks on the final ranking when
// SearchOptions.UseDocumentRanks is enabled. The default is 0.5 * 9000 (half the zoekt default),
// to match existing behavior where ranks are given half the priority as existing scoring signals.
// We plan to eventually remove this, once we experiment on real data to find a good default.
func SearchDocumentRanksWeight() float64 {
	ranking := ExperimentalFeatures().Ranking
	if ranking != nil && ranking.DocumentRanksWeight != nil {
		return *ranking.DocumentRanksWeight
	} else {
		return 4500
	}
}

// SearchFlushWallTime controls the amount of time that Zoekt shards collect and rank results when
// the 'search-ranking' feature is enabled. We plan to eventually remove this, once we experiment
// on real data to find a good default.
func SearchFlushWallTime() time.Duration {
	ranking := ExperimentalFeatures().Ranking
	if ranking != nil && ranking.FlushWallTimeMS > 0 {
		return time.Duration(ranking.FlushWallTimeMS) * time.Millisecond
	} else {
		return 500 * time.Millisecond
	}
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

// GenericPasswordPolicy is a generic password policy that defines password requirements.
type GenericPasswordPolicy struct {
	Enabled                   bool
	MinimumLength             int
	NumberOfSpecialCharacters int
	RequireAtLeastOneNumber   bool
	RequireUpperandLowerCase  bool
}

// AuthPasswordPolicy returns a GenericPasswordPolicy for password validation
func AuthPasswordPolicy() GenericPasswordPolicy {
	ml := Get().AuthMinPasswordLength

	if p := Get().AuthPasswordPolicy; p != nil {
		return GenericPasswordPolicy{
			Enabled:                   p.Enabled,
			MinimumLength:             ml,
			NumberOfSpecialCharacters: p.NumberOfSpecialCharacters,
			RequireAtLeastOneNumber:   p.RequireAtLeastOneNumber,
			RequireUpperandLowerCase:  p.RequireUpperandLowerCase,
		}
	}

	if ep := ExperimentalFeatures().PasswordPolicy; ep != nil {
		return GenericPasswordPolicy{
			Enabled:                   ep.Enabled,
			MinimumLength:             ml,
			NumberOfSpecialCharacters: ep.NumberOfSpecialCharacters,
			RequireAtLeastOneNumber:   ep.RequireAtLeastOneNumber,
			RequireUpperandLowerCase:  ep.RequireUpperandLowerCase,
		}
	}

	return GenericPasswordPolicy{
		Enabled:                   false,
		MinimumLength:             0,
		NumberOfSpecialCharacters: 0,
		RequireAtLeastOneNumber:   false,
		RequireUpperandLowerCase:  false,
	}
}

func PasswordPolicyEnabled() bool {
	pc := AuthPasswordPolicy()
	return pc.Enabled
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

// AuthLockout populates and returns the *schema.AuthLockout with default values
// for fields that are not initialized.
func AuthLockout() *schema.AuthLockout {
	val := Get().AuthLockout
	if val == nil {
		return &schema.AuthLockout{
			ConsecutivePeriod:      3600,
			FailedAttemptThreshold: 5,
			LockoutPeriod:          1800,
		}
	}

	if val.ConsecutivePeriod <= 0 {
		val.ConsecutivePeriod = 3600
	}
	if val.FailedAttemptThreshold <= 0 {
		val.FailedAttemptThreshold = 5
	}
	if val.LockoutPeriod <= 0 {
		val.LockoutPeriod = 1800
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
