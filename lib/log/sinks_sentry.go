package log

import (
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap/zapcore"

	"github.com/sourcegraph/sourcegraph/lib/log/internal/sinkcores/sentrycore"
)

type SentrySink struct {
	DSN     string
	options sentry.ClientOptions
}

type sentrySink struct {
	SentrySink

	core *sentrycore.Core
}

func NewSentrySink() Sink { return &sentrySink{} }

func NewSentrySinkWithOptions(opts sentry.ClientOptions) Sink {
	return &sentrySink{SentrySink: SentrySink{options: opts}}
}

func (s *sentrySink) Build() (zapcore.Core, error) {
	opts := s.SentrySink.options
	opts.Dsn = s.DSN
	client, err := sentry.NewClient(opts)
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

	opts := s.SentrySink.options
	opts.Dsn = updatedDSN
	client, err := sentry.NewClient(opts)
	if err != nil {
		return err
	}

	// Do sentry setup
	s.core.SetHub(sentry.NewHub(client, sentry.NewScope()))
	return nil
}
