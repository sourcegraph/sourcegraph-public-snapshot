package observation

import (
	otlog "github.com/opentracing/opentracing-go/log"
	"go.uber.org/zap"

	sglog "github.com/sourcegraph/sourcegraph/lib/log"
)

func toLogFields(otFields []otlog.Field) []sglog.Field {
	fields := make([]sglog.Field, len(otFields))
	for i, field := range otFields {
		// Allow usage of zap.Any here for ease of interop.
		fields[i] = zap.Any(field.Key(), field.Value())
	}
	return fields
}
