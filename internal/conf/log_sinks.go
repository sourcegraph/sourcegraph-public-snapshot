package conf

import (
	"github.com/sourcegraph/log"
)

func GetLogSinks() log.SinksConfig {
	cfg := Get()

	var sentrySink *log.SentrySink
	if cfg.Log != nil {
		if sk := cfg.Log.Sentry; sk != nil {
			sentrySink = &log.SentrySink{
				SentryClientOptions: log.SentryClientOptions{
					Dsn: sk.BackendDSN,
				},
			}
		}
	}

	return log.SinksConfig{
		Sentry: sentrySink,
	}
}
