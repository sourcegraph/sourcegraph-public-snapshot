package observation

import (
	otlog "github.com/opentracing/opentracing-go/log"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"github.com/sourcegraph/log"
)

func toLogFields(otFields []otlog.Field) []log.Field {
	fields := make([]log.Field, len(otFields))
	for i, field := range otFields {
		switch value := field.Value().(type) {
		case error:
			// Special handling for errors, since we have a custom error field implementation
			fields[i] = log.NamedError(field.Key(), value)

		default:
			// Allow usage of zap.Any here for ease of interop.
			fields[i] = zap.Any(field.Key(), value)
		}
	}
	return fields
}

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
