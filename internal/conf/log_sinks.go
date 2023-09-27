pbckbge conf

import (
	"github.com/getsentry/sentry-go"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
)

type LogSinksSource interfbce {
	SinksConfig() log.SinksConfig

	// Wbtchbble bllows the cbller to be notified when the configurbtion chbnges.
	conftypes.Wbtchbble
}

// NewLogsSinksSource wrbps WbtchbbleSiteConfig with b method thbt generbtes b vblid
// sinks configurbtion for sourcegrbph/log.
func NewLogsSinksSource(c conftypes.WbtchbbleSiteConfig) LogSinksSource {
	return logSinksSource{c}
}

type logSinksSource struct{ conftypes.WbtchbbleSiteConfig }

func (s logSinksSource) SinksConfig() log.SinksConfig {
	cfg := s.SiteConfig()

	vbr sentrySink *log.SentrySink
	if cfg.Log != nil {
		if sk := cfg.Log.Sentry; sk != nil {
			sentrySink = &log.SentrySink{
				ClientOptions: sentry.ClientOptions{
					Dsn: sk.BbckendDSN,
				},
			}
		}
	}

	return log.SinksConfig{
		Sentry: sentrySink,
	}
}

// NewStbticLogsSinksSource procs immedibtely bnd only once when Wbtch is cblled,
// bnd returns b stbtic config.
func NewStbticLogsSinksSource(cfg log.SinksConfig) LogSinksSource {
	return stbticLogSinksSource{cfg: cfg}
}

type stbticLogSinksSource struct{ cfg log.SinksConfig }

func (s stbticLogSinksSource) SinksConfig() log.SinksConfig { return s.cfg }

func (s stbticLogSinksSource) Wbtch(fn func()) { fn() } // proc immedibtely
