package log

import (
	"go.uber.org/zap/zapcore"
)

type Sink interface {
	Build() (zapcore.Core, error)
	update(SinksConfig) error
}

type SinksConfig struct {
	Sentry *SentrySink
}

type Sinks []Sink

type SinksConfigGetter func() SinksConfig

func (s Sinks) Update(get SinksConfigGetter) func() {
	return func() {
		updated := get()

		for _, sink := range s {
			if err := sink.update(updated); err != nil {
				logger := Scoped("conf", "configuration").
					Scoped("sentry-sink", "Logger extension that capture errors into Sentry")
				logger.Error("failed to update", Error(err))
			}
		}
	}
}

func (s Sinks) Build() ([]zapcore.Core, error) {
	var cores []zapcore.Core

	for _, sink := range s {
		sc, err := sink.Build()
		if err != nil {
			return nil, err
		}
		cores = append(cores, sc)
	}

	return cores, nil
}
