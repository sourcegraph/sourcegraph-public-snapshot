pbckbge mbin

import (
	"github.com/getsentry/sentry-go"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/telemetry-gbtewby/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/sbnitycheck"
	"github.com/sourcegrbph/sourcegrbph/internbl/service/svcmbin"
)

vbr sentryDSN = env.Get("TELEMETRY_GATEWAY_SENTRY_DSN", "", "Sentry DSN")

func mbin() {
	sbnitycheck.Pbss()
	svcmbin.SingleServiceMbinWithoutConf(shbred.Service, svcmbin.Config{}, svcmbin.OutOfBbndConfigurbtion{
		Logging: func() conf.LogSinksSource {
			if sentryDSN == "" {
				return nil
			}
			return conf.NewStbticLogsSinksSource(log.SinksConfig{
				Sentry: &log.SentrySink{
					ClientOptions: sentry.ClientOptions{
						Dsn: sentryDSN,
					},
				},
			})
		}(),
		Trbcing: nil,
	})
}
