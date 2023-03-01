package bg

import (
	"github.com/pkg/browser"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
)

// AppReady is called once the frontend has reported it is ready to serve
// requests. It contains tasks related to Sourcegraph App (single binary).
func AppReady(logger log.Logger) {
	if !deploy.IsDeployTypeSingleProgram(deploy.Type()) {
		return
	}

	u := globals.ExternalURL().String()
	if err := browser.OpenURL(u); err != nil {
		logger.Error("failed to open browser", log.String("url", u), log.Error(err))
	}
}
