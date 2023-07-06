package trace

import (
	"fmt"
	"unicode/utf8"

	"go.opentelemetry.io/otel/attribute"
)

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
	err = truncateError(err, defaultErrorRuneLimit)
	if err != nil {
		return attribute.String("error", err.Error())
	}
	return attribute.String("error", "<nil>")
}

const defaultErrorRuneLimit = 512

func truncateError(err error, maxRunes int) error {
	if err == nil {
		return nil
	}
	return truncatedError{err, maxRunes}
}

type truncatedError struct {
	err      error
	maxRunes int
}

func (e truncatedError) Error() string {
	errString := e.err.Error()
	if utf8.RuneCountInString(errString) > e.maxRunes {
		runes := []rune(errString)
		errString = string(runes[:e.maxRunes/2]) + " ...truncated... " + string(runes[len(runes)-e.maxRunes/2:])
	}
	return errString
}
