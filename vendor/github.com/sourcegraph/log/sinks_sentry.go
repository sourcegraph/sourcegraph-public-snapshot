package log

import (
	"github.com/getsentry/sentry-go"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap/zapcore"

	"github.com/sourcegraph/log/internal/sinkcores/sentrycore"
)

// SentrySink reports all warning-level and above log messages that contain an error field
// (via the `log.Error(err)` or `log.NamedError(name, err)` field constructors) to Sentry,
// complete with stacktrace data and any additional context logged in the corresponding
// log message (including anything accumulated on a sub-logger).
type SentrySink struct {
	// ClientOptions expose various options to configure the Sentry client
	sentry.ClientOptions
}

type sentrySink struct {
	SentrySink

	core *sentrycore.Core
}

// NewSentrySink instantiates a Sentry sink to provide to `log.Init` with the following default values:
// - SampleRate: 0.1
// To provide different values see `NewSentrySinkWith`
func NewSentrySink() Sink {
	return &sentrySink{SentrySink: SentrySink{sentrycore.DefaultSentryClientOptions}}
}

// NewSentrySinkWith instantiates a Sentry sink to provide to `log.Init` with the values provided in SentrySink.
func NewSentrySinkWith(s SentrySink) Sink {
	return &sentrySink{SentrySink: SentrySink{s.ClientOptions}}
}

func (s *sentrySink) Name() string { return "SentrySink" }

func (s *sentrySink) build() (zapcore.Core, error) {
	client, err := sentry.NewClient(s.ClientOptions)
	if err != nil {
		return nil, err
	}
	s.core = sentrycore.NewCore(sentry.NewHub(client, sentry.NewScope()))
	return s.core, nil
}

func (s *sentrySink) update(updated SinksConfig) error {
	if updated.Sentry == nil {
		// use zero-value, effectively disabling sentry next
		updated.Sentry = &SentrySink{}
	}

	if cmp.Equal(s.ClientOptions, updated.Sentry.ClientOptions) {
		return nil
	}

	s.ClientOptions = updated.Sentry.ClientOptions
	client, err := sentry.NewClient(s.ClientOptions)
	if err != nil {
		return err
	}

	// Do sentry setup
	s.core.SetHub(sentry.NewHub(client, sentry.NewScope()))
	return nil
}
