package extsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
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
func (s *Accounts) TracingFields() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("Accounts.ServiceType", s.ServiceType),
		attribute.String("Accounts.Perm", s.ServiceID),
		attribute.Int("Accounts.AccountIDs.Count", len(s.AccountIDs)),
	}
}

// ServiceVariant enumerates different types/kinds of external services.
// It replaces both the Type... and Kind... variables.
// Types and Kinds are exposed through AsKind and AsType functions
// so that usages relying on the particular string of Type vs Kind
// will continue to behave correctly.
// The Type... and Kind... variables are left in place to avoid edge-case issues and to support
// commits that come in while the switch to Variant is ongoing.
// The Type... and Kind... variables are turned from consts into vars and use
// the corresponding Variant's AsType()/AsKind() functions.
// Consolidating Type... and Kind... into a single enum should decrease the smell
// and increase the usability and maintainability of this code.
// Note that Go Packages and Modules seem to have been a victim of the confusion engendered by having both Type and Kind:
// There was `KindGoPackages` and `TypeGoModules`, both with the value of (case insensitivly) "gomodules".
// I think that they both refer to the same thing and can be merged into one variant.
// Olaf confirms that the name `...GoPackages` should be used to align naming conventions with the other `...Packages` variables.

// To add another external service variant
// 1. Add the name to the enum
// 2. Add an entry to the `variantValues` map, containing the appropriate values for `AsType`, `AsKind`, and the other values, if applicable
// 3. Use that Variant elsewhere in code, using the `AsType` and `AsKind` functions as necessary.

type ServiceVariant int64

const (
	// VariantAWSCodeCommit is the (api.ExternalRepoSpec).ServiceType value for AWS CodeCommit
	// repositories. The ServiceID value is the ARN (Amazon Resource Name) omitting the repository name
	// suffix (e.g., "arn:aws:codecommit:us-west-1:123456789:").
	VariantAWSCodeCommit ServiceVariant = iota

	// VariantBitbucketServer is the (api.ExternalRepoSpec).ServiceType value for Bitbucket Server projects. The
	// ServiceID value is the base URL to the Bitbucket Server instance.
	VariantBitbucketServer

	// VariantBitbucketCloud is the (api.ExternalRepoSpec).ServiceType value for Bitbucket Cloud projects. The
	// ServiceID value is the base URL to the Bitbucket Cloud.
	VariantBitbucketCloud

	// VariantGerrit is the (api.ExternalRepoSpec).ServiceType value for Gerrit projects.
	VariantGerrit

	// VariantGitHub is the (api.ExternalRepoSpec).ServiceType value for GitHub repositories. The ServiceID value
	// is the base URL to the GitHub instance (https://github.com or the GitHub Enterprise URL).
	VariantGitHub

	// VariantGitLab is the (api.ExternalRepoSpec).ServiceType value for GitLab projects. The ServiceID
	// value is the base URL to the GitLab instance (https://gitlab.com or self-hosted GitLab URL).
	VariantGitLab

	// VariantGitolite is the (api.ExternalRepoSpec).ServiceType value for Gitolite projects.
	VariantGitolite

	// VariantPerforce is the (api.ExternalRepoSpec).ServiceType value for Perforce projects.
	VariantPerforce

	// VariantPhabricator is the (api.ExternalRepoSpec).ServiceType value for Phabricator projects.
	VariantPhabricator

	// VariangGoPackages is the (api.ExternalRepoSpec).ServiceType value for Golang packages.
	// Duplicate of VariantGoModules?
	VariantGoPackages

	// VariantJVMPackages is the (api.ExternalRepoSpec).ServiceType value for Maven packages (Java/JVM ecosystem libraries).
	VariantJVMPackages

	// VariantPagure is the (api.ExternalRepoSpec).ServiceType value for Pagure projects.
	VariantPagure

	// VariantAzureDevOps is the (api.ExternalRepoSpec).ServiceType value for ADO projects.
	VariantAzureDevOps

	// VariantAzureDevOps is the (api.ExternalRepoSpec).ServiceType value for ADO projects.
	VariantSCIM

	// VariantNpmPackages is the (api.ExternalRepoSpec).ServiceType value for Npm packages (JavaScript/VariantScript ecosystem libraries).
	VariantNpmPackages

	// VariantGoModules is the (api.ExternalRepoSpec).ServiceType value for Go modules.
	VariantGoModules

	// VariantPythonPackages is the (api.ExternalRepoSpec).ServiceType value for Python packages.
	VariantPythonPackages

	// VariantRustPackages is the (api.ExternalRepoSpec).ServiceType value for Rust packages.
	VariantRustPackages

	// VariantRubyPackages is the (api.ExternalRepoSpec).ServiceType value for Ruby packages.
	VariantRubyPackages

	// VariantOther is the (api.ExternalRepoSpec).ServiceType value for other projects.
	VariantOther
)

type serviceVariantValues struct {
	AsKind                string
	AsType                string
	ConfigPrototype       interface{}
	WebhookURLPath        string
	SupportsRepoExclusion bool
}

var variantValues = map[ServiceVariant]serviceVariantValues{
	VariantAWSCodeCommit:   {AsKind: "AWSCODECOMMIT", AsType: "awscodecommit", ConfigPrototype: &schema.AWSCodeCommitConnection{}, SupportsRepoExclusion: true},
	VariantAzureDevOps:     {AsKind: "AZUREDEVOPS", AsType: "azuredevops", ConfigPrototype: &schema.AzureDevOpsConnection{}, SupportsRepoExclusion: true},
	VariantBitbucketCloud:  {AsKind: "BITBUCKETCLOUD", AsType: "bitbucketCloud", ConfigPrototype: &schema.BitbucketCloudConnection{}, WebhookURLPath: "bitbucket-cloud-webhooks", SupportsRepoExclusion: true},
	VariantBitbucketServer: {AsKind: "BITBUCKETSERVER", AsType: "bitbucketServer", ConfigPrototype: &schema.BitbucketServerConnection{}, WebhookURLPath: "bitbucket-server-webhooks", SupportsRepoExclusion: true},
	VariantGerrit:          {AsKind: "GERRIT", AsType: "gerrit", ConfigPrototype: &schema.GerritConnection{}},
	VariantGitHub:          {AsKind: "GITHUB", AsType: "github", ConfigPrototype: &schema.GitHubConnection{}, WebhookURLPath: "github-webhooks", SupportsRepoExclusion: true},
	VariantGitLab:          {AsKind: "GITLAB", AsType: "gitlab", ConfigPrototype: &schema.GitLabConnection{}, WebhookURLPath: "gitlab-webhooks", SupportsRepoExclusion: true},
	VariantGitolite:        {AsKind: "GITOLITE", AsType: "gitolite", ConfigPrototype: &schema.GitoliteConnection{}, SupportsRepoExclusion: true},
	VariantGoModules:       {AsKind: "GOMODULES", AsType: "goModules", ConfigPrototype: &schema.GoModulesConnection{}},
	VariantGoPackages:      {AsKind: "GOMODULES", AsType: "goModules", ConfigPrototype: &schema.GoModulesConnection{}},
	VariantJVMPackages:     {AsKind: "JVMPACKAGES", AsType: "jvmPackages", ConfigPrototype: &schema.JVMPackagesConnection{}},
	VariantNpmPackages:     {AsKind: "NPMPACKAGES", AsType: "npmPackages", ConfigPrototype: &schema.NpmPackagesConnection{}},
	VariantOther:           {AsKind: "OTHER", AsType: "other", ConfigPrototype: &schema.OtherExternalServiceConnection{}},
	VariantPagure:          {AsKind: "PAGURE", AsType: "pagure", ConfigPrototype: &schema.PagureConnection{}},
	VariantPerforce:        {AsKind: "PERFORCE", AsType: "perforce", ConfigPrototype: &schema.PerforceConnection{}},
	VariantPhabricator:     {AsKind: "PHABRICATOR", AsType: "phabricator", ConfigPrototype: &schema.PhabricatorConnection{}},
	VariantPythonPackages:  {AsKind: "PYTHONPACKAGES", AsType: "pythonPackages", ConfigPrototype: &schema.PythonPackagesConnection{}},
	VariantRubyPackages:    {AsKind: "RUBYPACKAGES", AsType: "rubyPackages", ConfigPrototype: &schema.RubyPackagesConnection{}},
	VariantRustPackages:    {AsKind: "RUSTPACKAGES", AsType: "rustPackages", ConfigPrototype: &schema.RustPackagesConnection{}},
	VariantSCIM:            {AsKind: "SCIM", AsType: "scim"},
}

func (sv ServiceVariant) AsKind() string {
	return variantValues[sv].AsKind
}

func (sv ServiceVariant) AsType() string {
	// Returns the values used in the external_service_type column of the repo table.
	return variantValues[sv].AsType
}

func (sv ServiceVariant) ConfigPrototype() interface{} {
	return variantValues[sv].ConfigPrototype
}

func (sv ServiceVariant) WebhookURLPath() string {
	return variantValues[sv].WebhookURLPath
}

func (sv ServiceVariant) SupportsRepoExclusion() bool {
	return variantValues[sv].SupportsRepoExclusion
}

// case-insensitive matching of an input string against the ServiceVariant kinds and types
// returns the matching ServiceVariant or an error if the given value is not a kind or type value
func ServiceVariantValueOf(input string) (ServiceVariant, error) {
	for variant, value := range variantValues {
		if strings.EqualFold(value.AsKind, input) || strings.EqualFold(value.AsType, input) {
			return variant, nil
		}
	}
	return 0, errors.Newf("no ServiceVariant found for %s", input)
}

// TODO: DEPRECATE/REMOVE ONCE CONVERSION TO Variants IS COMPLETE (2023-05-18)
// the Kind... and Type... variables have been superceded by the ServiceVariant enum
// DO NOT ADD MORE VARIABLES TO THE TYPE AND KIND VARIABLES
// instead, follow the instructions above for adding and using Variant variables

var (
	// The constants below represent the different kinds of external service we support and should be used
	// in preference to the Type values below.

	KindAWSCodeCommit   = VariantAWSCodeCommit.AsKind()
	KindBitbucketServer = VariantBitbucketServer.AsKind()
	KindBitbucketCloud  = VariantBitbucketCloud.AsKind()
	KindGerrit          = VariantGerrit.AsKind()
	KindGitHub          = VariantGitHub.AsKind()
	KindGitLab          = VariantGitLab.AsKind()
	KindGitolite        = VariantGitolite.AsKind()
	KindPerforce        = VariantPerforce.AsKind()
	KindPhabricator     = VariantPhabricator.AsKind()
	KindGoPackages      = VariantGoPackages.AsKind()
	KindJVMPackages     = VariantJVMPackages.AsKind()
	KindPythonPackages  = VariantPythonPackages.AsKind()
	KindRustPackages    = VariantRustPackages.AsKind()
	KindRubyPackages    = VariantRubyPackages.AsKind()
	KindNpmPackages     = VariantNpmPackages.AsKind()
	KindPagure          = VariantPagure.AsKind()
	KindAzureDevOps     = VariantAzureDevOps.AsKind()
	KindSCIM            = VariantSCIM.AsKind()
	KindOther           = VariantOther.AsKind()
)

var (
	// The constants below represent the values used for the external_service_type column of the repo table.

	// TypeAWSCodeCommit is the (api.ExternalRepoSpec).ServiceType value for AWS CodeCommit
	// repositories. The ServiceID value is the ARN (Amazon Resource Name) omitting the repository name
	// suffix (e.g., "arn:aws:codecommit:us-west-1:123456789:").
	TypeAWSCodeCommit = VariantAWSCodeCommit.AsType()

	// TypeBitbucketServer is the (api.ExternalRepoSpec).ServiceType value for Bitbucket Server projects. The
	// ServiceID value is the base URL to the Bitbucket Server instance.
	TypeBitbucketServer = VariantBitbucketServer.AsType()

	// TypeBitbucketCloud is the (api.ExternalRepoSpec).ServiceType value for Bitbucket Cloud projects. The
	// ServiceID value is the base URL to the Bitbucket Cloud.
	TypeBitbucketCloud = VariantBitbucketCloud.AsType()

	// TypeGerrit is the (api.ExternalRepoSpec).ServiceType value for Gerrit projects.
	TypeGerrit = VariantGerrit.AsType()

	// TypeGitHub is the (api.ExternalRepoSpec).ServiceType value for GitHub repositories. The ServiceID value
	// is the base URL to the GitHub instance (https://github.com or the GitHub Enterprise URL).
	TypeGitHub = VariantGitHub.AsType()

	// TypeGitLab is the (api.ExternalRepoSpec).ServiceType value for GitLab projects. The ServiceID
	// value is the base URL to the GitLab instance (https://gitlab.com or self-hosted GitLab URL).
	TypeGitLab = VariantGitLab.AsType()

	// TypeGitolite is the (api.ExternalRepoSpec).ServiceType value for Gitolite projects.
	TypeGitolite = VariantGitolite.AsType()

	// TypePerforce is the (api.ExternalRepoSpec).ServiceType value for Perforce projects.
	TypePerforce = VariantPerforce.AsType()

	// TypePhabricator is the (api.ExternalRepoSpec).ServiceType value for Phabricator projects.
	TypePhabricator = VariantPhabricator.AsType()

	// TypeJVMPackages is the (api.ExternalRepoSpec).ServiceType value for Maven packages (Java/JVM ecosystem libraries).
	TypeJVMPackages = VariantJVMPackages.AsType()

	// TypePagure is the (api.ExternalRepoSpec).ServiceType value for Pagure projects.
	TypePagure = VariantPagure.AsType()

	// TypeAzureDevOps is the (api.ExternalRepoSpec).ServiceType value for ADO projects.
	TypeAzureDevOps = VariantAzureDevOps.AsType()

	// TypeNpmPackages is the (api.ExternalRepoSpec).ServiceType value for Npm packages (JavaScript/TypeScript ecosystem libraries).
	TypeNpmPackages = VariantNpmPackages.AsType()

	// TypeGoModules is the (api.ExternalRepoSpec).ServiceType value for Go modules.
	TypeGoModules = VariantGoModules.AsType()

	// TypePythonPackages is the (api.ExternalRepoSpec).ServiceType value for Python packages.
	TypePythonPackages = VariantPythonPackages.AsType()

	// TypeRustPackages is the (api.ExternalRepoSpec).ServiceType value for Rust packages.
	TypeRustPackages = VariantRustPackages.AsType()

	// TypeRubyPackages is the (api.ExternalRepoSpec).ServiceType value for Ruby packages.
	TypeRubyPackages = VariantRubyPackages.AsType()

	// TypeOther is the (api.ExternalRepoSpec).ServiceType value for other projects.
	TypeOther = VariantOther.AsType()
)

// END TODO: DEPRECATE/REMOVE

// TODO: handle in a less smelly way with Variants
// KindToType returns a Type constant given a Kind
// It will panic when given an unknown kind
func KindToType(kind string) string {
	sv, err := ServiceVariantValueOf(kind)
	if err != nil {
		panic(fmt.Sprintf("unknown kind: %q", kind))
	}
	return sv.AsType()
}

// TODO: handle in a less smelly way with Variants
// TypeToKind returns a Kind constant given a Type
// It will panic when given an unknown type.
func TypeToKind(t string) string {
	sv, err := ServiceVariantValueOf(t)
	if err != nil {
		panic(fmt.Sprintf("unknown type: %q", t))
	}
	return sv.AsKind()
}

var (
	// Precompute these for use in ParseServiceType below since the constants are mixed case
	bbsLower    = strings.ToLower(VariantBitbucketServer.AsType())
	bbcLower    = strings.ToLower(VariantBitbucketCloud.AsType())
	jvmLower    = strings.ToLower(VariantJVMPackages.AsType())
	npmLower    = strings.ToLower(VariantNpmPackages.AsType())
	goLower     = strings.ToLower(VariantGoModules.AsType())
	pythonLower = strings.ToLower(VariantPythonPackages.AsType())
	rustLower   = strings.ToLower(VariantRustPackages.AsType())
	rubyLower   = strings.ToLower(VariantRubyPackages.AsType())
)

// ParseServiceType will return a ServiceType constant after doing a case insensitive match on s.
// It returns ("", false) if no match was found.
func ParseServiceType(s string) (string, bool) {
	sv, err := ServiceVariantValueOf(s)
	if err != nil {
		return "", false
	}
	return sv.AsType(), true
}

// ParseServiceKind will return a ServiceKind constant after doing a case insensitive match on s.
// It returns ("", false) if no match was found.
func ParseServiceKind(s string) (string, bool) {
	sv, err := ServiceVariantValueOf(s)
	if err != nil {
		return "", false
	}
	return sv.AsKind(), true
}

var supportsRepoExclusion = map[string]bool{
	VariantAWSCodeCommit.AsKind():   true,
	VariantBitbucketCloud.AsKind():  true,
	VariantBitbucketServer.AsKind(): true,
	VariantGitHub.AsKind():          true,
	VariantGitLab.AsKind():          true,
	VariantGitolite.AsKind():        true,
	VariantAzureDevOps.AsKind():     true,
}

// SupportsRepoExclusion returns true when given external service kind supports
// repo exclusion.
func SupportsRepoExclusion(extSvcKind string) bool {
	sv, err := ServiceVariantValueOf(extSvcKind)
	if err != nil {
		// no mechanism for percolating errors, so just return false
		return false
	}
	return sv.SupportsRepoExclusion()
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
	sv, err := ServiceVariantValueOf(kind)
	if err != nil {
		return nil, errors.Errorf("unknown external service kind %q", kind)
	}
	if sv.ConfigPrototype() == nil {
		return nil, errors.Errorf("no config prototype for %q", kind)
	}
	return sv.ConfigPrototype(), nil
}

const IDParam = "externalServiceID"

// WebhookURL returns an endpoint URL for the given external service. If the kind
// of external service does not support webhooks it returns an empty string.
func WebhookURL(kind string, externalServiceID int64, cfg any, externalURL string) (string, error) {
	sv, err := ServiceVariantValueOf(kind)
	if err != nil {
		return "", errors.Errorf("unknown external service kind %q", kind)
	}

	path := sv.WebhookURLPath()
	if path == "" {
		// If not a supported kind, bail out.
		return "", nil
	}

	u, err := url.Parse(externalURL)
	if err != nil {
		return "", err
	}
	u.Path = ".api/" + path
	q := u.Query()
	q.Set(IDParam, strconv.FormatInt(externalServiceID, 10))

	if sv == VariantBitbucketCloud {
		// Unlike other external service kinds, Bitbucket Cloud doesn't support
		// a shared secret defined as part of the webhook. As a result, we need
		// to include it as an explicit part of the URL that we construct.
		switch c := cfg.(type) {
		case *schema.BitbucketCloudConnection:
			q.Set("secret", url.QueryEscape(c.WebhookSecret))
		default:
			return "", errors.Newf("external service with id=%d claims to be a Bitbucket Cloud service, but the configuration is of type %T", externalServiceID, cfg)
		}
	}

	u.RawQuery = q.Encode()

	// eg. https://example.com/.api/github-webhooks?externalServiceID=1
	return u.String(), nil
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
	case *schema.AzureDevOpsConnection:
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
	// 1. Not defined: Some infinite, some limited, depending on code host.
	// 2. Defined and enabled: We use their defined limit.
	// 3. Defined and disabled: We use an infinite limiter.

	var limit rate.Limit
	switch c := config.(type) {
	case *schema.GitLabConnection:
		limit = rate.Inf
		if c != nil && c.RateLimit != nil {
			limit = limitOrInf(c.RateLimit.Enabled, c.RateLimit.RequestsPerHour)
		}
	case *schema.GitHubConnection:
		// Use an infinite rate limiter. GitHub has an external rate limiter we obey.
		limit = rate.Inf
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
		if c != nil && c.Maven.RateLimit != nil {
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

	// AbsFilePath is an optional field which is the absolute path to the
	// repository on the src git-serve server. Notably this is only
	// implemented for Sourcegraph App's implementation of src git-serve.
	AbsFilePath string
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
	case *schema.AzureDevOpsConnection:
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
		return VariantGoPackages.AsKind(), nil
	case *schema.JVMPackagesConnection:
		return VariantJVMPackages.AsKind(), nil
	case *schema.NpmPackagesConnection:
		return VariantNpmPackages.AsKind(), nil
	case *schema.PythonPackagesConnection:
		return VariantPythonPackages.AsKind(), nil
	case *schema.RustPackagesConnection:
		return VariantRustPackages.AsKind(), nil
	case *schema.RubyPackagesConnection:
		return VariantRubyPackages.AsKind(), nil
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
