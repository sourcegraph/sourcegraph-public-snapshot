package log

import (
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap/zapcore"

	"github.com/sourcegraph/sourcegraph/lib/log/internal/sinkcores/sentrycore"
)

type SentrySink struct {
	DSN string
}

type sentrySink struct {
	SentrySink

	core *sentrycore.Core
}

func NewSentrySink() Sink { return &sentrySink{} }

func (s *sentrySink) build() (zapcore.Core, error) {
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn: s.DSN,
	})
	if err != nil {
		return nil, err
	}
	s.core = sentrycore.NewCore(sentry.NewHub(client, sentry.NewScope()))
	return s.core, nil
}

func (s *sentrySink) update(updated SinksConfig) error {
	var updatedDSN string
	if updated.Sentry != nil {
		updatedDSN = updated.Sentry.DSN
	}

	if s.DSN == updatedDSN {
		return nil
	}

	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn: updatedDSN,
	})
	if err != nil {
		return err
	}

	// Do sentry setup
	s.core.SetHub(sentry.NewHub(client, sentry.NewScope()))
	return nil
}
