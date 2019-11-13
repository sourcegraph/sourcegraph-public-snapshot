package lsifserver

import (
	"github.com/sourcegraph/sourcegraph/internal/env"
)

var ServerURLFromEnv = env.Get("LSIF_SERVER_URL", "http://lsif-server:3186", "URL at which the lsif-server service can be reached")
