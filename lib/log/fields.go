package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/sourcegraph/sourcegraph/lib/log/internal/encoders"
	"github.com/sourcegraph/sourcegraph/lib/log/privacy"
)

// A Field is a marshaling operation used to add a key-value pair to a logger's context.
//
// Field is an aliased import that is intentionally restricted so as to not allow overly
// liberal use of log fields, namely 'Any()'.
type Field = zapcore.Field

var (
	// Int constructs a field with the given key and value.
	Int = zap.Int
	// Ints constructs a field that carries a slice of integers.
	Ints = zap.Ints
	// Int64 constructs a field with the given key and value.
	Int64 = zap.Int64

	// Uint constructs a field with the given key and value.
	Uint = zap.Uint
	// Uint64 constructs a field with the given key and value.
	Uint64 = zap.Uint64

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

	// Namespace creates a named, isolated scope within the logger's context. All subsequent
	// fields will be added to the new namespace.
	//
	// This helps prevent key collisions when injecting loggers into sub-components or
	// third-party libraries.
	Namespace = zap.Namespace
)

func shouldRedact(p privacy.Privacy) bool {
	return p <= privacy.Unknown
}

func Text(key string, t privacy.Text) Field {
	if shouldRedact(t.Privacy()) {
		return zap.String(key, "<redacted>")
	}
	return zap.String(key, t.GetDataUnchecked())
}

func Texts(key string, ts []privacy.Text) Field {
	out := make([]string, 0, len(ts))
	for _, t := range ts {
		if shouldRedact(t.Privacy()) {
			out = append(out, "<redacted>")
		} else {
			out = append(out, t.GetDataUnchecked())
		}
	}
	return zap.Strings(key, out)
}

// String is a convenient wrapper around Text.
func String(key string, value string, p privacy.Privacy) Field {
	return Text(key, privacy.NewText(value, p))
}

// Strings constructs a field with multiple strings with the same privacy level.
func Strings(key string, values []string, p privacy.Privacy) Field {
	if !shouldRedact(p) {
		return zap.Strings(key, values)
	}
	out := make([]string, 0, len(values))
	for i := 0; i < len(values); i++ {
		out = append(out, "<redacted>")
	}
	return zap.Strings(key, out)
}

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
		return Text(key, privacy.NewText("<nil>", privacy.Public))
	}
	return Text(key, privacy.NewText(err.Error(), privacy.Public))
}
