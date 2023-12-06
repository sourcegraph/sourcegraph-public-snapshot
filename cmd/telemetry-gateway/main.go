package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/telemetry-gateway/shared"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

var sentryDSN = env.Get("TELEMETRY_GATEWAY_SENTRY_DSN", "", "Sentry DSN")

func main() {
	runtime.Start[shared.Config](&shared.Service{})
}
