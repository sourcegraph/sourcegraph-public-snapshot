package hostname

import (
	"os"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var envHostname = env.Get("HOSTNAME", "", "Hostname override")

// Get derives an OS hostname to return. If the `HOSTNAME` env var
// is set, it will return that, else falling back to `os.Hostname()`
func Get() string {
	if envHostname != "" {
		return envHostname
	}
	h, _ := os.Hostname()
	return h
}
