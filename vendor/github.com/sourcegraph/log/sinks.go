package log

import (
	"go.uber.org/zap/zapcore"
)

// Sink describes additional destinations that github.com/sourcegraph/log can send log
// entries to. It can only be implemented directly within the package.
type Sink interface {
	Name() string

	// build creates the core to attach to the root logger. The implementation should
	// maintain a reference to anything needed to update this core, as a Sink will only
	// ever be built once.
	build() (zapcore.Core, error)
	// update is called on the `Update` callback from `log.Init` with new configuration.
	update(SinksConfig) error
}

// SinksConfig describes unified configuration for all sinks.
type SinksConfig struct {
	Sentry *SentrySink
}

type sinks []Sink

// SinksConfigGetter should provide the latest SinksConfig to update sink configuration.
type SinksConfigGetter func() SinksConfig

func (s sinks) update(get SinksConfigGetter) func() {
	return func() {
		updated := get()
		for _, sink := range s {
			if err := sink.update(updated); err != nil {
				Scoped("log.sinks.update").
					Error("failed to update", String("sink", sink.Name()), Error(err))
			}
		}
	}
}

func (s sinks) build() ([]zapcore.Core, error) {
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
