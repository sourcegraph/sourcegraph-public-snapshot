package conf

import (
	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/log"
)

func GetLogSinks() log.SinksConfig {
	cfg := Get()

	var sentrySink *log.SentrySink
	if cfg.Log != nil {
		if sk := cfg.Log.Sentry; sk != nil {
			sentrySink = &log.SentrySink{
				ClientOptions: sentry.ClientOptions{
					Dsn: sk.BackendDSN,
				},
			}
		}
	}

	return log.SinksConfig{
		Sentry: sentrySink,
	}
}
