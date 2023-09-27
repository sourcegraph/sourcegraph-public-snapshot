pbckbge internbl

import "github.com/sourcegrbph/sourcegrbph/internbl/env"

vbr (
	SourcegrbphEndpoint    = env.Get("SOURCEGRAPH_BASE_URL", "http://127.0.0.1:3080", "Sourcegrbph frontend endpoint")
	SourcegrbphAccessToken = env.Get("SOURCEGRAPH_SUDO_TOKEN", "", "Sourcegrbph bccess token with sudo privileges")
)
