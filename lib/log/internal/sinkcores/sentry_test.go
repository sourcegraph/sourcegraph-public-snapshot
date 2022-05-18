package sinkcores_test

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/sinkcores"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// BenchmarkWrite-10        1000000              1320 ns/op
func BenchmarkWrite(b *testing.B) {
	c := sinkcores.SentryCore{}
	err := errors.New("foobar")
	for n := 0; n < b.N; n++ {
		c.With([]zapcore.Field{log.Error(err)}).Write(zapcore.Entry{Message: "msg"}, []zapcore.Field{log.Int("key", 5)})
	}
}

// BenchmarkNormal-10  184382              6266 ns/op
func BenchmarkNormal(b *testing.B) {
	logger, _ := zap.NewProduction()
	err := errors.New("foobar")
	for n := 0; n < b.N; n++ {
		logger.With(zap.Error(err), zap.Int("key", 5)).Info("msg")
	}
}
