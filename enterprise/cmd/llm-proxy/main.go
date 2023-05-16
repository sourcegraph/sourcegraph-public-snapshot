package main

import (
	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/service/svcmain"
)

func main() {
	svcmain.SingleServiceMainWithoutConf(shared.Service, svcmain.Config{}, svcmain.OutOfBandConfiguration{
		Logging: conf.NewStaticLogsSinksSource(log.SinksConfig{
			Sentry: &log.SentrySink{
				ClientOptions: sentry.ClientOptions{
					Dsn: env.Get("LLM_PROXY_SENTRY_DSN", "", "Sentry DSN"),
				},
			},
		}),
		Tracing: nil,
	})
}
