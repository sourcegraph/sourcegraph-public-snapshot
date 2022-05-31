package log

import (
	"go.uber.org/zap/zapcore"
)

type Sink interface {
	build() (zapcore.Core, error)
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
				// TODO
			}
		}
	}
}

func (s Sinks) build() ([]zapcore.Core, error) {
	var cores []zapcore.Core

	for _, sink := range s {
		sc, err := sink.build()
		if err != nil {
			return nil, err
		}
		cores = append(cores, sc)
	}

	return cores, nil
}
