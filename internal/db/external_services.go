package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/xeipuuv/gojsonschema"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
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

	// PreCreateExternalService (if set) is invoked as a hook prior to creating a
	// new external service in the database.
	PreCreateExternalService func(context.Context) error

	mu sync.Mutex
}

// NewExternalServicesStoreWithDB instantiates and returns a new ExternalServicesStore with prepared statements.
func NewExternalServicesStoreWithDB(db dbutil.DB) *ExternalServiceStore {
	return &ExternalServiceStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

func (e *ExternalServiceStore) With(other basestore.ShareableStore) *ExternalServiceStore {
	return &ExternalServiceStore{Store: e.Store.With(other)}
}

func (e *ExternalServiceStore) Transact(ctx context.Context) (*ExternalServiceStore, error) {
	txBase, err := e.Store.Transact(ctx)
	return &ExternalServiceStore{Store: txBase}, err
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
	// When true, only include external services not under any namespace (i.e. owned by all site admins),
	// and value of NamespaceUserID is ignored.
	NoNamespace bool
	// When specified, only include external services under given user namespace.
	NamespaceUserID int32
	// When specified, only include external services with given list of kinds.
	Kinds []string
	// When specified, only include external services with ID below this number
	// (because we're sorting results by ID in descending order).
	AfterID int64
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
	return conds
}

type ValidateExternalServiceConfigOptions struct {
	// The ID of the external service, 0 is a valid value for not-yet-created external service.
	ID int64
	// The kind of external service.
	Kind string
	// The actual config of the external service.
	Config string
	// The list of authN providers configured on the instance.
	AuthProviders []schema.AuthProviders
	// When true, indicates this is a user-added the external service.
	HasNamespace bool
}

var errAuthorizationRequired = errors.New("authorization required")

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

	// For user-added external services, we need to prevent them from using disallowed fields.
	if opt.HasNamespace {
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
		if err = json.Unmarshal(normalized, &c); err != nil {
			return nil, err
		}
		if opt.HasNamespace && c.Authorization == nil {
			errs = multierror.Append(errs, errAuthorizationRequired)
		}
		err = e.validateGitHubConnection(ctx, opt.ID, &c)

	case extsvc.KindGitLab:
		var c schema.GitLabConnection
		if err = json.Unmarshal(normalized, &c); err != nil {
			return nil, err
		}
		if opt.HasNamespace && c.Authorization == nil {
			errs = multierror.Append(errs, errAuthorizationRequired)
		}
		err = e.validateGitLabConnection(ctx, opt.ID, &c, opt.AuthProviders)

	case extsvc.KindBitbucketServer:
		var c schema.BitbucketServerConnection
		if err = json.Unmarshal(normalized, &c); err != nil {
			return nil, err
		}
		err = e.validateBitbucketServerConnection(ctx, opt.ID, &c)

	case extsvc.KindBitbucketCloud:
		var c schema.BitbucketCloudConnection
		if err = json.Unmarshal(normalized, &c); err != nil {
			return nil, err
		}
		err = e.validateBitbucketCloudConnection(ctx, opt.ID, &c)

	case extsvc.KindOther:
		var c schema.OtherExternalServiceConnection
		if err = json.Unmarshal(normalized, &c); err != nil {
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

	if envvar.SourcegraphDotComMode() && c.CloudGlobal {
		// We're setting this one to global, make sure it's the only one
		err = multierror.Append(err, e.validateSingleGlobalConnection(ctx, id, extsvc.KindGitHub))
	}

	return err.ErrorOrNil()
}

func (e *ExternalServiceStore) validateGitLabConnection(ctx context.Context, id int64, c *schema.GitLabConnection, ps []schema.AuthProviders) error {
	err := new(multierror.Error)
	for _, validate := range e.GitLabValidators {
		err = multierror.Append(err, validate(c, ps))
	}

	err = multierror.Append(err, e.validateDuplicateRateLimits(ctx, id, extsvc.KindGitLab, c))

	if envvar.SourcegraphDotComMode() && c.CloudGlobal {
		// We're setting this one to global, make sure it's the only one
		err = multierror.Append(err, e.validateSingleGlobalConnection(ctx, id, extsvc.KindGitLab))
	}

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

// validateSingleGlobalConnection returns an error if more than one external service for the given kind has its
// CloudGlobal flag set.
func (e *ExternalServiceStore) validateSingleGlobalConnection(ctx context.Context, id int64, kind string) error {
	opt := ExternalServicesListOptions{
		Kinds: []string{kind},
		// We only care about site admin external services
		NoNamespace: true,
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
			// No more results, exiting
			return nil
		}
		opt.AfterID = svcs[len(svcs)-1].ID // Advance the cursor

		for _, svc := range svcs {
			c, err := extsvc.ParseConfig(svc.Kind, svc.Config)
			if err != nil {
				return errors.Wrap(err, "parsing config")
			}
			var storedIsGlobal bool
			switch x := c.(type) {
			case *schema.GitHubConnection:
				storedIsGlobal = x.CloudGlobal
			case *schema.GitLabConnection:
				storedIsGlobal = x.CloudGlobal
			}
			if svc.ID != id && storedIsGlobal {
				return fmt.Errorf("existing external service, %q, already set as global", svc.DisplayName)
			}
		}

		if len(svcs) < opt.Limit {
			break // Less results than limit means we've reached end
		}
	}
	return nil
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

// Create creates an external service.
//
// Since this method is used before the configuration server has started
// (search for "EXTSVC_CONFIG_FILE") you must pass the conf.Get function in so
// that an alternative can be used when the configuration server has not
// started, otherwise a panic would occur once pkg/conf's deadlock detector
// determines a deadlock occurred.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
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
		Kind:          es.Kind,
		Config:        es.Config,
		AuthProviders: confGet().AuthProviders,
		HasNamespace:  es.NamespaceUserID != 0,
	})
	if err != nil {
		return err
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

	return e.Store.Handle().DB().QueryRowContext(
		ctx,
		"INSERT INTO external_services(kind, display_name, config, created_at, updated_at, namespace_user_id, unrestricted) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		es.Kind, es.DisplayName, es.Config, es.CreatedAt, es.UpdatedAt, nullInt32Column(es.NamespaceUserID), es.Unrestricted,
	).Scan(&es.ID)
}

// Upsert updates or inserts the given ExternalServices.
//
// 🚨 SECURITY: The value of `Unrestricted` field is disregarded and will always
// be recalculated based on whether `"authorization"` is presented in `Config`.
func (e *ExternalServiceStore) Upsert(ctx context.Context, svcs ...*types.ExternalService) error {
	if len(svcs) == 0 {
		return nil
	}
	e.ensureStore()

	for _, s := range svcs {
		s.Unrestricted = !gjson.Get(s.Config, "authorization").Exists()
	}

	q := upsertExternalServicesQuery(svcs)
	rows, err := e.Query(ctx, q)
	if err != nil {
		return err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	i := 0
	for rows.Next() {
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
		)
		if err != nil {
			return err
		}

		i++
	}

	return err
}

func upsertExternalServicesQuery(svcs []*types.ExternalService) *sqlf.Query {
	vals := make([]*sqlf.Query, 0, len(svcs))
	for _, s := range svcs {
		vals = append(vals, sqlf.Sprintf(
			upsertExternalServicesQueryValueFmtstr,
			s.ID,
			s.Kind,
			s.DisplayName,
			s.Config,
			s.CreatedAt.UTC(),
			s.UpdatedAt.UTC(),
			nullTimeColumn(s.DeletedAt),
			nullTimeColumn(s.LastSyncAt),
			nullTimeColumn(s.NextSyncAt),
			nullInt32Column(s.NamespaceUserID),
			s.Unrestricted,
		))
	}

	return sqlf.Sprintf(
		upsertExternalServicesQueryFmtstr,
		sqlf.Join(vals, ",\n"),
	)
}

const upsertExternalServicesQueryValueFmtstr = `
  (COALESCE(NULLIF(%s, 0), (SELECT nextval('external_services_id_seq'))), UPPER(%s), %s, %s, %s, %s, %s, %s, %s, %s, %s)
`

const upsertExternalServicesQueryFmtstr = `
-- source: internal/repos/store.go:DBStore.UpsertExternalServices
INSERT INTO external_services (
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
  unrestricted
)
VALUES %s
ON CONFLICT(id) DO UPDATE
SET
  kind         = UPPER(excluded.kind),
  display_name = excluded.display_name,
  config       = excluded.config,
  created_at   = excluded.created_at,
  updated_at   = excluded.updated_at,
  deleted_at   = excluded.deleted_at,
  last_sync_at = excluded.last_sync_at,
  next_sync_at = excluded.next_sync_at,
  namespace_user_id = excluded.namespace_user_id,
  unrestricted = excluded.unrestricted
RETURNING *
`

// ExternalServiceUpdate contains optional fields to update.
type ExternalServiceUpdate struct {
	DisplayName *string
	Config      *string
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

	var normalized []byte
	if update.Config != nil {
		// Query to get the kind (which is immutable) so we can validate the new config.
		externalService, err := e.GetByID(ctx, id)
		if err != nil {
			return err
		}

		normalized, err = e.ValidateConfig(ctx, ValidateExternalServiceConfigOptions{
			ID:            id,
			Kind:          externalService.Kind,
			Config:        *update.Config,
			AuthProviders: ps,
			HasNamespace:  externalService.NamespaceUserID != 0,
		})
		if err != nil {
			return err
		}
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
		q := sqlf.Sprintf(`config = %s, next_sync_at = NOW(), unrestricted = %s`, update.Config, unrestricted)
		if err := execUpdate(ctx, tx.DB(), q); err != nil {
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
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (e *ExternalServiceStore) Delete(ctx context.Context, id int64) error {
	if Mocks.ExternalServices.Delete != nil {
		return Mocks.ExternalServices.Delete(ctx, id)
	}
	e.ensureStore()

	res, err := dbconn.Global.ExecContext(ctx, "UPDATE external_services SET deleted_at=now() WHERE id=$1 AND deleted_at IS NULL", id)
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
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (e *ExternalServiceStore) GetByID(ctx context.Context, id int64) (*types.ExternalService, error) {
	if Mocks.ExternalServices.GetByID != nil {
		return Mocks.ExternalServices.GetByID(id)
	}
	e.ensureStore()

	conds := []*sqlf.Query{
		sqlf.Sprintf("deleted_at IS NULL"),
		sqlf.Sprintf("id=%d", id),
	}
	ess, err := e.list(ctx, conds, nil)
	if err != nil {
		return nil, err
	}
	if len(ess) == 0 {
		return nil, externalServiceNotFoundError{id: id}
	}
	return ess[0], nil
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

	return e.list(ctx, opt.sqlConditions(), opt.LimitOffset)
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

func (e *ExternalServiceStore) list(ctx context.Context, conds []*sqlf.Query, limitOffset *LimitOffset) ([]*types.ExternalService, error) {
	q := sqlf.Sprintf(`
		SELECT id, kind, display_name, config, created_at, updated_at, deleted_at, last_sync_at, next_sync_at, namespace_user_id, unrestricted
		FROM external_services
		WHERE (%s)
		ORDER BY id DESC
		%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
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
		)
		if err := rows.Scan(&h.ID, &h.Kind, &h.DisplayName, &h.Config, &h.CreatedAt, &h.UpdatedAt, &deletedAt, &lastSyncAt, &nextSyncAt, &namespaceUserID, &h.Unrestricted); err != nil {
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
		results = append(results, &h)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// Count counts all external services that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
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

// MockExternalServices mocks the external services store.
type MockExternalServices struct {
	Create  func(ctx context.Context, confGet func() *conf.Unified, externalService *types.ExternalService) error
	Delete  func(ctx context.Context, id int64) error
	GetByID func(id int64) (*types.ExternalService, error)
	List    func(opt ExternalServicesListOptions) ([]*types.ExternalService, error)
	Update  func(ctx context.Context, ps []schema.AuthProviders, id int64, update *ExternalServiceUpdate) error
	Count   func(ctx context.Context, opt ExternalServicesListOptions) (int, error)
}
