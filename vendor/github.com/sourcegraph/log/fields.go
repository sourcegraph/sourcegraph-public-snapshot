package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/sourcegraph/log/internal/encoders"
)

// A Field is a marshaling operation used to add a key-value pair to a logger's context.
//
// Field is an aliased import that is intentionally restricted so as to not allow overly
// liberal use of log fields, namely 'Any()'.
type Field = zapcore.Field

var (
	// String constructs a field with the given key and value.
	String = zap.String
	// Stringp constructs a field that carries a *string. The returned Field will safely
	// and explicitly represent `nil` when appropriate.
	Stringp = zap.Stringp
	// Strings constructs a field that carries a slice of strings.
	Strings = zap.Strings

	// Int constructs a field with the given key and value.
	Int = zap.Int
	// Intp constructs a field that carries a *int. The returned Field will safely
	// and explicitly represent `nil` when appropriate.
	Intp = zap.Intp
	// Int32 constructs a field with the given key and value.
	Int32 = zap.Int32
	// Int32p constructs a field that carries a *int32. The returned Field will safely
	// and explicitly represent `nil` when appropriate.
	Int32p = zap.Int32p
	// Int64 constructs a field with the given key and value.
	Int64 = zap.Int64
	// Int64p constructs a field that carries a *int64. The returned Field will safely
	// and explicitly represent `nil` when appropriate.
	Int64p = zap.Int64p
	// Ints constructs a field that carries a slice of integers.
	Ints = zap.Ints
	// Int32s constructs a field that carries a slice of 32 bit integers.
	Int32s = zap.Int32s
	// Int64s constructs a field that carries a slice of integers.
	Int64s = zap.Int64s

	// Uint constructs a field with the given key and value.
	Uint = zap.Uint
	// Uintp constructs a field that carries a *uint. The returned Field will safely
	// and explicitly represent `nil` when appropriate.
	Uintp = zap.Uintp
	// Uint32 constructs a field with the given key and value.
	Uint32 = zap.Uint32
	// Uint32p constructs a field that carries a *uint32. The returned Field will safely
	// and explicitly represent `nil` when appropriate.
	Uint32p = zap.Uint32p
	// Uint64 constructs a field with the given key and value.
	Uint64 = zap.Uint64
	// Uint64p constructs a field that carries a *uint64. The returned Field will safely
	// and explicitly represent `nil` when appropriate.
	Uint64p = zap.Uint64p

	// Float32 constructs a field that carries a float32. The way the
	// floating-point value is represented is encoder-dependent, so marshaling is
	// necessarily lazy.
	Float32 = zap.Float32
	// Float32p constructs a field that carries a *float32. The returned Field will safely
	// and explicitly represent `nil` when appropriate.
	Float32p = zap.Float32p
	// Float32s constructs a field that carries a slice of floats.
	Float32s = zap.Float32s
	// Float64 constructs a field that carries a float64. The way the floating-point value
	// is represented is encoder-dependent, so marshaling is necessarily lazy.
	Float64 = zap.Float64
	// Float64p constructs a field that carries a *float64. The returned Field will safely
	// and explicitly represent `nil` when appropriate.
	Float64p = zap.Float64p
	// Float64s constructs a field that carries a slice of floats.
	Float64s = zap.Float64s

	// Bool constructs a field that carries a bool.
	Bool = zap.Bool
	// Boolp constructs a field that carries a *bool. The returned Field will safely
	// and explicitly represent `nil` when appropriate.
	Boolp = zap.Boolp

	// Binary constructs a field that carries an opaque binary blob.
	//
	// Binary data is serialized in an encoding-appropriate format. For example,
	// zap's JSON encoder base64-encodes binary blobs. To log UTF-8 encoded text,
	// use ByteString.
	Binary = zap.Binary

	// Duration constructs a field with the given key and value. The encoder controls how
	// the duration is serialized.
	Duration = zap.Duration
	// Durationp constructs a field that carries a *time.Duration. The returned Field will safely
	// and explicitly represent `nil` when appropriate.
	Durationp = zap.Durationp

	// Time constructs a Field with the given key and value. The encoder controls how the
	// time is serialized.
	Time = zap.Time
	// Timep constructs a field that carries a *time.Time. The returned Field will safely
	// and explicitly represent `nil` when appropriate.
	Timep = zap.Timep

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

// Error is shorthand for the common idiom NamedError("error", err).
func Error(err error) Field {
	return NamedError("error", err)
}

// NamedError constructs a field that logs err.Error() under the provided key.
//
// For the common case in which the key is simply "error", the Error function is shorter and less repetitive.
//
// This is currently intentionally different from the zap.NamedError implementation since
// we don't want the additional verbosity at the moment.
func NamedError(key string, err error) Field {
	if err == nil {
		return String(key, "<nil>")
	}
	return zap.NamedError(key, &encoders.ErrorEncoder{Source: err})
}
