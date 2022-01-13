package database

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	jsoniter "github.com/json-iterator/go"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/tidwall/gjson"
	"github.com/xeipuuv/gojsonschema"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// BeforeCreateExternalService (if set) is invoked as a hook prior to creating a
// new external service in the database.
var BeforeCreateExternalService func(context.Context, DB) error

type ExternalServiceStore interface {
	// Count counts all external services that satisfy the options (ignoring limit and offset).
	//
	// 🚨 SECURITY: The caller must ensure that the actor is a site admin or owner of the external service.
	Count(ctx context.Context, opt ExternalServicesListOptions) (int, error)

	// Create creates an external service.
	//
	// Since this method is used before the configuration server has started (search
	// for "EXTSVC_CONFIG_FILE") you must pass the conf.Get function in so that an
	// alternative can be used when the configuration server has not started,
	// otherwise a panic would occur once pkg/conf's deadlock detector determines a
	// deadlock occurred.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is a site admin or owner
	// of the external service. Otherwise, `es.NamespaceUserID` must be specified
	// (i.e. non-nil) for a user-added external service.
	//
	// 🚨 SECURITY: The value of `es.Unrestricted` is disregarded and will always be
	// recalculated based on whether "authorization" field is presented in
	// `es.Config`. For Sourcegraph Cloud, the `es.Unrestricted` will always be
	// false (i.e. enforce permissions).
	Create(ctx context.Context, confGet func() *conf.Unified, es *types.ExternalService) error

	// Delete deletes an external service.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is a site admin or owner of the external service.
	Delete(ctx context.Context, id int64) (err error)

	// DistinctKinds returns the distinct list of external services kinds that are stored in the database.
	DistinctKinds(ctx context.Context) ([]string, error)

	// GetAffiliatedSyncErrors returns the most recent sync failure message for each
	// external service affiliated with the supplied user. If the latest run did not
	// have an error, the string will be empty. We fetch external services owned by
	// the supplied user and if they are a site admin we additionally return site
	// level external services. We exclude cloud_default repos as they are never
	// synced.
	GetAffiliatedSyncErrors(ctx context.Context, u *types.User) (map[int64]string, error)

	// GetByID returns the external service for id.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is a site admin or owner of the external service.
	GetByID(ctx context.Context, id int64) (*types.ExternalService, error)

	// GetLastSyncError returns the error associated with the latest sync of the
	// supplied external service.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is a site admin or owner of the external service
	GetLastSyncError(ctx context.Context, id int64) (string, error)

	// GetSyncJobs gets all sync jobs
	GetSyncJobs(ctx context.Context) ([]*types.ExternalServiceSyncJob, error)

	// List returns external services under given namespace.
	// If no namespace is given, it returns all external services.
	//
	// 🚨 SECURITY: The caller must ensure one of the following:
	// 	- The actor is a site admin
	// 	- The opt.NamespaceUserID is same as authenticated user ID (i.e. actor.UID)
	List(ctx context.Context, opt ExternalServicesListOptions) ([]*types.ExternalService, error)

	// RepoCount returns the number of repos synced by the external service with the
	// given id.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is a site admin or owner of the external service.
	RepoCount(ctx context.Context, id int64) (int32, error)

	// SyncDue returns true if any of the supplied external services are due to sync
	// now or within given duration from now.
	SyncDue(ctx context.Context, intIDs []int64, d time.Duration) (bool, error)

	// Update updates an external service.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is a site admin,
	// or has the legitimate access to the external service (i.e. the owner).
	Update(ctx context.Context, ps []schema.AuthProviders, id int64, update *ExternalServiceUpdate) (err error)

	// Upsert updates or inserts the given ExternalServices.
	//
	// NOTE: Deletion of an external service via Upsert is not allowed. Use Delete()
	// instead.
	//
	// 🚨 SECURITY: The value of `es.Unrestricted` is disregarded and will always be
	// recalculated based on whether "authorization" field is presented in
	// `es.Config`. For Sourcegraph Cloud, the `es.Unrestricted` will always be
	// false (i.e. enforce permissions).
	Upsert(ctx context.Context, svcs ...*types.ExternalService) (err error)

	// ValidateConfig validates the given external service configuration, and returns a normalized
	// version of the configuration (i.e. valid JSON without comments).
	// A positive opt.ID indicates we are updating an existing service, adding a new one otherwise.
	ValidateConfig(ctx context.Context, opt ValidateExternalServiceConfigOptions) (normalized []byte, err error)

	WithEncryptionKey(key encryption.Key) ExternalServiceStore

	Transact(ctx context.Context) (ExternalServiceStore, error)
	With(other basestore.ShareableStore) ExternalServiceStore
	Done(err error) error
	basestore.ShareableStore
}

// An externalServiceStore stores external services and their configuration.
// Before updating or creating a new external service, validation is performed.
// The enterprise code registers additional validators at run-time and sets the
// global instance in stores.go
type externalServiceStore struct {
	*basestore.Store

	gitHubValidators          []func(*schema.GitHubConnection) error
	gitLabValidators          []func(*schema.GitLabConnection, []schema.AuthProviders) error
	bitbucketServerValidators []func(*schema.BitbucketServerConnection) error
	perforceValidators        []func(*schema.PerforceConnection) error

	key encryption.Key
}

func (e *externalServiceStore) copy() *externalServiceStore {
	return &externalServiceStore{
		Store:                     e.Store,
		key:                       e.key,
		gitHubValidators:          e.gitHubValidators,
		gitLabValidators:          e.gitLabValidators,
		bitbucketServerValidators: e.bitbucketServerValidators,
		perforceValidators:        e.perforceValidators,
	}
}

// ExternalServices instantiates and returns a new ExternalServicesStore with prepared statements.
var ExternalServices = NewExternalServiceStore

func NewExternalServiceStore(db dbutil.DB) ExternalServiceStore {
	return &externalServiceStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

func NewExternalServiceStoreWithValidators(
	db dbutil.DB,
	gitHubValidators []func(*schema.GitHubConnection) error,
	gitLabValidators []func(*schema.GitLabConnection, []schema.AuthProviders) error,
	bitbucketServerValidators []func(*schema.BitbucketServerConnection) error,
	perforceValidators []func(*schema.PerforceConnection) error,
) ExternalServiceStore {
	return &externalServiceStore{
		Store:                     basestore.NewWithDB(db, sql.TxOptions{}),
		gitHubValidators:          gitHubValidators,
		gitLabValidators:          gitLabValidators,
		bitbucketServerValidators: bitbucketServerValidators,
		perforceValidators:        perforceValidators,
	}
}

// ExternalServicesWith instantiates and returns a new ExternalServicesStore with prepared statements.
func ExternalServicesWith(other basestore.ShareableStore) ExternalServiceStore {
	return &externalServiceStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (e *externalServiceStore) With(other basestore.ShareableStore) ExternalServiceStore {
	s := e.copy()
	s.Store = e.Store.With(other)
	return s
}

func (e *externalServiceStore) WithEncryptionKey(key encryption.Key) ExternalServiceStore {
	s := e.copy()
	s.key = key
	return s
}

func (e *externalServiceStore) Transact(ctx context.Context) (ExternalServiceStore, error) {
	if Mocks.ExternalServices.Transact != nil {
		return Mocks.ExternalServices.Transact(ctx)
	}
	return e.transact(ctx)
}

func (e *externalServiceStore) transact(ctx context.Context) (*externalServiceStore, error) {
	txBase, err := e.Store.Transact(ctx)
	s := e.copy()
	s.Store = txBase
	return s, err
}

func (e *externalServiceStore) Done(err error) error {
	if Mocks.ExternalServices.Done != nil {
		return Mocks.ExternalServices.Done(err)
	}
	return e.Store.Done(err)
}

// ExternalServiceKinds contains a map of all supported kinds of
// external services.
var ExternalServiceKinds = map[string]ExternalServiceKind{
	extsvc.KindAWSCodeCommit:   {CodeHost: true, JSONSchema: schema.AWSCodeCommitSchemaJSON},
	extsvc.KindBitbucketCloud:  {CodeHost: true, JSONSchema: schema.BitbucketCloudSchemaJSON},
	extsvc.KindBitbucketServer: {CodeHost: true, JSONSchema: schema.BitbucketServerSchemaJSON},
	extsvc.KindGitHub:          {CodeHost: true, JSONSchema: schema.GitHubSchemaJSON},
	extsvc.KindGitLab:          {CodeHost: true, JSONSchema: schema.GitLabSchemaJSON},
	extsvc.KindGitolite:        {CodeHost: true, JSONSchema: schema.GitoliteSchemaJSON},
	extsvc.KindJVMPackages:     {CodeHost: true, JSONSchema: schema.JVMPackagesSchemaJSON},
	extsvc.KindOther:           {CodeHost: true, JSONSchema: schema.OtherExternalServiceSchemaJSON},
	extsvc.KindPagure:          {CodeHost: true, JSONSchema: schema.PagureSchemaJSON},
	extsvc.KindNPMPackages:     {CodeHost: true, JSONSchema: schema.NPMPackagesSchemaJSON},
	extsvc.KindPerforce:        {CodeHost: true, JSONSchema: schema.PerforceSchemaJSON},
	extsvc.KindPhabricator:     {CodeHost: true, JSONSchema: schema.PhabricatorSchemaJSON},
}

// ExternalServiceKind describes a kind of external service.
type ExternalServiceKind struct {
	// True if the external service can host repositories.
	CodeHost bool

	JSONSchema string // JSON Schema for the external service's configuration
}

// ExternalServicesListOptions contains options for listing external services.
type ExternalServicesListOptions struct {
	// When specified, only include external services with the given IDs.
	IDs []int64
	// When true, only include external services not under any namespace (i.e. owned
	// by all site admins), and values of ExcludeNamespaceUser, NamespaceUserID and
	// NamespaceOrgID are ignored.
	NoNamespace bool
	// When true, will exclude external services under any user namespace, and the
	// value of NamespaceUserID is ignored.
	ExcludeNamespaceUser bool
	// When specified, only include external services under given user namespace.
	NamespaceUserID int32
	// When specified, only include external services under given organization namespace.
	NamespaceOrgID int32
	// When specified, only include external services with given list of kinds.
	Kinds []string
	// When specified, only include external services with ID below this number
	// (because we're sorting results by ID in descending order).
	AfterID int64
	// Possible values are ASC or DESC. Defaults to DESC.
	OrderByDirection string
	// When true, will only return services that have the cloud_default flag set to
	// true.
	OnlyCloudDefault bool

	*LimitOffset

	// When true, only external services without has_webhooks set will be
	// returned. For use by ExternalServiceWebhookMigrator only.
	NoCachedWebhooks bool
	// When true, records will be locked. For use by
	// ExternalServiceWebhookMigrator only.
	ForUpdate bool
}

func (o ExternalServicesListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("deleted_at IS NULL")}
	if len(o.IDs) > 0 {
		ids := make([]*sqlf.Query, 0, len(o.IDs))
		for _, id := range o.IDs {
			ids = append(ids, sqlf.Sprintf("%s", id))
		}
		conds = append(conds, sqlf.Sprintf("id IN (%s)", sqlf.Join(ids, ",")))
	}

	if o.NoNamespace {
		conds = append(conds, sqlf.Sprintf(`namespace_user_id IS NULL AND namespace_org_id IS NULL`))
	} else {
		if o.ExcludeNamespaceUser {
			conds = append(conds, sqlf.Sprintf(`namespace_user_id IS NULL`))
		} else if o.NamespaceUserID > 0 {
			conds = append(conds, sqlf.Sprintf(`namespace_user_id = %d`, o.NamespaceUserID))
		}

		if o.NamespaceOrgID > 0 {
			conds = append(conds, sqlf.Sprintf(`namespace_org_id = %d`, o.NamespaceOrgID))
		}
	}
	if len(o.Kinds) > 0 {
		kinds := make([]*sqlf.Query, 0, len(o.Kinds))
		for _, kind := range o.Kinds {
			kinds = append(kinds, sqlf.Sprintf("%s", kind))
		}
		conds = append(conds, sqlf.Sprintf("kind IN (%s)", sqlf.Join(kinds, ",")))
	}
	if o.AfterID > 0 {
		conds = append(conds, sqlf.Sprintf(`id < %d`, o.AfterID))
	}
	if o.OnlyCloudDefault {
		conds = append(conds, sqlf.Sprintf("cloud_default = true"))
	}
	if o.NoCachedWebhooks {
		conds = append(conds, sqlf.Sprintf("has_webhooks IS NULL"))
	}
	return conds
}

type ValidateExternalServiceConfigOptions struct {
	// The ID of the external service, 0 is a valid value for not-yet-created external service.
	ExternalServiceID int64
	// The kind of external service.
	Kind string
	// The actual config of the external service.
	Config string
	// The list of authN providers configured on the instance.
	AuthProviders []schema.AuthProviders
	// If non zero, indicates the user that owns the external service.
	NamespaceUserID int32
	// If non zero, indicates the organization that owns the codehost connection.
	NamespaceOrgID int32
}

// IsSiteOwned returns true if the external service is owned by the site.
func (e *ValidateExternalServiceConfigOptions) IsSiteOwned() bool {
	return e.NamespaceUserID == 0 && e.NamespaceOrgID == 0
}

func (e *externalServiceStore) ValidateConfig(ctx context.Context, opt ValidateExternalServiceConfigOptions) (normalized []byte, err error) {
	ext, ok := ExternalServiceKinds[opt.Kind]
	if !ok {
		return nil, errors.Errorf("invalid external service kind: %s", opt.Kind)
	}

	// All configs must be valid JSON.
	// If this requirement is ever changed, you will need to update
	// serveExternalServiceConfigs to handle this case.

	sl := gojsonschema.NewSchemaLoader()
	sc, err := sl.Compile(gojsonschema.NewStringLoader(ext.JSONSchema))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to compile schema for external service of kind %q", opt.Kind)
	}

	normalized, err = jsonc.Parse(opt.Config)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to normalize JSON")
	}

	// Check for any redacted secrets, in
	// graphqlbackend/external_service.go:externalServiceByID() we call
	// svc.RedactConfigSecrets() replacing any secret fields in the JSON with
	// types.RedactedSecret, this is to prevent us leaking tokens that users add.
	// Here we check that the config we've been passed doesn't contain any redacted
	// secrets in order to avoid breaking configs by writing the redacted version to
	// the database. we should have called svc.UnredactConfig(oldSvc) before this
	// point, e.g. in the Update method of the ExternalServiceStore.
	if bytes.Contains(normalized, []byte(types.RedactedSecret)) {
		return nil, errors.Errorf(
			"unable to write external service config as it contains redacted fields, this is likely a bug rather than a problem with your config",
		)
	}

	// For user-added and org-added external services, we need to prevent them from using disallowed fields.
	if !opt.IsSiteOwned() {
		// We do not allow users to add external service other than GitHub.com and GitLab.com
		result := gjson.GetBytes(normalized, "url")
		baseURL, err := url.Parse(result.String())
		if err != nil {
			return nil, errors.Wrap(err, "parse base URL")
		}
		normalizedURL := extsvc.NormalizeBaseURL(baseURL).String()
		if normalizedURL != "https://github.com/" &&
			normalizedURL != "https://gitlab.com/" {
			return nil, errors.New("external service only allowed for https://github.com/ and https://gitlab.com/")
		}

		disallowedFields := []string{"repositoryPathPattern", "nameTransformations", "rateLimit"}
		results := gjson.GetManyBytes(normalized, disallowedFields...)
		for i, r := range results {
			if r.Exists() {
				return nil, errors.Errorf("field %q is not allowed in a user-added external service", disallowedFields[i])
			}
		}

		// Allow only create one external service per kind
		if err := e.validateSingleKindPerNamespace(ctx, opt.ExternalServiceID, opt.Kind, opt.NamespaceUserID, opt.NamespaceOrgID); err != nil {
			return nil, err
		}
	}

	res, err := sc.Validate(gojsonschema.NewBytesLoader(normalized))
	if err != nil {
		return nil, errors.Wrap(err, "unable to validate config against schema")
	}

	var errs *multierror.Error
	for _, err := range res.Errors() {
		e := err.String()
		// Remove `(root): ` from error formatting since these errors are
		// presented to users.
		e = strings.TrimPrefix(e, "(root): ")
		errs = multierror.Append(errs, errors.New(e))
	}

	// Extra validation not based on JSON Schema.
	switch opt.Kind {
	case extsvc.KindGitHub:
		var c schema.GitHubConnection
		if err = jsoniter.Unmarshal(normalized, &c); err != nil {
			return nil, err
		}
		err = e.validateGitHubConnection(ctx, opt.ExternalServiceID, &c)

	case extsvc.KindGitLab:
		var c schema.GitLabConnection
		if err = jsoniter.Unmarshal(normalized, &c); err != nil {
			return nil, err
		}
		err = e.validateGitLabConnection(ctx, opt.ExternalServiceID, &c, opt.AuthProviders)

	case extsvc.KindBitbucketServer:
		var c schema.BitbucketServerConnection
		if err = jsoniter.Unmarshal(normalized, &c); err != nil {
			return nil, err
		}
		err = e.validateBitbucketServerConnection(ctx, opt.ExternalServiceID, &c)

	case extsvc.KindBitbucketCloud:
		var c schema.BitbucketCloudConnection
		if err = jsoniter.Unmarshal(normalized, &c); err != nil {
			return nil, err
		}
		err = e.validateBitbucketCloudConnection(ctx, opt.ExternalServiceID, &c)

	case extsvc.KindPerforce:
		var c schema.PerforceConnection
		if err = jsoniter.Unmarshal(normalized, &c); err != nil {
			return nil, err
		}
		err = e.validatePerforceConnection(ctx, opt.ExternalServiceID, &c)

	case extsvc.KindOther:
		var c schema.OtherExternalServiceConnection
		if err = jsoniter.Unmarshal(normalized, &c); err != nil {
			return nil, err
		}
		err = validateOtherExternalServiceConnection(&c)
	}

	return normalized, multierror.Append(errs, err).ErrorOrNil()
}

// Neither our JSON schema library nor the Monaco editor we use supports
// object dependencies well, so we must validate here that repo items
// match the uri-reference format when url is set, instead of uri when
// it isn't.
func validateOtherExternalServiceConnection(c *schema.OtherExternalServiceConnection) error {
	parseRepo := url.Parse
	if c.Url != "" {
		// We ignore the error because this already validated by JSON Schema.
		baseURL, _ := url.Parse(c.Url)
		parseRepo = baseURL.Parse
	}

	for i, repo := range c.Repos {
		cloneURL, err := parseRepo(repo)
		if err != nil {
			return errors.Errorf(`repos.%d: %s`, i, err)
		}

		switch cloneURL.Scheme {
		case "git", "http", "https", "ssh":
			continue
		default:
			return errors.Errorf("repos.%d: scheme %q not one of git, http, https or ssh", i, cloneURL.Scheme)
		}
	}

	return nil
}

func (e *externalServiceStore) validateGitHubConnection(ctx context.Context, id int64, c *schema.GitHubConnection) error {
	err := new(multierror.Error)
	for _, validate := range e.gitHubValidators {
		err = multierror.Append(err, validate(c))
	}

	if c.Token == "" && c.GithubAppInstallationID == "" {
		err = multierror.Append(err, errors.New("at least one of token or githubAppInstallationID must be set"))
	}
	if c.Repos == nil && c.RepositoryQuery == nil && c.Orgs == nil {
		err = multierror.Append(err, errors.New("at least one of repositoryQuery, repos or orgs must be set"))
	}

	err = multierror.Append(err, e.validateDuplicateRateLimits(ctx, id, extsvc.KindGitHub, c))

	return err.ErrorOrNil()
}

func (e *externalServiceStore) validateGitLabConnection(ctx context.Context, id int64, c *schema.GitLabConnection, ps []schema.AuthProviders) error {
	err := new(multierror.Error)
	for _, validate := range e.gitLabValidators {
		err = multierror.Append(err, validate(c, ps))
	}

	err = multierror.Append(err, e.validateDuplicateRateLimits(ctx, id, extsvc.KindGitLab, c))

	return err.ErrorOrNil()
}

func (e *externalServiceStore) validateBitbucketServerConnection(ctx context.Context, id int64, c *schema.BitbucketServerConnection) error {
	err := new(multierror.Error)
	for _, validate := range e.bitbucketServerValidators {
		err = multierror.Append(err, validate(c))
	}

	if c.Repos == nil && c.RepositoryQuery == nil {
		err = multierror.Append(err, errors.New("at least one of repositoryQuery or repos must be set"))
	}

	err = multierror.Append(err, e.validateDuplicateRateLimits(ctx, id, extsvc.KindBitbucketServer, c))

	return err.ErrorOrNil()
}

func (e *externalServiceStore) validateBitbucketCloudConnection(ctx context.Context, id int64, c *schema.BitbucketCloudConnection) error {
	return e.validateDuplicateRateLimits(ctx, id, extsvc.KindBitbucketCloud, c)
}

func (e *externalServiceStore) validatePerforceConnection(ctx context.Context, id int64, c *schema.PerforceConnection) error {
	err := new(multierror.Error)
	for _, validate := range e.perforceValidators {
		err = multierror.Append(err, validate(c))
	}

	if c.Depots == nil {
		err = multierror.Append(err, errors.New("depots must be set"))
	}

	err = multierror.Append(err, e.validateDuplicateRateLimits(ctx, id, extsvc.KindPerforce, c))

	return err.ErrorOrNil()
}

// validateDuplicateRateLimits returns an error if given config has duplicated non-default rate limit
// with another external service for the same code host.
func (e *externalServiceStore) validateDuplicateRateLimits(ctx context.Context, id int64, kind string, parsedConfig interface{}) error {
	// Check if rate limit is already defined for this code host on another external service
	rlc, err := extsvc.GetLimitFromConfig(kind, parsedConfig)
	if err != nil {
		return errors.Wrap(err, "getting rate limit config")
	}

	// Default implies that no overriding rate limit has been set so it can't conflict with anything
	if rlc.IsDefault {
		return nil
	}

	baseURL := rlc.BaseURL
	opt := ExternalServicesListOptions{
		Kinds: []string{kind},
		LimitOffset: &LimitOffset{
			Limit: 500, // The number is randomly chosen
		},
	}
	for {
		svcs, err := e.List(ctx, opt)
		if err != nil {
			return errors.Wrap(err, "list")
		}
		if len(svcs) == 0 {
			break // No more results, exiting
		}
		opt.AfterID = svcs[len(svcs)-1].ID // Advance the cursor

		for _, svc := range svcs {
			rlc, err := extsvc.ExtractRateLimitConfig(svc.Config, svc.Kind, svc.DisplayName)
			if err != nil {
				return errors.Wrap(err, "extracting rate limit config")
			}
			if rlc.BaseURL == baseURL && svc.ID != id && !rlc.IsDefault {
				return errors.Errorf("existing external service, %q, already has a rate limit set", rlc.DisplayName)
			}
		}

		if len(svcs) < opt.Limit {
			break // Less results than limit means we've reached end
		}
	}
	return nil
}

// validateSingleKindPerNamespace returns an error if the user/org attempts to add more than one external service of the same kind.
func (e *externalServiceStore) validateSingleKindPerNamespace(ctx context.Context, id int64, kind string, userID int32, orgID int32) error {

	opt := ExternalServicesListOptions{
		Kinds: []string{kind},
		LimitOffset: &LimitOffset{
			Limit: 500, // The number is randomly chosen
		},
	}
	if userID > 0 {
		opt.NamespaceUserID = userID
	} else if orgID > 0 {
		opt.NamespaceOrgID = orgID
	}
	for {
		svcs, err := e.List(ctx, opt)
		if err != nil {
			return errors.Wrap(err, "list")
		}
		if len(svcs) == 0 {
			break // No more results, exiting
		}
		opt.AfterID = svcs[len(svcs)-1].ID // Advance the cursor

		// Fail if a service already exists that is not the current service
		for _, svc := range svcs {
			if svc.ID != id {
				return errors.Errorf("existing external service, %q, of same kind already added", svc.DisplayName)
			}
		}
		if len(svcs) < opt.Limit {
			break // Less results than limit means we've reached end
		}
	}
	return nil
}

// upsertAuthorizationToExternalService adds "authorization" field to the
// external service config when not yet present for GitHub and GitLab.
func upsertAuthorizationToExternalService(kind, config string) (string, error) {
	switch kind {
	case extsvc.KindGitHub:
		return jsonc.Edit(config, &schema.GitHubAuthorization{}, "authorization")

	case extsvc.KindGitLab:
		return jsonc.Edit(config,
			&schema.GitLabAuthorization{
				IdentityProvider: schema.IdentityProvider{
					Oauth: &schema.OAuthIdentity{
						Type: "oauth",
					},
				},
			},
			"authorization")
	}
	return config, nil
}

func (e *externalServiceStore) Create(ctx context.Context, confGet func() *conf.Unified, es *types.ExternalService) error {
	if Mocks.ExternalServices.Create != nil {
		return Mocks.ExternalServices.Create(ctx, confGet, es)
	}

	normalized, err := e.ValidateConfig(ctx, ValidateExternalServiceConfigOptions{
		Kind:            es.Kind,
		Config:          es.Config,
		AuthProviders:   confGet().AuthProviders,
		NamespaceUserID: es.NamespaceUserID,
		NamespaceOrgID:  es.NamespaceOrgID,
	})
	if err != nil {
		return err
	}

	// 🚨 SECURITY: For all GitHub and GitLab code host connections on Sourcegraph
	// Cloud, we always want to enforce repository permissions using OAuth to
	// prevent unexpected resource leaking.
	if envvar.SourcegraphDotComMode() {
		es.Config, err = upsertAuthorizationToExternalService(es.Kind, es.Config)
		if err != nil {
			return err
		}
	}

	es.CreatedAt = timeutil.Now()
	es.UpdatedAt = es.CreatedAt

	// Prior to saving the record, run a validation hook.
	if BeforeCreateExternalService != nil {
		if err := BeforeCreateExternalService(ctx, NewDB(e.Store.Handle().DB())); err != nil {
			return err
		}
	}

	// Ensure the calculated fields in the external service are up to date.
	if err := e.recalculateFields(es, string(normalized)); err != nil {
		return err
	}

	config, keyID, err := e.maybeEncryptConfig(ctx, es.Config)
	if err != nil {
		return err
	}

	return e.Store.Handle().DB().QueryRowContext(
		ctx,
		"INSERT INTO external_services(kind, display_name, config, encryption_key_id, created_at, updated_at, namespace_user_id, namespace_org_id, unrestricted, cloud_default, has_webhooks) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id",
		es.Kind, es.DisplayName, config, keyID, es.CreatedAt, es.UpdatedAt, nullInt32Column(es.NamespaceUserID), nullInt32Column(es.NamespaceOrgID), es.Unrestricted, es.CloudDefault, es.HasWebhooks,
	).Scan(&es.ID)
}

// maybeEncryptConfig encrypts and returns externals service config if an encryption.Key is configured
func (e *externalServiceStore) maybeEncryptConfig(ctx context.Context, config string) (string, string, error) {
	// encrypt the config before writing if we have a key configured
	var keyVersion string
	key := e.key
	if key == nil {
		key = keyring.Default().ExternalServiceKey
	}
	if key != nil {
		encrypted, err := key.Encrypt(ctx, []byte(config))
		if err != nil {
			return "", "", err
		}
		config = string(encrypted)
		version, err := key.Version(ctx)
		if err != nil {
			return "", "", err
		}
		keyVersion = version.JSON()
	}
	return config, keyVersion, nil
}

func (e *externalServiceStore) maybeDecryptConfig(ctx context.Context, config string, keyID string) (string, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "ExternalServiceStore.maybeDecryptConfig")
	defer span.Finish()

	if keyID == "" {
		// config is not encrypted, return plaintext
		return config, nil
	}
	key := e.key
	if key == nil {
		key = keyring.Default().ExternalServiceKey
	}
	if key == nil {
		return config, errors.Errorf("couldn't decrypt encrypted config, key is nil")
	}
	decryptSpan, ctx := ot.StartSpanFromContext(ctx, "key.Decrypt")
	decrypted, err := key.Decrypt(ctx, []byte(config))
	decryptSpan.Finish()
	if err != nil {
		return config, err
	}
	return decrypted.Secret(), nil
}

func (e *externalServiceStore) Upsert(ctx context.Context, svcs ...*types.ExternalService) (err error) {
	if Mocks.ExternalServices.Upsert != nil {
		return Mocks.ExternalServices.Upsert(ctx, svcs...)
	}
	if len(svcs) == 0 {
		return nil
	}

	for _, s := range svcs {
		// 🚨 SECURITY: For all GitHub and GitLab code host connections on Sourcegraph
		// Cloud, we always want to enforce repository permissions using OAuth to
		// prevent unexpected resource leaking.
		if envvar.SourcegraphDotComMode() {
			s.Config, err = upsertAuthorizationToExternalService(s.Kind, s.Config)
			if err != nil {
				return err
			}
		}

		if err := e.recalculateFields(s, s.Config); err != nil {
			return err
		}
	}

	tx, err := e.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Get the list services that are marked as deleted. We don't know at this point
	// whether they are marked as deleted in the DB too.
	var deleted []int64
	for _, es := range svcs {
		if es.ID != 0 && es.IsDeleted() {
			deleted = append(deleted, es.ID)
		}
	}

	// Fetch any services marked for deletion. list() only fetches non deleted
	// services so if we find anything here it indicates that we are marking a
	// service as deleted that is NOT deleted in the DB
	if len(deleted) > 0 {
		existing, err := tx.List(ctx, ExternalServicesListOptions{IDs: deleted})
		if err != nil {
			return errors.Wrap(err, "fetching services marked for deletion")
		}
		if len(existing) > 0 {
			// We found services marked for deletion that are currently not deleted in the
			// DB.
			return errors.New("deletion via Upsert() not allowed, use Delete()")
		}
	}

	q, err := tx.upsertExternalServicesQuery(ctx, svcs)
	if err != nil {
		return err
	}

	rows, err := tx.Query(ctx, q)
	if err != nil {
		return err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	i := 0
	for rows.Next() {
		var encryptionKeyID string
		err = rows.Scan(
			&svcs[i].ID,
			&svcs[i].Kind,
			&svcs[i].DisplayName,
			&svcs[i].Config,
			&svcs[i].CreatedAt,
			&dbutil.NullTime{Time: &svcs[i].UpdatedAt},
			&dbutil.NullTime{Time: &svcs[i].DeletedAt},
			&dbutil.NullTime{Time: &svcs[i].LastSyncAt},
			&dbutil.NullTime{Time: &svcs[i].NextSyncAt},
			&dbutil.NullInt32{N: &svcs[i].NamespaceUserID},
			&dbutil.NullInt32{N: &svcs[i].NamespaceOrgID},
			&svcs[i].Unrestricted,
			&svcs[i].CloudDefault,
			&encryptionKeyID,
			&dbutil.NullBool{B: svcs[i].HasWebhooks},
		)
		if err != nil {
			return err
		}

		svcs[i].Config, err = tx.maybeDecryptConfig(ctx, svcs[i].Config, encryptionKeyID)
		if err != nil {
			return err
		}

		i++
	}

	return nil
}

func (e *externalServiceStore) upsertExternalServicesQuery(ctx context.Context, svcs []*types.ExternalService) (*sqlf.Query, error) {
	vals := make([]*sqlf.Query, 0, len(svcs))
	for _, s := range svcs {
		config, keyID, err := e.maybeEncryptConfig(ctx, s.Config)
		if err != nil {
			return nil, err
		}
		vals = append(vals, sqlf.Sprintf(
			upsertExternalServicesQueryValueFmtstr,
			s.ID,
			s.Kind,
			s.DisplayName,
			config,
			keyID,
			s.CreatedAt.UTC(),
			s.UpdatedAt.UTC(),
			nullTimeColumn(s.DeletedAt),
			nullTimeColumn(s.LastSyncAt),
			nullTimeColumn(s.NextSyncAt),
			nullInt32Column(s.NamespaceUserID),
			nullInt32Column(s.NamespaceOrgID),
			s.Unrestricted,
			s.CloudDefault,
			s.HasWebhooks,
		))
	}

	return sqlf.Sprintf(
		upsertExternalServicesQueryFmtstr,
		sqlf.Join(vals, ",\n"),
	), nil
}

const upsertExternalServicesQueryValueFmtstr = `
  (COALESCE(NULLIF(%s, 0), (SELECT nextval('external_services_id_seq'))), UPPER(%s), %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
`

const upsertExternalServicesQueryFmtstr = `
-- source: internal/database/external_services.go:ExternalServiceStore.Upsert
INSERT INTO external_services (
  id,
  kind,
  display_name,
  config,
  encryption_key_id,
  created_at,
  updated_at,
  deleted_at,
  last_sync_at,
  next_sync_at,
  namespace_user_id,
  namespace_org_id,
  unrestricted,
  cloud_default,
  has_webhooks
)
VALUES %s
ON CONFLICT(id) DO UPDATE
SET
  kind               = UPPER(excluded.kind),
  display_name       = excluded.display_name,
  config             = excluded.config,
  encryption_key_id  = excluded.encryption_key_id,
  created_at         = excluded.created_at,
  updated_at         = excluded.updated_at,
  deleted_at         = excluded.deleted_at,
  last_sync_at       = excluded.last_sync_at,
  next_sync_at       = excluded.next_sync_at,
  namespace_user_id  = excluded.namespace_user_id,
  namespace_org_id   = excluded.namespace_org_id,
  unrestricted       = excluded.unrestricted,
  cloud_default      = excluded.cloud_default,
  has_webhooks       = excluded.has_webhooks
RETURNING
	id,
	kind,
	display_name,
	config,
	created_at,
	updated_at,
	deleted_at,
	last_sync_at,
	next_sync_at,
	namespace_user_id,
	namespace_org_id,
	unrestricted,
	cloud_default,
	encryption_key_id,
	has_webhooks
`

// ExternalServiceUpdate contains optional fields to update.
type ExternalServiceUpdate struct {
	DisplayName  *string
	Config       *string
	CloudDefault *bool
}

func (e *externalServiceStore) Update(ctx context.Context, ps []schema.AuthProviders, id int64, update *ExternalServiceUpdate) (err error) {
	var (
		normalized  []byte
		keyID       string
		hasWebhooks bool
	)
	if update.Config != nil {
		// Query to get the kind (which is immutable) so we can validate the new config.
		externalService, err := e.GetByID(ctx, id)
		if err != nil {
			return err
		}
		newSvc := types.ExternalService{
			Kind:   externalService.Kind,
			Config: *update.Config,
		}
		err = newSvc.UnredactConfig(externalService)
		if err != nil {
			return errors.Wrapf(err, "error unredacting config")
		}
		cfg, err := newSvc.Configuration()
		if err == nil {
			hasWebhooks = configurationHasWebhooks(cfg)
		} else {
			// Legacy configurations might not be valid JSON; in that case, they
			// also can't have webhooks, so we'll just log the issue and move
			// on.
			log15.Warn("cannot parse external service configuration as JSON", "err", err, "id", id)
			hasWebhooks = false
		}
		update.Config = &newSvc.Config

		normalized, err = e.ValidateConfig(ctx, ValidateExternalServiceConfigOptions{
			ExternalServiceID: id,
			Kind:              externalService.Kind,
			Config:            *update.Config,
			AuthProviders:     ps,
			NamespaceUserID:   externalService.NamespaceUserID,
			NamespaceOrgID:    externalService.NamespaceOrgID,
		})
		if err != nil {
			return err
		}

		// 🚨 SECURITY: For all GitHub and GitLab code host connections on Sourcegraph
		// Cloud, we always want to enforce repository permissions using OAuth to
		// prevent unexpected resource leaking.
		if envvar.SourcegraphDotComMode() {
			config, err := upsertAuthorizationToExternalService(externalService.Kind, *update.Config)
			if err != nil {
				return err
			}
			update.Config = &config
		}

		var config string
		config, keyID, err = e.maybeEncryptConfig(ctx, *update.Config)
		if err != nil {
			return err
		}
		update.Config = &config
	}

	execUpdate := func(ctx context.Context, tx dbutil.DB, update *sqlf.Query) error {
		q := sqlf.Sprintf("UPDATE external_services SET %s, updated_at=now() WHERE id=%d AND deleted_at IS NULL", update, id)
		res, err := tx.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return externalServiceNotFoundError{id: id}
		}
		return nil
	}
	tx, err := e.Store.Handle().Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if update.DisplayName != nil {
		if err := execUpdate(ctx, tx.DB(), sqlf.Sprintf("display_name=%s", update.DisplayName)); err != nil {
			return err
		}
	}

	if update.Config != nil {
		unrestricted := !envvar.SourcegraphDotComMode() && !gjson.GetBytes(normalized, "authorization").Exists()
		q := sqlf.Sprintf(`config = %s, encryption_key_id = %s, next_sync_at = NOW(), unrestricted = %s, has_webhooks = %s`, update.Config, keyID, unrestricted, hasWebhooks)
		if err := execUpdate(ctx, tx.DB(), q); err != nil {
			return err
		}
	}

	if update.CloudDefault != nil {
		if err := execUpdate(ctx, tx.DB(), sqlf.Sprintf("cloud_default=%s", update.CloudDefault)); err != nil {
			return err
		}
	}
	return nil
}

type externalServiceNotFoundError struct {
	id int64
}

func (e externalServiceNotFoundError) Error() string {
	return fmt.Sprintf("external service not found: %v", e.id)
}

func (e externalServiceNotFoundError) NotFound() bool {
	return true
}

func (e *externalServiceStore) Delete(ctx context.Context, id int64) (err error) {
	tx, err := e.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Create a temporary table where we'll store repos affected by the deletion of
	// the external service
	if err := tx.Exec(ctx, sqlf.Sprintf(`
CREATE TEMPORARY TABLE IF NOT EXISTS
    deleted_repos_temp(
    repo_id int
) ON COMMIT DROP`)); err != nil {
		return errors.Wrap(err, "creating temporary table")
	}

	// Delete external service <-> repo relationships, storing the affected repos
	if err := tx.Exec(ctx, sqlf.Sprintf(`
	WITH deleted AS (
	   DELETE FROM external_service_repos
	       WHERE external_service_id = %s
	       RETURNING repo_id
	)
	INSERT INTO deleted_repos_temp
	SELECT repo_id from deleted
`, id)); err != nil {
		return errors.Wrap(err, "populating temporary table")
	}

	// Soft delete orphaned repos
	if err := tx.Exec(ctx, sqlf.Sprintf(`
	UPDATE repo
	SET name       = soft_deleted_repository_name(name),
	   deleted_at = TRANSACTION_TIMESTAMP()
	WHERE deleted_at IS NULL
	 AND EXISTS (SELECT FROM deleted_repos_temp WHERE repo.id = deleted_repos_temp.repo_id)
	 AND NOT EXISTS (
	       SELECT FROM external_service_repos
	       WHERE repo_id = repo.id
	   );
`)); err != nil {
		return errors.Wrap(err, "cleaning up potentially orphaned repos")
	}

	// Clear temporary table in case delete is called multiple times within the same
	// transaction
	if err := tx.Exec(ctx, sqlf.Sprintf(`
    DELETE FROM deleted_repos_temp;
`)); err != nil {
		return errors.Wrap(err, "clearing temporary table")
	}

	// Soft delete external service
	res, err := tx.ExecResult(ctx, sqlf.Sprintf(`
	-- Soft delete external service
	UPDATE external_services
	SET deleted_at=TRANSACTION_TIMESTAMP()
	WHERE id = %s
	 AND deleted_at IS NULL;
	`, id))
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return externalServiceNotFoundError{id: id}
	}
	return nil
}

func (e *externalServiceStore) GetByID(ctx context.Context, id int64) (*types.ExternalService, error) {
	opt := ExternalServicesListOptions{
		IDs: []int64{id},
	}

	ess, err := e.List(ctx, opt)
	if err != nil {
		return nil, err
	}
	if len(ess) == 0 {
		return nil, externalServiceNotFoundError{id: id}
	}
	return ess[0], nil
}

func (e *externalServiceStore) GetSyncJobs(ctx context.Context) ([]*types.ExternalServiceSyncJob, error) {
	q := sqlf.Sprintf(`SELECT id, state, failure_message, started_at, finished_at, process_after, num_resets, external_service_id, num_failures
FROM external_service_sync_jobs ORDER BY started_at desc
`)

	rows, err := e.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	var jobs []*types.ExternalServiceSyncJob
	for rows.Next() {
		var job types.ExternalServiceSyncJob
		if err := rows.Scan(
			&job.ID,
			&job.State,
			&dbutil.NullString{S: &job.FailureMessage},
			&dbutil.NullTime{Time: &job.StartedAt},
			&dbutil.NullTime{Time: &job.FinishedAt},
			&dbutil.NullTime{Time: &job.ProcessAfter},
			&job.NumResets,
			&dbutil.NullInt64{N: &job.ExternalServiceID},
			&job.NumFailures,
		); err != nil {
			return nil, errors.Wrap(err, "scanning external service job row")
		}
		jobs = append(jobs, &job)
	}
	if rows.Err() != nil {
		return nil, errors.Wrap(err, "row scanning error")
	}

	return jobs, nil
}

func (e *externalServiceStore) GetLastSyncError(ctx context.Context, id int64) (string, error) {
	q := sqlf.Sprintf(`
SELECT failure_message from external_service_sync_jobs
WHERE external_service_id = %d
AND state IN ('completed','errored','failed')
ORDER BY finished_at DESC
LIMIT 1
`, id)

	lastError, _, err := basestore.ScanFirstNullString(e.Query(ctx, q))
	return lastError, err
}

func (e *externalServiceStore) GetAffiliatedSyncErrors(ctx context.Context, u *types.User) (map[int64]string, error) {
	if Mocks.ExternalServices.ListSyncErrors != nil {
		return Mocks.ExternalServices.ListSyncErrors(ctx)
	}
	if u == nil {
		return nil, errors.New("nil user")
	}
	q := sqlf.Sprintf(`
SELECT DISTINCT ON (es.id) es.id, essj.failure_message
FROM external_services es
         LEFT JOIN external_service_sync_jobs essj
                   ON es.id = essj.external_service_id
                       AND essj.state IN ('completed', 'errored', 'failed')
                       AND essj.finished_at IS NOT NULL
WHERE ((es.namespace_user_id = %s) OR (%s AND es.namespace_user_id IS NULL))
  AND es.deleted_at IS NULL
  AND NOT es.cloud_default
ORDER BY es.id, essj.finished_at DESC
`, u.ID, u.SiteAdmin)

	rows, err := e.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	messages := make(map[int64]string)

	for rows.Next() {
		var svcID int64
		var message sql.NullString
		if err := rows.Scan(&svcID, &message); err != nil {
			return nil, err
		}
		messages[svcID] = message.String
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (e *externalServiceStore) List(ctx context.Context, opt ExternalServicesListOptions) ([]*types.ExternalService, error) {
	if Mocks.ExternalServices.List != nil {
		return Mocks.ExternalServices.List(opt)
	}

	span, _ := ot.StartSpanFromContext(ctx, "ExternalServiceStore.list")
	defer span.Finish()

	if opt.OrderByDirection != "ASC" {
		opt.OrderByDirection = "DESC"
	}

	var forUpdate *sqlf.Query
	if opt.ForUpdate {
		forUpdate = sqlf.Sprintf("FOR UPDATE SKIP LOCKED")
	} else {
		forUpdate = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(`
		SELECT id, kind, display_name, config, encryption_key_id, created_at, updated_at, deleted_at, last_sync_at, next_sync_at, namespace_user_id, namespace_org_id, unrestricted, cloud_default, has_webhooks
		FROM external_services
		WHERE (%s)
		ORDER BY id `+opt.OrderByDirection+`
		%s
		%s`,
		sqlf.Join(opt.sqlConditions(), ") AND ("),
		opt.LimitOffset.SQL(),
		forUpdate,
	)

	rows, err := e.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keyIDs := make(map[int64]string)

	var results []*types.ExternalService
	for rows.Next() {
		var (
			h               types.ExternalService
			deletedAt       sql.NullTime
			lastSyncAt      sql.NullTime
			nextSyncAt      sql.NullTime
			namespaceUserID sql.NullInt32
			namespaceOrgID  sql.NullInt32
			keyID           string
			hasWebhooks     sql.NullBool
		)
		if err := rows.Scan(&h.ID, &h.Kind, &h.DisplayName, &h.Config, &keyID, &h.CreatedAt, &h.UpdatedAt, &deletedAt, &lastSyncAt, &nextSyncAt, &namespaceUserID, &namespaceOrgID, &h.Unrestricted, &h.CloudDefault, &hasWebhooks); err != nil {
			return nil, err
		}

		if deletedAt.Valid {
			h.DeletedAt = deletedAt.Time
		}
		if lastSyncAt.Valid {
			h.LastSyncAt = lastSyncAt.Time
		}
		if nextSyncAt.Valid {
			h.NextSyncAt = nextSyncAt.Time
		}
		if namespaceUserID.Valid {
			h.NamespaceUserID = namespaceUserID.Int32
		}
		if namespaceOrgID.Valid {
			h.NamespaceOrgID = namespaceOrgID.Int32
		}
		if hasWebhooks.Valid {
			h.HasWebhooks = &hasWebhooks.Bool
		}

		keyIDs[h.ID] = keyID

		results = append(results, &h)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Now we may need to decrypt config. Since each decrypt operation could make an
	// API call we should run them in parallel
	group, ctx := errgroup.WithContext(ctx)
	for i := range results {
		s := results[i]
		var groupErr error
		group.Go(func() error {
			keyID := keyIDs[s.ID]
			s.Config, groupErr = e.maybeDecryptConfig(ctx, s.Config, keyID)
			if groupErr != nil {
				return groupErr
			}
			return nil
		})
	}

	err = group.Wait()
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (e *externalServiceStore) DistinctKinds(ctx context.Context) ([]string, error) {
	q := sqlf.Sprintf(`
SELECT ARRAY_AGG(DISTINCT(kind)::TEXT)
FROM external_services
WHERE deleted_at IS NULL
`)

	var kinds []string
	err := e.QueryRow(ctx, q).Scan(pq.Array(&kinds))
	if err != nil {
		if err == sql.ErrNoRows {
			return []string{}, nil
		}
		return nil, err
	}

	return kinds, nil
}

func (e *externalServiceStore) Count(ctx context.Context, opt ExternalServicesListOptions) (int, error) {
	if Mocks.ExternalServices.Count != nil {
		return Mocks.ExternalServices.Count(ctx, opt)
	}

	q := sqlf.Sprintf("SELECT COUNT(*) FROM external_services WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := e.QueryRow(ctx, q).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (e *externalServiceStore) RepoCount(ctx context.Context, id int64) (int32, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM external_service_repos WHERE external_service_id = %s", id)
	var count int32

	if err := e.QueryRow(ctx, q).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (e *externalServiceStore) SyncDue(ctx context.Context, intIDs []int64, d time.Duration) (bool, error) {
	if len(intIDs) == 0 {
		return false, nil
	}
	ids := make([]*sqlf.Query, 0, len(intIDs))
	for _, id := range intIDs {
		ids = append(ids, sqlf.Sprintf("%s", id))
	}
	idFilter := sqlf.Sprintf("IN (%s)", sqlf.Join(ids, ","))
	deadline := time.Now().Add(d)

	q := sqlf.Sprintf(`
SELECT TRUE
WHERE EXISTS(
        SELECT
        FROM external_services
        WHERE id %s
          AND (
                next_sync_at IS NULL
                OR next_sync_at <= %s)
    )
   OR EXISTS(
        SELECT
        FROM external_service_sync_jobs
        WHERE external_service_id %s
          AND state IN ('queued', 'processing')
    );
`, idFilter, deadline, idFilter)

	v, exists, err := basestore.ScanFirstBool(e.Query(ctx, q))
	if err != nil {
		return false, err
	}
	return v && exists, nil
}

// recalculateFields updates the value of the external service fields that are
// calculated depending on the external service configuration, namely
// `Unrestricted` and `HasWebhooks`.
func (e *externalServiceStore) recalculateFields(es *types.ExternalService, rawConfig string) error {
	es.Unrestricted = !envvar.SourcegraphDotComMode() && !gjson.Get(es.Config, "authorization").Exists()

	hasWebhooks := false
	cfg, err := extsvc.ParseConfig(es.Kind, rawConfig)
	if err == nil {
		hasWebhooks = configurationHasWebhooks(cfg)
	} else {
		// Legacy configurations might not be valid JSON; in that case, they
		// also can't have webhooks, so we'll just log the issue and move on.
		log15.Warn("cannot parse external service configuration as JSON", "err", err, "id", es.ID)
	}
	es.HasWebhooks = &hasWebhooks

	return nil
}

func configurationHasWebhooks(config interface{}) bool {
	switch v := config.(type) {
	case *schema.GitHubConnection:
		return len(v.Webhooks) > 0
	case *schema.GitLabConnection:
		return len(v.Webhooks) > 0
	case *schema.BitbucketServerConnection:
		return v.WebhookSecret() != ""
	}

	return false
}

// MockExternalServices mocks the external services store.
type MockExternalServices struct {
	Create         func(ctx context.Context, confGet func() *conf.Unified, externalService *types.ExternalService) error
	ListSyncErrors func(ctx context.Context) (map[int64]string, error)
	List           func(opt ExternalServicesListOptions) ([]*types.ExternalService, error)
	Count          func(ctx context.Context, opt ExternalServicesListOptions) (int, error)
	Upsert         func(ctx context.Context, services ...*types.ExternalService) error
	Transact       func(ctx context.Context) (ExternalServiceStore, error)
	Done           func(error) error
}
