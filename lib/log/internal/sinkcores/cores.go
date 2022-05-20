package sinkcores

import (
	"github.com/sourcegraph/sourcegraph/lib/log/sinks"
	"go.uber.org/zap/zapcore"
)

func Build(s *sinks.Sinks) []zapcore.Core {
	cores := []zapcore.Core{}
	if s.SentryHub != nil {
		c := NewSentryCore(s.SentryHub)
		c.Start()
		cores = append(cores, c)
	}
	return cores
}
