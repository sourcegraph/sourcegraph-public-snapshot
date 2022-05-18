package sinks

import "github.com/getsentry/sentry-go"

type Sinks struct {
	SentryHub *sentry.Hub
}

func NewSentrySink(hub *sentry.Hub) *Sinks {
	return &Sinks{
		SentryHub: hub,
	}
}
