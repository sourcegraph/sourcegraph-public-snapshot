package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/sourcegraph/sourcegraph/lib/log/internal/encoders"
)

// A Field is a marshaling operation used to add a key-value pair to a logger's context.
//
// Field is an aliased import that is intentionally restricted so as to not allow overly
// liberal use of log fields, namely 'Any()'.
type Field = zapcore.Field

var (
	// String constructs a field with the given key and value.
	String = zap.String
	// Strings constructs a field that carries a slice of strings.
	Strings = zap.Strings

	// Int constructs a field with the given key and value.
	Int = zap.Int
	// Ints constructs a field that carries a slice of integers.
	Ints = zap.Ints

	// Float64 constructs a field that carries a float64. The way the floating-point value
	// is represented is encoder-dependent, so marshaling is necessarily lazy.
	Float64 = zap.Float64

	// Bool constructs a field that carries a bool.
	Bool = zap.Bool

	// Duration constructs a field with the given key and value. The encoder controls how
	// the duration is serialized.
	Duration = zap.Duration

	// Time constructs a Field with the given key and value. The encoder controls how the
	// time is serialized.
	Time = zap.Time

	// Error is shorthand for the common idiom NamedError("error", err).
	Error = zap.Error
	// NamedError constructs a field that lazily stores err.Error() under the provided key.
	// Errors which also implement fmt.Formatter (like those produced by github.com/pkg/errors)
	// will also have their verbose representation stored under key+"Verbose". If passed a
	// nil error, the field is a no-op.
	//
	// For the common case in which the key is simply "error", the Error function is shorter and less repetitive.
	NamedError = zap.NamedError

	// Namespace creates a named, isolated scope within the logger's context. All subsequent
	// fields will be added to the new namespace.
	//
	// This helps prevent key collisions when injecting loggers into sub-components or
	// third-party libraries.
	Namespace = zap.Namespace
)

// Object constructs a field that places all the given fields within the given key's
// namespace.
func Object(key string, fields ...Field) Field {
	return zap.Object(key, encoders.FieldsObjectEncoder(fields))
}
