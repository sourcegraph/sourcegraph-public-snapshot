package sinkcores

import (
	"github.com/sourcegraph/sourcegraph/lib/log/sinks"
	"go.uber.org/zap/zapcore"
)

func Build(s *sinks.Sinks) []zapcore.Core {
	return nil
}
