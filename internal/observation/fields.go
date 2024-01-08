package observation

import (
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
