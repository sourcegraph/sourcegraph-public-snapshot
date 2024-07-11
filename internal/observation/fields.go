package observation

import (
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap" //nolint:logging // This is an expected usage

	"github.com/sourcegraph/log"
)

func attributesToLogFields(attributes []attribute.KeyValue) []log.Field {
	fields := make([]log.Field, len(attributes))
	for i, field := range attributes {
		switch value := field.Value.AsInterface().(type) {
		case error:
			// Special handling for errors, since we have a custom error field implementation
			fields[i] = log.NamedError(string(field.Key), value)

		default:
			// Allow usage of zap.Any here for ease of interop.
			fields[i] = zap.Any(string(field.Key), value)
		}
	}
	return fields
}

// MergeAttributes merges a list of attributes into a single list of
// attributes, with last-preferred semantics.
func MergeAttributes(attributes []attribute.KeyValue, more ...attribute.KeyValue) []attribute.KeyValue {
	m := orderedmap.New[string, attribute.Value]()
	for _, attr := range attributes {
		m.Set(string(attr.Key), attr.Value)
	}
	for _, attr := range more {
		m.Set(string(attr.Key), attr.Value)
	}
	var out []attribute.KeyValue
	for p := m.Oldest(); p != nil; p = p.Next() {
		out = append(out, attribute.KeyValue{
			Key:   attribute.Key(p.Key),
			Value: p.Value,
		})
	}
	return out
}
