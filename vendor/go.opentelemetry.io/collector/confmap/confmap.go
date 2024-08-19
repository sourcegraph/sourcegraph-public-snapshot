// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package confmap // import "go.opentelemetry.io/collector/confmap"

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/maps"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"

	"go.opentelemetry.io/collector/confmap/internal"
	encoder "go.opentelemetry.io/collector/confmap/internal/mapstructure"
)

const (
	// KeyDelimiter is used as the default key delimiter in the default koanf instance.
	KeyDelimiter = "::"
)

// New creates a new empty confmap.Conf instance.
func New() *Conf {
	return &Conf{k: koanf.New(KeyDelimiter)}
}

// NewFromStringMap creates a confmap.Conf from a map[string]any.
func NewFromStringMap(data map[string]any) *Conf {
	p := New()
	// Cannot return error because the koanf instance is empty.
	_ = p.k.Load(confmap.Provider(data, KeyDelimiter), nil)
	return p
}

// Conf represents the raw configuration map for the OpenTelemetry Collector.
// The confmap.Conf can be unmarshalled into the Collector's config using the "service" package.
type Conf struct {
	k *koanf.Koanf
	// If true, upon unmarshaling do not call the Unmarshal function on the struct
	// if it implements Unmarshaler and is the top-level struct.
	// This avoids running into an infinite recursion where Unmarshaler.Unmarshal and
	// Conf.Unmarshal would call each other.
	skipTopLevelUnmarshaler bool
}

// AllKeys returns all keys holding a value, regardless of where they are set.
// Nested keys are returned with a KeyDelimiter separator.
func (l *Conf) AllKeys() []string {
	return l.k.Keys()
}

type UnmarshalOption interface {
	apply(*unmarshalOption)
}

type unmarshalOption struct {
	ignoreUnused bool
}

// WithIgnoreUnused sets an option to ignore errors if existing
// keys in the original Conf were unused in the decoding process
// (extra keys).
func WithIgnoreUnused() UnmarshalOption {
	return unmarshalOptionFunc(func(uo *unmarshalOption) {
		uo.ignoreUnused = true
	})
}

type unmarshalOptionFunc func(*unmarshalOption)

func (fn unmarshalOptionFunc) apply(set *unmarshalOption) {
	fn(set)
}

// Unmarshal unmarshalls the config into a struct using the given options.
// Tags on the fields of the structure must be properly set.
func (l *Conf) Unmarshal(result any, opts ...UnmarshalOption) error {
	set := unmarshalOption{}
	for _, opt := range opts {
		opt.apply(&set)
	}
	return decodeConfig(l, result, !set.ignoreUnused, l.skipTopLevelUnmarshaler)
}

type marshalOption struct{}

type MarshalOption interface {
	apply(*marshalOption)
}

// Marshal encodes the config and merges it into the Conf.
func (l *Conf) Marshal(rawVal any, _ ...MarshalOption) error {
	enc := encoder.New(encoderConfig(rawVal))
	data, err := enc.Encode(rawVal)
	if err != nil {
		return err
	}
	out, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid config encoding")
	}
	return l.Merge(NewFromStringMap(out))
}

// Get can retrieve any value given the key to use.
func (l *Conf) Get(key string) any {
	return l.k.Get(key)
}

// IsSet checks to see if the key has been set in any of the data locations.
func (l *Conf) IsSet(key string) bool {
	return l.k.Exists(key)
}

// Merge merges the input given configuration into the existing config.
// Note that the given map may be modified.
func (l *Conf) Merge(in *Conf) error {
	return l.k.Merge(in.k)
}

// Sub returns new Conf instance representing a sub-config of this instance.
// It returns an error is the sub-config is not a map[string]any (use Get()), and an empty Map if none exists.
func (l *Conf) Sub(key string) (*Conf, error) {
	// Code inspired by the koanf "Cut" func, but returns an error instead of empty map for unsupported sub-config type.
	data := l.Get(key)
	if data == nil {
		return New(), nil
	}

	if v, ok := data.(map[string]any); ok {
		return NewFromStringMap(v), nil
	}

	return nil, fmt.Errorf("unexpected sub-config value kind for key:%s value:%v kind:%v)", key, data, reflect.TypeOf(data).Kind())
}

// ToStringMap creates a map[string]any from a Parser.
func (l *Conf) ToStringMap() map[string]any {
	return maps.Unflatten(l.k.All(), KeyDelimiter)
}

// decodeConfig decodes the contents of the Conf into the result argument, using a
// mapstructure decoder with the following notable behaviors. Ensures that maps whose
// values are nil pointer structs resolved to the zero value of the target struct (see
// expandNilStructPointers). Converts string to []string by splitting on ','. Ensures
// uniqueness of component IDs (see mapKeyStringToMapKeyTextUnmarshalerHookFunc).
// Decodes time.Duration from strings. Allows custom unmarshaling for structs implementing
// encoding.TextUnmarshaler. Allows custom unmarshaling for structs implementing confmap.Unmarshaler.
func decodeConfig(m *Conf, result any, errorUnused bool, skipTopLevelUnmarshaler bool) error {
	dc := &mapstructure.DecoderConfig{
		ErrorUnused:      errorUnused,
		Result:           result,
		TagName:          "mapstructure",
		WeaklyTypedInput: !internal.StrictlyTypedInputGate.IsEnabled(),
		MatchName:        caseSensitiveMatchName,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			expandNilStructPointersHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			mapKeyStringToMapKeyTextUnmarshalerHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.TextUnmarshallerHookFunc(),
			unmarshalerHookFunc(result, skipTopLevelUnmarshaler),
			// after the main unmarshaler hook is called,
			// we unmarshal the embedded structs if present to merge with the result:
			unmarshalerEmbeddedStructsHookFunc(),
			zeroSliceHookFunc(),
			negativeUintHookFunc(),
		),
	}
	decoder, err := mapstructure.NewDecoder(dc)
	if err != nil {
		return err
	}
	if err = decoder.Decode(m.ToStringMap()); err != nil {
		if strings.HasPrefix(err.Error(), "error decoding ''") {
			return errors.Unwrap(err)
		}
		return err
	}
	return nil
}

// encoderConfig returns a default encoder.EncoderConfig that includes
// an EncodeHook that handles both TextMarshaller and Marshaler
// interfaces.
func encoderConfig(rawVal any) *encoder.EncoderConfig {
	return &encoder.EncoderConfig{
		EncodeHook: mapstructure.ComposeDecodeHookFunc(
			encoder.YamlMarshalerHookFunc(),
			encoder.TextMarshalerHookFunc(),
			marshalerHookFunc(rawVal),
		),
	}
}

// case-sensitive version of the callback to be used in the MatchName property
// of the DecoderConfig. The default for MatchEqual is to use strings.EqualFold,
// which is case-insensitive.
func caseSensitiveMatchName(a, b string) bool {
	return a == b
}

// In cases where a config has a mapping of something to a struct pointers
// we want nil values to resolve to a pointer to the zero value of the
// underlying struct just as we want nil values of a mapping of something
// to a struct to resolve to the zero value of that struct.
//
// e.g. given a config type:
// type Config struct { Thing *SomeStruct `mapstructure:"thing"` }
//
// and yaml of:
// config:
//
//	thing:
//
// we want an unmarshaled Config to be equivalent to
// Config{Thing: &SomeStruct{}} instead of Config{Thing: nil}
func expandNilStructPointersHookFunc() mapstructure.DecodeHookFuncValue {
	return func(from reflect.Value, to reflect.Value) (any, error) {
		// ensure we are dealing with map to map comparison
		if from.Kind() == reflect.Map && to.Kind() == reflect.Map {
			toElem := to.Type().Elem()
			// ensure that map values are pointers to a struct
			// (that may be nil and require manual setting w/ zero value)
			if toElem.Kind() == reflect.Ptr && toElem.Elem().Kind() == reflect.Struct {
				fromRange := from.MapRange()
				for fromRange.Next() {
					fromKey := fromRange.Key()
					fromValue := fromRange.Value()
					// ensure that we've run into a nil pointer instance
					if fromValue.IsNil() {
						newFromValue := reflect.New(toElem.Elem())
						from.SetMapIndex(fromKey, newFromValue)
					}
				}
			}
		}
		return from.Interface(), nil
	}
}

// mapKeyStringToMapKeyTextUnmarshalerHookFunc returns a DecodeHookFuncType that checks that a conversion from
// map[string]any to map[encoding.TextUnmarshaler]any does not overwrite keys,
// when UnmarshalText produces equal elements from different strings (e.g. trims whitespaces).
//
// This is needed in combination with ComponentID, which may produce equal IDs for different strings,
// and an error needs to be returned in that case, otherwise the last equivalent ID overwrites the previous one.
func mapKeyStringToMapKeyTextUnmarshalerHookFunc() mapstructure.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.Map || from.Key().Kind() != reflect.String {
			return data, nil
		}

		if to.Kind() != reflect.Map {
			return data, nil
		}

		// Checks that the key type of to implements the TextUnmarshaler interface.
		if _, ok := reflect.New(to.Key()).Interface().(encoding.TextUnmarshaler); !ok {
			return data, nil
		}

		// Create a map with key value of to's key to bool.
		fieldNameSet := reflect.MakeMap(reflect.MapOf(to.Key(), reflect.TypeOf(true)))
		for k := range data.(map[string]any) {
			// Create a new value of the to's key type.
			tKey := reflect.New(to.Key())

			// Use tKey to unmarshal the key of the map.
			if err := tKey.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(k)); err != nil {
				return nil, err
			}
			// Checks if the key has already been decoded in a previous iteration.
			if fieldNameSet.MapIndex(reflect.Indirect(tKey)).IsValid() {
				return nil, fmt.Errorf("duplicate name %q after unmarshaling %v", k, tKey)
			}
			fieldNameSet.SetMapIndex(reflect.Indirect(tKey), reflect.ValueOf(true))
		}
		return data, nil
	}
}

// unmarshalerEmbeddedStructsHookFunc provides a mechanism for embedded structs to define their own unmarshal logic,
// by implementing the Unmarshaler interface.
func unmarshalerEmbeddedStructsHookFunc() mapstructure.DecodeHookFuncValue {
	return func(from reflect.Value, to reflect.Value) (any, error) {
		if to.Type().Kind() != reflect.Struct {
			return from.Interface(), nil
		}
		fromAsMap, ok := from.Interface().(map[string]any)
		if !ok {
			return from.Interface(), nil
		}
		for i := 0; i < to.Type().NumField(); i++ {
			// embedded structs passed in via `squash` cannot be pointers. We just check if they are structs:
			f := to.Type().Field(i)
			if f.IsExported() && slices.Contains(strings.Split(f.Tag.Get("mapstructure"), ","), "squash") {
				if unmarshaler, ok := to.Field(i).Addr().Interface().(Unmarshaler); ok {
					c := NewFromStringMap(fromAsMap)
					c.skipTopLevelUnmarshaler = true
					if err := unmarshaler.Unmarshal(c); err != nil {
						return nil, err
					}
					// the struct we receive from this unmarshaling only contains fields related to the embedded struct.
					// we merge this partially unmarshaled struct with the rest of the result.
					// note we already unmarshaled the main struct earlier, and therefore merge with it.
					conf := New()
					if err := conf.Marshal(unmarshaler); err != nil {
						return nil, err
					}
					resultMap := conf.ToStringMap()
					for k, v := range resultMap {
						fromAsMap[k] = v
					}
				}
			}
		}
		return fromAsMap, nil
	}
}

// Provides a mechanism for individual structs to define their own unmarshal logic,
// by implementing the Unmarshaler interface, unless skipTopLevelUnmarshaler is
// true and the struct matches the top level object being unmarshaled.
func unmarshalerHookFunc(result any, skipTopLevelUnmarshaler bool) mapstructure.DecodeHookFuncValue {
	return func(from reflect.Value, to reflect.Value) (any, error) {
		if !to.CanAddr() {
			return from.Interface(), nil
		}

		toPtr := to.Addr().Interface()
		// Need to ignore the top structure to avoid running into an infinite recursion
		// where Unmarshaler.Unmarshal and Conf.Unmarshal would call each other.
		if toPtr == result && skipTopLevelUnmarshaler {
			return from.Interface(), nil
		}

		unmarshaler, ok := toPtr.(Unmarshaler)
		if !ok {
			return from.Interface(), nil
		}

		if _, ok = from.Interface().(map[string]any); !ok {
			return from.Interface(), nil
		}

		// Use the current object if not nil (to preserve other configs in the object), otherwise zero initialize.
		if to.Addr().IsNil() {
			unmarshaler = reflect.New(to.Type()).Interface().(Unmarshaler)
		}

		c := NewFromStringMap(from.Interface().(map[string]any))
		c.skipTopLevelUnmarshaler = true
		if err := unmarshaler.Unmarshal(c); err != nil {
			return nil, err
		}

		return unmarshaler, nil
	}
}

// marshalerHookFunc returns a DecodeHookFuncValue that checks structs that aren't
// the original to see if they implement the Marshaler interface.
func marshalerHookFunc(orig any) mapstructure.DecodeHookFuncValue {
	origType := reflect.TypeOf(orig)
	return func(from reflect.Value, _ reflect.Value) (any, error) {
		if from.Kind() != reflect.Struct {
			return from.Interface(), nil
		}

		// ignore original to avoid infinite loop.
		if from.Type() == origType && reflect.DeepEqual(from.Interface(), orig) {
			return from.Interface(), nil
		}
		marshaler, ok := from.Interface().(Marshaler)
		if !ok {
			return from.Interface(), nil
		}
		conf := New()
		if err := marshaler.Marshal(conf); err != nil {
			return nil, err
		}
		return conf.ToStringMap(), nil
	}
}

// Unmarshaler interface may be implemented by types to customize their behavior when being unmarshaled from a Conf.
type Unmarshaler interface {
	// Unmarshal a Conf into the struct in a custom way.
	// The Conf for this specific component may be nil or empty if no config available.
	// This method should only be called by decoding hooks when calling Conf.Unmarshal.
	Unmarshal(component *Conf) error
}

// Marshaler defines an optional interface for custom configuration marshaling.
// A configuration struct can implement this interface to override the default
// marshaling.
type Marshaler interface {
	// Marshal the config into a Conf in a custom way.
	// The Conf will be empty and can be merged into.
	Marshal(component *Conf) error
}

// This hook is used to solve the issue: https://github.com/open-telemetry/opentelemetry-collector/issues/4001
// We adopt the suggestion provided in this issue: https://github.com/mitchellh/mapstructure/issues/74#issuecomment-279886492
// We should empty every slice before unmarshalling unless user provided slice is nil.
// Assume that we had a struct with a field of type slice called `keys`, which has default values of ["a", "b"]
//
//	type Config struct {
//	  Keys []string `mapstructure:"keys"`
//	}
//
// The configuration provided by users may have following cases
// 1. configuration have `keys` field and have a non-nil values for this key, the output should be overrided
//   - for example, input is {"keys", ["c"]}, then output is Config{ Keys: ["c"]}
//
// 2. configuration have `keys` field and have an empty slice for this key, the output should be overrided by empty slics
//   - for example, input is {"keys", []}, then output is Config{ Keys: []}
//
// 3. configuration have `keys` field and have nil value for this key, the output should be default config
//   - for example, input is {"keys": nil}, then output is Config{ Keys: ["a", "b"]}
//
// 4. configuration have no `keys` field specified, the output should be default config
//   - for example, input is {}, then output is Config{ Keys: ["a", "b"]}
func zeroSliceHookFunc() mapstructure.DecodeHookFuncValue {
	return func(from reflect.Value, to reflect.Value) (interface{}, error) {
		if to.CanSet() && to.Kind() == reflect.Slice && from.Kind() == reflect.Slice {
			to.Set(reflect.MakeSlice(to.Type(), from.Len(), from.Cap()))
		}

		return from.Interface(), nil
	}
}

// This hook is used to solve the issue: https://github.com/open-telemetry/opentelemetry-collector/issues/9060
// Decoding should fail when converting a negative integer to any type of unsigned integer. This prevents
// negative values being decoded as large uint values.
// TODO: This should be removed as a part of https://github.com/open-telemetry/opentelemetry-collector/issues/9532
func negativeUintHookFunc() mapstructure.DecodeHookFuncValue {
	return func(from reflect.Value, to reflect.Value) (interface{}, error) {
		if from.CanInt() && from.Int() < 0 && to.CanUint() {
			return nil, fmt.Errorf("cannot convert negative value %v to an unsigned integer", from.Int())
		}
		return from.Interface(), nil
	}
}

type moduleFactory[T any, S any] interface {
	Create(s S) T
}

type createConfmapFunc[T any, S any] func(s S) T

type confmapModuleFactory[T any, S any] struct {
	f createConfmapFunc[T, S]
}

func (c confmapModuleFactory[T, S]) Create(s S) T {
	return c.f(s)
}

func newConfmapModuleFactory[T any, S any](f createConfmapFunc[T, S]) moduleFactory[T, S] {
	return confmapModuleFactory[T, S]{
		f: f,
	}
}
