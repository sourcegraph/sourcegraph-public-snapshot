package conf

import "github.com/sourcegraph/sourcegraph/lib/log"

func GetLogSinks() log.SinksConfig {
	cfg := Get()

	var sentry *log.SentrySink
	if cfg.Log != nil {
		if sk := cfg.Log.Sentry; sk != nil {
			sentry = &log.SentrySink{
				DSN: sk.BackendDSN,
			}
		}
	}

	return log.SinksConfig{
		Sentry: sentry,
	}
}
