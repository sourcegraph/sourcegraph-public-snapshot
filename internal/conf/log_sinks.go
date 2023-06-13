package conf

import (
	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

type LogSinksSource interface {
	SinksConfig() log.SinksConfig

	// Watchable allows the caller to be notified when the configuration changes.
	conftypes.Watchable
}

// NewLogsSinksSource wraps WatchableSiteConfig with a method that generates a valid
// sinks configuration for sourcegraph/log.
func NewLogsSinksSource(c conftypes.WatchableSiteConfig) LogSinksSource {
	return logSinksSource{c}
}

type logSinksSource struct{ conftypes.WatchableSiteConfig }

func (s logSinksSource) SinksConfig() log.SinksConfig {
	cfg := s.SiteConfig()

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

// NewStaticLogsSinksSource procs immediately and only once when Watch is called,
// and returns a static config.
func NewStaticLogsSinksSource(cfg log.SinksConfig) LogSinksSource {
	return staticLogSinksSource{cfg: cfg}
}

type staticLogSinksSource struct{ cfg log.SinksConfig }

func (s staticLogSinksSource) SinksConfig() log.SinksConfig { return s.cfg }

func (s staticLogSinksSource) Watch(fn func()) { fn() } // proc immediately
