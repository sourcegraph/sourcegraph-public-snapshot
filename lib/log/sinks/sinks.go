package sinks

import (
	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/sinkcores/sentrycore"
	"go.uber.org/zap/zapcore"
)

type SinkCore interface {
	Core() zapcore.Core
}

func NewSentrySinkCore(hub *sentry.Hub) (SinkCore, func(hub *sentry.Hub)) {
	c := sentrycore.NewCore(hub)
	c.Start()
	return c, func(hub *sentry.Hub) {
		c.SetHub(hub)
	}
}
