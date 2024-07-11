package contract

import (
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RequestedEnvVar struct {
	Name         string
	DefaultValue string
	Description  string
}

// Env should be used to reference all environment variables used by the service.
// It should be used instead of directly calling os.Getenv() to access environment
// variables.
//
// Env should only be used during service startup, and should not be used after
// (*Env).Validate() has been called. In services using the MSP runtime, this
// should only be used in a service's runtime.ConfigLoader implementation.
//
// All environment variables referenced are reported when running a service
// using the MSP runtime with the '-help' flag.
type Env struct {
	errs []error
	env  map[string]string

	// validated indicates if Env.Validate() has been called, after which
	// further calls to getters should panic.
	validated bool

	// requestedEnvVars are only available after ConfigLoader is used on this
	// Env instance.
	requestedEnvVars []RequestedEnvVar
}

// ParseEnv parses os.Environ() once for various contracts to reference for
// expected configuration values (the 'MSP contract'). The 'environ' argument
// is expected to be the output of os.Environ(), or an equivalent format.
//
// After using Env to retrieve all expected variables, callers are expected to
// use (*Env).Validate() to ensure all required variables were provided and
// collect any parsing errors.
//
// Env is not concurrency-safe.
func ParseEnv(environ []string) (*Env, error) {
	envMap := make(map[string]string, len(environ))
	for _, e := range environ {
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

func (e *Env) get(name, defaultValue, description string) string {
	if e.validated {
		panic("contract.Env was used for retrieving configuration after Validate() was called")
	}

	e.requestedEnvVars = append(e.requestedEnvVars, RequestedEnvVar{
		Name:         name,
		DefaultValue: defaultValue,
		Description:  description,
	})

	v, ok := e.env[name]
	if !ok {
		return defaultValue
	}
	return v
}

// Validate returns any errors constructed from a Get* method after the values have
// been loaded from the environment, including errors from env.AddError(...).
func (e *Env) Validate() error {
	e.validated = true

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
//
// The name, defaultValue, and description are reported when running a service
// using the MSP runtime with the '-help' flag.
func (e *Env) Get(name, defaultValue, description string) string {
	rawValue := e.get(name, defaultValue, description)
	if rawValue == "" {
		e.AddError(errors.Errorf("invalid value %q for %s: no value supplied, description: %s",
			rawValue, name, description))
		return ""
	}

	return rawValue
}

// GetOptional returns the value with the given name.
//
// The name and description are reported when running a service using the MSP
// runtime with the '-help' flag.
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
//
// The name, defaultValue, and description are reported when running a service
// using the MSP runtime with the '-help' flag.
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
//
// The name, defaultValue, and description are reported when running a service
// using the MSP runtime with the '-help' flag.
func (e *Env) GetPercent(name, defaultValue, description string) int {
	value := e.GetInt(name, defaultValue, description)
	if value < 0 || value > 100 {
		e.AddError(errors.Errorf("invalid percent %q for %s: must be 0 <= p <= 100", value, name))
		return 0
	}

	return value
}

// GetInterval parses a duration string using 'time.ParseDuration', expecting
// formats such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us"
// (or "µs"), "ms", "s", "m", "h".
//
// The name, defaultValue, and description are reported when running a service
// using the MSP runtime with the '-help' flag.
func (e *Env) GetInterval(name, defaultValue, description string) time.Duration {
	rawValue := e.get(name, defaultValue, description)
	d, err := time.ParseDuration(rawValue)
	if err != nil {
		e.AddError(errors.Errorf("invalid duration %q for %s: %s", rawValue, name, err))
		return 0
	}

	return d
}

// GetInterval parses a duration string using 'time.ParseDuration', expecting
// formats such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us"
// (or "µs"), "ms", "s", "m", "h".
//
// The name and description are reported when running a service using the MSP
// runtime with the '-help' flag.
func (e *Env) GetOptionalInterval(name, description string) *time.Duration {
	rawValue := e.GetOptional(name, description)
	if rawValue == nil {
		return nil
	}
	d, err := time.ParseDuration(*rawValue)
	if err != nil {
		e.AddError(errors.Errorf("invalid duration %q for %s: %s", *rawValue, name, err))
		return nil
	}
	return &d
}

// GetBool returns the value with the given name interpreted as a boolean. If no value was
// supplied in the environment, the given default is used in its place. If no value is available,
// or if the given value or default cannot be converted to a boolean, an error is added to the
// validation errors list.
//
// The name, defaultValue, and description are reported when running a service
// using the MSP runtime with the '-help' flag.
func (e *Env) GetBool(name, defaultValue, description string) bool {
	rawValue := e.get(name, defaultValue, description)
	v, err := strconv.ParseBool(rawValue)
	if err != nil {
		e.AddError(errors.Errorf("invalid bool %q for %s: %s", rawValue, name, err))
		return false
	}

	return v
}

// GetFloat returns the value with the given name interpreted as a float64. If no
// value was supplied in the environment, the given default is used in its place.
// If no value is available, or if the given value or default cannot be converted
// to a float64, an error is added to the validation errors list.
//
// The name, defaultValue, and description are reported when running a service
// using the MSP runtime with the '-help' flag.
func (e *Env) GetFloat(name, defaultValue, description string) float64 {
	rawValue := e.get(name, defaultValue, description)
	v, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		e.AddError(errors.Errorf("invalid float %q for %s: %s", rawValue, name, err))
		return 0
	}
	return v
}

// AddError adds a validation error to the configuration object. This should be
// called from within the Load method of a decorated configuration object to have
// any effect. The error is silently collected and can be retrieved using
// env.Validate()
func (e *Env) AddError(err error) {
	e.errs = append(e.errs, err)
}

// GetRequestedEnvVars returns the list of environment variables that were
// requested so far through Env's various Get methods.
func (e *Env) GetRequestedEnvVars() []RequestedEnvVar {
	return e.requestedEnvVars
}
