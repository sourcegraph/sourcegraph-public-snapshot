// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package component // import "go.opentelemetry.io/collector/component"

import (
	"fmt"
	"reflect"
	"regexp"

	"go.uber.org/multierr"
)

// Config defines the configuration for a component.Component.
//
// Implementations and/or any sub-configs (other types embedded or included in the Config implementation)
// MUST implement the ConfigValidator if any validation is required for that part of the configuration
// (e.g. check if a required field is present).
//
// A valid implementation MUST pass the check componenttest.CheckConfigStruct (return nil error).
type Config any

// As interface types are only used for static typing, a common idiom to find the reflection Type
// for an interface type Foo is to use a *Foo value.
var configValidatorType = reflect.TypeOf((*ConfigValidator)(nil)).Elem()

// ConfigValidator defines an optional interface for configurations to implement to do validation.
type ConfigValidator interface {
	// Validate the configuration and returns an error if invalid.
	Validate() error
}

// ValidateConfig validates a config, by doing this:
//   - Call Validate on the config itself if the config implements ConfigValidator.
func ValidateConfig(cfg Config) error {
	return validate(reflect.ValueOf(cfg))
}

func validate(v reflect.Value) error {
	// Validate the value itself.
	switch v.Kind() {
	case reflect.Invalid:
		return nil
	case reflect.Ptr:
		return validate(v.Elem())
	case reflect.Struct:
		var errs error
		errs = multierr.Append(errs, callValidateIfPossible(v))
		// Reflect on the pointed data and check each of its fields.
		for i := 0; i < v.NumField(); i++ {
			if !v.Type().Field(i).IsExported() {
				continue
			}
			errs = multierr.Append(errs, validate(v.Field(i)))
		}
		return errs
	case reflect.Slice, reflect.Array:
		var errs error
		errs = multierr.Append(errs, callValidateIfPossible(v))
		// Reflect on the pointed data and check each of its fields.
		for i := 0; i < v.Len(); i++ {
			errs = multierr.Append(errs, validate(v.Index(i)))
		}
		return errs
	case reflect.Map:
		var errs error
		errs = multierr.Append(errs, callValidateIfPossible(v))
		iter := v.MapRange()
		for iter.Next() {
			errs = multierr.Append(errs, validate(iter.Key()))
			errs = multierr.Append(errs, validate(iter.Value()))
		}
		return errs
	default:
		return callValidateIfPossible(v)
	}
}

func callValidateIfPossible(v reflect.Value) error {
	// If the value type implements ConfigValidator just call Validate
	if v.Type().Implements(configValidatorType) {
		return v.Interface().(ConfigValidator).Validate()
	}

	// If the pointer type implements ConfigValidator call Validate on the pointer to the current value.
	if reflect.PointerTo(v.Type()).Implements(configValidatorType) {
		// If not addressable, then create a new *V pointer and set the value to current v.
		if !v.CanAddr() {
			pv := reflect.New(reflect.PtrTo(v.Type()).Elem())
			pv.Elem().Set(v)
			v = pv.Elem()
		}
		return v.Addr().Interface().(ConfigValidator).Validate()
	}

	return nil
}

// Type is the component type as it is used in the config.
type Type struct {
	name string
}

// String returns the string representation of the type.
func (t Type) String() string {
	return t.name
}

// MarshalText marshals returns the Type name.
func (t Type) MarshalText() ([]byte, error) {
	return []byte(t.name), nil
}

// typeRegexp is used to validate the type of a component.
// A type must start with an ASCII alphabetic character and
// can only contain ASCII alphanumeric characters and '_'.
// This must be kept in sync with the regex in cmd/mdatagen/validate.go.
var typeRegexp = regexp.MustCompile(`^[a-zA-Z][0-9a-zA-Z_]{0,62}$`)

// NewType creates a type. It returns an error if the type is invalid.
// A type must
// - have at least one character,
// - start with an ASCII alphabetic character and
// - can only contain ASCII alphanumeric characters and '_'.
func NewType(ty string) (Type, error) {
	if len(ty) == 0 {
		return Type{}, fmt.Errorf("id must not be empty")
	}
	if !typeRegexp.MatchString(ty) {
		return Type{}, fmt.Errorf("invalid character(s) in type %q", ty)
	}
	return Type{name: ty}, nil
}

// MustNewType creates a type. It panics if the type is invalid.
// A type must
// - have at least one character,
// - start with an ASCII alphabetic character and
// - can only contain ASCII alphanumeric characters and '_'.
func MustNewType(strType string) Type {
	ty, err := NewType(strType)
	if err != nil {
		panic(err)
	}
	return ty
}

// DataType is a special Type that represents the data types supported by the collector. We currently support
// collecting metrics, traces and logs, this can expand in the future.
type DataType = Type

func mustNewDataType(strType string) DataType {
	return MustNewType(strType)
}

// Currently supported data types. Add new data types here when new types are supported in the future.
var (
	// DataTypeTraces is the data type tag for traces.
	DataTypeTraces = mustNewDataType("traces")

	// DataTypeMetrics is the data type tag for metrics.
	DataTypeMetrics = mustNewDataType("metrics")

	// DataTypeLogs is the data type tag for logs.
	DataTypeLogs = mustNewDataType("logs")
)
