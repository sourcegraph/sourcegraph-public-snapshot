package extsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Account represents a row in the `user_external_accounts` table. See the GraphQL API's
// corresponding fields in "ExternalAccount" for documentation.
type Account struct {
	ID          int32
	UserID      int32
	AccountSpec // ServiceType, ServiceID, ClientID, AccountID
	AccountData // AuthData, Data
	PublicAccountData
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AccountSpec specifies a user external account by its external identifier (i.e., by the
// identifier provided by the account's owner service), instead of by our database's serial
// ID. See the GraphQL API's corresponding fields in "ExternalAccount" for documentation.
type AccountSpec struct {
	ServiceType string
	ServiceID   string
	ClientID    string
	AccountID   string
}

// AccountData contains data that can be freely updated in the user external account after it
// has been created. See the GraphQL API's corresponding fields for documentation.
type AccountData struct {
	AuthData *EncryptableData
	Data     *EncryptableData
}

// PublicAccountData contains a few fields from the AccountData.Data mentioned above.
// We only expose publicly available fields in this struct.
// See the GraphQL API's corresponding fields for documentation.
type PublicAccountData struct {
	DisplayName *string `json:"displayName,omitempty"`
	Login       *string `json:"login,omitempty"`
	URL         *string `json:"url,omitempty"`
}

type EncryptableData = encryption.JSONEncryptable[any]

func NewUnencryptedData(value json.RawMessage) *EncryptableData {
	return &EncryptableData{Encryptable: encryption.NewUnencrypted(string(value))}
}

func NewEncryptedData(cipher, keyID string, key encryption.Key) *EncryptableData {
	if cipher == "" && keyID == "" {
		return nil
	}

	return &EncryptableData{Encryptable: encryption.NewEncrypted(cipher, keyID, key)}
}

// Repository contains necessary information to identify an external repository on the code host.
type Repository struct {
	// URI is the full name for this repository, e.g. "github.com/user/repo".
	URI string
	api.ExternalRepoSpec
}

// Accounts contains a list of accounts that belong to the same external service.
// All fields have a same meaning to AccountSpec. See GraphQL API's corresponding fields
// in "ExternalAccount" for documentation.
type Accounts struct {
	ServiceType string
	ServiceID   string
	AccountIDs  []string
}

type EncryptableConfig = encryption.Encryptable

func NewEmptyConfig() *EncryptableConfig {
	return NewUnencryptedConfig("{}")
}

func NewEmptyGitLabConfig() *EncryptableConfig {
	return NewUnencryptedConfig(`{"url": "https://gitlab.com", "token": "abdef", "projectQuery":["none"]}`)
}

func NewUnencryptedConfig(value string) *EncryptableConfig {
	return encryption.NewUnencrypted(value)
}

func NewEncryptedConfig(cipher, keyID string, key encryption.Key) *EncryptableConfig {
	if cipher == "" && keyID == "" {
		return nil
	}

	return encryption.NewEncrypted(cipher, keyID, key)
}

// TracingFields returns tracing fields for the opentracing log.
func (s *Accounts) TracingFields() []otlog.Field {
	return []otlog.Field{
		otlog.String("Accounts.ServiceType", s.ServiceType),
		otlog.String("Accounts.Perm", s.ServiceID),
		otlog.Int("Accounts.AccountIDs.Count", len(s.AccountIDs)),
	}
}

const (
	// The constants below represent the different kinds of external service we support and should be used
	// in preference to the Type values below.

	KindAWSCodeCommit   = "AWSCODECOMMIT"
	KindBitbucketServer = "BITBUCKETSERVER"
	KindBitbucketCloud  = "BITBUCKETCLOUD"
	KindGerrit          = "GERRIT"
	KindGitHub          = "GITHUB"
	KindGitLab          = "GITLAB"
	KindGitolite        = "GITOLITE"
	KindPerforce        = "PERFORCE"
	KindPhabricator     = "PHABRICATOR"
	KindGoPackages      = "GOMODULES"
	KindJVMPackages     = "JVMPACKAGES"
	KindPythonPackages  = "PYTHONPACKAGES"
	KindRustPackages    = "RUSTPACKAGES"
	KindRubyPackages    = "RUBYPACKAGES"
	KindNpmPackages     = "NPMPACKAGES"
	KindPagure          = "PAGURE"
	KindOther           = "OTHER"
)

const (
	// The constants below represent the values used for the external_service_type column of the repo table.

	// TypeAWSCodeCommit is the (api.ExternalRepoSpec).ServiceType value for AWS CodeCommit
	// repositories. The ServiceID value is the ARN (Amazon Resource Name) omitting the repository name
	// suffix (e.g., "arn:aws:codecommit:us-west-1:123456789:").
	TypeAWSCodeCommit = "awscodecommit"

	// TypeBitbucketServer is the (api.ExternalRepoSpec).ServiceType value for Bitbucket Server projects. The
	// ServiceID value is the base URL to the Bitbucket Server instance.
	TypeBitbucketServer = "bitbucketServer"

	// TypeBitbucketCloud is the (api.ExternalRepoSpec).ServiceType value for Bitbucket Cloud projects. The
	// ServiceID value is the base URL to the Bitbucket Cloud.
	TypeBitbucketCloud = "bitbucketCloud"

	// TypeGerrit is the (api.ExternalRepoSpec).ServiceType value for Gerrit projects.
	TypeGerrit = "gerrit"

	// TypeGitHub is the (api.ExternalRepoSpec).ServiceType value for GitHub repositories. The ServiceID value
	// is the base URL to the GitHub instance (https://github.com or the GitHub Enterprise URL).
	TypeGitHub = "github"

	// TypeGitLab is the (api.ExternalRepoSpec).ServiceType value for GitLab projects. The ServiceID
	// value is the base URL to the GitLab instance (https://gitlab.com or self-hosted GitLab URL).
	TypeGitLab = "gitlab"

	// TypeGitolite is the (api.ExternalRepoSpec).ServiceType value for Gitolite projects.
	TypeGitolite = "gitolite"

	// TypePerforce is the (api.ExternalRepoSpec).ServiceType value for Perforce projects.
	TypePerforce = "perforce"

	// TypePhabricator is the (api.ExternalRepoSpec).ServiceType value for Phabricator projects.
	TypePhabricator = "phabricator"

	// TypeJVMPackages is the (api.ExternalRepoSpec).ServiceType value for Maven packages (Java/JVM ecosystem libraries).
	TypeJVMPackages = "jvmPackages"

	// TypePagure is the (api.ExternalRepoSpec).ServiceType value for Pagure projects.
	TypePagure = "pagure"

	// TypeNpmPackages is the (api.ExternalRepoSpec).ServiceType value for Npm packages (JavaScript/TypeScript ecosystem libraries).
	TypeNpmPackages = "npmPackages"

	// TypeGoModules is the (api.ExternalRepoSpec).ServiceType value for Go modules.
	TypeGoModules = "goModules"

	// TypePythonPackages is the (api.ExternalRepoSpec).ServiceType value for Python packages.
	TypePythonPackages = "pythonPackages"

	// TypeRustPackages is the (api.ExternalRepoSpec).ServiceType value for Rust packages.
	TypeRustPackages = "rustPackages"

	// TypeRubyPackages is the (api.ExternalRepoSpec).ServiceType value for Ruby packages.
	TypeRubyPackages = "rubyPackages"

	// TypeOther is the (api.ExternalRepoSpec).ServiceType value for other projects.
	TypeOther = "other"
)

// KindToType returns a Type constants given a Kind
// It will panic when given an unknown kind
func KindToType(kind string) string {
	switch kind {
	case KindAWSCodeCommit:
		return TypeAWSCodeCommit
	case KindBitbucketServer:
		return TypeBitbucketServer
	case KindBitbucketCloud:
		return TypeBitbucketCloud
	case KindGerrit:
		return TypeGerrit
	case KindGitHub:
		return TypeGitHub
	case KindGitLab:
		return TypeGitLab
	case KindGitolite:
		return TypeGitolite
	case KindPhabricator:
		return TypePhabricator
	case KindPerforce:
		return TypePerforce
	case KindJVMPackages:
		return TypeJVMPackages
	case KindPythonPackages:
		return TypePythonPackages
	case KindRustPackages:
		return TypeRustPackages
	case KindRubyPackages:
		return TypeRubyPackages
	case KindNpmPackages:
		return TypeNpmPackages
	case KindGoPackages:
		return TypeGoModules
	case KindPagure:
		return TypePagure
	case KindOther:
		return TypeOther
	default:
		panic(fmt.Sprintf("unknown kind: %q", kind))
	}
}

// TypeToKind returns a Kind constants given a Type
// It will panic when given an unknown type.
func TypeToKind(t string) string {
	switch t {
	case TypeAWSCodeCommit:
		return KindAWSCodeCommit
	case TypeBitbucketServer:
		return KindBitbucketServer
	case TypeBitbucketCloud:
		return KindBitbucketCloud
	case TypeGerrit:
		return KindGerrit
	case TypeGitHub:
		return KindGitHub
	case TypeGitLab:
		return KindGitLab
	case TypeGitolite:
		return KindGitolite
	case TypePerforce:
		return KindPerforce
	case TypePhabricator:
		return KindPhabricator
	case TypeNpmPackages:
		return KindNpmPackages
	case TypeJVMPackages:
		return KindJVMPackages
	case TypePythonPackages:
		return KindPythonPackages
	case TypeRustPackages:
		return KindRustPackages
	case TypeRubyPackages:
		return KindRubyPackages
	case TypeGoModules:
		return KindGoPackages
	case TypePagure:
		return KindPagure
	case TypeOther:
		return KindOther
	default:
		panic(fmt.Sprintf("unknown type: %q", t))
	}
}

var (
	// Precompute these for use in ParseServiceType below since the constants are mixed case
	bbsLower    = strings.ToLower(TypeBitbucketServer)
	bbcLower    = strings.ToLower(TypeBitbucketCloud)
	jvmLower    = strings.ToLower(TypeJVMPackages)
	npmLower    = strings.ToLower(TypeNpmPackages)
	goLower     = strings.ToLower(TypeGoModules)
	pythonLower = strings.ToLower(TypePythonPackages)
	rustLower   = strings.ToLower(TypeRustPackages)
	rubyLower   = strings.ToLower(TypeRubyPackages)
)

// ParseServiceType will return a ServiceType constant after doing a case insensitive match on s.
// It returns ("", false) if no match was found.
func ParseServiceType(s string) (string, bool) {
	switch strings.ToLower(s) {
	case TypeAWSCodeCommit:
		return TypeAWSCodeCommit, true
	case bbsLower:
		return TypeBitbucketServer, true
	case bbcLower:
		return TypeBitbucketCloud, true
	case TypeGerrit:
		return TypeGerrit, true
	case TypeGitHub:
		return TypeGitHub, true
	case TypeGitLab:
		return TypeGitLab, true
	case TypeGitolite:
		return TypeGitolite, true
	case TypePerforce:
		return TypePerforce, true
	case TypePhabricator:
		return TypePhabricator, true
	case goLower:
		return TypeGoModules, true
	case jvmLower:
		return TypeJVMPackages, true
	case npmLower:
		return TypeNpmPackages, true
	case pythonLower:
		return TypePythonPackages, true
	case rustLower:
		return TypeRustPackages, true
	case rubyLower:
		return TypeRubyPackages, true
	case TypePagure:
		return TypePagure, true
	case TypeOther:
		return TypeOther, true
	default:
		return "", false
	}
}

// ParseServiceKind will return a ServiceKind constant after doing a case insensitive match on s.
// It returns ("", false) if no match was found.
func ParseServiceKind(s string) (string, bool) {
	switch strings.ToUpper(s) {
	case KindAWSCodeCommit:
		return KindAWSCodeCommit, true
	case KindBitbucketServer:
		return KindBitbucketServer, true
	case KindBitbucketCloud:
		return KindBitbucketCloud, true
	case KindGerrit:
		return KindGerrit, true
	case KindGitHub:
		return KindGitHub, true
	case KindGitLab:
		return KindGitLab, true
	case KindGitolite:
		return KindGitolite, true
	case KindPerforce:
		return KindPerforce, true
	case KindPhabricator:
		return KindPhabricator, true
	case KindGoPackages:
		return KindGoPackages, true
	case KindJVMPackages:
		return KindJVMPackages, true
	case KindPythonPackages:
		return KindPythonPackages, true
	case KindRustPackages:
		return KindRustPackages, true
	case KindRubyPackages:
		return KindRubyPackages, true
	case KindPagure:
		return KindPagure, true
	case KindOther:
		return KindOther, true
	default:
		return "", false
	}
}

var supportsRepoExclusion = map[string]bool{
	KindAWSCodeCommit:   true,
	KindBitbucketCloud:  true,
	KindBitbucketServer: true,
	KindGitHub:          true,
	KindGitLab:          true,
	KindGitolite:        true,
}

// SupportsRepoExclusion returns true when given external service kind supports
// repo exclusion.
func SupportsRepoExclusion(extSvcKind string) bool {
	return supportsRepoExclusion[extSvcKind]
}

// AccountID is a descriptive type for the external identifier of an external account on the
// code host. It can be the string representation of an integer (e.g. GitLab), a GraphQL ID
// (e.g. GitHub), or a username (e.g. Bitbucket Server) depends on the code host type.
type AccountID string

// RepoID is a descriptive type for the external identifier of an external repository on the
// code host. It can be the string representation of an integer (e.g. GitLab and Bitbucket
// Server) or a GraphQL ID (e.g. GitHub) depends on the code host type.
type RepoID string

// RepoIDType indicates the type of the RepoID.
type RepoIDType string

func ParseEncryptableConfig(ctx context.Context, kind string, config *EncryptableConfig) (any, error) {
	cfg, err := getConfigPrototype(kind)
	if err != nil {
		return nil, err
	}

	rawConfig, err := config.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	if err := jsonc.Unmarshal(rawConfig, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// ParseConfig attempts to unmarshal the given JSON config into a configuration struct defined in the schema package.
func ParseConfig(kind, config string) (any, error) {
	cfg, err := getConfigPrototype(kind)
	if err != nil {
		return nil, err
	}

	return cfg, jsonc.Unmarshal(config, cfg)
}

func getConfigPrototype(kind string) (any, error) {
	switch strings.ToUpper(kind) {
	case KindAWSCodeCommit:
		return &schema.AWSCodeCommitConnection{}, nil
	case KindBitbucketServer:
		return &schema.BitbucketServerConnection{}, nil
	case KindBitbucketCloud:
		return &schema.BitbucketCloudConnection{}, nil
	case KindGerrit:
		return &schema.GerritConnection{}, nil
	case KindGitHub:
		return &schema.GitHubConnection{}, nil
	case KindGitLab:
		return &schema.GitLabConnection{}, nil
	case KindGitolite:
		return &schema.GitoliteConnection{}, nil
	case KindPerforce:
		return &schema.PerforceConnection{}, nil
	case KindPhabricator:
		return &schema.PhabricatorConnection{}, nil
	case KindGoPackages:
		return &schema.GoModulesConnection{}, nil
	case KindJVMPackages:
		return &schema.JVMPackagesConnection{}, nil
	case KindPagure:
		return &schema.PagureConnection{}, nil
	case KindNpmPackages:
		return &schema.NpmPackagesConnection{}, nil
	case KindPythonPackages:
		return &schema.PythonPackagesConnection{}, nil
	case KindRustPackages:
		return &schema.RustPackagesConnection{}, nil
	case KindRubyPackages:
		return &schema.RubyPackagesConnection{}, nil
	case KindOther:
		return &schema.OtherExternalServiceConnection{}, nil
	default:
		return nil, errors.Errorf("unknown external service kind %q", kind)
	}
}

const IDParam = "externalServiceID"

// WebhookURL returns an endpoint URL for the given external service. If the kind
// of external service does not support webhooks it returns an empty string.
func WebhookURL(kind string, externalServiceID int64, cfg any, externalURL string) (string, error) {
	var path, extra string
	switch strings.ToUpper(kind) {
	case KindGitHub:
		path = "github-webhooks"
	case KindBitbucketServer:
		path = "bitbucket-server-webhooks"
	case KindGitLab:
		path = "gitlab-webhooks"
	case KindBitbucketCloud:
		path = "bitbucket-cloud-webhooks"

		// Unlike other external service kinds, Bitbucket Cloud doesn't support
		// a shared secret defined as part of the webhook. As a result, we need
		// to include it as an explicit part of the URL that we construct.
		switch c := cfg.(type) {
		case *schema.BitbucketCloudConnection:
			extra = "&secret=" + url.QueryEscape(c.WebhookSecret)
		default:
			return "", errors.Newf("external service with id=%d claims to be a Bitbucket Cloud service, but the configuration is of type %T", cfg)
		}
	default:
		// If not a supported kind, bail out.
		return "", nil
	}
	// eg. https://example.com/.api/github-webhooks?externalServiceID=1
	return fmt.Sprintf("%s/.api/%s?%s=%d%s", externalURL, path, IDParam, externalServiceID, extra), nil
}

func ExtractEncryptableToken(ctx context.Context, config *EncryptableConfig, kind string) (string, error) {
	parsed, err := ParseEncryptableConfig(ctx, kind, config)
	if err != nil {
		return "", errors.Wrap(err, "loading service configuration")
	}

	return extractToken(parsed, kind)
}

// ExtractToken attempts to extract the token from the supplied args
func ExtractToken(config string, kind string) (string, error) {
	parsed, err := ParseConfig(kind, config)
	if err != nil {
		return "", errors.Wrap(err, "loading service configuration")
	}

	return extractToken(parsed, kind)
}

func extractToken(parsed any, kind string) (string, error) {
	switch c := parsed.(type) {
	case *schema.GitHubConnection:
		return c.Token, nil
	case *schema.GitLabConnection:
		return c.Token, nil
	case *schema.BitbucketServerConnection:
		return c.Token, nil
	case *schema.PhabricatorConnection:
		return c.Token, nil
	case *schema.PagureConnection:
		return c.Token, nil
	default:
		return "", errors.Errorf("unable to extract token for service kind %q", kind)
	}
}

func ExtractEncryptableRateLimit(ctx context.Context, config *EncryptableConfig, kind string) (rate.Limit, error) {
	parsed, err := ParseEncryptableConfig(ctx, kind, config)
	if err != nil {
		return rate.Inf, errors.Wrap(err, "loading service configuration")
	}

	rlc, err := GetLimitFromConfig(kind, parsed)
	if err != nil {
		return rate.Inf, err
	}

	return rlc, nil
}

// ExtractRateLimit extracts the rate limit from the given args. If rate limiting is not
// supported the error returned will be an ErrRateLimitUnsupported.
func ExtractRateLimit(config, kind string) (rate.Limit, error) {
	parsed, err := ParseConfig(kind, config)
	if err != nil {
		return rate.Inf, errors.Wrap(err, "loading service configuration")
	}

	rlc, err := GetLimitFromConfig(kind, parsed)
	if err != nil {
		return rate.Inf, err
	}

	return rlc, nil
}

// GetLimitFromConfig gets RateLimitConfig from an already parsed config schema.
func GetLimitFromConfig(kind string, config any) (rate.Limit, error) {
	// Rate limit config can be in a few states:
	// 1. Not defined: We fall back to default specified in code.
	// 2. Defined and enabled: We use their defined limit.
	// 3. Defined and disabled: We use an infinite limiter.

	var limit rate.Limit
	switch c := config.(type) {
	case *schema.GitLabConnection:
		// 10/s is the default enforced by GitLab on their end
		limit = rate.Limit(10)
		if c != nil && c.RateLimit != nil {
			limit = limitOrInf(c.RateLimit.Enabled, c.RateLimit.RequestsPerHour)
		}
	case *schema.GitHubConnection:
		// 5000 per hour is the default enforced by GitHub on their end
		limit = rate.Limit(5000.0 / 3600.0)
		if c != nil && c.RateLimit != nil {
			limit = limitOrInf(c.RateLimit.Enabled, c.RateLimit.RequestsPerHour)
		}
	case *schema.BitbucketServerConnection:
		// 8/s is the default limit we enforce
		limit = rate.Limit(8)
		if c != nil && c.RateLimit != nil {
			limit = limitOrInf(c.RateLimit.Enabled, c.RateLimit.RequestsPerHour)
		}
	case *schema.BitbucketCloudConnection:
		limit = rate.Limit(2)
		if c != nil && c.RateLimit != nil {
			limit = limitOrInf(c.RateLimit.Enabled, c.RateLimit.RequestsPerHour)
		}
	case *schema.PerforceConnection:
		limit = rate.Limit(5000.0 / 3600.0)
		if c != nil && c.RateLimit != nil {
			limit = limitOrInf(c.RateLimit.Enabled, c.RateLimit.RequestsPerHour)
		}
	case *schema.JVMPackagesConnection:
		limit = rate.Limit(2)
		if c != nil && c.Maven != nil && c.Maven.RateLimit != nil {
			limit = limitOrInf(c.Maven.RateLimit.Enabled, c.Maven.RateLimit.RequestsPerHour)
		}
	case *schema.PagureConnection:
		// 8/s is the default limit we enforce
		limit = rate.Limit(8)
		if c != nil && c.RateLimit != nil {
			limit = limitOrInf(c.RateLimit.Enabled, c.RateLimit.RequestsPerHour)
		}
	case *schema.NpmPackagesConnection:
		limit = rate.Limit(6000 / 3600.0) // Same as the default in npm-packages.schema.json
		if c != nil && c.RateLimit != nil {
			limit = limitOrInf(c.RateLimit.Enabled, c.RateLimit.RequestsPerHour)
		}
	case *schema.GoModulesConnection:
		// Unlike the GitHub or GitLab APIs, the public npm registry (i.e. https://proxy.golang.org)
		// doesn't document an enforced req/s rate limit AND we do a lot more individual
		// requests in comparison since they don't offer enough batch APIs.
		limit = rate.Limit(57600.0 / 3600.0) // Same as default in go-modules.schema.json
		if c != nil && c.RateLimit != nil {
			limit = limitOrInf(c.RateLimit.Enabled, c.RateLimit.RequestsPerHour)
		}
	case *schema.PythonPackagesConnection:
		// Unlike the GitHub or GitLab APIs, the pypi.org doesn't
		// document an enforced req/s rate limit.
		limit = rate.Limit(57600.0 / 3600.0) // 16/second same as default in python-packages.schema.json
		if c != nil && c.RateLimit != nil {
			limit = limitOrInf(c.RateLimit.Enabled, c.RateLimit.RequestsPerHour)
		}
	case *schema.RustPackagesConnection:
		// The crates.io CDN has no rate limits https://www.pietroalbini.org/blog/downloading-crates-io/
		limit = rate.Limit(100)
		if c != nil && c.RateLimit != nil {
			limit = limitOrInf(c.RateLimit.Enabled, c.RateLimit.RequestsPerHour)
		}
	case *schema.RubyPackagesConnection:
		// The rubygems.org API allows 10 rps https://guides.rubygems.org/rubygems-org-rate-limits/
		limit = rate.Limit(10)
		if c != nil && c.RateLimit != nil {
			limit = limitOrInf(c.RateLimit.Enabled, c.RateLimit.RequestsPerHour)
		}
	default:
		return limit, ErrRateLimitUnsupported{codehostKind: kind}
	}

	return limit, nil
}

func limitOrInf(enabled bool, perHour float64) rate.Limit {
	if enabled {
		return rate.Limit(perHour / 3600)
	}
	return rate.Inf
}

type ErrRateLimitUnsupported struct {
	codehostKind string
}

func (e ErrRateLimitUnsupported) Error() string {
	return fmt.Sprintf("internal rate limiting not supported for %s", e.codehostKind)
}

const (
	URNGitHubApp   = "GitHubApp"
	URNGitHubOAuth = "GitHubOAuth"
	URNGitLabOAuth = "GitLabOAuth"
	URNCodeIntel   = "CodeIntel"
)

// URN returns a unique resource identifier of an external service by given kind and ID.
func URN(kind string, id int64) string {
	return "extsvc:" + strings.ToLower(kind) + ":" + strconv.FormatInt(id, 10)
}

// DecodeURN returns the kind of the external service and its ID.
func DecodeURN(urn string) (kind string, id int64) {
	fields := strings.Split(urn, ":")
	if len(fields) != 3 {
		return "", 0
	}

	id, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return "", 0
	}
	return fields[1], id
}

type OtherRepoMetadata struct {
	// RelativePath is relative to ServiceID which is usually the host URL.
	// Joining them gives you the clone url.
	RelativePath string
}

func UniqueEncryptableCodeHostIdentifier(ctx context.Context, kind string, config *EncryptableConfig) (string, error) {
	cfg, err := ParseEncryptableConfig(ctx, kind, config)
	if err != nil {
		return "", err
	}

	return uniqueCodeHostIdentifier(kind, cfg)
}

// UniqueCodeHostIdentifier returns a string that uniquely identifies the
// instance of a code host an external service is pointing at.
//
// E.g.: multiple external service configurations might point at the same
// GitHub Enterprise instance. All of them would return the normalized base URL
// as a unique identifier.
//
// In case an external service doesn't have a base URL (e.g. AWS Code Commit)
// another unique identifier is returned.
//
// This function can be used to group external services by the code host
// instance they point at.
func UniqueCodeHostIdentifier(kind, config string) (string, error) {
	cfg, err := ParseConfig(kind, config)
	if err != nil {
		return "", err
	}

	return uniqueCodeHostIdentifier(kind, cfg)
}

func uniqueCodeHostIdentifier(kind string, cfg any) (string, error) {
	var rawURL string
	switch c := cfg.(type) {
	case *schema.GitLabConnection:
		rawURL = c.Url
	case *schema.GitHubConnection:
		rawURL = c.Url
	case *schema.BitbucketServerConnection:
		rawURL = c.Url
	case *schema.BitbucketCloudConnection:
		rawURL = c.Url
	case *schema.GerritConnection:
		rawURL = c.Url
	case *schema.PhabricatorConnection:
		rawURL = c.Url
	case *schema.OtherExternalServiceConnection:
		rawURL = c.Url
	case *schema.GitoliteConnection:
		rawURL = c.Host
	case *schema.AWSCodeCommitConnection:
		// AWS Code Commit does not have a URL in the config, so we return a
		// unique string here and return early:
		return c.Region + ":" + c.AccessKeyID, nil
	case *schema.PerforceConnection:
		// Perforce uses the P4PORT to specify the instance, so we use that
		return c.P4Port, nil
	case *schema.GoModulesConnection:
		return KindGoPackages, nil
	case *schema.JVMPackagesConnection:
		return KindJVMPackages, nil
	case *schema.NpmPackagesConnection:
		return KindNpmPackages, nil
	case *schema.PythonPackagesConnection:
		return KindPythonPackages, nil
	case *schema.RustPackagesConnection:
		return KindRustPackages, nil
	case *schema.RubyPackagesConnection:
		return KindRubyPackages, nil
	case *schema.PagureConnection:
		rawURL = c.Url
	default:
		return "", errors.Errorf("unknown external service kind: %s", kind)
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	return NormalizeBaseURL(u).String(), nil
}

// CodeHostBaseURL is an identifier for a unique code host in the form
// of its host URL.
// To create a new CodeHostBaseURL, use NewCodeHostURN.
// e.g. NewCodeHostURN("https://github.com")
// To use the string value again, use codeHostURN.String()
type CodeHostBaseURL struct {
	baseURL string
}

// NewCodeHostBaseURL takes a code host URL string, normalizes it,
// and returns a CodeHostURN. If the string is required, use the
// .String() method on the CodeHostURN.
func NewCodeHostBaseURL(baseURL string) (CodeHostBaseURL, error) {
	codeHostURL, err := url.Parse(baseURL)
	if err != nil {
		return CodeHostBaseURL{}, err
	}

	return CodeHostBaseURL{baseURL: NormalizeBaseURL(codeHostURL).String()}, nil
}

// String returns the stored, normalized code host URN as a string.
func (c CodeHostBaseURL) String() string {
	return c.baseURL
}
