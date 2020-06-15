package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/xeipuuv/gojsonschema"
)

// An ExternalServicesStore stores external services and their configuration.
// Before updating or creating a new external service, validation is performed.
// The enterprise code registers additional validators at run-time and sets the
// global instance in stores.go
type ExternalServicesStore struct {
	GitHubValidators          []func(*schema.GitHubConnection) error
	GitLabValidators          []func(*schema.GitLabConnection, []schema.AuthProviders) error
	BitbucketServerValidators []func(*schema.BitbucketServerConnection) error
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
	Kinds []string
	*LimitOffset
}

func (o ExternalServicesListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("deleted_at IS NULL")}
	if len(o.Kinds) > 0 {
		kinds := make([]*sqlf.Query, 0, len(o.Kinds))
		for _, kind := range o.Kinds {
			kinds = append(kinds, sqlf.Sprintf("%s", kind))
		}
		conds = append(conds, sqlf.Sprintf("kind IN (%s)", sqlf.Join(kinds, ",")))
	}
	return conds
}

// ValidateConfig validates the given external service configuration.
// A non zero id indicates we are updating an existing service, 0 indicates we are adding a new one.
func (e *ExternalServicesStore) ValidateConfig(ctx context.Context, id int64, kind, config string, ps []schema.AuthProviders) error {
	ext, ok := ExternalServiceKinds[kind]
	if !ok {
		return fmt.Errorf("invalid external service kind: %s", kind)
	}

	// All configs must be valid JSON.
	// If this requirement is ever changed, you will need to update
	// serveExternalServiceConfigs to handle this case.

	sl := gojsonschema.NewSchemaLoader()
	sc, err := sl.Compile(gojsonschema.NewStringLoader(ext.JSONSchema))
	if err != nil {
		return errors.Wrapf(err, "failed to compile schema for external service of kind %q", kind)
	}

	normalized, err := jsonc.Parse(config)
	if err != nil {
		return errors.Wrapf(err, "failed to normalize JSON")
	}

	res, err := sc.Validate(gojsonschema.NewBytesLoader(normalized))
	if err != nil {
		return errors.Wrap(err, "failed to validate config against schema")
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
	switch kind {
	case extsvc.KindGitHub:
		var c schema.GitHubConnection
		if err = json.Unmarshal(normalized, &c); err != nil {
			return err
		}
		err = e.validateGitHubConnection(ctx, id, &c)

	case extsvc.KindGitLab:
		var c schema.GitLabConnection
		if err = json.Unmarshal(normalized, &c); err != nil {
			return err
		}
		err = e.validateGitLabConnection(ctx, id, &c, ps)

	case extsvc.KindBitbucketServer:
		var c schema.BitbucketServerConnection
		if err = json.Unmarshal(normalized, &c); err != nil {
			return err
		}
		err = e.validateBitbucketServerConnection(ctx, id, &c)

	case extsvc.KindBitbucketCloud:
		var c schema.BitbucketCloudConnection
		if err = json.Unmarshal(normalized, &c); err != nil {
			return err
		}
		err = e.validateBitbucketCloudConnection(ctx, id, &c)

	case extsvc.KindOther:
		var c schema.OtherExternalServiceConnection
		if err = json.Unmarshal(normalized, &c); err != nil {
			return err
		}
		err = validateOtherExternalServiceConnection(&c)
	}

	return multierror.Append(errs, err).ErrorOrNil()
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

func (e *ExternalServicesStore) validateGitHubConnection(ctx context.Context, id int64, c *schema.GitHubConnection) error {
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

func (e *ExternalServicesStore) validateGitLabConnection(ctx context.Context, id int64, c *schema.GitLabConnection, ps []schema.AuthProviders) error {
	err := new(multierror.Error)
	for _, validate := range e.GitLabValidators {
		err = multierror.Append(err, validate(c, ps))
	}

	err = multierror.Append(err, e.validateDuplicateRateLimits(ctx, id, extsvc.KindGitLab, c))

	return err.ErrorOrNil()
}

func (e *ExternalServicesStore) validateBitbucketServerConnection(ctx context.Context, id int64, c *schema.BitbucketServerConnection) error {
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

func (e *ExternalServicesStore) validateBitbucketCloudConnection(ctx context.Context, id int64, c *schema.BitbucketCloudConnection) error {
	return e.validateDuplicateRateLimits(ctx, id, extsvc.KindBitbucketCloud, c)
}

func (e *ExternalServicesStore) validateDuplicateRateLimits(ctx context.Context, id int64, kind string, parsedConfig interface{}) error {
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
	// A rate limit has been defined
	services, err := e.List(ctx, ExternalServicesListOptions{
		Kinds: []string{kind},
	})
	if err != nil {
		return errors.Wrap(err, "listing existing services")
	}

	for _, svc := range services {
		rlc, err := extsvc.ExtractRateLimitConfig(svc.Config, svc.Kind, svc.DisplayName)
		if err != nil {
			return errors.Wrap(err, "extracting rate limit config")
		}
		if rlc.BaseURL == baseURL && svc.ID != id && !rlc.IsDefault {
			return fmt.Errorf("existing external service, %q, already has a rate limit set", rlc.DisplayName)
		}
	}

	return nil
}

// Create creates a external service.
//
// Since this method is used before the configuration server has started
// (search for "EXTSVC_CONFIG_FILE") you must pass the conf.Get function in so
// that an alternative can be used when the configuration server has not
// started, otherwise a panic would occur once pkg/conf's deadlock detector
// determines a deadlock occurred.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (e *ExternalServicesStore) Create(ctx context.Context, confGet func() *conf.Unified, externalService *types.ExternalService) error {
	if Mocks.ExternalServices.Create != nil {
		return Mocks.ExternalServices.Create(ctx, confGet, externalService)
	}

	ps := confGet().AuthProviders
	if err := e.ValidateConfig(ctx, 0, externalService.Kind, externalService.Config, ps); err != nil {
		return err
	}

	externalService.CreatedAt = time.Now().UTC().Truncate(time.Microsecond)
	externalService.UpdatedAt = externalService.CreatedAt

	return dbconn.Global.QueryRowContext(
		ctx,
		"INSERT INTO external_services(kind, display_name, config, created_at, updated_at) VALUES($1, $2, $3, $4, $5) RETURNING id",
		externalService.Kind, externalService.DisplayName, externalService.Config, externalService.CreatedAt, externalService.UpdatedAt,
	).Scan(&externalService.ID)
}

// ExternalServiceUpdate contains optional fields to update.
type ExternalServiceUpdate struct {
	DisplayName *string
	Config      *string
}

// Update updates a external service.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (e *ExternalServicesStore) Update(ctx context.Context, ps []schema.AuthProviders, id int64, update *ExternalServiceUpdate) error {
	if Mocks.ExternalServices.Update != nil {
		return Mocks.ExternalServices.Update(ctx, ps, id, update)
	}

	if update.Config != nil {
		// Query to get the kind (which is immutable) so we can validate the new config.
		externalService, err := e.GetByID(ctx, id)
		if err != nil {
			return err
		}

		if err := e.ValidateConfig(ctx, id, externalService.Kind, *update.Config, ps); err != nil {
			return err
		}
	}

	execUpdate := func(ctx context.Context, tx *sql.Tx, update *sqlf.Query) error {
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
	return dbutil.Transaction(ctx, dbconn.Global, func(tx *sql.Tx) error {
		if update.DisplayName != nil {
			if err := execUpdate(ctx, tx, sqlf.Sprintf("display_name=%s", update.DisplayName)); err != nil {
				return err
			}
		}
		if update.Config != nil {
			if err := execUpdate(ctx, tx, sqlf.Sprintf("config=%s", update.Config)); err != nil {
				return err
			}
		}
		return nil
	})
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
func (*ExternalServicesStore) Delete(ctx context.Context, id int64) error {
	if Mocks.ExternalServices.Delete != nil {
		return Mocks.ExternalServices.Delete(ctx, id)
	}

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
func (e *ExternalServicesStore) GetByID(ctx context.Context, id int64) (*types.ExternalService, error) {
	if Mocks.ExternalServices.GetByID != nil {
		return Mocks.ExternalServices.GetByID(id)
	}

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

// List returns all external services.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (e *ExternalServicesStore) List(ctx context.Context, opt ExternalServicesListOptions) ([]*types.ExternalService, error) {
	if Mocks.ExternalServices.List != nil {
		return Mocks.ExternalServices.List(opt)
	}
	return e.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

// listConfigs decodes the list of configs into result. In addition to populating
// loaded configs into the given result, it also calls the "SetURN(string)" method
// of elements in result when the method exists.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (e *ExternalServicesStore) listConfigs(ctx context.Context, kind string, result interface{}) error {
	services, err := e.List(ctx, ExternalServicesListOptions{Kinds: []string{kind}})
	if err != nil {
		return err
	}

	// Decode the jsonc configs into Go objects.
	cfgs := make([]interface{}, 0, len(services))
	urns := make([]string, 0, len(services))
	for _, service := range services {
		var cfg interface{}
		if err := jsonc.Unmarshal(service.Config, &cfg); err != nil {
			return err
		}
		cfgs = append(cfgs, cfg)
		urns = append(urns, service.URN())
	}

	// Now move our untyped config list into the typed list (result). We could
	// do this using reflection, but JSON marshaling and unmarshaling is easier
	// and fast enough for our purposes. Note that service.Config is jsonc, not
	// plain JSON so we could not simply treat it as json.RawMessage.
	buf, err := json.Marshal(cfgs)
	if err != nil {
		return err
	}

	err = json.Unmarshal(buf, result)
	if err != nil {
		return err
	}

	conns := reflect.ValueOf(result).Elem()
	for i := 0; i < conns.Len(); i++ {
		field, ok := conns.Index(i).Interface().(interface{ SetURN(string) })
		if ok {
			field.SetURN(urns[i])
		}
	}

	return nil
}

// ListAWSCodeCommitConnections returns a list of AWSCodeCommit configs.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (e *ExternalServicesStore) ListAWSCodeCommitConnections(ctx context.Context) ([]*types.AWSCodeCommitConnection, error) {
	var connections []*types.AWSCodeCommitConnection
	if err := e.listConfigs(ctx, extsvc.KindAWSCodeCommit, &connections); err != nil {
		return nil, err
	}
	return connections, nil
}

// ListBitbucketCloudConnections returns a list of BitbucketCloud configs.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (e *ExternalServicesStore) ListBitbucketCloudConnections(ctx context.Context) ([]*types.BitbucketCloudConnection, error) {
	var connections []*types.BitbucketCloudConnection
	if err := e.listConfigs(ctx, extsvc.KindBitbucketCloud, &connections); err != nil {
		return nil, err
	}
	return connections, nil
}

// ListBitbucketServerConnections returns a list of BitbucketServer configs.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (e *ExternalServicesStore) ListBitbucketServerConnections(ctx context.Context) ([]*types.BitbucketServerConnection, error) {
	var connections []*types.BitbucketServerConnection
	if err := e.listConfigs(ctx, extsvc.KindBitbucketServer, &connections); err != nil {
		return nil, err
	}
	return connections, nil
}

// ListGitHubConnections returns a list of GitHubConnection configs.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (e *ExternalServicesStore) ListGitHubConnections(ctx context.Context) ([]*types.GitHubConnection, error) {
	var connections []*types.GitHubConnection
	if err := e.listConfigs(ctx, extsvc.KindGitHub, &connections); err != nil {
		return nil, err
	}
	return connections, nil
}

// ListGitLabConnections returns a list of GitLabConnection configs.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (e *ExternalServicesStore) ListGitLabConnections(ctx context.Context) ([]*types.GitLabConnection, error) {
	var connections []*types.GitLabConnection
	if err := e.listConfigs(ctx, extsvc.KindGitLab, &connections); err != nil {
		return nil, err
	}
	return connections, nil
}

// ListGitoliteConnections returns a list of GitoliteConnection configs.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (e *ExternalServicesStore) ListGitoliteConnections(ctx context.Context) ([]*types.GitoliteConnection, error) {
	var connections []*types.GitoliteConnection
	if err := e.listConfigs(ctx, extsvc.KindGitolite, &connections); err != nil {
		return nil, err
	}
	return connections, nil
}

// ListPhabricatorConnections returns a list of PhabricatorConnection configs.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (e *ExternalServicesStore) ListPhabricatorConnections(ctx context.Context) ([]*types.PhabricatorConnection, error) {
	var connections []*types.PhabricatorConnection
	if err := e.listConfigs(ctx, extsvc.KindPhabricator, &connections); err != nil {
		return nil, err
	}
	return connections, nil
}

// ListOtherExternalServicesConnections returns a list of OtherExternalServiceConnection configs.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (e *ExternalServicesStore) ListOtherExternalServicesConnections(ctx context.Context) ([]*types.OtherExternalServiceConnection, error) {
	var connections []*types.OtherExternalServiceConnection
	if err := e.listConfigs(ctx, extsvc.KindOther, &connections); err != nil {
		return nil, err
	}
	return connections, nil
}

func (*ExternalServicesStore) list(ctx context.Context, conds []*sqlf.Query, limitOffset *LimitOffset) ([]*types.ExternalService, error) {
	q := sqlf.Sprintf(`
		SELECT id, kind, display_name, config, created_at, updated_at
		FROM external_services
		WHERE (%s)
		ORDER BY id DESC
		%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)

	rows, err := dbconn.Global.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*types.ExternalService
	for rows.Next() {
		var h types.ExternalService
		if err := rows.Scan(&h.ID, &h.Kind, &h.DisplayName, &h.Config, &h.CreatedAt, &h.UpdatedAt); err != nil {
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
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (*ExternalServicesStore) Count(ctx context.Context, opt ExternalServicesListOptions) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM external_services WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
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
}
