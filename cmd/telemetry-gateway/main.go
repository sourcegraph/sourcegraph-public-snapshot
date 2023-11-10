package main

import (
	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/telemetry-gateway/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/sanitycheck"
	"github.com/sourcegraph/sourcegraph/internal/service/svcmain"
)

var sentryDSN = env.Get("TELEMETRY_GATEWAY_SENTRY_DSN", "", "Sentry DSN")

func main() {
	sanitycheck.Pass()
	svcmain.SingleServiceMainWithoutConf(shared.Service, svcmain.OutOfBandConfiguration{
		Logging: func() conf.LogSinksSource {
			if sentryDSN == "" {
				return nil
			}
			return conf.NewStaticLogsSinksSource(log.SinksConfig{
				Sentry: &log.SentrySink{
					ClientOptions: sentry.ClientOptions{
						Dsn: sentryDSN,
					},
				},
			})
		}(),
		Tracing: nil,
	})
}
