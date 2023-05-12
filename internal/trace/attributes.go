package trace

import (
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

type attributesStringer []attribute.KeyValue

func (a attributesStringer) String() string {
	var b strings.Builder
	for i, attr := range a {
		if i > 0 {
			b.WriteString("\n")
		}
		var (
			key   = string(attr.Key)
			value = attr.Value.Emit()
		)
		b.Grow(len(key) + 1 + len(value))
		b.WriteString(key)
		b.WriteString(":")
		b.WriteString(value)
	}
	return b.String()
}

type stringerFunc func() string

func (s stringerFunc) String() string { return s() }

// Scoped wraps a set of opentelemetry attributes with a prefixed key.
func Scoped(scope string, kvs ...attribute.KeyValue) []attribute.KeyValue {
	res := make([]attribute.KeyValue, len(kvs))
	for i, kv := range kvs {
		res[i] = attribute.KeyValue{
			Key:   attribute.Key(fmt.Sprintf("%s.%s", scope, kv.Key)),
			Value: kv.Value,
		}
	}
	return res
}

// Stringers creates a set of key values from a slice of elements that implement Stringer.
func Stringers[T fmt.Stringer](key string, values []T) attribute.KeyValue {
	strs := make([]string, 0, len(values))
	for _, value := range values {
		strs = append(strs, value.String())
	}
	return attribute.StringSlice(key, strs)
}

func Error(err error) attribute.KeyValue {
	if err != nil {
		return attribute.String("error", err.Error())
	}
	return attribute.String("error", "<nil>")
}
