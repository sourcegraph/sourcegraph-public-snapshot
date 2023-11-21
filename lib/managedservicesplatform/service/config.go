package service

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Contract loads standardized MSP-provisioned (Managed Services Platform)
// configuration.
type Contract struct {
	// Indicate if we are running in a MSP environment.
	MSP bool
	// Port is the port the service must listen on.
	Port int
	// ExternalDNSName is the DNS name the service uses, if one is configured.
	ExternalDNSName *string

	postgreSQLContract
}

type postgreSQLContract struct {
	customDSNTemplate *string

	instanceConnectionName *string
	user                   *string
}

func newContract(env *Env) Contract {
	return Contract{
		MSP:             env.GetBool("MSP", "false", "indicates if we are running in a MSP environment"),
		Port:            env.GetInt("PORT", "", "service port"),
		ExternalDNSName: env.GetOptional("EXTERNAL_DNS_NAME", "external DNS name provisioned for the service"),
		postgreSQLContract: postgreSQLContract{
			customDSNTemplate: env.GetOptional("PGDSN",
				"custom PostgreSQL DSN with templatized database, e.g. 'user=foo database={{ .Database }}'"),

			instanceConnectionName: env.GetOptional("PGINSTANCE", "Cloud SQL instance connection name"),
			user:                   env.GetOptional("PGUSER", "Cloud SQL user"),
		},
	}
}

// GetPostgreSQLDB returns a standard library DB pointing to the configured
// PostgreSQL database. In MSP, we connect to a Cloud SQL instance over IAM auth.
//
// In development, the connection can be overridden with the PGDSN environment
// variable.
func (c postgreSQLContract) GetPostgreSQLDB(ctx context.Context, database string) (*sql.DB, error) {
	if c.customDSNTemplate != nil {
		tmpl, err := template.New("PGDSN").Parse(*c.customDSNTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "PGDSN is not a valid template")
		}
		var dsn bytes.Buffer
		if err := tmpl.Execute(&dsn, struct{ Database string }{Database: database}); err != nil {
			return nil, errors.Wrap(err, "PGDSN template is invalid")
		}
		return sql.Open("pgx", dsn.String())
	}

	config, err := c.getCloudSQLConnConfig(ctx, database)
	if err != nil {
		return nil, errors.Wrap(err, "get CloudSQL connection config")
	}
	return sql.Open("pgx", stdlib.RegisterConnConfig(config))
}

// getCloudSQLConnConfig generates a pgx connection configuration for using
// a Cloud SQL instance using IAM auth.
func (c postgreSQLContract) getCloudSQLConnConfig(ctx context.Context, database string) (*pgx.ConnConfig, error) {
	if c.instanceConnectionName == nil || c.user == nil {
		return nil, errors.New("missing required PostgreSQL configuration")
	}

	// https://github.com/GoogleCloudPlatform/cloud-sql-go-connector?tab=readme-ov-file#automatic-iam-database-authentication
	dsn := fmt.Sprintf("user=%s database=%s", *c.user, database)
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "pgx.ParseConfig")
	}
	d, err := cloudsqlconn.NewDialer(ctx,
		cloudsqlconn.WithIAMAuthN(),
		// MSP uses private IP
		cloudsqlconn.WithDefaultDialOptions(cloudsqlconn.WithPrivateIP()))
	if err != nil {
		return nil, errors.Wrap(err, "cloudsqlconn.NewDialer")
	}
	// Use the Cloud SQL connector to handle connecting to the instance.
	// This approach does *NOT* require the Cloud SQL proxy.
	config.DialFunc = func(ctx context.Context, network, instance string) (net.Conn, error) {
		return d.Dial(ctx, *c.instanceConnectionName)
	}
	return config, nil
}

type ConfigLoader[ConfigT any] interface {
	*ConfigT

	// Load should populate ConfigT with values from env. Errors should be added
	// to env using env.AddError().
	Load(env *Env)
}

type Env struct {
	errs []error
	env  map[string]string
}

func newEnv() (*Env, error) {
	env := os.Environ()
	envMap := make(map[string]string, len(env))
	for _, e := range env {
		kv := strings.SplitN(e, "=", 2)
		if len(kv) != 2 {
			return nil, errors.Errorf("unable to parse environment variable %q", e)
		}
		envMap[kv[0]] = kv[1]
	}
	return &Env{
		errs: make([]error, 0),
		env:  envMap,
	}, nil
}

func (e *Env) get(name, defaultValue, _ string) string {
	v, ok := e.env[name]
	if !ok {
		return defaultValue
	}
	return v
}

// validate returns any errors constructed from a Get* method after the values have
// been loaded from the environment.
func (e *Env) validate() error {
	if len(e.errs) == 0 {
		return nil
	}

	err := e.errs[0]
	for i := 1; i < len(e.errs); i++ {
		err = errors.Append(err, e.errs[i])
	}

	return err
}

// Get returns the value with the given name. If no value was supplied in the
// environment, the given default is used in its place. If no value is available,
// an error is added to the validation errors list.
func (e *Env) Get(name, defaultValue, description string) string {
	rawValue := e.get(name, defaultValue, description)
	if rawValue == "" {
		e.AddError(errors.Errorf("invalid value %q for %s: no value supplied", rawValue, name))
		return ""
	}

	return rawValue
}

// GetOptional returns the value with the given name.
func (e *Env) GetOptional(name, description string) *string {
	v, ok := e.env[name]
	if !ok {
		return nil
	}
	return &v
}

// GetInt returns the value with the given name interpreted as an integer. If no
// value was supplied in the environment, the given default is used in its place.
// If no value is available, or if the given value or default cannot be converted
// to an integer, an error is added to the validation errors list.
func (e *Env) GetInt(name, defaultValue, description string) int {
	rawValue := e.get(name, defaultValue, description)
	i, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		e.AddError(errors.Errorf("invalid int %q for %s: %s", rawValue, name, err))
		return 0
	}

	return int(i)
}

// GetPercent returns the value with the given name interpreted as an integer between
// 0 and 100. If no value was supplied in the environment, the given default is used
// in its place. If no value is available, if the given value or default cannot be
// converted to an integer, or if the value is out of the expected range, an error
// is added to the validation errors list.
func (e *Env) GetPercent(name, defaultValue, description string) int {
	value := e.GetInt(name, defaultValue, description)
	if value < 0 || value > 100 {
		e.AddError(errors.Errorf("invalid percent %q for %s: must be 0 <= p <= 100", value, name))
		return 0
	}

	return value
}

func (e *Env) GetInterval(name, defaultValue, description string) time.Duration {
	rawValue := e.get(name, defaultValue, description)
	d, err := time.ParseDuration(rawValue)
	if err != nil {
		e.AddError(errors.Errorf("invalid duration %q for %s: %s", rawValue, name, err))
		return 0
	}

	return d
}

// GetBool returns the value with the given name interpreted as a boolean. If no value was
// supplied in the environment, the given default is used in its place. If no value is available,
// or if the given value or default cannot be converted to a boolean, an error is added to the
// validation errors list.
func (e *Env) GetBool(name, defaultValue, description string) bool {
	rawValue := e.get(name, defaultValue, description)
	v, err := strconv.ParseBool(rawValue)
	if err != nil {
		e.AddError(errors.Errorf("invalid bool %q for %s: %s", rawValue, name, err))
		return false
	}

	return v
}

// AddError adds a validation error to the configuration object. This should be
// called from within the Load method of a decorated configuration object to have
// any effect.
func (e *Env) AddError(err error) {
	e.errs = append(e.errs, err)
}
