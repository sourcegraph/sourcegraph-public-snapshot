package spanattrprocessor

import (
	"go.opentelemetry.io/collector/processor"
)

type Config struct {
	processor.CreateSettings `mapstructure:",squash"`
}
