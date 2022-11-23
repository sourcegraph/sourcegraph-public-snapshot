package database

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/tidwall/gjson"
	"github.com/xeipuuv/gojsonschema"

	"github.com/sourcegraph/log"

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
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// BeforeCreateExternalService (if set) is invoked as a hook prior to creating a
// new external service in the database.
var BeforeCreateExternalService func(context.Context, ExternalServiceStore, *types.ExternalService) error

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

	// GetLatestSyncErrors returns the most recent sync failure message for
	// each external service. If the latest sync did not have an error, the
	// string will be empty. We exclude cloud_default external services as they
	// are never synced.
	GetLatestSyncErrors(ctx context.Context) (map[int64]string, error)

	// GetByID returns the external service for id.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is a site admin or owner of the external service.
	GetByID(ctx context.Context, id int64) (*types.ExternalService, error)

	// GetLastSyncError returns the error associated with the latest sync of the
	// supplied external service.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is a site admin or owner of the external service
	GetLastSyncError(ctx context.Context, id int64) (string, error)

	// GetSyncJobByID gets a sync job by its ID.
	GetSyncJobByID(ctx context.Context, id int64) (job *types.ExternalServiceSyncJob, err error)

	// GetSyncJobs gets all sync jobs.
	GetSyncJobs(ctx context.Context, opt ExternalServicesGetSyncJobsOptions) ([]*types.ExternalServiceSyncJob, error)

	// CountSyncJobs counts all sync jobs.
	CountSyncJobs(ctx context.Context, opt ExternalServicesGetSyncJobsOptions) (int64, error)

	// CancelSyncJob cancels a given sync job. It returns an error when the job was not
	// found or not in processing or queued state.
	CancelSyncJob(ctx context.Context, opts ExternalServicesCancelSyncJobOptions) error

	// UpdateSyncJobCounters persists only the sync job counters for the supplied job.
	UpdateSyncJobCounters(ctx context.Context, job *types.ExternalServiceSyncJob) error

	// List returns external services under given namespace.
	// If no namespace is given, it returns all external services.
	//
	// 🚨 SECURITY: The caller must ensure one of the following:
	// 	- The actor is a site admin
	// 	- The opt.NamespaceUserID is same as authenticated user ID (i.e. actor.UID)
	List(ctx context.Context, opt ExternalServicesListOptions) ([]*types.ExternalService, error)

	// ListRepos returns external service repos for given externalServiceID.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is a site admin or owner of the external service.
	ListRepos(ctx context.Context, opt ExternalServiceReposListOptions) ([]*types.ExternalServiceRepo, error)

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
	logger log.Logger
	*basestore.Store

	key encryption.Key
}

func (e *externalServiceStore) copy() *externalServiceStore {
	return &externalServiceStore{
		Store: e.Store,
		key:   e.key,
	}
}

// ExternalServicesWith instantiates and returns a new ExternalServicesStore with prepared statements.
func ExternalServicesWith(logger log.Logger, other basestore.ShareableStore) ExternalServiceStore {
	return &externalServiceStore{
		logger: logger,
		Store:  basestore.NewWithHandle(other.Handle()),
	}
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
	return e.transact(ctx)
}

func (e *externalServiceStore) transact(ctx context.Context) (*externalServiceStore, error) {
	txBase, err := e.Store.Transact(ctx)
	s := e.copy()
	s.Store = txBase
	return s, err
}

func (e *externalServiceStore) Done(err error) error {
	return e.Store.Done(err)
}

// ExternalServiceKinds contains a map of all supported kinds of
// external services.
var ExternalServiceKinds = map[string]ExternalServiceKind{
	extsvc.KindAWSCodeCommit:   {CodeHost: true, JSONSchema: schema.AWSCodeCommitSchemaJSON},
	extsvc.KindBitbucketCloud:  {CodeHost: true, JSONSchema: schema.BitbucketCloudSchemaJSON},
	extsvc.KindBitbucketServer: {CodeHost: true, JSONSchema: schema.BitbucketServerSchemaJSON},
	extsvc.KindGerrit:          {CodeHost: true, JSONSchema: schema.GerritSchemaJSON},
	extsvc.KindGitHub:          {CodeHost: true, JSONSchema: schema.GitHubSchemaJSON},
	extsvc.KindGitLab:          {CodeHost: true, JSONSchema: schema.GitLabSchemaJSON},
	extsvc.KindGitolite:        {CodeHost: true, JSONSchema: schema.GitoliteSchemaJSON},
	extsvc.KindGoPackages:      {CodeHost: true, JSONSchema: schema.GoModulesSchemaJSON},
	extsvc.KindJVMPackages:     {CodeHost: true, JSONSchema: schema.JVMPackagesSchemaJSON},
	extsvc.KindNpmPackages:     {CodeHost: true, JSONSchema: schema.NpmPackagesSchemaJSON},
	extsvc.KindOther:           {CodeHost: true, JSONSchema: schema.OtherExternalServiceSchemaJSON},
	extsvc.KindPagure:          {CodeHost: true, JSONSchema: schema.PagureSchemaJSON},
	extsvc.KindPerforce:        {CodeHost: true, JSONSchema: schema.PerforceSchemaJSON},
	extsvc.KindPhabricator:     {CodeHost: true, JSONSchema: schema.PhabricatorSchemaJSON},
	extsvc.KindPythonPackages:  {CodeHost: true, JSONSchema: schema.PythonPackagesSchemaJSON},
	extsvc.KindRustPackages:    {CodeHost: true, JSONSchema: schema.RustPackagesSchemaJSON},
	extsvc.KindRubyPackages:    {CodeHost: true, JSONSchema: schema.RubyPackagesSchemaJSON},
}

// ExternalServiceKind describes a kind of external service.
type ExternalServiceKind struct {
	// True if the external service can host repositories.
	CodeHost bool

	JSONSchema string // JSON Schema for the external service's configuration
}

type ExternalServiceReposListOptions ExternalServicesGetSyncJobsOptions

type ExternalServicesGetSyncJobsOptions struct {
	ExternalServiceID int64

	*LimitOffset
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
	// When specified, only include external services with that were updated after
	// the specified time.
	UpdatedAfter time.Time
	// Possible values are ASC or DESC. Defaults to DESC.
	OrderByDirection string
	// When true, will only return services that have the cloud_default flag set to
	// true.
	OnlyCloudDefault bool

	*LimitOffset

	// When true, soft-deleted external services will also be included in the results.
	IncludeDeleted bool
}

func (o ExternalServicesListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{}
	if !o.IncludeDeleted {
		conds = append(conds, sqlf.Sprintf("deleted_at IS NULL"))
	}
	if len(o.IDs) > 0 {
		conds = append(conds, sqlf.Sprintf("id = ANY(%s)", pq.Array(o.IDs)))
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
		conds = append(conds, sqlf.Sprintf("kind = ANY(%s)", pq.Array(o.Kinds)))
	}
	if o.AfterID > 0 {
		conds = append(conds, sqlf.Sprintf(`id < %d`, o.AfterID))
	}
	if !o.UpdatedAfter.IsZero() {
		conds = append(conds, sqlf.Sprintf(`updated_at > %s`, o.UpdatedAfter))
	}
	if o.OnlyCloudDefault {
		conds = append(conds, sqlf.Sprintf("cloud_default = true"))
	}
	if len(conds) == 0 {
		conds = append(conds, sqlf.Sprintf("TRUE"))
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

type ValidateExternalServiceConfigFunc = func(ctx context.Context, e ExternalServiceStore, opt ValidateExternalServiceConfigOptions) (normalized []byte, err error)

// ValidateExternalServiceConfig is the default non-enterprise version of our validation function
var ValidateExternalServiceConfig = MakeValidateExternalServiceConfigFunc(nil, nil, nil, nil)

func MakeValidateExternalServiceConfigFunc(gitHubValidators []func(*types.GitHubConnection) error, gitLabValidators []func(*schema.GitLabConnection, []schema.AuthProviders) error, bitbucketServerValidators []func(*schema.BitbucketServerConnection) error, perforceValidators []func(*schema.PerforceConnection) error) ValidateExternalServiceConfigFunc {
	return func(ctx context.Context, e ExternalServiceStore, opt ValidateExternalServiceConfigOptions) (normalized []byte, err error) {
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
		}

		res, err := sc.Validate(gojsonschema.NewBytesLoader(normalized))
		if err != nil {
			return nil, errors.Wrap(err, "unable to validate config against schema")
		}

		var errs error
		for _, err := range res.Errors() {
			errString := err.String()
			// Remove `(root): ` from error formatting since these errors are
			// presented to users.
			errString = strings.TrimPrefix(errString, "(root): ")
			errs = errors.Append(errs, errors.New(errString))
		}

		// Extra validation not based on JSON Schema.
		switch opt.Kind {
		case extsvc.KindGitHub:
			var c schema.GitHubConnection
			if err = jsoniter.Unmarshal(normalized, &c); err != nil {
				return nil, err
			}
			err = validateGitHubConnection(gitHubValidators, opt.ExternalServiceID, &c)

		case extsvc.KindGitLab:
			var c schema.GitLabConnection
			if err = jsoniter.Unmarshal(normalized, &c); err != nil {
				return nil, err
			}
			err = validateGitLabConnection(gitLabValidators, opt.ExternalServiceID, &c, opt.AuthProviders)

		case extsvc.KindBitbucketServer:
			var c schema.BitbucketServerConnection
			if err = jsoniter.Unmarshal(normalized, &c); err != nil {
				return nil, err
			}
			err = validateBitbucketServerConnection(bitbucketServerValidators, opt.ExternalServiceID, &c)

		case extsvc.KindBitbucketCloud:
			var c schema.BitbucketCloudConnection
			if err = jsoniter.Unmarshal(normalized, &c); err != nil {
				return nil, err
			}

		case extsvc.KindPerforce:
			var c schema.PerforceConnection
			if err = jsoniter.Unmarshal(normalized, &c); err != nil {
				return nil, err
			}
			err = validatePerforceConnection(perforceValidators, opt.ExternalServiceID, &c)

		case extsvc.KindOther:
			var c schema.OtherExternalServiceConnection
			if err = jsoniter.Unmarshal(normalized, &c); err != nil {
				return nil, err
			}
			err = validateOtherExternalServiceConnection(&c)
		}

		return normalized, errors.Append(errs, err)
	}
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

func validateGitHubConnection(githubValidators []func(*types.GitHubConnection) error, id int64, c *schema.GitHubConnection) error {
	var err error
	for _, validate := range githubValidators {
		err = errors.Append(err,
			validate(&types.GitHubConnection{
				URN:              extsvc.URN(extsvc.KindGitHub, id),
				GitHubConnection: c,
			}),
		)
	}

	if c.Token == "" && c.GithubAppInstallationID == "" {
		err = errors.Append(err, errors.New("at least one of token or githubAppInstallationID must be set"))
	}
	if c.Repos == nil && c.RepositoryQuery == nil && c.Orgs == nil {
		err = errors.Append(err, errors.New("at least one of repositoryQuery, repos or orgs must be set"))
	}
	return err
}

func validateGitLabConnection(gitLabValidators []func(*schema.GitLabConnection, []schema.AuthProviders) error, _ int64, c *schema.GitLabConnection, ps []schema.AuthProviders) error {
	var err error
	for _, validate := range gitLabValidators {
		err = errors.Append(err, validate(c, ps))
	}
	return err
}

func validateBitbucketServerConnection(bitbucketServerValidators []func(connection *schema.BitbucketServerConnection) error, _ int64, c *schema.BitbucketServerConnection) error {
	var err error
	for _, validate := range bitbucketServerValidators {
		err = errors.Append(err, validate(c))
	}

	if c.Repos == nil && c.RepositoryQuery == nil && c.ProjectKeys == nil {
		err = errors.Append(err, errors.New("at least one of: repositoryQuery, projectKeys, or repos must be set"))
	}
	return err
}

func validatePerforceConnection(perforceValidators []func(*schema.PerforceConnection) error, _ int64, c *schema.PerforceConnection) error {
	var err error
	for _, validate := range perforceValidators {
		err = errors.Append(err, validate(c))
	}

	if c.Depots == nil {
		err = errors.Append(err, errors.New("depots must be set"))
	}

	if strings.Contains(c.P4Passwd, ":") {
		err = errors.Append(err, errors.New("p4.passwd must not contain a colon. It must be the ticket generated by `p4 login -p`, not a full ticket from the `.p4tickets` file."))
	}

	return err
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
	rawConfig, err := es.Config.Decrypt(ctx)
	if err != nil {
		return err
	}

	normalized, err := ValidateExternalServiceConfig(ctx, e, ValidateExternalServiceConfigOptions{
		Kind:            es.Kind,
		Config:          rawConfig,
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
		rawConfig, err = upsertAuthorizationToExternalService(es.Kind, rawConfig)
		if err != nil {
			return err
		}

		es.Config.Set(rawConfig)
	}

	es.CreatedAt = timeutil.Now()
	es.UpdatedAt = es.CreatedAt

	// Prior to saving the record, run a validation hook.
	if BeforeCreateExternalService != nil {
		if err = BeforeCreateExternalService(ctx, NewDBWith(e.logger, e.Store).ExternalServices(), es); err != nil {
			return err
		}
	}

	// Ensure the calculated fields in the external service are up to date.
	if err := e.recalculateFields(es, string(normalized)); err != nil {
		return err
	}

	encryptedConfig, keyID, err := es.Config.Encrypt(ctx, e.getEncryptionKey())
	if err != nil {
		return err
	}

	return e.QueryRow(
		ctx,
		sqlf.Sprintf(
			createExternalServiceQueryFmtstr,
			es.Kind,
			es.DisplayName,
			encryptedConfig,
			keyID,
			es.CreatedAt,
			es.UpdatedAt,
			dbutil.NullInt32Column(es.NamespaceUserID),
			dbutil.NullInt32Column(es.NamespaceOrgID),
			es.Unrestricted,
			es.CloudDefault,
			es.HasWebhooks,
		),
	).Scan(&es.ID)
}

const createExternalServiceQueryFmtstr = `
INSERT INTO external_services
	(kind, display_name, config, encryption_key_id, created_at, updated_at, namespace_user_id, namespace_org_id, unrestricted, cloud_default, has_webhooks)
	VALUES(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING id
`

func (e *externalServiceStore) getEncryptionKey() encryption.Key {
	if e.key != nil {
		return e.key
	}

	return keyring.Default().ExternalServiceKey
}

func (e *externalServiceStore) Upsert(ctx context.Context, svcs ...*types.ExternalService) (err error) {
	if len(svcs) == 0 {
		return nil
	}

	authProviders := conf.Get().AuthProviders
	for _, s := range svcs {
		rawConfig, err := s.Config.Decrypt(ctx)
		if err != nil {
			return err
		}

		normalized, err := ValidateExternalServiceConfig(ctx, e, ValidateExternalServiceConfigOptions{
			Kind:            s.Kind,
			Config:          rawConfig,
			AuthProviders:   authProviders,
			NamespaceUserID: s.NamespaceUserID,
			NamespaceOrgID:  s.NamespaceOrgID,
		})
		if err != nil {
			return errors.Wrapf(err, "validating service of kind %q", s.Kind)
		}

		// 🚨 SECURITY: For all GitHub and GitLab code host connections on Sourcegraph
		// Cloud, we always want to enforce repository permissions using OAuth to
		// prevent unexpected resource leaking.
		if envvar.SourcegraphDotComMode() {
			rawConfig, err = upsertAuthorizationToExternalService(s.Kind, rawConfig)
			if err != nil {
				return err
			}

			s.Config.Set(rawConfig)
		}

		if err := e.recalculateFields(s, string(normalized)); err != nil {
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
		var encryptedConfig, keyID string
		err = rows.Scan(
			&svcs[i].ID,
			&svcs[i].Kind,
			&svcs[i].DisplayName,
			&encryptedConfig,
			&svcs[i].CreatedAt,
			&dbutil.NullTime{Time: &svcs[i].UpdatedAt},
			&dbutil.NullTime{Time: &svcs[i].DeletedAt},
			&dbutil.NullTime{Time: &svcs[i].LastSyncAt},
			&dbutil.NullTime{Time: &svcs[i].NextSyncAt},
			&dbutil.NullInt32{N: &svcs[i].NamespaceUserID},
			&dbutil.NullInt32{N: &svcs[i].NamespaceOrgID},
			&svcs[i].Unrestricted,
			&svcs[i].CloudDefault,
			&keyID,
			&dbutil.NullBool{B: svcs[i].HasWebhooks},
		)
		if err != nil {
			return err
		}

		svcs[i].Config = extsvc.NewEncryptedConfig(encryptedConfig, keyID, e.getEncryptionKey())
		i++
	}

	return nil
}

func (e *externalServiceStore) upsertExternalServicesQuery(ctx context.Context, svcs []*types.ExternalService) (*sqlf.Query, error) {
	vals := make([]*sqlf.Query, 0, len(svcs))
	for _, s := range svcs {
		encryptedConfig, keyID, err := s.Config.Encrypt(ctx, e.getEncryptionKey())
		if err != nil {
			return nil, err
		}
		vals = append(vals, sqlf.Sprintf(
			upsertExternalServicesQueryValueFmtstr,
			s.ID,
			s.Kind,
			s.DisplayName,
			encryptedConfig,
			keyID,
			s.CreatedAt.UTC(),
			s.UpdatedAt.UTC(),
			dbutil.NullTimeColumn(s.DeletedAt),
			dbutil.NullTimeColumn(s.LastSyncAt),
			dbutil.NullTimeColumn(s.NextSyncAt),
			dbutil.NullInt32Column(s.NamespaceUserID),
			dbutil.NullInt32Column(s.NamespaceOrgID),
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
	DisplayName    *string
	Config         *string
	CloudDefault   *bool
	TokenExpiresAt *time.Time
}

func (e *externalServiceStore) Update(ctx context.Context, ps []schema.AuthProviders, id int64, update *ExternalServiceUpdate) (err error) {
	var (
		normalized      []byte
		encryptedConfig string
		keyID           string
		hasWebhooks     bool
	)
	if update.Config != nil {
		rawConfig := *update.Config

		// Query to get the kind (which is immutable) so we can validate the new config.
		externalService, err := e.GetByID(ctx, id)
		if err != nil {
			return err
		}
		newSvc := types.ExternalService{
			Kind:   externalService.Kind,
			Config: extsvc.NewUnencryptedConfig(rawConfig),
		}

		if err := newSvc.UnredactConfig(ctx, externalService); err != nil {
			return errors.Wrapf(err, "error unredacting config")
		}
		unredactedConfig, err := newSvc.Config.Decrypt(ctx)
		if err != nil {
			return err
		}

		cfg, err := newSvc.Configuration(ctx)
		if err == nil {
			hasWebhooks = configurationHasWebhooks(cfg)
		} else {
			// Legacy configurations might not be valid JSON; in that case, they
			// also can't have webhooks, so we'll just log the issue and move
			// on.
			e.logger.Warn("cannot parse external service configuration as JSON", log.Error(err), log.Int64("id", id))
			hasWebhooks = false
		}

		normalized, err = ValidateExternalServiceConfig(ctx, e, ValidateExternalServiceConfigOptions{
			ExternalServiceID: id,
			Kind:              externalService.Kind,
			Config:            unredactedConfig,
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
			unredactedConfig, err = upsertAuthorizationToExternalService(externalService.Kind, unredactedConfig)
			if err != nil {
				return err
			}
			newSvc.Config.Set(unredactedConfig)
		}

		encryptedConfig, keyID, err = newSvc.Config.Encrypt(ctx, e.getEncryptionKey())
		if err != nil {
			return err
		}
	}

	// 4 is the number of fields of the ExternalServiceUpdate
	updates := make([]*sqlf.Query, 0, 4)

	if update.DisplayName != nil {
		updates = append(updates, sqlf.Sprintf("display_name = %s", update.DisplayName))
	}

	if update.Config != nil {
		unrestricted := !envvar.SourcegraphDotComMode() && !gjson.GetBytes(normalized, "authorization").Exists()
		updates = append(updates,
			sqlf.Sprintf(
				"config = %s, encryption_key_id = %s, next_sync_at = NOW(), unrestricted = %s, has_webhooks = %s",
				encryptedConfig, keyID, unrestricted, hasWebhooks,
			))
	}

	if update.CloudDefault != nil {
		updates = append(updates, sqlf.Sprintf("cloud_default = %s", update.CloudDefault))
	}

	if update.TokenExpiresAt != nil {
		updates = append(updates, sqlf.Sprintf("token_expires_at = %s", update.TokenExpiresAt))
	}

	if len(updates) == 0 {
		return nil
	}

	q := sqlf.Sprintf("UPDATE external_services SET %s, updated_at = NOW() WHERE id = %d AND deleted_at IS NULL", sqlf.Join(updates, ","), id)
	res, err := e.Store.Handle().ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
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

	// Load the external service *for update* so that no sync job can be created
	if err := tx.selectForUpdate(ctx, id); err != nil {
		return err
	}

	// Cancel all currently running sync jobs, *outside* the transaction.
	err = e.CancelSyncJob(ctx, ExternalServicesCancelSyncJobOptions{ExternalServiceID: id})
	if err != nil {
		return err
	}

	// Wait until all the sync jobs we just canceled are done executing to
	// ensure that we delete all repositories and no new ones are inserted.
	runningJobsCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	for {
		if err := runningJobsCtx.Err(); err != nil {
			return err
		}

		ok, err := e.hasRunningSyncJobs(runningJobsCtx, id)
		if err != nil {
			return err
		}

		if !ok {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

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

// selectForUpdate loads an external service with FOR UPDATE with the given ID
// and that is not deleted. It's used by Delete.
func (e *externalServiceStore) selectForUpdate(ctx context.Context, id int64) error {
	q := sqlf.Sprintf(
		`SELECT id FROM external_services WHERE id = %s AND deleted_at IS NULL FOR UPDATE`,
		id,
	)
	_, ok, err := basestore.ScanFirstInt(e.Query(ctx, q))
	if err != nil {
		return err
	}
	if !ok {
		return &externalServiceNotFoundError{id: id}
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

const getSyncJobsQueryFmtstr = `
SELECT
	id,
	state,
	failure_message,
	queued_at,
	started_at,
	finished_at,
	process_after,
	num_resets,
	external_service_id,
	num_failures,
	cancel,
	repos_synced,
	repo_sync_errors,
	repos_added,
	repos_modified,
	repos_unmodified,
	repos_deleted
FROM
	external_service_sync_jobs
WHERE %s
ORDER BY
	started_at DESC
%s
`

func (e *externalServiceStore) GetSyncJobs(ctx context.Context, opt ExternalServicesGetSyncJobsOptions) (_ []*types.ExternalServiceSyncJob, err error) {
	var preds []*sqlf.Query

	if opt.ExternalServiceID != 0 {
		preds = append(preds, sqlf.Sprintf("external_service_id = %s", opt.ExternalServiceID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	q := sqlf.Sprintf(getSyncJobsQueryFmtstr, sqlf.Join(preds, "AND"), opt.LimitOffset.SQL())

	rows, err := e.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = basestore.CloseRows(rows, err)
	}()

	jobs := make([]*types.ExternalServiceSyncJob, 0)
	for rows.Next() {
		var job types.ExternalServiceSyncJob
		if err := scanExternalServiceSyncJob(rows, &job); err != nil {
			return nil, errors.Wrap(err, "scanning external service job row")
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

const countSyncJobsQueryFmtstr = `
SELECT
	COUNT(*)
FROM
	external_service_sync_jobs
WHERE %s
`

func (e *externalServiceStore) CountSyncJobs(ctx context.Context, opt ExternalServicesGetSyncJobsOptions) (int64, error) {
	var preds []*sqlf.Query

	if opt.ExternalServiceID != 0 {
		preds = append(preds, sqlf.Sprintf("external_service_id = %s", opt.ExternalServiceID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	q := sqlf.Sprintf(countSyncJobsQueryFmtstr, sqlf.Join(preds, "AND"))

	count, _, err := basestore.ScanFirstInt64(e.Query(ctx, q))
	return count, err
}

type errSyncJobNotFound struct {
	id, externalServiceID int64
}

func (e errSyncJobNotFound) Error() string {
	if e.id != 0 {
		return fmt.Sprintf("sync job with id %d not found", e.id)
	} else if e.externalServiceID != 0 {
		return fmt.Sprintf("sync job with external service id %d not found", e.externalServiceID)
	}
	return "sync job not found"
}

func (errSyncJobNotFound) NotFound() bool {
	return true
}

func (e *externalServiceStore) GetSyncJobByID(ctx context.Context, id int64) (*types.ExternalServiceSyncJob, error) {
	q := sqlf.Sprintf(getSyncJobsQueryFmtstr, sqlf.Sprintf("id = %s", id), (&LimitOffset{Limit: 1}).SQL())

	var job types.ExternalServiceSyncJob
	if err := scanExternalServiceSyncJob(e.QueryRow(ctx, q), &job); err != nil {
		if err == sql.ErrNoRows {
			return nil, &errSyncJobNotFound{id: id}
		}
		return nil, errors.Wrap(err, "scanning external service job row")
	}

	return &job, nil
}

// UpdateSyncJobCounters persists only the sync job counters for the supplied job.
func (e *externalServiceStore) UpdateSyncJobCounters(ctx context.Context, job *types.ExternalServiceSyncJob) error {
	q := sqlf.Sprintf(updateSyncJobQueryFmtstr, job.ReposSynced, job.RepoSyncErrors, job.ReposAdded, job.ReposModified, job.ReposUnmodified, job.ReposDeleted, job.ID)
	result, err := e.ExecResult(ctx, q)
	if err != nil {
		return errors.Wrap(err, "updating sync job counters")
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "checking affected rows")
	}
	if affected == 0 {
		return &errSyncJobNotFound{id: job.ID}
	}
	return nil
}

const updateSyncJobQueryFmtstr = `
UPDATE external_service_sync_jobs
SET
	repos_synced = %d,
	repo_sync_errors = %d,
	repos_added = %d,
	repos_modified = %d,
	repos_unmodified = %d,
	repos_deleted = %d
WHERE
    id = %d
`

func scanExternalServiceSyncJob(sc dbutil.Scanner, job *types.ExternalServiceSyncJob) error {
	return sc.Scan(
		&job.ID,
		&job.State,
		&dbutil.NullString{S: &job.FailureMessage},
		&job.QueuedAt,
		&dbutil.NullTime{Time: &job.StartedAt},
		&dbutil.NullTime{Time: &job.FinishedAt},
		&dbutil.NullTime{Time: &job.ProcessAfter},
		&job.NumResets,
		&dbutil.NullInt64{N: &job.ExternalServiceID},
		&job.NumFailures,
		&job.Cancel,
		&job.ReposSynced,
		&job.RepoSyncErrors,
		&job.ReposAdded,
		&job.ReposModified,
		&job.ReposUnmodified,
		&job.ReposDeleted,
	)
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

type ExternalServicesCancelSyncJobOptions struct {
	ID                int64
	ExternalServiceID int64
}

func buildCancelSyncJobQuery(opts ExternalServicesCancelSyncJobOptions) (*sqlf.Query, error) {
	var conds []*sqlf.Query
	if opts.ID != 0 {
		conds = append(conds, sqlf.Sprintf("id = %s", opts.ID))
	}
	if opts.ExternalServiceID != 0 {
		conds = append(conds, sqlf.Sprintf("external_service_id = %s", opts.ExternalServiceID))
	}

	if len(conds) == 0 {
		return nil, errors.New("not enough conditions given to build query to cancel external service sync job")
	}

	now := timeutil.Now()
	q := sqlf.Sprintf(`
UPDATE
	external_service_sync_jobs
SET
	cancel = TRUE,
	-- If the sync job is still queued, we directly abort, otherwise we keep the
	-- state, so the worker can do teardown and, at some point, mark it failed itself.
	state = CASE WHEN external_service_sync_jobs.state = 'processing' THEN external_service_sync_jobs.state ELSE 'canceled' END,
	finished_at = CASE WHEN external_service_sync_jobs.state = 'processing' THEN external_service_sync_jobs.finished_at ELSE %s END
WHERE
	%s
	AND
	state IN ('queued', 'processing')
	AND
	cancel IS FALSE
`, now, sqlf.Join(conds, " AND "))

	return q, nil
}

func (e *externalServiceStore) CancelSyncJob(ctx context.Context, opts ExternalServicesCancelSyncJobOptions) error {
	q, err := buildCancelSyncJobQuery(opts)
	if err != nil {
		return err
	}

	res, err := e.ExecResult(ctx, q)
	if err != nil {
		return err
	}
	af, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if opts.ID != 0 && af != 1 {
		return &errSyncJobNotFound{id: opts.ID, externalServiceID: opts.ExternalServiceID}
	}
	return nil
}

func (e *externalServiceStore) hasRunningSyncJobs(ctx context.Context, id int64) (bool, error) {
	q := sqlf.Sprintf(`
SELECT 1
FROM external_service_sync_jobs
WHERE
	external_service_id = %s
	AND
	state IN ('queued', 'processing')
LIMIT 1
`, id)

	_, ok, err := basestore.ScanFirstInt(e.Query(ctx, q))
	return ok, err
}

func (e *externalServiceStore) GetLatestSyncErrors(ctx context.Context) (map[int64]string, error) {
	q := sqlf.Sprintf(`
SELECT DISTINCT ON (es.id) es.id, essj.failure_message
FROM external_services es
         LEFT JOIN external_service_sync_jobs essj
                   ON es.id = essj.external_service_id
                       AND essj.state IN ('completed', 'errored', 'failed')
                       AND essj.finished_at IS NOT NULL
WHERE es.deleted_at IS NULL AND NOT es.cloud_default
ORDER BY es.id, essj.finished_at DESC
`)

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
	span, _ := ot.StartSpanFromContext(ctx, "ExternalServiceStore.list")
	defer span.Finish()

	if opt.OrderByDirection != "ASC" {
		opt.OrderByDirection = "DESC"
	}

	q := sqlf.Sprintf(`
		SELECT
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
			has_webhooks,
			token_expires_at
		FROM external_services
		WHERE (%s)
		ORDER BY id `+opt.OrderByDirection+`
		%s`,
		sqlf.Join(opt.sqlConditions(), ") AND ("),
		opt.LimitOffset.SQL(),
	)

	rows, err := e.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*types.ExternalService
	for rows.Next() {
		var (
			h               types.ExternalService
			deletedAt       sql.NullTime
			lastSyncAt      sql.NullTime
			nextSyncAt      sql.NullTime
			namespaceUserID sql.NullInt32
			namespaceOrgID  sql.NullInt32
			encryptedConfig string
			keyID           string
			hasWebhooks     sql.NullBool
			tokenExpiresAt  sql.NullTime
		)
		if err := rows.Scan(
			&h.ID,
			&h.Kind,
			&h.DisplayName,
			&encryptedConfig,
			&keyID,
			&h.CreatedAt,
			&h.UpdatedAt,
			&deletedAt,
			&lastSyncAt,
			&nextSyncAt,
			&namespaceUserID,
			&namespaceOrgID,
			&h.Unrestricted,
			&h.CloudDefault,
			&hasWebhooks,
			&tokenExpiresAt,
		); err != nil {
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
		if tokenExpiresAt.Valid {
			h.TokenExpiresAt = &tokenExpiresAt.Time
		}
		h.Config = extsvc.NewEncryptedConfig(encryptedConfig, keyID, e.getEncryptionKey())

		results = append(results, &h)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (e *externalServiceStore) ListRepos(ctx context.Context, opt ExternalServiceReposListOptions) ([]*types.ExternalServiceRepo, error) {
	span, _ := ot.StartSpanFromContext(ctx, "ExternalServiceStore.listRepos")
	defer span.Finish()

	predicate := sqlf.Sprintf("TRUE")

	if opt.ExternalServiceID != 0 {
		predicate = sqlf.Sprintf("external_service_id = %s", opt.ExternalServiceID)
	}

	q := sqlf.Sprintf(`
SELECT
	external_service_id,
	repo_id,
	clone_url,
	user_id,
	org_id,
	created_at
FROM external_service_repos
WHERE %s
%s`,
		predicate,
		opt.LimitOffset.SQL(),
	)

	rows, err := e.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []*types.ExternalServiceRepo
	for rows.Next() {
		var (
			repo   types.ExternalServiceRepo
			userID sql.NullInt32
			orgID  sql.NullInt32
		)

		if err := rows.Scan(
			&repo.ExternalServiceID,
			&repo.RepoID,
			&repo.CloneURL,
			&userID,
			&orgID,
			&repo.CreatedAt,
		); err != nil {
			return nil, err
		}

		if userID.Valid {
			repo.UserID = userID.Int32
		}
		if orgID.Valid {
			repo.OrgID = orgID.Int32
		}

		repos = append(repos, &repo)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return repos, nil
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
	es.Unrestricted = !envvar.SourcegraphDotComMode() && !gjson.Get(rawConfig, "authorization").Exists()

	hasWebhooks := false
	cfg, err := extsvc.ParseConfig(es.Kind, rawConfig)
	if err == nil {
		hasWebhooks = configurationHasWebhooks(cfg)
	} else {
		// Legacy configurations might not be valid JSON; in that case, they
		// also can't have webhooks, so we'll just log the issue and move on.
		e.logger.Warn("cannot parse external service configuration as JSON", log.Error(err), log.Int64("id", es.ID))
	}
	es.HasWebhooks = &hasWebhooks

	return nil
}

func configurationHasWebhooks(config any) bool {
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
