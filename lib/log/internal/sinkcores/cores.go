package sinkcores

import (
	"github.com/sourcegraph/sourcegraph/lib/log/sinks"
	"go.uber.org/zap/zapcore"
)

func Build(s *sinks.Sinks) []zapcore.Core {
	cores := []zapcore.Core{}
	println("build")
	if s.SentryHub != nil {
		println("sentry")
		c := SentryCore{hub: s.SentryHub}
		c.Start()
		cores = append(cores, &c)
	}

	return cores
}
