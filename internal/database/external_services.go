package database

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	jsoniter "github.com/json-iterator/go"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/xeipuuv/gojsonschema"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// An ExternalServiceStore stores external services and their configuration.
// Before updating or creating a new external service, validation is performed.
// The enterprise code registers additional validators at run-time and sets the
// global instance in stores.go
type ExternalServiceStore struct {
	*basestore.Store

	GitHubValidators          []func(*schema.GitHubConnection) error
	GitLabValidators          []func(*schema.GitLabConnection, []schema.AuthProviders) error
	BitbucketServerValidators []func(*schema.BitbucketServerConnection) error

	key encryption.Key

	// PreCreateExternalService (if set) is invoked as a hook prior to creating a
	// new external service in the database.
	PreCreateExternalService func(context.Context) error

	mu sync.Mutex
}

// ExternalServices instantiates and returns a new ExternalServicesStore with prepared statements.
func ExternalServices(db dbutil.DB) *ExternalServiceStore {
	return &ExternalServiceStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

// NewExternalServicesStoreWithDB instantiates and returns a new ExternalServicesStore with prepared statements.
func ExternalServicesWith(other basestore.ShareableStore) *ExternalServiceStore {
	return &ExternalServiceStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (e *ExternalServiceStore) With(other basestore.ShareableStore) *ExternalServiceStore {
	return &ExternalServiceStore{Store: e.Store.With(other), key: e.key}
}

func (e *ExternalServiceStore) WithEncryptionKey(key encryption.Key) *ExternalServiceStore {
	return &ExternalServiceStore{Store: e.Store, key: key}
}

func (e *ExternalServiceStore) Transact(ctx context.Context) (*ExternalServiceStore, error) {
	txBase, err := e.Store.Transact(ctx)
	return &ExternalServiceStore{Store: txBase, key: e.key}, err
}

// ensureStore instantiates a basestore.Store if necessary, using the dbconn.Global handle.
// This function ensures access to dbconn happens after the rest of the code or tests have
// initialized it.
func (e *ExternalServiceStore) ensureStore() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.Store == nil {
		e.Store = basestore.NewWithDB(dbconn.Global, sql.TxOptions{})
	}
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
	extsvc.KindPerforce:        {CodeHost: true, JSONSchema: schema.PerforceSchemaJSON},
	extsvc.KindPhabricator:     {CodeHost: true, JSONSchema: schema.PhabricatorSchemaJSON},
	extsvc.KindOther:           {CodeHost: true, JSONSchema: schema.OtherExternalServiceSchemaJSON},
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
	// by all site admins), and value of NamespaceUserID is ignored.
	NoNamespace bool
	// When specified, only include external services under given user namespace.
	NamespaceUserID int32
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
		conds = append(conds, sqlf.Sprintf(`namespace_user_id IS NULL`))
	} else if o.NamespaceUserID > 0 {
		conds = append(conds, sqlf.Sprintf(`namespace_user_id = %d`, o.NamespaceUserID))
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
}

// ValidateConfig validates the given external service configuration, and returns a normalized
// version of the configuration (i.e. valid JSON without comments).
// A positive opt.ID indicates we are updating an existing service, adding a new one otherwise.
func (e *ExternalServiceStore) ValidateConfig(ctx context.Context, opt ValidateExternalServiceConfigOptions) (normalized []byte, err error) {
	ext, ok := ExternalServiceKinds[opt.Kind]
	if !ok {
		return nil, fmt.Errorf("invalid external service kind: %s", opt.Kind)
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

	// Check for any redacted secrets, in graphqlbackend/external_service.go:externalServiceByID() we call
	// svc.RedactConfigSecrets() replacing any secret fields in the JSON with types.RedactedSecret, this is to
	// prevent us leaking tokens that users add. Here we check that the config we've been passed doesn't
	// contain any redacted secrets in order to avoid breaking configs by writing the redacted version to
	// the database. we should have called svc.UnredactConfig(oldSvc) before this point, eg in the Update
	// method of the ExternalServiceStore. talk to @arussellsaw or the cloud team if you have any questions
	if bytes.Contains(normalized, []byte(types.RedactedSecret)) {
		return nil, errors.Errorf(
			"unable to write external service config as it contains redacted fields, this is likely a bug rather than a problem with your config",
		)
	}

	// For user-added external services, we need to prevent them from using disallowed fields.
	if opt.NamespaceUserID > 0 {
		// We do not allow users to add external service other than GitHub.com and GitLab.com
		result := gjson.GetBytes(normalized, "url")
		baseURL, err := url.Parse(result.String())
		if err != nil {
			return nil, errors.Wrap(err, "parse base URL")
		}
		normalizedURL := extsvc.NormalizeBaseURL(baseURL).String()
		if normalizedURL != "https://github.com/" &&
			normalizedURL != "https://gitlab.com/" {
			return nil, errors.New("users are only allowed to add external service for https://github.com/ and https://gitlab.com/")
		}

		disallowedFields := []string{"repositoryPathPattern", "nameTransformations", "rateLimit"}
		results := gjson.GetManyBytes(normalized, disallowedFields...)
		for i, r := range results {
			if r.Exists() {
				return nil, errors.Errorf("field %q is not allowed in a user-added external service", disallowedFields[i])
			}
		}

		// A user can only create one external service per kind
		if err := e.validateSingleKindPerUser(ctx, opt.ExternalServiceID, opt.Kind, opt.NamespaceUserID); err != nil {
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
			return fmt.Errorf(`repos.%d: %s`, i, err)
		}

		switch cloneURL.Scheme {
		case "git", "http", "https", "ssh":
			continue
		default:
			return fmt.Errorf("repos.%d: scheme %q not one of git, http, https or ssh", i, cloneURL.Scheme)
		}
	}

	return nil
}

func (e *ExternalServiceStore) validateGitHubConnection(ctx context.Context, id int64, c *schema.GitHubConnection) error {
	err := new(multierror.Error)
	for _, validate := range e.GitHubValidators {
		err = multierror.Append(err, validate(c))
	}

	if c.Repos == nil && c.RepositoryQuery == nil && c.Orgs == nil {
		err = multierror.Append(err, errors.New("at least one of repositoryQuery, repos or orgs must be set"))
	}

	err = multierror.Append(err, e.validateDuplicateRateLimits(ctx, id, extsvc.KindGitHub, c))

	return err.ErrorOrNil()
}

func (e *ExternalServiceStore) validateGitLabConnection(ctx context.Context, id int64, c *schema.GitLabConnection, ps []schema.AuthProviders) error {
	err := new(multierror.Error)
	for _, validate := range e.GitLabValidators {
		err = multierror.Append(err, validate(c, ps))
	}

	err = multierror.Append(err, e.validateDuplicateRateLimits(ctx, id, extsvc.KindGitLab, c))

	return err.ErrorOrNil()
}

func (e *ExternalServiceStore) validateBitbucketServerConnection(ctx context.Context, id int64, c *schema.BitbucketServerConnection) error {
	err := new(multierror.Error)
	for _, validate := range e.BitbucketServerValidators {
		err = multierror.Append(err, validate(c))
	}

	if c.Repos == nil && c.RepositoryQuery == nil {
		err = multierror.Append(err, errors.New("at least one of repositoryQuery or repos must be set"))
	}

	err = multierror.Append(err, e.validateDuplicateRateLimits(ctx, id, extsvc.KindBitbucketServer, c))

	return err.ErrorOrNil()
}

func (e *ExternalServiceStore) validateBitbucketCloudConnection(ctx context.Context, id int64, c *schema.BitbucketCloudConnection) error {
	return e.validateDuplicateRateLimits(ctx, id, extsvc.KindBitbucketCloud, c)
}

// validateDuplicateRateLimits returns an error if given config has duplicated non-default rate limit
// with another external service for the same code host.
func (e *ExternalServiceStore) validateDuplicateRateLimits(ctx context.Context, id int64, kind string, parsedConfig interface{}) error {
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
				return fmt.Errorf("existing external service, %q, already has a rate limit set", rlc.DisplayName)
			}
		}

		if len(svcs) < opt.Limit {
			break // Less results than limit means we've reached end
		}
	}
	return nil
}

// validateSingleKindPerUser returns an error if the user attempts to add more than one external service of the same kind.
func (e *ExternalServiceStore) validateSingleKindPerUser(ctx context.Context, id int64, kind string, userID int32) error {
	opt := ExternalServicesListOptions{
		Kinds: []string{kind},
		LimitOffset: &LimitOffset{
			Limit: 500, // The number is randomly chosen
		},
		NamespaceUserID: userID,
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
				return fmt.Errorf("existing external service, %q, of same kind already added", svc.DisplayName)
			}
		}
		if len(svcs) < opt.Limit {
			break // Less results than limit means we've reached end
		}
	}
	return nil
}

// Create creates an external service.
//
// Since this method is used before the configuration server has started
// (search for "EXTSVC_CONFIG_FILE") you must pass the conf.Get function in so
// that an alternative can be used when the configuration server has not
// started, otherwise a panic would occur once pkg/conf's deadlock detector
// determines a deadlock occurred.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin or owner of the external service.
// Otherwise, `es.NamespaceUserID` must be specified (i.e. non-nil) for
// a user-added external service.
//
// 🚨 SECURITY: The value of `es.Unrestricted` is disregarded and will always
// be recalculated based on whether `"authorization"` is presented in `es.Config`.
func (e *ExternalServiceStore) Create(ctx context.Context, confGet func() *conf.Unified, es *types.ExternalService) error {
	if Mocks.ExternalServices.Create != nil {
		return Mocks.ExternalServices.Create(ctx, confGet, es)
	}
	e.ensureStore()

	normalized, err := e.ValidateConfig(ctx, ValidateExternalServiceConfigOptions{
		Kind:            es.Kind,
		Config:          es.Config,
		AuthProviders:   confGet().AuthProviders,
		NamespaceUserID: es.NamespaceUserID,
	})
	if err != nil {
		return err
	}

	// NOTE: For GitHub and GitLab user code host connections on Sourcegraph Cloud,
	//  we always want to enforce repository permissions using OAuth to prevent
	//  unexpected resource leaking.
	if envvar.SourcegraphDotComMode() && es.NamespaceUserID != 0 {
		switch es.Kind {
		case extsvc.KindGitHub:
			var c schema.GitHubConnection
			if err = jsoniter.Unmarshal(normalized, &c); err != nil {
				return err
			}

			if c.Authorization == nil {
				c.Authorization = &schema.GitHubAuthorization{}

				normalized, err = jsoniter.Marshal(c)
				if err != nil {
					return err
				}
			}

		case extsvc.KindGitLab:
			var c schema.GitLabConnection
			if err = jsoniter.Unmarshal(normalized, &c); err != nil {
				return err
			}

			if c.Authorization == nil {
				c.Authorization = &schema.GitLabAuthorization{
					IdentityProvider: schema.IdentityProvider{
						Oauth: &schema.OAuthIdentity{
							Type: "oauth",
						},
					},
				}

				normalized, err = jsoniter.Marshal(c)
				if err != nil {
					return err
				}
			}
		}

		// We expect users to edit code host connections via our UI so no JSON with
		// comments should appear, thus OK to set config as normalized.
		es.Config = string(normalized)
	}

	es.CreatedAt = timeutil.Now()
	es.UpdatedAt = es.CreatedAt

	// Prior to saving the record, run a validation hook.
	if e.PreCreateExternalService != nil {
		if err := e.PreCreateExternalService(ctx); err != nil {
			return err
		}
	}
	es.Unrestricted = !gjson.GetBytes(normalized, "authorization").Exists()

	config, keyID, err := e.maybeEncryptConfig(ctx, es.Config)
	if err != nil {
		return err
	}

	return e.Store.Handle().DB().QueryRowContext(
		ctx,
		"INSERT INTO external_services(kind, display_name, config, encryption_key_id, created_at, updated_at, namespace_user_id, unrestricted, cloud_default) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id",
		es.Kind, es.DisplayName, config, keyID, es.CreatedAt, es.UpdatedAt, nullInt32Column(es.NamespaceUserID), es.Unrestricted, es.CloudDefault,
	).Scan(&es.ID)
}

// maybeEncryptConfig encrypts and returns externals service config if an encryption.Key is configured
func (e *ExternalServiceStore) maybeEncryptConfig(ctx context.Context, config string) (string, string, error) {
	// encrypt the config before writing if we have a key configured
	var (
		keyIdent string
		key      = keyring.Default().ExternalServiceKey
	)
	if e.key != nil {
		key = e.key
	}
	if key != nil {
		encrypted, err := key.Encrypt(ctx, []byte(config))
		if err != nil {
			return "", "", err
		}
		config = string(encrypted)
		keyIdent, err = key.ID(ctx)
		if err != nil {
			return "", "", err
		}
	}
	return config, keyIdent, nil
}

func (e *ExternalServiceStore) maybeDecryptConfig(ctx context.Context, config string, keyIdent string) (string, error) {
	if keyIdent == "" {
		// config is not encrypted, return plaintext
		return config, nil
	}
	var key = keyring.Default().ExternalServiceKey
	if e.key != nil {
		key = e.key
	}
	if key == nil {
		return config, fmt.Errorf("couldn't decrypt encrypted config, key is nil")
	}
	decrypted, err := key.Decrypt(ctx, []byte(config))
	if err != nil {
		return config, err
	}
	return decrypted.Secret(), nil
}

// Upsert updates or inserts the given ExternalServices.
//
// 🚨 SECURITY: The value of `Unrestricted` field is disregarded and will always
// be recalculated based on whether `"authorization"` is presented in `Config`.
func (e *ExternalServiceStore) Upsert(ctx context.Context, svcs ...*types.ExternalService) error {
	if Mocks.ExternalServices.Upsert != nil {
		return Mocks.ExternalServices.Upsert(ctx, svcs...)
	}
	if len(svcs) == 0 {
		return nil
	}
	e.ensureStore()

	for _, s := range svcs {
		s.Unrestricted = !gjson.Get(s.Config, "authorization").Exists()
	}

	q, err := e.upsertExternalServicesQuery(ctx, svcs)
	if err != nil {
		return err
	}
	rows, err := e.Query(ctx, q)
	if err != nil {
		return err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	i := 0
	for rows.Next() {
		var keyIdent string
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
			&svcs[i].Unrestricted,
			&svcs[i].CloudDefault,
			&keyIdent,
		)
		if err != nil {
			return err
		}

		svcs[i].Config, err = e.maybeDecryptConfig(ctx, svcs[i].Config, keyIdent)
		if err != nil {
			return err
		}

		i++
	}

	return err
}

func (e *ExternalServiceStore) upsertExternalServicesQuery(ctx context.Context, svcs []*types.ExternalService) (*sqlf.Query, error) {
	vals := make([]*sqlf.Query, 0, len(svcs))
	for _, s := range svcs {
		config, keyIdent, err := e.maybeEncryptConfig(ctx, s.Config)
		if err != nil {
			return nil, err
		}
		vals = append(vals, sqlf.Sprintf(
			upsertExternalServicesQueryValueFmtstr,
			s.ID,
			s.Kind,
			s.DisplayName,
			config,
			keyIdent,
			s.CreatedAt.UTC(),
			s.UpdatedAt.UTC(),
			nullTimeColumn(s.DeletedAt),
			nullTimeColumn(s.LastSyncAt),
			nullTimeColumn(s.NextSyncAt),
			nullInt32Column(s.NamespaceUserID),
			s.Unrestricted,
			s.CloudDefault,
		))
	}

	return sqlf.Sprintf(
		upsertExternalServicesQueryFmtstr,
		sqlf.Join(vals, ",\n"),
	), nil
}

const upsertExternalServicesQueryValueFmtstr = `
  (COALESCE(NULLIF(%s, 0), (SELECT nextval('external_services_id_seq'))), UPPER(%s), %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
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
  unrestricted,
  cloud_default
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
  unrestricted       = excluded.unrestricted,
  cloud_default      = excluded.cloud_default
RETURNING *
`

// ExternalServiceUpdate contains optional fields to update.
type ExternalServiceUpdate struct {
	DisplayName  *string
	Config       *string
	CloudDefault *bool
}

// Update updates an external service.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin,
// or has the legitimate access to the external service (i.e. the owner).
func (e *ExternalServiceStore) Update(ctx context.Context, ps []schema.AuthProviders, id int64, update *ExternalServiceUpdate) (err error) {
	if Mocks.ExternalServices.Update != nil {
		return Mocks.ExternalServices.Update(ctx, ps, id, update)
	}
	e.ensureStore()

	var (
		normalized []byte
		keyIdent   string
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
		update.Config = &newSvc.Config

		normalized, err = e.ValidateConfig(ctx, ValidateExternalServiceConfigOptions{
			ExternalServiceID: id,
			Kind:              externalService.Kind,
			Config:            *update.Config,
			AuthProviders:     ps,
			NamespaceUserID:   externalService.NamespaceUserID,
		})
		if err != nil {
			return err
		}
		var config string
		config, keyIdent, err = e.maybeEncryptConfig(ctx, *update.Config)
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
		unrestricted := !gjson.GetBytes(normalized, "authorization").Exists()
		q := sqlf.Sprintf(`config = %s, encryption_key_id = %s, next_sync_at = NOW(), unrestricted = %s`, update.Config, keyIdent, unrestricted)
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

// Delete deletes an external service.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin or owner of the external service.
func (e *ExternalServiceStore) Delete(ctx context.Context, id int64) error {
	if Mocks.ExternalServices.Delete != nil {
		return Mocks.ExternalServices.Delete(ctx, id)
	}
	e.ensureStore()

	res, err := e.Handle().DB().ExecContext(ctx, "UPDATE external_services SET deleted_at=now() WHERE id=$1 AND deleted_at IS NULL", id)
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

// GetByID returns the external service for id.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin or owner of the external service.
func (e *ExternalServiceStore) GetByID(ctx context.Context, id int64) (*types.ExternalService, error) {
	if Mocks.ExternalServices.GetByID != nil {
		return Mocks.ExternalServices.GetByID(id)
	}
	e.ensureStore()

	opt := ExternalServicesListOptions{
		IDs: []int64{id},
	}

	ess, err := e.list(ctx, opt)
	if err != nil {
		return nil, err
	}
	if len(ess) == 0 {
		return nil, externalServiceNotFoundError{id: id}
	}
	return ess[0], nil
}

// GetLastSyncError returns the error associated with the latest sync of the
// supplied external service.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin or owner of the external service
func (e *ExternalServiceStore) GetLastSyncError(ctx context.Context, id int64) (string, error) {
	if Mocks.ExternalServices.GetLastSyncError != nil {
		return Mocks.ExternalServices.GetLastSyncError(id)
	}
	e.ensureStore()

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

// GetAffiliatedSyncErrors returns the most recent failure message for each
// external service. If the latest run did not have an error, it will be excluded
// from the map. We fetch external services owned by the supplied user and if
// they are a site admin we also return site level external services.
func (e *ExternalServiceStore) GetAffiliatedSyncErrors(ctx context.Context, u *types.User) (map[int64]string, error) {
	if Mocks.ExternalServices.ListSyncErrors != nil {
		return Mocks.ExternalServices.ListSyncErrors(ctx)
	}
	if u == nil {
		return nil, errors.New("nil user")
	}
	q := sqlf.Sprintf(`SELECT DISTINCT ON(external_service_id) external_service_id, failure_message
FROM external_service_sync_jobs sj
JOIN external_services es ON sj.external_service_id = es.id
WHERE
  state IN ('completed','errored','failed')
  AND finished_at IS NOT NULL
  AND ((es.namespace_user_id = %s) OR (%s AND es.namespace_user_id IS NULL))
ORDER BY external_service_id, finished_at DESC
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
		if message.Valid {
			messages[svcID] = message.String
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

// List returns external services under given namespace.
// If no namespace is given, it returns all external services.
//
// 🚨 SECURITY: The caller must ensure one of the following:
// 	- The actor is a site admin
// 	- The opt.NamespaceUserID is same as authenticated user ID (i.e. actor.UID)
func (e *ExternalServiceStore) List(ctx context.Context, opt ExternalServicesListOptions) ([]*types.ExternalService, error) {
	if Mocks.ExternalServices.List != nil {
		return Mocks.ExternalServices.List(opt)
	}
	e.ensureStore()

	return e.list(ctx, opt)
}

// DistinctKinds returns the distinct list of external services kinds that are stored in the database.
func (e *ExternalServiceStore) DistinctKinds(ctx context.Context) ([]string, error) {
	e.ensureStore()

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

func (e *ExternalServiceStore) list(ctx context.Context, opt ExternalServicesListOptions) ([]*types.ExternalService, error) {
	if opt.OrderByDirection != "ASC" {
		opt.OrderByDirection = "DESC"
	}

	q := sqlf.Sprintf(`
		SELECT id, kind, display_name, config, encryption_key_id, created_at, updated_at, deleted_at, last_sync_at, next_sync_at, namespace_user_id, unrestricted, cloud_default
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
			keyIdent        string
		)
		if err := rows.Scan(&h.ID, &h.Kind, &h.DisplayName, &h.Config, &keyIdent, &h.CreatedAt, &h.UpdatedAt, &deletedAt, &lastSyncAt, &nextSyncAt, &namespaceUserID, &h.Unrestricted, &h.CloudDefault); err != nil {
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

		h.Config, err = e.maybeDecryptConfig(ctx, h.Config, keyIdent)
		if err != nil {
			return nil, err
		}
		results = append(results, &h)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// Count counts all external services that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin or owner of the external service.
func (e *ExternalServiceStore) Count(ctx context.Context, opt ExternalServicesListOptions) (int, error) {
	if Mocks.ExternalServices.Count != nil {
		return Mocks.ExternalServices.Count(ctx, opt)
	}
	e.ensureStore()

	q := sqlf.Sprintf("SELECT COUNT(*) FROM external_services WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := e.QueryRow(ctx, q).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// RepoCount returns the number of repos synced by the external service with the
// given id.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin or owner of the external service.
func (e *ExternalServiceStore) RepoCount(ctx context.Context, id int64) (int32, error) {
	e.ensureStore()

	q := sqlf.Sprintf("SELECT COUNT(*) FROM external_service_repos WHERE external_service_id = %s", id)
	var count int32

	if err := e.QueryRow(ctx, q).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// MockExternalServices mocks the external services store.
type MockExternalServices struct {
	Create           func(ctx context.Context, confGet func() *conf.Unified, externalService *types.ExternalService) error
	Delete           func(ctx context.Context, id int64) error
	GetByID          func(id int64) (*types.ExternalService, error)
	GetLastSyncError func(id int64) (string, error)
	ListSyncErrors   func(ctx context.Context) (map[int64]string, error)
	List             func(opt ExternalServicesListOptions) ([]*types.ExternalService, error)
	Update           func(ctx context.Context, ps []schema.AuthProviders, id int64, update *ExternalServiceUpdate) error
	Count            func(ctx context.Context, opt ExternalServicesListOptions) (int, error)
	Upsert           func(ctx context.Context, services ...*types.ExternalService) error
}
