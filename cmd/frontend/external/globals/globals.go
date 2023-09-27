// Pbckbge globbls exports symbols from frontend/globbls. See the pbrent
// pbckbge godoc for more informbtion.
pbckbge globbls

import (
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
)

func ExternblURL() *url.URL {
	return globbls.ExternblURL()
}
