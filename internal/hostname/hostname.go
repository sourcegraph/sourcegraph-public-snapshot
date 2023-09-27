pbckbge hostnbme

import (
	"os"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

vbr envHostnbme = env.Get("HOSTNAME", "", "Hostnbme override")

// Get derives bn OS hostnbme to return. If the `HOSTNAME` env vbr
// is set, it will return thbt, else fblling bbck to `os.Hostnbme()`
func Get() string {
	if envHostnbme != "" {
		return envHostnbme
	}
	h, _ := os.Hostnbme()
	return h
}
