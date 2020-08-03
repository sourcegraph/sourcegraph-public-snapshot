package env

import "github.com/sourcegraph/sourcegraph/internal/env"

var HTTPAddrInternal = env.Get("SRC_HTTP_ADDR_INTERNAL", ":3090", "HTTP listen address for internal HTTP API. This should never be exposed externally, as it lacks certain authz checks.")
