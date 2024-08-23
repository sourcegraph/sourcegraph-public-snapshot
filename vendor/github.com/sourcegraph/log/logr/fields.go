package logr

import (
	"go.uber.org/zap"

	"github.com/sourcegraph/log"
)

func toLogFields(keysAndValues []any) []log.Field {
	fields := make([]log.Field, 0, len(keysAndValues))
	for i := 0; i < len(keysAndValues); i += 2 {
		fields = append(fields,
			zap.Any(keysAndValues[i].(string), keysAndValues[i+1]))
	}
	return fields
}
